package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/graph"
	"github.com/raphaelreyna/graphqld/internal/reload"
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
		c, err = config.ParseFromEnv()
		if err != nil {
			log.Fatal().Err(err).
				Msg("error reading configuration from environment")
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

	var (
		graphHosts  = make(map[string]*graphHost)
		singleGraph *graphHost
	)
	for _, g := range c.Graphs {
		gh, err := newGraphHost(c.Addr, g)
		if err != nil {
			log.Fatal().Err(err).
				Msg("error creating new graph host")
		}

		graphHosts[gh.config.ServerName] = gh

		singleGraph = gh
	}
	if 1 < len(graphHosts) {
		singleGraph = nil
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gh = singleGraph

		if gh == nil {
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			gh = graphHosts[host]

			if host != gh.config.ServerName {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else if c.Hostname != "" {
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if host != c.Hostname {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else if gh.config.ServerName != "" {
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if host != gh.config.ServerName {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		gh.ServeHTTP(w, r)
	}))

	if err := http.ListenAndServe(c.Addr, nil); err != nil {
		log.Fatal().Err(err).
			Msg("error listening and serving")
	}

	if singleGraph != nil {
		singleGraph.stop()
	} else {
		for _, gh := range graphHosts {
			gh.stop()
		}
	}

	retCode = 0
}

type graphHost struct {
	config config.GraphConf

	graph   graph.Graph
	watcher reload.Watcher
	server  httputil.Server
}

func newGraphHost(addr string, config config.GraphConf) (*graphHost, error) {
	var gh = graphHost{
		config: config,
		server: httputil.Server{},
		graph: graph.Graph{
			Dir: config.DocumentRoot,
		},
	}

	gh.server.Schema = make(chan graphql.Schema, 1)
	gh.server.Addr = addr
	gh.server.CtxPath = gh.config.ContextExecPath
	gh.server.CtxFilesDir = gh.config.ContextFilesDir

	if gh.config.Graphiql {
		gh.server.GraphiQL = "/graphiql"
	}

	if err := gh.graph.Build(); err != nil {
		var logEvent *zerolog.Event
		if gh.config.HotReload {
			logEvent = log.Error()
		} else {
			logEvent = log.Fatal()
		}
		logEvent.Err(err).
			Msg("unable to build graph schema config")
	} else {
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: graphql.NewObject(gh.graph.Query.ObjectConf),
		})
		if err != nil {
			return nil, err
		}

		gh.server.Schema <- schema
	}

	if gh.config.HotReload {
		gh.watcher = reload.Watcher{
			RootDir:  gh.config.DocumentRoot,
			Interval: time.Second,
			Schema:   gh.server.Schema,
			ServerMu: &gh.server.RWMutex,
		}

		go gh.watcher.Run()
	}

	gh.server.Start()

	return &gh, nil
}

func (gh *graphHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.server.ServeHTTP(w, r)
}

func (gh *graphHost) stop() {
	gh.server.Stop()
}
