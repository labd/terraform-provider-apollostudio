package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
	"strings"
)

var _ datasource.DataSource = &ValidationDataSource{}

func NewValidationDataSource() datasource.DataSource {
	return &ValidationDataSource{}
}

// ValidationDataSource defines the data source implementation.
type ValidationDataSource struct {
	client *apollostudio.Client
}

// ValidationDataSourceModel describes the data source data model.
type ValidationDataSourceModel struct {
	Schema  types.String `tfsdk:"schema"`
	Name    types.String `tfsdk:"name"`
	Changes types.String `tfsdk:"changes"`
}

func (d *ValidationDataSource) Metadata(
	ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sub_graph_validation"
}

func (d *ValidationDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Fields required to validate sub graph",

		Attributes: map[string]tfsdk.Attribute{
			"sub_graph_schema": {
				MarkdownDescription: "Sub Graph SDL schema",
				Type:                types.StringType,
				Required:            true,
			},
			"sub_graph_name": {
				MarkdownDescription: "Sub Graph name",
				Type:                types.StringType,
				Optional:            true,
			},
			"changes": {
				MarkdownDescription: "Last generated changes",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
}

func (d *ValidationDataSource) Configure(
	ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apollostudio.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *apollostudio.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	d.client = client
}

func (d *ValidationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ValidationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	schema := data.Schema.ValueString()
	name := data.Name.ValueString()

	result, err := d.client.ValidateSubGraph(
		ctx, &apollostudio.ValidateOptions{
			SubGraphSchema: []byte(schema),
			SubGraphName:   name,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable validate schema, got error: %s", err))
		return
	}

	if !result.IsValid() {
		errs := result.Errors()
		if len(errs) == 0 {
			resp.Diagnostics.AddError("Validation Error", "Unable to validate schema, but got no errors")
			return
		}
		for _, e := range errs {
			resp.Diagnostics.AddError(e.Code, e.Message)
		}
		return
	}

	if len(result.Changes()) > 0 {
		data.Changes = types.StringValue(strings.Join(result.Changes(), ","))
		resp.Diagnostics.AddWarning("Changes Detected", "Changes detected in sub graph")
		for _, c := range result.Changes() {
			resp.Diagnostics.AddWarning(c, "")
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
