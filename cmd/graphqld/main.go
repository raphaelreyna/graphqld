package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/radovskyb/watcher"
	"github.com/raphaelreyna/graphqld/internal/graph"
)

func main() {
	var (
		retCode = 1
		rootDir string
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
		os.Exit(retCode)
	}()

	if len(os.Args) < 2 {
		fmt.Println("no root dir given")
		return
	}
	rootDir = os.Args[1]

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	g := graph.Graph{
		Dir: rootDir,
	}

	if err := g.Build(); err != nil {
		log.Fatal(err)
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(g.Query.ObjectConf),
	})
	if err != nil {
		log.Fatal(err)
	}

	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(
		watcher.Rename, watcher.Move,
		watcher.Create, watcher.Chmod,
		watcher.Write,
	)

	lock := sync.RWMutex{}
	go func() {
		for {
			select {
			case <-w.Event:
				g := graph.Graph{
					Dir: rootDir,
				}

				if err := g.Build(); err != nil {
					log.Fatal(err)
				}

				schm, err := graphql.NewSchema(graphql.SchemaConfig{
					Query: graphql.NewObject(g.Query.ObjectConf),
				})
				if err != nil {
					log.Fatal(err)
				}

				lock.Lock()
				schema = schm
				lock.Unlock()
				fmt.Println("updated schema")

			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(g.Dir); err != nil {
		log.Fatalln(err)
	}
	go func() {
		if err := w.Start(time.Second); err != nil {
			log.Fatalln(err)
		}
	}()

	http.HandleFunc("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(result)
	}))

	port := "8080"
	if x := os.Getenv("PORT"); x != "" {
		port = x
	}
	fmt.Println("starting...")
	http.ListenAndServe(":"+port, nil)
}
