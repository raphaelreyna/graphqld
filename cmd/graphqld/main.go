package main

import (
	"flag"
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
		c = config.ParseFromEnv()
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

	server := httputil.Server{}
	server.Schema = make(chan graphql.Schema, 1)
	server.Addr = ":" + c.Port

	if c.Graphiql {
		server.GraphiQL = "/graphiql"
	}

	g := graph.Graph{
		Dir: c.RootDir,
	}

	if err := g.Build(); err != nil {
		var logEvent *zerolog.Event
		if c.HotReload {
			logEvent = log.Error()
		} else {
			logEvent = log.Fatal()
		}
		logEvent.Err(err).
			Msg("unable to build graph schema config")
	} else {
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: graphql.NewObject(g.Query.ObjectConf),
		})
		if err != nil {
			log.Error().Err(err).
				Msg("unable to build graph schema")
		}

		server.Schema <- schema
	}

	if c.HotReload {
		w := reload.Watcher{
			RootDir:  c.RootDir,
			Interval: time.Second,
			Schema:   server.Schema,
			ServerMu: &server.RWMutex,
		}

		go w.Run()
	}

	log.Info().Msg("starting ...")
	if err := server.Run(); err != nil {
		log.Error().Err(err).
			Msg("error serving http")
	}

	retCode = 0
}
