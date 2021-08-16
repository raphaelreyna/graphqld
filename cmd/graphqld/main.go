package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
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

	http.HandleFunc("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts := handler.NewRequestOptions(r)
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  opts.Query,
			VariableValues: opts.Variables,
			OperationName:  opts.OperationName,
			Context:        r.Context(),
		})
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))

	port := "8080"
	if x := os.Getenv("PORT"); x != "" {
		port = x
	}
	http.ListenAndServe(":"+port, nil)
}
