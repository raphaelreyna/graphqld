<img src="https://raw.githubusercontent.com/raphaelreyna/graphqld/master/logo/graphqld.png" width="500" height="130">


# graphqld
Do you miss being able to throw CGI scripts into an FTP server and call it a day?
Is GraphQL the only thing holding you back from living in the past?

If so, then try graphqld, the graphql "CGI" server.

Still an experiment/poc.

# Working example
Consider the following schema:
```graphql
type Query {
  python(ssss: PythonInput): String!
  javascript: String
  charCount(string: String!): CharCountResponse!
}
 
type CharCountResponse {
  string: String!
  count: Int!
  isEven: Boolean!
}

input PythonInput {
  Name: String!
  LevelTwo: PythonInputTwo
}

input PythonInputTwo {
  IncludeVersion: Boolean!
}
 ```
To serve up this graph with graphqld we create a directory with the following structure and run `graphqld -f ./example/graphqld.yaml`:
```
example/graph
├── charCount.py
├── CharCountResponse
│   ├── CharCountResponse.graphql
│   ├── isEven.py
│   └── IsEvenResponse
│       └── IsEvenResponse.graphql
├── javascript.js
├── PythonInput
│   ├── PythonInput.graphql
│   └── PythonInputTwo
│       └── PythonInputTwo.graphql
└── python.py
```

See [the example directory](https://github.com/raphaelreyna/graphqld/tree/master/example/graph) to check out examples of these scripts.


# Contexts
While full blown support for contexts including passing around functions and objects is tricky (thoughts/suggestions are welcome!).


What is currently supported is a static JSON context.
If the `GRAPHQLD_CTX_EXEC` env var points to an executable, that executble will be ran at the beginning of every request, before graph resolution happens. Its output will be copied into a file which will be made available to each resolver at `/dev/fd/3`. The HTTP header from the incoming request will be made available to this context providing executable, just as with CGI.


This at least allows for some level of auth.


See [the example auth python script](https://github.com/raphaelreyna/graphqld/tree/master/example/auth.py) to check out example of a context providing executable.

# Extras
### Hot Server / Live Reloading
The graph can be rebuilt whenever theres a change in the root directory.
To enable this, simply set `GRAPHQLD_HOT_RELOAD=TRUE`.

### GraphiQL
To enable the built-in GraphiQL server at `/graphiql`, simply set `GRAPHQLD_GRAPHIQL=TRUE`.


# How it works
The graph is built scanning the given directory and querying each executable for its fields:
```bash
$ ./example/charCount.py --cggi-fields 
["charCount(string: String!): CharCountResponse!"]

$ ./example/python.py --cggi-fields
["python: String!"]

$ ./example/javascript.js --cggi-fields
["javascript: String"]
```
The the top level executables are understood to collectively define the root `query` type; complex types are only added when referenced by the graph.

Field arguments are provided via os args:
```bash
$ ./example/charCount.py --string hello
{"string": "hello", "count": 5}
```

Each executable is used as the resolver for the field it reports.
Complex types with no resolvers may be definied with a `{{TYPE_NAME}}/{{TYPE_NAME}}.graphql` file.

# Still missing...
- support for defining interfaces
- support for defining enums
- a lot of other things, this is still a pretty early stage project
