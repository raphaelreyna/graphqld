package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/friendsofgo/graphiql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Addr string
	Name string

	Schema chan graphql.Schema

	CtxPath     string
	CtxFilesDir string

	HotReload bool

	GraphiQL        string
	graphiqlHandler *graphiql.Handler

	close chan struct{}

	schema graphql.Schema
	port   string

	sync.RWMutex
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/graphiql" {
		if s.graphiqlHandler != nil {
			s.graphiqlHandler.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var (
		ctx = r.Context()

		opts = handler.NewRequestOptions(r)

		env = getEnv(s.port, r)
	)
	ctx = context.WithValue(ctx, keyHeaderFunc, w.Header)
	ctx = context.WithValue(ctx, keyEnv, env)

	var params = graphql.Params{
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}

	s.RLock()
	params.Schema = s.schema
	s.RUnlock()

	if s.CtxPath != "" {
		ctxFile, err := ioutil.TempFile(s.CtxFilesDir, "")
		if err != nil {
			log.Error().Err(err).
				Msg("unable to create temporary context file")

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
		defer func() {
			ctxFile.Close()
			os.Remove(ctxFile.Name())
		}()

		cmd := exec.Cmd{
			Path: s.CtxPath,
			Env:  env,
		}

		ctxData, err := cmd.Output()
		if err != nil {
			log.Error().Err(err).
				Msg("unable to create a context from the ctx handler")

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if _, err := ctxFile.Write(ctxData); err != nil {
			log.Error().Err(err).
				Msg("unable to write context to the context file")

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		params.Context = context.WithValue(params.Context, keyCtxFile, ctxFile)
	}

	result := graphql.Do(params)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error().Err(err).
			Interface("result", *result).
			Msg("unable to encode result")
	}
}

func (s *Server) Start() error {
	s.close = make(chan struct{})

	_, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return err
	}
	s.port = port

	go func() {
		for run := true; run; {
			select {
			case schema := <-s.Schema:
				s.Lock()
				s.schema = schema
				s.Unlock()
			case <-s.close:
				return
			}

			run = s.HotReload
		}
	}()

	if s.GraphiQL != "" {
		s.graphiqlHandler, err = graphiql.NewGraphiqlHandler("/")
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.close <- struct{}{}
}
