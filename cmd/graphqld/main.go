package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/friendsofgo/graphiql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/radovskyb/watcher"
	"github.com/raphaelreyna/graphqld/internal/graph"
	httputil "github.com/raphaelreyna/graphqld/internal/transport/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

var (
	rootDir    string
	liveReload bool
)

func init() {
	flag.StringVarP(&rootDir, "root-dir", "d", "", "Will default to the current working directory.")
	flag.BoolVarP(&liveReload, "live-graph", "l", false, "Set to true the graph will be rebuilt when a change is made in the root directory.")
}

func main() {
	flag.Parse()

	var (
		retCode = 1
		err     error
	)

	if os.Getenv("DEV") != "" {
		zerolog.SetGlobalLevel(-1)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("recovered", r)
		}
		os.Exit(retCode)
	}()

	if rootDir == "" {
		rootDir, err = os.Getwd()
		if err != nil {
			log.Error().Err(err).
				Msg("unable to obtain a root directory")
		}
		return
	}

	g := graph.Graph{
		Dir: rootDir,
	}

	if err := g.Build(); err != nil {
		var logEvent *zerolog.Event
		if liveReload {
			logEvent = log.Error()
		} else {
			logEvent = log.Fatal()
		}
		logEvent.Err(err).
			Msg("unable to build graph schema config")
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(g.Query.ObjectConf),
	})
	if err != nil {
		log.Error().Err(err).
			Msg("unable to build graph schema")
	}

	lock := sync.RWMutex{}
	if liveReload {
		w := watcher.New()
		w.SetMaxEvents(1)
		w.FilterOps(
			watcher.Rename, watcher.Move,
			watcher.Create, watcher.Chmod,
			watcher.Write,
		)

		go func() {
			for {
				select {
				case <-w.Event:
					g := graph.Graph{
						Dir: rootDir,
					}

					if err := g.Build(); err != nil {
						log.Error().Err(err).
							Msg("unable to rebuild graph schema config")
						continue
					}

					schm, err := graphql.NewSchema(graphql.SchemaConfig{
						Query: graphql.NewObject(g.Query.ObjectConf),
					})
					if err != nil {
						log.Error().Err(err).
							Msg("unable to rebuild schema")
						continue
					}

					lock.Lock()
					schema = schm
					lock.Unlock()

					log.Info().Msg("rebuild schema")
				case err := <-w.Error:
					log.Fatal().Err(err).
						Msg("unable to watch root directory")
				case <-w.Closed:
					return
				}
			}
		}()

		if err := w.AddRecursive(g.Dir); err != nil {
			log.Fatal().Err(err).
				Msg("unable to recursvely watch root directory")
		}
		go func() {
			if err := w.Start(time.Second); err != nil {
				log.Fatal().Err(err).
					Msg("error watching root directory")
			}
		}()
	}

	http.Handle("/", httputil.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts := handler.NewRequestOptions(r)
		lock.RLock()
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  opts.Query,
			VariableValues: opts.Variables,
			OperationName:  opts.OperationName,
			Context:        r.Context(),
		})
		lock.RUnlock()
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})))

	if x := os.Getenv("GRAPHQLD_GRAPHIQL"); x != "" {
		graphiqlServer, err := graphiql.NewGraphiqlHandler("/")
		if err != nil {
			panic(err)
		}

		http.Handle("/graphiql", graphiqlServer)

		log.Info().Msg("enabled graphiql")
	}

	port := "8080"
	if x := os.Getenv("PORT"); x != "" {
		port = x
	}

	log.Info().Msg("starting ...")
	http.ListenAndServe(":"+port, nil)
}
