package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"computer_models": schema.SetAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":    types.StringType,
						"release": types.StringType,
						"id":      types.StringType,
					},
				},
				Optional: true,
			},
		},
	}
}

type companyResourceData struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Location       types.String `tfsdk:"location"`
	ComputerModels types.Set    `tfsdk:"computer_models"`
}
type computerModelResourceData struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Release types.String `tfsdk:"release"`
}

func (tfCompanyComputerModel computerModelResourceData) ComputerModel() *cdb.ComputerModel {
	return &cdb.ComputerModel{
		ID:      tfCompanyComputerModel.ID.ValueString(),
		Name:    tfCompanyComputerModel.Name.ValueString(),
		Release: tfCompanyComputerModel.Release.ValueString(),
	}
}
func (tfCompany *companyResourceData) setDefaults() {
	if tfCompany.Location.IsNull() || tfCompany.Location.IsUnknown() {
		tfCompany.Location = types.StringValue("global")
	}
}
func (tfCompany *companyResourceData) Company(ctx context.Context) (company *cdb.Company, diags diag.Diagnostics) {
	tfCompany.setDefaults()

	tfCompanyComputerModels := []computerModelResourceData{}
	diags = tfCompany.ComputerModels.ElementsAs(ctx, &tfCompanyComputerModels, false)
	if diags.HasError() {
		return
	}
	computerModels := make([]cdb.ComputerModel, len(tfCompanyComputerModels))
	for idx, model := range tfCompanyComputerModels {
		computerModels[idx] = *model.ComputerModel()
	}
	company = &cdb.Company{
		ID:             tfCompany.ID.ValueString(),
		Name:           tfCompany.Name.ValueString(),
		Location:       tfCompany.Location.ValueString(),
		ComputerModels: &computerModels,
	}
	return
}

func (r *companyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tfCompany companyResourceData

	diags := req.Config.Get(ctx, &tfCompany)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	company, diags := tfCompany.Company(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateCompany(
		company,
	)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't create company.", fmt.Sprint(err))
		return
	}

	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &tfCompany)
	resp.Diagnostics.Append(diags...)
}

func (r *companyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateTFCompanyID string
	diags := req.State.GetAttribute(ctx, path.Root("id"), &stateTFCompanyID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read resource using 3rd party API.
	company, err := r.client.GetCompany(stateTFCompanyID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("couldn't update company", err.Error())
		return
	}

	tfCompany := companyResourceData{
		ID:       types.StringValue(company.ID),
		Name:     types.StringValue(company.Name),
		Location: types.StringValue(company.Location),
	}

	stateTFCompanyComputerModels := []computerModelResourceData{}
	diags = req.State.GetAttribute(ctx, path.Root("computer_models"), &stateTFCompanyComputerModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	otype := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"name":    types.StringType,
			"release": types.StringType,
		},
	}
	computerModelsAsObjects := []types.Object{}
	for _, model := range stateTFCompanyComputerModels {
		computerModel, err := r.client.GetComputerModel(company.ID, model.ID.ValueString())
		if err != nil {
			continue
		}
		obj, diags := types.ObjectValueFrom(ctx, otype.AttrTypes, computerModelResourceData{
			ID:      types.StringValue(computerModel.ID),
			Name:    types.StringValue(computerModel.Name),
			Release: types.StringValue(computerModel.Release),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		computerModelsAsObjects = append(computerModelsAsObjects, obj)
	}
	tfCompany.ComputerModels, diags = types.SetValueFrom(ctx, otype, computerModelsAsObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &tfCompany)
	resp.Diagnostics.Append(diags...)
}

func (r *companyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var tfCompany companyResourceData

	diags := req.Plan.Get(ctx, &tfCompany)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	company, diags := tfCompany.Company(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateCompany(
		company,
	)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't update company.", fmt.Sprint(err))
		return
	}

	diags = resp.State.Set(ctx, &tfCompany)
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
