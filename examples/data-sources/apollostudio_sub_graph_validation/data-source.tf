data "apollostudio_sub_graph_validation" "example" {
  name   = "sub-graph-name"
  schema = "schema { query: Query } type Query { hello: String }"
}
