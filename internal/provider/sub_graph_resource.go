package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-apollostudio/internal/utils"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
	"time"
)

var (
	_ resource.Resource                = &SubGraphResource{}
	_ resource.ResourceWithConfigure   = &SubGraphResource{}
	_ resource.ResourceWithImportState = &SubGraphResource{}
)

const retryTimeout = 5 * time.Second

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
	ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sub_graph"
}

func (r *SubGraphResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Fields required to submit sub graph",

		Attributes: map[string]tfsdk.Attribute{
			"url": {
				MarkdownDescription: "The URL of the sub graph endpoint",
				Type:                types.StringType,
				Optional:            true,
			},
			"schema": {
				MarkdownDescription: "Sub Graph SDL schema",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "Sub Graph name",
				Type:                types.StringType,
				Required:            true,
			},
			"id": {
				MarkdownDescription: "Resource identifier for terraform",
				Type:                types.StringType,
				Computed:            true,
			},
			"revision": {
				MarkdownDescription: "Sub Graph revision",
				Type:                types.NumberType,
				Computed:            true,
			},
			"created_at": {
				MarkdownDescription: "Schema creation date",
				Type:                types.StringType,
				Computed:            true,
			},
			"updated_at": {
				MarkdownDescription: "Schema update date",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
}

func (r *SubGraphResource) Configure(
	ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
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

	schema := plan.Schema.ValueString()
	name := plan.Name.ValueString()
	url := plan.URL.ValueString()

	graph, err := r.client.ReadSubGraph(ctx, name)
	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.client.GetLatestSchemaBuild(ctx)
	utils.ProcessError(&resp.Diagnostics, err, "Federation schema contains errors", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if graph.Name != "" {
		resp.Diagnostics.AddError(
			"Sub Graph already exists",
			fmt.Sprintf(
				"Sub Graph `%s` already exits, if you want to submit pre-existing graph, "+
					"please import the resource",
				name,
			),
		)
		return
	}

	result, err := r.client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
			SubGraphSchema: []byte(schema),
			SubGraphName:   name,
			SubGraphURL:    url,
		},
	)

	utils.ProcessError(&resp.Diagnostics, err, "Federation schema error while submitting sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if !result.WasCreated && graph.Name != "" {
		resp.Diagnostics.AddWarning(
			"No new subgraph was created",
			"New sub graph was not created, submitted only",
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
	result, err := r.client.ReadSubGraph(ctx, name)

	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		state.ID = types.StringValue(name)
	}
	if state.URL.IsNull() {
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

	schema := plan.Schema.ValueString()
	name := plan.Name.ValueString()
	url := plan.URL.ValueString()

	if !plan.Name.Equal(state.ID) {
		err := r.client.RemoveSubGraph(ctx, state.Name.ValueString())

		utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	err := sdkresource.RetryContext(
		ctx, retryTimeout, func() *sdkresource.RetryError {
			var err error
			_, err = r.client.SubmitSubGraph(
				ctx, &apollostudio.SubmitOptions{
					SubGraphSchema: []byte(schema),
					SubGraphName:   name,
					SubGraphURL:    url,
				},
			)
			return utils.NewRetryableError(err)
		},
	)

	utils.ProcessError(&resp.Diagnostics, err, "Federation schema error while submitting sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	rr, err := r.client.ReadSubGraph(ctx, name)

	utils.ProcessError(&resp.Diagnostics, err, "Operational errors when reading sub graph", "Client Error")
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID != plan.Name {
		// new sub graph was created, we need to re-assign new id
		plan.ID = plan.Name
	}

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
