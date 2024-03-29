package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/labd/terraform-provider-apollostudio/internal/acctest"
)

var testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"apollostudio": func() (tfprotov5.ProviderServer, error) {
		p, err := providerserver.NewProtocol5WithError(New("test", true)())()
		if err != nil {
			return nil, err
		}
		err = acctest.ConfigureProvider(p)
		if err != nil {
			return nil, err
		}
		return p, nil
	},
}

func testAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"APOLLO_API_KEY",
		"APOLLO_GRAPH_REF",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}
}
