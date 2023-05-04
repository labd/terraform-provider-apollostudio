package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-apollostudio/internal/acctest"
	"github.com/labd/apollostudio-go-sdk/pkg/apollostudio"
	"github.com/labd/terraform-provider-apollostudio/internal/utils"
)

func TestAccSubGraphValidation_ChangesAndErrors(t *testing.T) {
	var graph apollostudio.SubGraphResult

	schema1 := "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name1: String price: Int weight: Int }"
	schema2 := "type Query extend type Query { topCucumbers(first: Int = 5): [Cucumber] } type Cucumber @key(fields: id) { id: String! name1: String total: Int weight: Int }"
	schema3 := "type Query extend type Query { topCucumbers(first: Int = 5): [Test] } type Cucumber @key(fields: id) { id: String! name1: String total: Int weight: Int }"
	name := "vegetables"
	url := "https://example.com/graphql"

	n := "apollostudio_sub_graph.vegetables"

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			CheckDestroy:             testAccCheckSubGraphResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccSubGraphConfig(name, schema1, name, url),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSubGraphResourceExists(n, &graph),
					),
				},
				{
					Config: testAccSubGraphValidateConfig(url, schema1, schema2, name),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSubGraphValidateChanges(n, name, schema2, false),
					),
				},
				{
					Config:      testAccSubGraphValidateConfig(url, schema1, schema3, name),
					ExpectError: regexp.MustCompile("INVALID_GRAPHQL"),
				},
			},
		},
	)
}

func testAccSubGraphValidateConfig(url, schema, newSchema, name string) string {
	return utils.HCLTemplate(
		`
		resource "apollostudio_sub_graph" "vegetables" {
		  schema = "{{ .schema }}"
		  name = "{{ .name }}"
		  url = "{{ .url }}"
		}

		data "apollostudio_sub_graph_validation" "vegetables" {
		  schema = "{{ .newSchema }}"
		  name = "{{ .name }}"
		}
		`,
		map[string]any{
			"schema":    schema,
			"newSchema": newSchema,
			"name":      name,
			"url":       url,
		},
	)
}

func testAccCheckSubGraphValidateChanges(n, name, schema string, errors bool) resource.TestCheckFunc {
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

		r, err := client.ValidateSubGraph(
			ctx, &apollostudio.ValidateOptions{
				SubGraphName:   name,
				SubGraphSchema: []byte(schema),
			},
		)

		if err != nil {
			return err
		}

		if errors {
			if len(r.Errors()) == 0 {
				return fmt.Errorf("expected validation errors, but got none")
			}
			codes := make([]string, len(r.Errors()))
			for i, e := range r.Errors() {
				codes[i] = e.Code
			}
		}

		if !errors {
			if len(r.Changes()) == 0 {
				return fmt.Errorf("expected validation changes, but got none")
			}
		}

		return nil
	}
}
