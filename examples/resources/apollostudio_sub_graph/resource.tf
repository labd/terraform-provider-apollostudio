resource "apollostudio_sub_graph" "example" {
  name   = "sub-graph-name"
  url    = "https://example.com/graphql"
  schema = "schema { query: Query } type Query { hello: String }"
}
