package graphhost

import (
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/graph"
	httputil "github.com/raphaelreyna/graphqld/internal/transport/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type GraphHost struct {
	Config config.GraphConf

	Graph graph.Graph

	watcher fileWatcher
	server  httputil.Server
}

func NewGraphHost(addr string, maxBodySize int64, config config.GraphConf) (*GraphHost, error) {
	var gh = GraphHost{
		Config: config,
		server: httputil.Server{
			HotReload:   config.HotReload,
			MaxBodySize: maxBodySize,
		},
		Graph: graph.Graph{
			DocumentRoot: config.DocumentRoot,
			ResolverDir:  config.ResolverDir,
		},
	}

	gh.server.Schema = make(chan graphql.Schema, 1)
	gh.server.Addr = addr
	gh.server.CtxPath = gh.Config.ContextExecPath
	gh.server.CtxFilesDir = gh.Config.ContextFilesDir

	if gh.Config.Graphiql {
		gh.server.GraphiQL = "/graphiql"
	}

	if err := gh.Graph.Build(); err != nil {
		var logEvent *zerolog.Event
		if gh.Config.HotReload {
			logEvent = log.Error()
		} else {
			logEvent = log.Fatal()
		}
		logEvent.Err(err).
			Msg("unable to build graph schema config")
	} else {
		var schemaConf graphql.SchemaConfig
		if q := gh.Graph.Query; q != nil {
			schemaConf.Query = q
		}
		if m := gh.Graph.Mutation; m != nil {
			schemaConf.Mutation = m
		}

		schema, err := graphql.NewSchema(schemaConf)
		if err != nil {
			return nil, err
		}

		gh.server.Schema <- schema
	}

	if gh.Config.HotReload {
		gh.watcher = fileWatcher{
			RootDir:  gh.Config.DocumentRoot,
			Interval: time.Second,
			Schema:   gh.server.Schema,
			ServerMu: &gh.server.RWMutex,
		}

		go gh.watcher.Run()
	}

	gh.server.Start()

	return &gh, nil
}

func (gh *GraphHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.server.ServeHTTP(w, r)
}

func (gh *GraphHost) Stop() {
	gh.server.Stop()
}
