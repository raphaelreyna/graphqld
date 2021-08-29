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
	Schema chan graphql.Schema

	CtxPath     string
	CtxFilesDir string

	GraphiQL string

	close chan struct{}

	schema graphql.Schema
	port   string

	sync.RWMutex
	http.Server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) Run() error {
	s.close = make(chan struct{})

	_, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return err
	}
	s.port = port

	mux := http.NewServeMux()
	mux.Handle("/", s)

	go func() {
		for {
			select {
			case schema := <-s.Schema:
				s.Lock()
				s.schema = schema
				s.Unlock()
				log.Info().
					Msg("serving new schema")
			case <-s.close:
				return
			}
		}
	}()

	if s.GraphiQL != "" {
		graphiqlServer, err := graphiql.NewGraphiqlHandler("/")
		if err != nil {
			return err
		}

		mux.Handle(s.GraphiQL, graphiqlServer)

		log.Info().
			Msg("enabled graphiql")
	}

	s.Handler = mux

	return s.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.close <- struct{}{}
	s.Shutdown(ctx)
	return nil
}
