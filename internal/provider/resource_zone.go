// Zone Resource

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &ZoneResource{}
	_ resource.ResourceWithImportState = &ZoneResource{}
)

// NewZoneResource creates a new zone resource
func NewZoneResource() resource.Resource {
	return &ZoneResource{}
}

// ZoneResource defines the resource implementation
type ZoneResource struct {
	client *Client
}

// ZoneResourceModel describes the resource data model
type ZoneResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	File          types.String `tfsdk:"file"`
	SOAMname      types.String `tfsdk:"soa_mname"`
	SOARname      types.String `tfsdk:"soa_rname"`
	SOARefresh    types.Int64  `tfsdk:"soa_refresh"`
	SOARetry      types.Int64  `tfsdk:"soa_retry"`
	SOAExpire     types.Int64  `tfsdk:"soa_expire"`
	SOAMinimum    types.Int64  `tfsdk:"soa_minimum"`
	DefaultTTL    types.Int64  `tfsdk:"default_ttl"`
	Nameservers   types.List   `tfsdk:"nameservers"`
	NSAddresses   types.Map    `tfsdk:"ns_addresses"`
	AllowTransfer types.List   `tfsdk:"allow_transfer"`
	AllowUpdate   types.List   `tfsdk:"allow_update"`
	AllowQuery    types.List   `tfsdk:"allow_query"`
	Notify        types.Bool   `tfsdk:"notify"`
	DeleteFile    types.Bool   `tfsdk:"delete_file_on_destroy"`
	Serial        types.Int64  `tfsdk:"serial"`
	Loaded        types.Bool   `tfsdk:"loaded"`
	DNSSECEnabled types.Bool   `tfsdk:"dnssec_enabled"`
}

// Metadata returns the resource type name
func (r *ZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

// Schema defines the schema for the resource
func (r *ZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS zone on BIND9 server.",
		MarkdownDescription: `
Manages a DNS zone on a BIND9 server.

## Example Usage

### Master Zone

` + "```hcl" + `
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
  
  soa_mname  = "ns1.example.com"
  soa_rname  = "hostmaster.example.com"
  default_ttl = 3600
  
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
  
  allow_transfer = ["192.168.1.0/24"]
  notify = true
}
` + "```" + `

### Slave Zone

` + "```hcl" + `
resource "bind9_zone" "slave" {
  name = "example.com"
  type = "slave"
  
  # Masters would be configured in BIND9 directly
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Zone identifier (same as name)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Zone name (e.g., example.com)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Zone type: master, slave, forward, stub",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file": schema.StringAttribute{
				Description: "Zone file path (auto-generated if not specified)",
				Optional:    true,
				Computed:    true,
			},
			"soa_mname": schema.StringAttribute{
				Description: "Primary nameserver for SOA record",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ns1"),
			},
			"soa_rname": schema.StringAttribute{
				Description: "Responsible person email for SOA (use . instead of @)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("hostmaster"),
			},
			"soa_refresh": schema.Int64Attribute{
				Description: "SOA refresh interval in seconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			"soa_retry": schema.Int64Attribute{
				Description: "SOA retry interval in seconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(7200),
			},
			"soa_expire": schema.Int64Attribute{
				Description: "SOA expire time in seconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600000),
			},
			"soa_minimum": schema.Int64Attribute{
				Description: "SOA minimum/negative TTL in seconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"default_ttl": schema.Int64Attribute{
				Description: "Default TTL for records",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"nameservers": schema.ListAttribute{
				Description: "List of authoritative nameservers",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"allow_transfer": schema.ListAttribute{
				Description: "ACL for zone transfers",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"allow_update": schema.ListAttribute{
				Description: "ACL for dynamic updates",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"allow_query": schema.ListAttribute{
				Description: "ACL for queries",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"notify": schema.BoolAttribute{
				Description: "Send NOTIFY to slaves on zone changes",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"delete_file_on_destroy": schema.BoolAttribute{
				Description: "Delete zone file when zone is destroyed",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ns_addresses": schema.MapAttribute{
				Description: "Map of nameserver names to IP addresses for glue records (e.g., {\"ns1\" = \"192.168.1.1\"})",
				Optional:    true,
				ElementType: types.StringType,
			},
			"serial": schema.Int64Attribute{
				Description: "Current zone serial number",
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
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state
func (r *ZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating zone", map[string]any{"name": plan.Name.ValueString()})

	// Build create request
	createReq := &ZoneCreateRequest{
		Name:       plan.Name.ValueString(),
		Type:       plan.Type.ValueString(),
		SOAMname:   plan.SOAMname.ValueString(),
		SOARname:   plan.SOARname.ValueString(),
		SOARefresh: int(plan.SOARefresh.ValueInt64()),
		SOARetry:   int(plan.SOARetry.ValueInt64()),
		SOAExpire:  int(plan.SOAExpire.ValueInt64()),
		SOAMinimum: int(plan.SOAMinimum.ValueInt64()),
		DefaultTTL: int(plan.DefaultTTL.ValueInt64()),
	}

	// Convert ns_addresses map
	if !plan.NSAddresses.IsNull() {
		nsAddresses := make(map[string]string)
		diags = plan.NSAddresses.ElementsAs(ctx, &nsAddresses, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.NSAddresses = nsAddresses
	}

	if !plan.File.IsNull() {
		createReq.File = plan.File.ValueString()
	}

	// Convert nameservers
	if !plan.Nameservers.IsNull() {
		var nameservers []string
		diags = plan.Nameservers.ElementsAs(ctx, &nameservers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Nameservers = nameservers
	}

	// Build options (allow_update, allow_transfer, allow_query)
	options := &ZoneOptions{}
	hasOptions := false

	if !plan.AllowUpdate.IsNull() {
		var allowUpdate []string
		diags = plan.AllowUpdate.ElementsAs(ctx, &allowUpdate, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		options.AllowUpdate = allowUpdate
		hasOptions = true
	}

	if !plan.AllowTransfer.IsNull() {
		var allowTransfer []string
		diags = plan.AllowTransfer.ElementsAs(ctx, &allowTransfer, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		options.AllowTransfer = allowTransfer
		hasOptions = true
	}

	if !plan.AllowQuery.IsNull() {
		var allowQuery []string
		diags = plan.AllowQuery.ElementsAs(ctx, &allowQuery, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		options.AllowQuery = allowQuery
		hasOptions = true
	}

	if hasOptions {
		createReq.Options = options
	}

	// Create zone
	zone, err := r.client.CreateZone(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Zone",
			"Could not create zone: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(zone.Name)
	plan.Serial = types.Int64Value(zone.Serial)
	plan.Loaded = types.BoolValue(zone.Loaded)
	plan.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	if zone.File != "" {
		plan.File = types.StringValue(zone.File)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data
func (r *ZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading zone", map[string]any{"name": state.Name.ValueString()})

	zone, err := r.client.GetZone(ctx, state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Zone",
			"Could not read zone: "+err.Error(),
		)
		return
	}

	// Update state with API response values
	state.Serial = types.Int64Value(zone.Serial)
	state.Loaded = types.BoolValue(zone.Loaded)
	state.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	if zone.File != "" {
		state.File = types.StringValue(zone.File)
	}
	if zone.Type != "" {
		// Normalize zone type (BIND9 uses "primary"/"secondary" in newer versions,
		// but "master"/"slave" are still commonly used synonyms)
		zoneType := zone.Type
		switch zoneType {
		case "primary":
			zoneType = "master"
		case "secondary":
			zoneType = "slave"
		}
		state.Type = types.StringValue(zoneType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state
func (r *ZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating zone", map[string]any{"name": plan.Name.ValueString()})

	// Reload zone to apply changes
	if err := r.client.ReloadZone(ctx, plan.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zone",
			"Could not reload zone: "+err.Error(),
		)
		return
	}

	// Read back the zone
	zone, err := r.client.GetZone(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Zone",
			"Could not read zone after update: "+err.Error(),
		)
		return
	}

	// Update all computed fields
	plan.Serial = types.Int64Value(zone.Serial)
	plan.Loaded = types.BoolValue(zone.Loaded)
	plan.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	if zone.File != "" {
		plan.File = types.StringValue(zone.File)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *ZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting zone", map[string]any{"name": state.Name.ValueString()})

	deleteFile := false
	if !state.DeleteFile.IsNull() {
		deleteFile = state.DeleteFile.ValueBool()
	}

	if err := r.client.DeleteZone(ctx, state.Name.ValueString(), deleteFile); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Zone",
			"Could not delete zone: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform
func (r *ZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
