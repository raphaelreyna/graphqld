package main

import (
	"github.com/raphaelreyna/graphqld/internal/config"
	"github.com/raphaelreyna/graphqld/internal/server"
	"github.com/rs/zerolog"
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

	if c.Context != nil {
		logEvent = logEvent.Interface("context", c.Context)
	}

	if c.CORS != nil {
		logEvent = logEvent.Interface("cors", c.CORS)
	}

	if c.BasicAuth != nil {
		logEvent = logEvent.Dict("basic-auth",
			zerolog.Dict().
				Str("username", c.BasicAuth.Username),
		)
	}

	if c.User != nil {
		logEvent = logEvent.Interface("user", c.User)
	}

	if c.TLS != nil {
		logEvent = logEvent.Interface("tls", c.TLS)
	}

	logEvent.Msg("graph default configuration")

	for _, g := range c.Graphs {
		logEvent := log.Info().
			Bool("hot", g.HotReload).
			Bool("graphiql", g.Graphiql).
			Str("document-root", g.DocumentRoot).
			Str("resolver-dir", g.ResolverDir)

		if g.ServerName != "" {
			logEvent = logEvent.Str("server-name", g.ServerName)
		}

		if g.Context != nil {
			logEvent = logEvent.Interface("context", g.Context)
		}

		if g.CORS != nil {
			logEvent = logEvent.Interface("cors", g.CORS)
		}

		if g.BasicAuth != nil {
			logEvent = logEvent.Dict("basic-auth",
				zerolog.Dict().
					Str("username", g.BasicAuth.Username),
			)
		}

		if g.User != nil {
			logEvent = logEvent.Interface("user", g.User)
		}

		logEvent.Msg("loaded graph configuration")
	}
}
