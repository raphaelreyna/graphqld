package graphhost

import (
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/graph"
	"github.com/raphaelreyna/graphqld/internal/reload"
	httputil "github.com/raphaelreyna/graphqld/internal/transport/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type GraphHost struct {
	Config config.GraphConf

	Graph   graph.Graph
	Watcher reload.Watcher
	Server  httputil.Server
}

func NewGraphHost(addr string, config config.GraphConf) (*GraphHost, error) {
	var gh = GraphHost{
		Config: config,
		Server: httputil.Server{},
		Graph: graph.Graph{
			Dir:        config.DocumentRoot,
			ResolverWD: config.ResolverDir,
		},
	}

	gh.Server.Schema = make(chan graphql.Schema, 1)
	gh.Server.Addr = addr
	gh.Server.CtxPath = gh.Config.ContextExecPath
	gh.Server.CtxFilesDir = gh.Config.ContextFilesDir

	if gh.Config.Graphiql {
		gh.Server.GraphiQL = "/graphiql"
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

		gh.Server.Schema <- schema
	}

	if gh.Config.HotReload {
		gh.Watcher = reload.Watcher{
			RootDir:  gh.Config.DocumentRoot,
			Interval: time.Second,
			Schema:   gh.Server.Schema,
			ServerMu: &gh.Server.RWMutex,
		}

		go gh.Watcher.Run()
	}

	gh.Server.Start()

	return &gh, nil
}

func (gh *GraphHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.Server.ServeHTTP(w, r)
}

func (gh *GraphHost) Stop() {
	gh.Server.Stop()
}
