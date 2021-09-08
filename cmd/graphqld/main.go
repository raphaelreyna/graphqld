package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/raphaelreyna/graphqld/internal/config"
	graphhost "github.com/raphaelreyna/graphqld/internal/graphHost"
	"github.com/rs/zerolog/log"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal().
				Interface("recovered", r)
		}
	}()

	var (
		c = config.Config

		graphHosts  = make(map[string]*graphhost.GraphHost)
		singleGraph *graphhost.GraphHost
	)

	{
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
	}

	for _, g := range c.Graphs {
		gh, err := graphhost.NewGraphHost(c.Addr, c.MaxBodyReadSize, g)
		if err != nil {
			log.Fatal().Err(err).
				Msg("error creating new graph host")
		}

		graphHosts[gh.Config.ServerName] = gh

		singleGraph = gh

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

	if 1 < len(graphHosts) {
		singleGraph = nil
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var gh = singleGraph

		if gh == nil {
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				if !strings.Contains(err.Error(), "missing port") {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				host = r.Host
			}

			gh = graphHosts[host]
			if gh == nil {
				panic(host)
			}

			if host != gh.Config.ServerName {
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
		} else if gh.Config.ServerName != "" {
			host, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if host != gh.Config.ServerName {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		gh.ServeHTTP(w, r)
	}))

	log.Info().Msg("listening for HTTP traffic ...")

	if err := http.ListenAndServe(c.Addr, nil); err != nil {
		log.Fatal().Err(err).
			Msg("error listening and serving")
	}

	log.Info().Msg("... stopped listening for HTTP traffic")

	if singleGraph != nil {
		singleGraph.Stop()
	} else {
		for _, gh := range graphHosts {
			gh.Stop()
		}
	}
}
