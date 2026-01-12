// Provider configuration and initialization

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var _ provider.Provider = &Bind9Provider{}

// Bind9Provider defines the provider implementation
type Bind9Provider struct {
	version string
}

// Bind9ProviderModel describes the provider data model
type Bind9ProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Bind9Provider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *Bind9Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bind9"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *Bind9Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing BIND9 DNS server via REST API",
		MarkdownDescription: `
The BIND9 provider allows you to manage DNS zones and records on a BIND9 server through its REST API.

## Authentication

The provider supports two authentication methods:

1. **API Key** (recommended): Set the ` + "`api_key`" + ` attribute or ` + "`BIND9_API_KEY`" + ` environment variable
2. **Username/Password**: Set ` + "`username`" + ` and ` + "`password`" + ` attributes or ` + "`BIND9_USERNAME`" + ` and ` + "`BIND9_PASSWORD`" + ` environment variables

## Example Usage

` + "```hcl" + `
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  api_key  = var.bind9_api_key
}

resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
}

resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["192.168.1.100"]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "BIND9 REST API endpoint URL (e.g., https://dns.example.com:8080). Can also be set via BIND9_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for authentication. Can also be set via BIND9_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "Username for JWT authentication. Can also be set via BIND9_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for JWT authentication. Can also be set via BIND9_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Default: false",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "API request timeout in seconds. Default: 30",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a BIND9 API client for data sources and resources
func (p *Bind9Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring BIND9 client")

	var config Bind9ProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check environment variables for defaults
	endpoint := os.Getenv("BIND9_ENDPOINT")
	apiKey := os.Getenv("BIND9_API_KEY")
	username := os.Getenv("BIND9_USERNAME")
	password := os.Getenv("BIND9_PASSWORD")

	// Override with config values if set
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// Validate required configuration
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing BIND9 API Endpoint",
			"The provider cannot create the BIND9 API client as there is a missing or empty value for the BIND9 API endpoint. "+
				"Set the endpoint value in the configuration or use the BIND9_ENDPOINT environment variable.",
		)
	}

	if apiKey == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"The provider requires either an API key or username/password for authentication. "+
				"Set api_key or username/password in the configuration, or use environment variables.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get optional config values
	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	timeout := int64(30)
	if !config.Timeout.IsNull() {
		timeout = config.Timeout.ValueInt64()
	}

	// Create the API client
	client, err := NewClient(endpoint, apiKey, username, password, insecure, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create BIND9 API Client",
			"An unexpected error occurred when creating the BIND9 API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created BIND9 client", map[string]any{"endpoint": endpoint})

	// Make the client available during DataSource and Resource type Configure methods
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources defines the resources implemented in the provider
func (p *Bind9Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewZoneResource,
		NewRecordResource,
		NewDNSSECKeyResource,
		NewACLResource,
	}
}

// DataSources defines the data sources implemented in the provider
func (p *Bind9Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewZoneDataSource,
		NewZonesDataSource,
		NewRecordDataSource,
		NewRecordsDataSource,
	}
}

