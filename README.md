![graphqld](https://raw.githubusercontent.com/raphaelreyna/graphqld/master/logo/graphqld.svg)
# graphqld
Do you miss being able to throw CGI scripts into an FTP server and call it a day?
Is GraphQL the only thing holding you back from living in the past?

If so, then try graphqld, the graphql "CGI" server.

Still an experiment/poc.

# Working example
Consider the following schema:
```graphql
type Query {
  python: String!
  javascript: String
  charCount(string: String!): CharCountResponse!
}
 
type CharCountResponse {
  string: String!
  count: Int!
}
 ```
To serve up this graph on port 8000 with graphqld we create a directory with the following structure and run `PORT=8000 graphqld ./example`:
```
example
├── charCount.py
├── CharCountResponse
│   └── CharCountResponse.graphql
├── javascript.js
└── python.py
```


See [the example directory](https://github.com/raphaelreyna/graphqld/tree/master/example) to check out examples of these scripts.


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
- support for defining input types
- support for defining interfaces
- support for defining enums
- context
- a lot of other things, this is still a pretty early stage project
