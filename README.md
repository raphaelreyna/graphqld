<img src="https://raw.githubusercontent.com/raphaelreyna/graphqld/master/logo/graphqld.png" width="500" height="130">

Build GraphQL backends using [CGI](https://en.wikipedia.org/wiki/Common_Gateway_Interface) style executables.

# Overview
Graphqld aims to bring the two main advantages of CGI to GraphQL:
- CGI executables are language independent, allowing developers to use their prefered language, even across teams.
- The CGI server (or graphqld) provides a simple interface for responding to HTTP (or GraphQL) requests.


## How it works
Graphqld relies on two types of files in its root directory, `.graphql` schema files and executables/scripts which are the resolvers.
All GraphQL types, except for objects, are defined using only the schema files (or one big schema file, its up to you). 
Objects are defined in 3 ways:
- In a schema file.
- By directories and executables/scripts. Each directory corresponds to an object with that same name; executables/scripts that output field definitions when passed the `--graphqld-fields` flag become resolvers for those fields. The fields defined by executables are added to the object defined by their parent directory.
- A combination of the first two.

Any directory that defines a graph must contains either a `Query` or `Mutation` directory (or both).

## Features
- Hot server / live reloading; rebuild a graph any time a file changes in its root directory.
- Multiple graphs; graphqld can serve up multiple graphs, each on its own domain name.
- Built in GraphiQL server; easily explore your graphs.
- Built in HTTP username and password authentication.
- TLS/HTTPS support.
- CORS support.
- Access HTTP request info through environment variables in resolvers; resolvers can access cookies, header values, and request info.
- Set HTTP header values from resolvers; just like with CGI, resolvers can set headers and write cookies.
- Flexible contexts; the graphql context passed to each resolver is availble as a JSON file at `/dev/fd/3` and can be statically set from a config file or dynamically created using a designated executable.
- Flexible logging; graphqld can do either structured logging or pretty-printed human-friendly logging (with color!)



### Serving multiple graphs
If graphqld finds directories named `Query` or `Mutation` in its root directoty, it will serve up a single graph with the Query object defined by `Query` and the Mutation object by `Mutation`.

If neither a `Query` or `Mutation` directory is found in the root directory, graphqld will assume multiple graphs are to be served.
Each graph is defined by a directory, where the directory name is the hostname (ex "mycoolgraph.io") for that particular graph.
Each of these directories should then each contain either a `Query` directory or a `Mutation` directory (or both)

### Still missing...
- support for defining abstract types (interfaces and unions)
- full blown context support (not just JSON), although this is most likely too difficult / not possible.

# Examples
Two examples are provided [here](https://github.com/raphaelreyna/graphqld/tree/master/examples/graphqld) and an example configuration file [here](https://github.com/raphaelreyna/graphqld/blob/master/examples/graphqld.yaml).

### Serving both example graphs
Run graphqld from the `./examples` directory. Done.

This directory contains a configration file `graphqld.yaml` which sets `./examples/graphqld` as the root directory.
Since `./examples/graphqld` contains two directories each defining a graph, they're each served and accessed via their directory name as the hostname (`http://example1.localhost/`) to reach the graph defined in `./examples/graphqld/example1.localhost`.

### Serving a single example graph
Modify the config file at `./examples/graphqld.yaml` by changing the `root` value from `./graphqld` to `./graphqld/example1`, then run graphqld from the `./examples`  directory. Done.

Since the root directory we just set in our config file contains a `Query` directory, graphqld recognizes the root dir as defining a single graph.

# Configuration
On startup, graphqld will look for a configuration file `graphqld.yaml` in `.`, `$HOME/.config/graphqld` or `/etc` (in that order).
Config values set via the environment will override those set by a config file.

## example `graphqld.yaml` configuration file
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
context:
  # execPath is the path to an executable that will read the request
  # and return some json as the context to be passed to the resolvers.
  execPath: ""

  # tmpDir is where the context file for each request will be written to
  # defaults to your tmp dir
  tmpDir: ""

  # context allows for a static context to be passed to the resolvers.
  # this should be marshalable as JSON.
  # this context is ignored if execPath is not empty.
  context:
    loggedIn: true
    user: "graphqld"

log:
  json: false
  # color does nothing if json is true
  color: true
  # acceptable levels: info | warn | error | fatal | disabled
  level: "info"

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
      execPath: "./graphqld/example1.localhost/auth.py"
    cors:
      allowCredentials: true
      allowedHeaders:
        - "Host"
      allowedOrigins:
        - "localhost"
        - "127.0.0.1"
      ignoreOptions: true
```

## environment variables
### `GRAPHQLD_CTX_EXEC_PATH`
- Description: Defines the default path to an executable that can provide a context for each incoming request.
- Default: ""

### `GRAPHQLD_CTX_TMP_DIR`
- Description: The directory where the temporary context files will be written to.
- Default: ""

### `GRAPHQLD_RESOLVER_DIR`
- Description: The working directory for resolvers.
- Default: "/"

### `GRAPHQLD_LOG_JSON`
- Description: Log structured JSON.
- Default: false

### `GRAPHQLD_LOG_COLOR`
- Description: Use color for human-friendly logging.
- Default: true

### `GRAPHQLD_MAX_BODY_SIZE`
- Description: The max size of an incoming request body.
- Default: 1048576

### `GRAPHQLD_HOSTNAME`
- Description: The hostname that graphqld will listen for.
- Default: ""

### `GRAPHQLD_ADDRESS`
- Description: The address that graphqld will bind to.
- Default: ":80"

### `GRAPHQLD_ROOT`
- Description: The root directory of graphqld, where it looks for graphs.
- Default: "/var/graphqld"

### `GRAPHQLD_HOT`
- Description: Rebuild graphs when a file in their root directory changes.
- Default: false

### `GRAPHQLD_GRAPHIQL`
- Description: Serve a GraphiQL client at "/graphiql"
- Default: false
