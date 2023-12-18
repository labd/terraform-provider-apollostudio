package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/apollostudio-go-sdk/apollostudio"
	"github.com/labd/terraform-provider-apollostudio/internal/utils"
)

var (
	_ resource.Resource                = &SubGraphResource{}
	_ resource.ResourceWithConfigure   = &SubGraphResource{}
	_ resource.ResourceWithImportState = &SubGraphResource{}
)

func NewSubGraphResource() resource.Resource {
	return &SubGraphResource{}
}

// SubGraphResource defines the resource implementation.
type SubGraphResource struct {
	client *apollostudio.Client
}

// SubGraphResourceModel describes the resource data model.
type SubGraphResourceModel struct {
	URL       types.String `tfsdk:"url"`
	Schema    types.String `tfsdk:"schema"`
	Name      types.String `tfsdk:"name"`
	ID        types.String `tfsdk:"id"`
	Revision  types.String `tfsdk:"revision"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *SubGraphResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sub_graph"
}

func (r *SubGraphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource is used to manage subgraphs within the Federated Apollo schema. " +
			"More information about the Apollo Federation subgraphs can be found " +
			"[here](https://www.apollographql.com/docs/federation/v1/subgraphs/).",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL of the sub graph endpoint",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The SDL schema of the sub graph",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the sub graph",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the sub graph",
				Computed:            true,
			},
			"revision": schema.StringAttribute{
				MarkdownDescription: "The revision of the sub graph",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation date of the sub graph",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last update date of the sub graph",
				Computed:            true,
			},
		},
	}
}

func (r *SubGraphResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
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

	r.client = client
}

func (r *SubGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubGraphResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.Schema.ValueString()
	name := plan.Name.ValueString()
	url := plan.URL.ValueString()

	graph, err := r.client.GetSubGraph(ctx, name)
	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.client.GetLatestSchemaBuild(ctx)
	utils.ProcessError(&resp.Diagnostics, err, "Federation s contains errors", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
			SubGraphSchema: []byte(s),
			SubGraphName:   name,
			SubGraphURL:    url,
		},
	)

	utils.ProcessError(&resp.Diagnostics, err, "Federation s error while submitting sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if !result.WasCreated && graph.Name != "" {
		resp.Diagnostics.AddWarning(
			"No new subgraph was created",
			fmt.Sprintf(
				"Sub Graph `%s` already exits, did not create new sub graph",
				name,
			),
		)
	}

	plan.ID = types.StringValue(name)
	plan.Revision = types.StringValue(graph.Revision)
	plan.CreatedAt = types.StringValue(graph.CreatedAt.Format(time.RFC850))
	plan.UpdatedAt = types.StringValue(graph.UpdatedAt.Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SubGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubGraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	result, err := r.client.GetSubGraph(ctx, name)

	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		state.ID = types.StringValue(name)
	}
	if state.URL.IsNull() && result.URL != "" {
		state.URL = types.StringValue(result.URL)
	}
	if state.Schema.IsNull() {
		state.Schema = types.StringValue(result.ActivePartialSchema.Sdl)
	}
	state.Revision = types.StringValue(result.Revision)
	state.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC850))
	state.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC850))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubGraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SubGraphResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.Schema.ValueString()
	name := plan.Name.ValueString()
	url := plan.URL.ValueString()

	err := retry.RetryContext(
		ctx, retryTimeout, func() *retry.RetryError {
			var err error
			_, err = r.client.SubmitSubGraph(
				ctx, &apollostudio.SubmitOptions{
					SubGraphSchema: []byte(s),
					SubGraphName:   name,
					SubGraphURL:    url,
				},
			)
			return utils.NewRetryableError(err)
		},
	)

	utils.ProcessError(&resp.Diagnostics, err, "Federation s error while submitting sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.ID) {
		err := r.client.RemoveSubGraph(ctx, state.Name.ValueString())

		utils.ProcessError(&resp.Diagnostics, err, "Operational errors when removing sub graph", "Client Error")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	rr, err := r.client.GetSubGraph(ctx, name)

	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID != plan.Name {
		// new sub graph was created, we need to re-assign new id
		plan.ID = plan.Name
	}

	plan.Revision = types.StringValue(rr.Revision)
	plan.CreatedAt = types.StringValue(rr.CreatedAt.Format(time.RFC850))
	plan.UpdatedAt = types.StringValue(rr.UpdatedAt.Format(time.RFC850))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan *SubGraphResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	err := r.client.RemoveSubGraph(ctx, name)

	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when removing sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
