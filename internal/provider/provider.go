package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
)

var (
	_ provider.Provider             = &ApollostudioProvider{}
	_ provider.ProviderWithMetadata = &ApollostudioProvider{}
)

// ApollostudioProvider defines the provider implementation.
type ApollostudioProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ApollostudioProviderModel describes the provider data model.
type ApollostudioProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApollostudioProvider{
			version: version,
		}
	}
}

func (p *ApollostudioProvider) Metadata(
	_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse,
) {
	resp.TypeName = "apollostudio"
	resp.Version = p.version
}

func (p *ApollostudioProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"api_key": {
				Type:                types.StringType,
				MarkdownDescription: "Apollo studio graph API key",
				Required:            true,
				Sensitive:           true,
			},
		},
	}, nil
}

func (p *ApollostudioProvider) Configure(
	ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse,
) {
	var data ApollostudioProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.ApiKey.ValueString()

	if key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Apollo Studio API key",
			"The provider cannot create the Apollo Studio API client as there is a missing or empty value for the Apollo Studio API key. "+
				"Set the API key value in the configuration. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := apollostudio.NewClient(
		apollostudio.ClientOpts{
			APIKey: key,
		},
	)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApollostudioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewValidationDataSource,
	}
}

func (p *ApollostudioProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubGraphResource,
	}
}
