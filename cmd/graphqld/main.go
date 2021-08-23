package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/friendsofgo/graphiql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/radovskyb/watcher"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/graph"
	httputil "github.com/raphaelreyna/graphqld/internal/transport/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "f", "", "Path to configuration file.")
	flag.Parse()
}

func main() {
	var (
		retCode = 1
		c       *config.Conf
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

	switch configPath {
	case "":
		c = &config.Conf{}
		if err := c.Default(); err != nil {
			log.Error().Err(err).
				Msg("error creating default configuration")
			return
		}
		log.Info().Interface("conf", *c).
			Msg("using default configuration")
	default:
		c, err = config.ParseYamlFile(configPath)
		if err != nil {
			log.Error().Err(err).
				Str("file", configPath).
				Msg("error reading configuration")
			return
		}
		log.Info().Interface("conf", *c).
			Msg("finished loading configuration")
	}

	if c.RootDir == "" {
		if err != nil {
			log.Error().Err(err).
				Msg("unable to obtain a root directory")
		}
		return
	}

	g := graph.Graph{
		Dir: c.RootDir,
	}

	if err := g.Build(); err != nil {
		var logEvent *zerolog.Event
		if c.LiveReload {
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
	if c.LiveReload {
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
						Dir: c.RootDir,
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
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Error().Err(err).
				Interface("result", *result).
				Msg("unable to encode result")
		}
	})))

	if c.Graphiql {
		graphiqlServer, err := graphiql.NewGraphiqlHandler("/")
		if err != nil {
			log.Error().Err(err).
				Msg("could not enable graphiql")
			return
		}

		http.Handle("/graphiql", graphiqlServer)

		log.Info().Msg("enabled graphiql")
	}

	log.Info().Msg("starting ...")
	if err := http.ListenAndServe(":"+c.Port, nil); err != nil {
		log.Error().Err(err).
			Msg("error serving http")
	}

	retCode = 0
}
