package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/apollostudio-go-sdk/apollostudio"
	"github.com/labd/terraform-provider-apollostudio/internal/utils"
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
	ID      types.String `tfsdk:"id"`
	Schema  types.String `tfsdk:"schema"`
	Name    types.String `tfsdk:"name"`
	Changes types.String `tfsdk:"changes"`
}

func (d *ValidationDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sub_graph_validation"
}

func (d *ValidationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source is used to apply schema validation checks. " +
			"It applies Composition and Operation checks to the provided schema. If the schema is invalid, " +
			"the data source will return an error. If the schema is valid, the data source will return the " +
			"schema name and a list of changes detected on the sub graph. " +
			"More information about schema validation can be found in the " +
			"[Apollo Studio documentation](https://www.apollographql.com/docs/graphos/delivery/schema-checks/).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the sub graph",
				Computed:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The sub graph SDL schema",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The sub graph name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"changes": schema.StringAttribute{
				MarkdownDescription: "The sub graph changes",
				Computed:            true,
			},
		},
	}
}

func (d *ValidationDataSource) Configure(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
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
	var state ValidationDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := state.Schema.ValueString()
	name := state.Name.ValueString()

	// we are checking if schema with provided name already exists
	// because validation may succeed even if provided schema does exist
	graph, err := d.client.GetSubGraph(ctx, name)
	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if graph.Name == "" {
		resp.Diagnostics.AddError("Sub graph not found", fmt.Sprintf("Sub graph \"%s\" not found", name))
		return
	}

	result, err := d.client.ValidateSubGraph(
		ctx, &apollostudio.ValidateOptions{
			SubGraphSchema: []byte(s),
			SubGraphName:   name,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable validate schema, got error: %s", err),
		)
		return
	}

	if !result.IsValid() {
		errs := result.Errors()
		if len(errs) == 0 {
			resp.Diagnostics.AddError("Validation Error", "Unable to validate schema, but got no errors")
			return
		}
		for _, e := range errs {
			resp.Diagnostics.AddError(fmt.Sprintf("Sub Graph validation failed: %s", e.Code), e.Message)
		}
		return
	}

	if len(result.Changes()) > 0 {
		state.Changes = types.StringValue(strings.Join(result.Changes(), ","))
		resp.Diagnostics.AddWarning(
			"Sub Graph changes detected",
			fmt.Sprintf("%d changes detected on \"%s\" sub graph", len(result.Changes()), name),
		)
		for _, c := range result.Changes() {
			resp.Diagnostics.AddWarning(c, "")
		}
	}

	state.ID = types.StringValue(name)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
