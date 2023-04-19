package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
	"time"
)

var _ resource.Resource = &SubGraphResource{}

func NewSubGraphResource() resource.Resource {
	return &SubGraphResource{}
}

// SubGraphResource defines the resource implementation.
type SubGraphResource struct {
	client *apollostudio.Client
}

// SubGraphResourceModel describes the resource data model.
type SubGraphResourceModel struct {
	SchemaID       types.String `tfsdk:"schema_id"`
	SchemaVariant  types.String `tfsdk:"schema_variant"`
	SubGraphURL    types.String `tfsdk:"sub_graph_url"`
	SubGraphSchema types.String `tfsdk:"sub_graph_schema"`
	SubGraphName   types.String `tfsdk:"sub_graph_name"`
	SchemaHash     types.String `tfsdk:"schema_hash"`
	SDL            types.String `tfsdk:"sdl"`
	Revision       types.String `tfsdk:"revision"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
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
			"sub_graph_url": {
				MarkdownDescription: "The URL of the sub graph API",
				Type:                types.StringType,
				Optional:            true,
			},
			"schema_id": {
				MarkdownDescription: "Schema ID to validate against",
				Required:            true,
				Type:                types.StringType,
			},
			"schema_variant": {
				MarkdownDescription: "Schema variant to validate against",
				Type:                types.StringType,
				Required:            true,
			},
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
			"schema_hash": {
				MarkdownDescription: "Schema hash",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"sdl": {
				MarkdownDescription: "SDL schema",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"revision": {
				MarkdownDescription: "Schema revision",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
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

	r.client = client
}

func (r *SubGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SubGraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	schemaId := data.SchemaID.ValueString()
	schemaVariant := data.SchemaVariant.ValueString()
	subGraphSchema := data.SubGraphSchema.ValueString()
	subGraphName := data.SubGraphName.ValueString()
	subGraphUrl := data.SubGraphURL.ValueString()

	tflog.Debug(
		ctx, "submitting sub graph schema", map[string]interface{}{
			"schema_id":        schemaId,
			"schema_variant":   schemaVariant,
			"sub_graph_schema": subGraphSchema,
			"sub_graph_name":   subGraphName,
		},
	)

	if subGraphName == "" {
		resp.Diagnostics.AddWarning(
			"Sub Graph name is empty",
			"Sub Graph name is empty",
		)
	}

	result, err := r.client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
			SchemaID:       schemaId,
			SchemaVariant:  schemaVariant,
			SubGraphSchema: []byte(subGraphSchema),
			SubGraphName:   subGraphName,
			SubGraphURL:    subGraphUrl,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable submit schema, got error: %s", err))
		return
	}

	if result.Errors != nil {
		for _, e := range result.Errors {
			resp.Diagnostics.AddError(e.Code, e.Message)
		}
	}

	data.SchemaHash = types.StringValue(result.CompositionConfig.SchemaHash)
	tflog.Trace(ctx, "submit sub graph applied")

	if result.WasCreated {
		tflog.Trace(ctx, "new sub graph created")
	}

	rr, err := r.client.ReadSubGraph(
		ctx, &apollostudio.ReadOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sub graph, got error: %s", err))
		return
	}

	data.Revision = types.StringValue(rr.Revision)
	data.SDL = types.StringValue(rr.ActivePartialSchema.Sdl)
	data.CreatedAt = types.StringValue(rr.CreatedAt.Format(time.RFC850))
	data.UpdatedAt = types.StringValue(rr.UpdatedAt.Format(time.RFC850))

	tflog.Trace(ctx, "updated sub graph schema")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SubGraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading sub graph info...")

	schemaId := data.SchemaID.ValueString()
	schemaVariant := data.SchemaVariant.ValueString()
	subGraphName := data.SubGraphName.ValueString()

	tflog.Debug(
		ctx, "reading sub graph schema", map[string]interface{}{
			"schema_id":      schemaId,
			"schema_variant": schemaVariant,
			"sub_graph_name": subGraphName,
		},
	)

	if subGraphName == "" {
		resp.Diagnostics.AddWarning(
			"Sub Graph name is empty",
			"Sub Graph name is empty",
		)
	}

	result, err := r.client.ReadSubGraph(
		ctx, &apollostudio.ReadOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sub graph, got error: %s", err))
		return
	}

	data.Revision = types.StringValue(result.Revision)
	data.SDL = types.StringValue(result.ActivePartialSchema.Sdl)
	data.CreatedAt = types.StringValue(result.CreatedAt.Format(time.RFC850))
	data.UpdatedAt = types.StringValue(result.UpdatedAt.Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SubGraphResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(
		ctx, "Applying sub graph", map[string]interface{}{
			"SDL": data.SDL,
		},
	)

	schemaId := data.SchemaID.ValueString()
	schemaVariant := data.SchemaVariant.ValueString()
	subGraphSchema := data.SubGraphSchema.ValueString()
	subGraphName := data.SubGraphName.ValueString()

	tflog.Debug(
		ctx, "submitting sub graph schema", map[string]interface{}{
			"schema_id":        schemaId,
			"schema_variant":   schemaVariant,
			"sub_graph_schema": subGraphSchema,
			"sub_graph_name":   subGraphName,
		},
	)

	if subGraphName == "" {
		resp.Diagnostics.AddWarning(
			"Sub Graph name is empty",
			"Sub Graph name is empty",
		)
	}

	result, err := r.client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
			SchemaID:       schemaId,
			SchemaVariant:  schemaVariant,
			SubGraphSchema: []byte(subGraphSchema),
			SubGraphName:   subGraphName,
		},
	)

	tflog.Info(
		ctx, "After submitting sub graph", map[string]interface{}{
			"SDL": data.SDL,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable submit schema, got error: %s", err))
		return
	}

	if result.Errors != nil {
		for _, e := range result.Errors {
			resp.Diagnostics.AddError(e.Code, e.Message)
		}
	}

	data.SchemaHash = types.StringValue(result.CompositionConfig.SchemaHash)
	tflog.Trace(ctx, "submit sub graph applied")

	rr, err := r.client.ReadSubGraph(
		ctx, &apollostudio.ReadOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sub graph, got error: %s", err))
		return
	}

	data.Revision = types.StringValue(rr.Revision)
	data.SDL = types.StringValue(rr.ActivePartialSchema.Sdl)
	data.CreatedAt = types.StringValue(rr.CreatedAt.Format(time.RFC850))
	data.UpdatedAt = types.StringValue(rr.UpdatedAt.Format(time.RFC850))

	tflog.Info(
		ctx, "Sub graph read after update", map[string]interface{}{
			"SDL": data.SDL,
		},
	)

	tflog.Trace(ctx, "updated sub graph schema")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SubGraphResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	schemaId := data.SchemaID.ValueString()
	schemaVariant := data.SchemaVariant.ValueString()
	subGraphName := data.SubGraphName.ValueString()

	tflog.Debug(
		ctx, "deleting sub graph", map[string]interface{}{
			"schema_id":      schemaId,
			"schema_variant": schemaVariant,
			"sub_graph_name": subGraphName,
		},
	)

	if subGraphName == "" {
		resp.Diagnostics.AddWarning(
			"Sub Graph name is empty",
			"Sub Graph name is empty",
		)
	}

	err := r.client.RemoveSubGraph(
		ctx, &apollostudio.RemoveOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove sub graph, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "removed sub graph")
}
