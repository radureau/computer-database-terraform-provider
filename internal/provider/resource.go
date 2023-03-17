package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/radureau/terraform-provider-computer-database/internal/cdb"
)

type companyResource struct {
	client *cdb.APIClient
}

func NewCompanyResource() resource.Resource {
	return &companyResource{}
}

func (r *companyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cdb.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cdb.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *companyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_company"
}

func (r *companyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"location": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "default to global",
			},
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// "computerModel": schema.ListNestedBlock{},
		},
	}
}

type companyResourceData struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
}

func (r *companyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data companyResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	if data.Location.IsNull() {
		data.Location = types.StringValue("global")
	}

	// Create resource using 3rd party API.
	err := r.client.CreateCompany(
		&cdb.Company{
			ID:       data.ID.ValueString(),
			Name:     data.Name.ValueString(),
			Location: data.Location.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create company.", fmt.Sprint(err))
		return
	}

	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *companyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data companyResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read resource using 3rd party API.
	company, err := r.client.GetCompany(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't read company.", fmt.Sprint(err))
		return
	}

	data = companyResourceData{
		ID:       types.StringValue(company.ID),
		Name:     types.StringValue(company.Name),
		Location: types.StringValue(company.Location),
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *companyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data companyResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update resource using 3rd party API.

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *companyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data companyResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource using 3rd party API.
	err := r.client.DeleteCompany(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Couldn't delete company.", fmt.Sprint(err))
	}
}
