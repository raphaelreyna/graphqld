package main

import (
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal().
				Interface("recovered", r)
		}
	}()

	var c = config.Config
	logConfig(c)

	s, err := server.NewServer(c)
	if err != nil {
		log.Error().Err(err).
			Msg("error creating new server")
	}

	if err := s.Start(); err != nil {
		log.Error().Err(err).
			Msg("error starting server")
	}

	if err := s.Stop(); err != nil {
		log.Error().Err(err).
			Msg("error stopping server")
	}
}

func logConfig(c config.Conf) {
	logEvent := log.Info().
		Str("address", c.Addr).
		Str("root", c.RootDir)

	if c.Hostname != "" {
		logEvent = logEvent.Str("hostname", c.Hostname)
	}

	logEvent.Msg("server configuration")

	logEvent = log.Info().
		Bool("hot", c.HotReload).
		Bool("graphiql", c.Graphiql).
		Str("resolver-wd", c.ResolverDir)
	if c.ContextExecPath != "" {
		logEvent = logEvent.Str("ctx-exec-path", c.ContextExecPath)
	}
	if c.ContextFilesDir != "" {
		logEvent = logEvent.Str("ctx-files-wd", c.ContextFilesDir)
	}

	logEvent.Msg("graph default configuration")

	for _, g := range c.Graphs {
		logEvent := log.Info().
			Bool("hot", g.HotReload).
			Bool("graphiql", g.Graphiql).
			Str("server-name", g.ServerName).
			Str("document-root", g.DocumentRoot).
			Str("resolver-dir", g.ResolverDir)
		if g.ContextExecPath != "" {
			logEvent = logEvent.Str("ctx-exec-path", g.ContextExecPath)
		}
		if g.ContextFilesDir != "" {
			logEvent = logEvent.Str("ctx-files-wd", g.ContextFilesDir)
		}
		logEvent.Msg("loaded graph configuration")
	}
}
