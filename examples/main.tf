terraform {
  required_providers {
    apollostudio = {
      source = "registry.terraform.io/labd/apollostudio"
    }
  }
}

provider "apollostudio" {
  api_key   = "service:deividas-pekunas-tea-4bjcrq:7LsCOuBlS8UIK-TVggdoIQ"
  graph_ref = "deividas-pekunas-tea-4bjcrq@main"
}

data "apollostudio_sub_graph_validation" "example" {
  schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name: String total: Int weight: Int }"
  // schema = "type Query  extend type Query { topProducts(first: Int = 5): [Product] }  type Product   @key(fields: upc) {   upc: String!   name: String   total: Int   weight: Int }"
  name = "ff"
}

resource "apollostudio_sub_graph" "example" {
  schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name: String price: Int weight: Int }"
  // schema = "type Query  extend type Query { topProducts(first: Int = 5): [Product] }  type Product   @key(fields: upc) {   upc: String!   name: String   total: Int   weight: Int }"
  name = "ff"
  url  = "https://example.com/graphql"
}

#resource "apollostudio_sub_graph" "example2" {
#  schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name: String total: Int weight: Int }"
#  // schema = "type Query  extend type Query { topProducts(first: Int = 5): [Product] }  type Product   @key(fields: upc) {   upc: String!   name: String   total: Int   weight: Int }"
#  name = "bb"
#  url = "https://example.com/graphql"
#}

#output "revision" {
#  value = apollostudio_sub_graph.example.revision
#}
#
#output "sub_graph_schema" {
#  value = apollostudio_sub_graph.example.schema
#}