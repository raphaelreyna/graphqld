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

# Configuration
```yaml
# hostname is ignored if serving multiple graphs;
# each graphs serverName will be used instead.
#
# Default: ""
hostname: "localhost"

# address is the address:port that graphqld will bind to.
#
# Default: 80
address: ":8082"

# root is the dir where graphqld will look for its graph(s)
# Single graph: the root dir should contain a dir name Query or Mutation or both
# Multiple graphs: each graph in its own dir with that graphs servername / hostname as the dir name.
#
# Default: "."
root: "./graphqld"

# hot enables global hot reloading; this can be overriden 
# by each graph config in the graphs section.
#
# Default: false
hot: false

# graphiql enables a graphiql server for each graph at "/graphiql"; this can be overriden
# by each graph config in the graphs section.
#
# Default: false
graphiql: false

# resolverWD is the default working directory for resolvers; this can be overriden
# by each graph config in the graphs section.
#
# Default: "/"
resolverWD: "."

# If basicAuth is set, graphqld will expect the HTTP header
# Authorization: Basic <BASE-64>
# where <BASE-64> is the base64 encoding of username:password
basicAuth:
  username: "test"
  password: "test"

#cors:
#  allowCredentials: true
#  allowedHeaders:
#    - "Host"
#  allowedOrigins:
#    - "localhost"
#    - "127.0.0.1"
#  ignoreOptions: true

#tls:
#  cert: "path/to/cert/file"
#  key: "path/to/key/file"

# context allows for a static context to be passed to the resolvers.
# this should be marshalable as JSON.
# this context is ignored if execPath is not empty.
#context:
   # execPath is the path to an executable that will read the request
   # and return some json as the context to be passed to the resolvers.
#  execPath: ""

   # tmpDir is where the context file for each request will be written to
   # defaults to your tmp dir
#  tmpDir: ""

   # context allows for a static context to be passed to the resolvers.
   # this should be marshalable as JSON.
   # this context is ignored if execPath is not empty.
#  context:

# graphs is a list of graph specific configurations.
# Each graphs serverName must match the name of the graphs directory in the root dir.
#
# This section is ignored if there is only one graph in the root dir; 
# root level defaults will be applied instead.
graphs:
  - serverName: "example1.localhost"
    graphiql: true
    hot: false
    workingDir: "."
    context:
      execPath: "./example1.localhost/auth.py"
#     context:
#       loggedIn: true
#       user:
#         name: "yaml"
    cors:
      allowCredentials: true
      allowedHeaders:
        - "Host"
      allowedOrigins:
        - "localhost"
        - "127.0.0.1"
      ignoreOptions: true

```

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
- a lot of other things, this is still a pretty early stage project
