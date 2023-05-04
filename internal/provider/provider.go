package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/apollostudio-go-sdk/pkg/apollostudio"
	"os"
	"regexp"
	"time"
)

var (
	_ provider.Provider = &ApollostudioProvider{}
)

const retryTimeout = 5 * time.Second

// ApollostudioProvider defines the provider implementation.
type ApollostudioProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ApollostudioProviderModel describes the provider data model.
type ApollostudioProviderModel struct {
	ApiKey   types.String `tfsdk:"api_key"`
	GraphRef types.String `tfsdk:"graph_ref"`
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

func (p *ApollostudioProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Apollo Studio provider allows you to manage your Apollo Studio Graphs and Subgraphs.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Apollo studio graph API key",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"graph_ref": schema.StringAttribute{
				MarkdownDescription: "Apollo studio graph ref",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompilePOSIX(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+$`),
						"should be in the format of <graph-name>@<variant-name>",
					),
				},
			},
		},
	}
}

func (p *ApollostudioProvider) Configure(
	ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse,
) {
	var data ApollostudioProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var key string
	if data.ApiKey.IsUnknown() || data.ApiKey.IsNull() {
		key = os.Getenv("APOLLO_API_KEY")
	} else {
		key = data.ApiKey.ValueString()
	}

	var ref string
	if data.GraphRef.IsUnknown() || data.GraphRef.IsNull() {
		ref = os.Getenv("APOLLO_GRAPH_REF")
	} else {
		ref = data.GraphRef.ValueString()
	}

	if key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Apollo Studio API key",
			"Please set the api_key attribute or the APOLLO_KEY environment variable",
		)
		return
	}

	if ref == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("graph_ref"),
			"Missing Apollo Studio Graph ref",
			"Please set the graph_ref attribute or the APOLLO_GRAPH_REF environment variable",
		)
		return
	}

	client, err := apollostudio.NewClient(
		apollostudio.ClientOpts{
			APIKey:   key,
			GraphRef: ref,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Apollo Studio client",
			"Please check your API key and Graph ref",
		)
		return
	}

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
