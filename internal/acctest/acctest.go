package acctest

import (
	"context"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"os"
)

func ConfigureProvider(p tfprotov5.ProviderServer) error {
	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":   tftypes.String,
			"graph_ref": tftypes.String,
		},
	}

	testValue := tftypes.NewValue(
		testType, map[string]tftypes.Value{
			"api_key":   tftypes.NewValue(tftypes.String, os.Getenv("APOLLO_API_KEY")),
			"graph_ref": tftypes.NewValue(tftypes.String, os.Getenv("APOLLO_GRAPH_REF")),
		},
	)

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)
	if err != nil {
		return err
	}

	_, err = p.ConfigureProvider(
		context.TODO(), &tfprotov5.ConfigureProviderRequest{
			TerraformVersion: "1.0.0",
			Config:           &testDynamicValue,
		},
	)

	return err
}
