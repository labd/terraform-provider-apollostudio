package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/apollostudio-go-sdk/apollostudio"
	"github.com/labd/terraform-provider-apollostudio/internal/acctest"
	"github.com/labd/terraform-provider-apollostudio/internal/utils"
)

func TestAccSubGraph_basic(t *testing.T) {
	var graph apollostudio.SubGraphResult

	schema := "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name1: String price: Int weight: Int }"
	name1 := "vegetables"
	name2 := "fruits"
	url := "https://example.com/graphql"
	n := "apollostudio_sub_graph.fruits_sub_graph"

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			CheckDestroy:             testAccCheckSubGraphResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccSubGraphConfig("fruits_sub_graph", schema, name1, url),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSubGraphResourceExists(n, &graph),
						resource.TestCheckResourceAttr(n, "schema", schema),
						resource.TestCheckResourceAttr(n, "name", name1),
						resource.TestCheckResourceAttr(n, "url", url),
					),
				},
				{
					Config: testAccSubGraphConfig("fruits_sub_graph", schema, name2, url),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSubGraphResourceExists(n, &graph),
						testAccCheckSubGraphAttributes(&graph, name2),
						testAccCheckSubGraphResourceNotExists(name1),
						resource.TestCheckResourceAttr(n, "schema", schema),
						resource.TestCheckResourceAttr(n, "name", name2),
						resource.TestCheckResourceAttr(n, "url", url),
					),
				},
			},
		},
	)
}

func testAccSubGraphConfig(res, schema, name, url string) string {
	return utils.HCLTemplate(
		`
		resource "apollostudio_sub_graph" {{ .res }} {
		  schema = "{{ .schema }}"
		  name = "{{ .name }}"
		  url = "{{ .url }}"
		}
		`,
		map[string]any{
			"res":    res,
			"schema": schema,
			"name":   name,
			"url":    url,
		},
	)
}

// testAccCheckSubGraphResourceDestroy verifies the Widget
// has been destroyed
func testAccCheckSubGraphResourceDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "apollostudio_sub_graph" {
			continue
		}
		graph, err := client.GetSubGraph(ctx, rs.Primary.ID)
		if err == nil && graph.Name == rs.Primary.ID {
			return fmt.Errorf("sub graph (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckSubGraphAttributes(graph *apollostudio.SubGraphResult, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if graph.Name != name {
			return fmt.Errorf("bad name: %s", graph.Name)
		}

		return nil
	}
}

func testAccCheckSubGraphResourceExists(n string, graph *apollostudio.SubGraphResult) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no sub graph ID is set")
		}

		client, err := acctest.GetClient()
		if err != nil {
			return err
		}

		ctx := context.Background()

		g, err := client.GetSubGraph(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		*graph = *g

		return nil
	}
}

func testAccCheckSubGraphResourceNotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := acctest.GetClient()
		if err != nil {
			return err
		}

		ctx := context.Background()

		g, err := client.GetSubGraph(ctx, name)
		if err != nil {
			return err
		}

		if g.Name != "" {
			return fmt.Errorf("sub graph (%s) still exists", name)
		}

		return nil
	}
}
