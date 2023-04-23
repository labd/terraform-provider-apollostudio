terraform {
  required_providers {
    apollostudio = {
      source = "hashicorp.com/edu/apollostudio"
    }
  }
}

provider "apollostudio" {
  api_key = "<...>"
  graph_ref = "deividas-pekunas-tea-4bjcrq@main"
}
#
data "apollostudio_sub_graph_validation" "example" {
  schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: \"id\") { id: String! name: String price: Int weight: Int }"
  name = "e"
}

resource "apollostudio_sub_graph" "example" {
  schema = "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: \"id\") { id: String! name: String total: Int weight: Int }"
  name = "e"
  url = "https://1193-78-58-36-253.ngrok-free.app"
}