package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/radureau/terraform-provider-computer-database/internal/cdb"
)

type cdbProvider struct {
	apiURL    string
	apiClient *cdb.APIClient
}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &cdbProvider{}
	}
}

func (p *cdbProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.Diagnostics.Append(
		req.Config.GetAttribute(ctx, path.Root("api_url"), &p.apiURL)...,
	)
	p.apiClient = cdb.NewAPIClient(p.apiURL)
	resp.ResourceData = p.apiClient
	resp.DataSourceData = p.apiClient
}

func (p *cdbProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "computer-database"
}

func (p *cdbProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataSource,
	}
}

func (p *cdbProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCompanyResource,
	}
}

func (p *cdbProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Required:    true,
				Description: "The url to the computer database api. Ex: http://localhost:8080/api/v1",
			},
		},
	}
}
