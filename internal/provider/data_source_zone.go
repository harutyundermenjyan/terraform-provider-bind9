// Zone Data Source

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var _ datasource.DataSource = &ZoneDataSource{}

// NewZoneDataSource creates a new zone data source
func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

// ZoneDataSource defines the data source implementation
type ZoneDataSource struct {
	client *Client
}

// ZoneDataSourceModel describes the data source data model
type ZoneDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	File          types.String `tfsdk:"file"`
	Serial        types.Int64  `tfsdk:"serial"`
	Loaded        types.Bool   `tfsdk:"loaded"`
	DNSSECEnabled types.Bool   `tfsdk:"dnssec_enabled"`
	RecordCount   types.Int64  `tfsdk:"record_count"`
}

// Metadata returns the data source type name
func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

// Schema defines the schema for the data source
func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a DNS zone.",
		MarkdownDescription: `
Retrieves information about a DNS zone on the BIND9 server.

## Example Usage

` + "```hcl" + `
data "bind9_zone" "example" {
  name = "example.com"
}

output "zone_serial" {
  value = data.bind9_zone.example.serial
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Zone identifier",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Zone name",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Zone type (master, slave, forward, stub)",
				Computed:    true,
			},
			"file": schema.StringAttribute{
				Description: "Zone file path",
				Computed:    true,
			},
			"serial": schema.Int64Attribute{
				Description: "Zone serial number",
				Computed:    true,
			},
			"loaded": schema.BoolAttribute{
				Description: "Whether zone is loaded",
				Computed:    true,
			},
			"dnssec_enabled": schema.BoolAttribute{
				Description: "Whether DNSSEC is enabled",
				Computed:    true,
			},
			"record_count": schema.Int64Attribute{
				Description: "Number of records in zone",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ZoneDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading zone data", map[string]any{"name": config.Name.ValueString()})

	zone, err := d.client.GetZone(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zone",
			"Could not read zone: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(zone.Name)
	config.Type = types.StringValue(zone.Type)
	config.Serial = types.Int64Value(zone.Serial)
	config.Loaded = types.BoolValue(zone.Loaded)
	config.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	config.RecordCount = types.Int64Value(int64(zone.RecordCount))

	if zone.File != "" {
		config.File = types.StringValue(zone.File)
	} else {
		config.File = types.StringNull()
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// ============================================================================
// Zones (plural) Data Source
// ============================================================================

var _ datasource.DataSource = &ZonesDataSource{}

// NewZonesDataSource creates a new zones data source
func NewZonesDataSource() datasource.DataSource {
	return &ZonesDataSource{}
}

// ZonesDataSource defines the data source implementation
type ZonesDataSource struct {
	client *Client
}

// ZonesDataSourceModel describes the data source data model
type ZonesDataSourceModel struct {
	ID    types.String          `tfsdk:"id"`
	Type  types.String          `tfsdk:"type"`
	Zones []ZoneDataSourceModel `tfsdk:"zones"`
}

// Metadata returns the data source type name
func (d *ZonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zones"
}

// Schema defines the schema for the data source
func (d *ZonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves list of all DNS zones.",
		MarkdownDescription: `
Retrieves a list of all DNS zones on the BIND9 server.

## Example Usage

` + "```hcl" + `
data "bind9_zones" "all" {}

output "zone_count" {
  value = length(data.bind9_zones.all.zones)
}

# Filter to master zones only
data "bind9_zones" "masters" {
  type = "master"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Data source identifier",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Filter by zone type",
				Optional:    true,
			},
			"zones": schema.ListNestedAttribute{
				Description: "List of zones",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"file": schema.StringAttribute{
							Computed: true,
						},
						"serial": schema.Int64Attribute{
							Computed: true,
						},
						"loaded": schema.BoolAttribute{
							Computed: true,
						},
						"dnssec_enabled": schema.BoolAttribute{
							Computed: true,
						},
						"record_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *ZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *ZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ZonesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading zones data")

	// Build filter params
	params := map[string]string{}
	filterType := ""
	if !config.Type.IsNull() {
		filterType = config.Type.ValueString()
		params["type"] = filterType
	}

	zones, err := d.client.ListZones(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zones",
			"Could not read zones: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue("zones")
	config.Zones = []ZoneDataSourceModel{}

	for _, zone := range zones {
		if filterType != "" && zone.Type != filterType {
			continue
		}

		zoneModel := ZoneDataSourceModel{
			ID:            types.StringValue(zone.Name),
			Name:          types.StringValue(zone.Name),
			Type:          types.StringValue(zone.Type),
			Serial:        types.Int64Value(zone.Serial),
			Loaded:        types.BoolValue(zone.Loaded),
			DNSSECEnabled: types.BoolValue(zone.DNSSECEnabled),
			RecordCount:   types.Int64Value(int64(zone.RecordCount)),
		}

		if zone.File != "" {
			zoneModel.File = types.StringValue(zone.File)
		} else {
			zoneModel.File = types.StringNull()
		}

		config.Zones = append(config.Zones, zoneModel)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
