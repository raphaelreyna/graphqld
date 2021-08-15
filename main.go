package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
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

	g := graph{
		rootDir: rootDir,
	}

	if err := g.synthesizeRootQueryConf(); err != nil {
		panic(err)
	}

	fmt.Printf("graph: %+v\n", g.uninstantiatedTypes)

	if err := g.instantiateTypesObjects(); err != nil {
		panic(err)
	}
	g.setTypes()
	if err := g.rootQuery.SetResolvers(); err != nil {
		panic(err)
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(g.rootQuery.ObjectConf),
	})
	if err != nil {
		panic(err)
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
	fmt.Println(http.ListenAndServe(":"+port, nil))
}
