package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"github.com/friendsofgo/graphiql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Schema   chan graphql.Schema
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
	)
	ctx = context.WithValue(ctx, keyHeaderFunc, w.Header)
	ctx = context.WithValue(ctx, keyHeader, r.Header)
	ctx = context.WithValue(ctx, keyEnv, getEnv(s.port, r))

	var params = graphql.Params{
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}

	s.RLock()
	params.Schema = s.schema
	s.RUnlock()

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
