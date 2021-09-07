package graphhost

import (
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/radovskyb/watcher"
	"github.com/raphaelreyna/graphqld/internal/graph"
	"github.com/rs/zerolog/log"
)

type fileWatcher struct {
	RootDir  string
	Interval time.Duration

	Schema   chan graphql.Schema
	ServerMu *sync.RWMutex

	w *watcher.Watcher
}

func (w *fileWatcher) Run() error {
	w.w = watcher.New()
	w.w.SetMaxEvents(1)
	w.w.FilterOps(
		watcher.Rename, watcher.Move,
		watcher.Create, watcher.Chmod,
		watcher.Write,
	)

	go func() {
		for {
			select {
			case <-w.w.Event:
				g := graph.Graph{
					DocumentRoot: w.RootDir,
				}

				if err := g.Build(); err != nil {
					log.Error().Err(err).
						Msg("unable to rebuild graph schema config")
					continue
				}

				var schemaConf graphql.SchemaConfig
				if q := g.Query; q != nil {
					schemaConf.Query = q
				}
				if m := g.Mutation; m != nil {
					schemaConf.Mutation = m
				}

				schm, err := graphql.NewSchema(schemaConf)
				if err != nil {
					log.Error().Err(err).
						Msg("unable to rebuild schema")
					continue
				}

				w.ServerMu.Lock()
				w.Schema <- schm
				w.ServerMu.Unlock()
			case err := <-w.w.Error:
				log.Fatal().Err(err).
					Msg("unable to watch root directory")
			case <-w.w.Closed:
				return
			}
		}
	}()

	if err := w.w.AddRecursive(w.RootDir); err != nil {
		return err
	}

	return w.w.Start(w.Interval)
}

func (w *fileWatcher) Stop() {
	w.w.Close()
}
