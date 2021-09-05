package main

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/raphaelreyna/graphqld/internal/config"
	graphhost "github.com/raphaelreyna/graphqld/internal/graphHost"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	if os.Getenv("DEV") != "" {
		zerolog.SetGlobalLevel(-1)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}

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

	log.Info().
		Interface("configuration", c).
		Msg("loaded configuration")

	for _, g := range c.Graphs {
		gh, err := graphhost.NewGraphHost(c.Addr, g)
		if err != nil {
			log.Fatal().Err(err).
				Msg("error creating new graph host")
		}

		graphHosts[gh.Config.ServerName] = gh

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

	if err := http.ListenAndServe(c.Addr, nil); err != nil {
		log.Fatal().Err(err).
			Msg("error listening and serving")
	}

	if singleGraph != nil {
		singleGraph.Stop()
	} else {
		for _, gh := range graphHosts {
			gh.Stop()
		}
	}
}
