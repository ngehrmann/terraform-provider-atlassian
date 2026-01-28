package main

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &AtlassianProvider{}
var _ provider.ProviderWithFunctions = &AtlassianProvider{}

// AtlassianProvider defines the provider implementation.
type AtlassianProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AtlassianProviderModel describes the provider data model.
type AtlassianProviderModel struct {
	ApiToken     types.String `tfsdk:"api_token"`
	Email        types.String `tfsdk:"email"`
	Organization types.String `tfsdk:"organization"`
	BaseUrl      types.String `tfsdk:"base_url"`
}

func (p *AtlassianProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "atlassian"
	resp.Version = p.version
}

func (p *AtlassianProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Atlassian API token for authentication. Can also be set via ATLASSIAN_API_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address associated with the Atlassian account. Can also be set via ATLASSIAN_EMAIL environment variable.",
				Optional:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Atlassian organization/site name. Can also be set via ATLASSIAN_ORGANIZATION environment variable.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for Atlassian API. Defaults to https://api.atlassian.com. Can also be set via ATLASSIAN_BASE_URL environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *AtlassianProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AtlassianProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// Example client configuration for data sources and resources
	if data.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Atlassian API Token",
			"The provider cannot create the Atlassian API client as there is an unknown configuration value for the Atlassian API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ATLASSIAN_API_TOKEN environment variable.",
		)
	}

	if data.Email.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Unknown Atlassian Email",
			"The provider cannot create the Atlassian API client as there is an unknown configuration value for the Atlassian email. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ATLASSIAN_EMAIL environment variable.",
		)
	}

	if data.Organization.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization"),
			"Unknown Atlassian Organization",
			"The provider cannot create the Atlassian API client as there is an unknown configuration value for the Atlassian organization. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ATLASSIAN_ORGANIZATION environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiToken := os.Getenv("ATLASSIAN_API_TOKEN")
	email := os.Getenv("ATLASSIAN_EMAIL")
	organization := os.Getenv("ATLASSIAN_ORGANIZATION")
	baseUrl := os.Getenv("ATLASSIAN_BASE_URL")

	if !data.ApiToken.IsNull() {
		apiToken = data.ApiToken.ValueString()
	}

	if !data.Email.IsNull() {
		email = data.Email.ValueString()
	}

	if !data.Organization.IsNull() {
		organization = data.Organization.ValueString()
	}

	if !data.BaseUrl.IsNull() {
		baseUrl = data.BaseUrl.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Atlassian API Token",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian API token. "+
				"Set the api_token value in the configuration or use the ATLASSIAN_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if email == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Missing Atlassian Email",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian email. "+
				"Set the email value in the configuration or use the ATLASSIAN_EMAIL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if organization == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization"),
			"Missing Atlassian Organization",
			"The provider cannot create the Atlassian API client as there is a missing or empty value for the Atlassian organization. "+
				"Set the organization value in the configuration or use the ATLASSIAN_ORGANIZATION environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if baseUrl == "" {
		baseUrl = "https://api.atlassian.com"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Atlassian client using the configuration values
	client, err := NewAtlassianClient(apiToken, email, organization, baseUrl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Atlassian API Client",
			"An unexpected error occurred when creating the Atlassian API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Atlassian Client Error: "+err.Error(),
		)
		return
	}

	// Make the Atlassian client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Atlassian client", map[string]any{"success": true})
}

func (p *AtlassianProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTeamResource,
	}
}

func (p *AtlassianProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add data sources here if needed
	}
}

func (p *AtlassianProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// Add functions here if needed
	}
}

func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AtlassianProvider{
			version: version,
		}
	}
}
