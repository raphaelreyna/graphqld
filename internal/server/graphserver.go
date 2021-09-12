package server

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/radovskyb/watcher"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/graph"
	"github.com/raphaelreyna/graphqld/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type server struct {
	conf config.GraphConf

	addr string
	port string

	schema graphql.Schema

	w *watcher.Watcher

	close chan struct{}

	sync.RWMutex
}

func newServer(addr string, mux *mux.Router, conf config.GraphConf) (*server, error) {
	var s = server{
		conf:  conf,
		addr:  addr,
		close: make(chan struct{}),
	}

	{
		_, port, err := net.SplitHostPort(s.addr)
		if err != nil {
			return nil, err
		}
		s.port = port
	}

	{
		if cc := s.conf.CORS; cc != nil {
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

			mux.Use(handlers.CORS(opts...))
		}
	}

	if conf.HotReload {
		s.w = watcher.New()
		s.w.SetMaxEvents(1)
		s.w.FilterOps(
			watcher.Rename, watcher.Move,
			watcher.Create, watcher.Chmod,
			watcher.Write,
		)

		if err := s.w.AddRecursive(s.conf.DocumentRoot); err != nil {
			return nil, err
		}
	}

	if conf.Graphiql {
		handler, err := graphiql.NewGraphiqlHandler("/")
		if err != nil {
			return nil, err
		}
		mux.Handle("/graphiql", handler)
	}

	mux.Use(middleware.Log)

	mux.Use(middleware.FromGraphConf(conf))

	if ba := conf.BasicAuth; ba != nil {
		mux.Use(
			middleware.BasicAuth(ba.Username, ba.Password),
		)
	}

	mux.HandleFunc("/", s.serveHTTP)

	return &s, nil
}

func (s *server) UpdateSchema() error {
	var (
		conf = s.conf

		g = graph.Graph{
			DocumentRoot: conf.DocumentRoot,
			ResolverDir:  conf.ResolverDir,
		}
	)

	if err := g.Build(); err != nil {
		var logEvent *zerolog.Event
		if conf.HotReload {
			logEvent = log.Error()
		} else {
			logEvent = log.Fatal()
		}
		logEvent.Err(err).
			Msg("unable to build graph schema config")
	} else {
		var schemaConf graphql.SchemaConfig

		if q := g.Query; q != nil {
			schemaConf.Query = q
		}
		if m := g.Mutation; m != nil {
			schemaConf.Mutation = m
		}

		schema, err := graphql.NewSchema(schemaConf)
		if err != nil {
			return err
		}

		s.Lock()
		s.schema = schema
		s.Unlock()
	}

	return nil
}

func (s *server) Start() error {
	var err = s.UpdateSchema()

	if s.conf.HotReload {
		go func() {
			for {
				select {
				case <-s.w.Event:
					if err := s.UpdateSchema(); err != nil {
						log.Error().Err(err).
							Msg("error updating schema")
					}
				case err := <-s.w.Error:
					log.Fatal().Err(err).
						Msg("unable to watch root directory")
				case <-s.w.Closed:
					return
				}
			}
		}()

		return s.w.Start(time.Second)
	}

	return err
}

func (s *server) Stop() {
	if !s.conf.HotReload {
		return
	}

	s.w.Close()
}

func (s *server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		logger = middleware.GetLogger(ctx)
		opts   = handler.NewRequestOptions(r)
	)

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
		logger.Error().Err(err).
			Interface("result", *result).
			Msg("unable to encode result")
	}
}
