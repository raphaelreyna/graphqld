package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/handlers"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/rs/zerolog/hlog"
)

type Server struct {
	Addr        string
	Name        string
	MaxBodySize int64

	CORS *config.CORSConfig

	Schema chan graphql.Schema

	CtxPath     string
	CtxFilesDir string

	HotReload bool

	GraphiQL        string
	graphiqlHandler *graphiql.Handler

	handler http.Handler

	close chan struct{}

	schema graphql.Schema
	port   string

	sync.RWMutex
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	var logger = hlog.FromRequest(r)
	logger.Info().Msg("got HTTP request")

	var (
		ctx = r.Context()

		opts = handler.NewRequestOptions(r)

		env = getEnv(s.port, r)
	)
	ctx = context.WithValue(ctx, keyHeaderFunc, w.Header)
	ctx = context.WithValue(ctx, keyEnv, env)
	ctx = context.WithValue(ctx, keyLog, logger)

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
			logger.Error().Err(err).
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
			logger.Error().Err(err).
				Msg("unable to create a context from the ctx handler")

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if _, err := ctxFile.Write(ctxData); err != nil {
			logger.Error().Err(err).
				Msg("unable to write context to the context file")

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		params.Context = context.WithValue(params.Context, keyCtxFile, ctxFile)
	}

	result := graphql.Do(params)

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Error().Err(err).
			Interface("result", *result).
			Msg("unable to encode result")
	}
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

	r.Body = &limitedReaderCloser{
		LimitedReader: io.LimitedReader{
			R: r.Body,
			N: s.MaxBodySize,
		},
	}

	s.handler.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	s.close = make(chan struct{})

	_, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return err
	}
	s.port = port

	s.handler = http.HandlerFunc(s.serveHTTP)

	if cc := s.CORS; cc != nil {
		var opts = make([]handlers.CORSOption, 0)

		if cc.AllowCredentials {
			opts = append(opts, handlers.AllowCredentials())
		}

		if cc.IgnoreOptions {
			opts = append(opts, handlers.IgnoreOptions())
		}

		if 0 < len(cc.AllowedHeaders) {
			opts = append(opts,
				handlers.AllowedHeaders(cc.AllowedHeaders),
			)
		}

		if 0 < len(cc.AllowedOrigins) {
			opts = append(opts,
				handlers.AllowedOrigins(cc.AllowedOrigins),
			)
		}

		s.handler = handlers.CORS(opts...)(s.handler)
	}

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

type limitedReaderCloser struct {
	io.LimitedReader
}

func (lrc *limitedReaderCloser) Read(p []byte) (int, error) {
	return lrc.LimitedReader.Read(p)
}

func (lrc *limitedReaderCloser) Close() error {
	x, ok := lrc.R.(io.Closer)
	if !ok {
		return errors.New("unable to convert from type io.Reader to io.Closer in limitedReaderCloser.Close")
	}

	return x.Close()
}
