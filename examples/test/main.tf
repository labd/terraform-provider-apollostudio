terraform {
  required_providers {
    apollostudio = {
      source = "hashicorp.com/edu/apollostudio"
    }
  }
}

provider "apollostudio" {
  api_key = "<...>"
}

#data "apollostudio_sub_graph_validation" "example" {
#  schema_id = "deividas-pekunas-tea-4bjcrq"
#  schema_variant = "main"
#  sub_graph_schema = "type Query  extend type Query { topProducts(first: Int = 5): [Product] }  type Product   @key(fields: \"upc\") {   upc: String!   name: String   total: Int   weight: Int }"
#  sub_graph_name = "reviews"
#}

#resource "apollostudio_sub_graph" "example" {
#  schema_id = "deividas-pekunas-tea-4bjcrq"
#  schema_variant = "main"
#  sub_graph_schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: \"id\") { id: String! name: String price: Int weight: Int }"
#  sub_graph_name = "cucumbers"
#  sub_graph_url = "https://1193-78-58-36-253.ngrok-free.app"
#}

#output "test" {
#  value = apollostudio_sub_graph.example.revision
#}