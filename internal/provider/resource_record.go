// Record Resource

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &RecordResource{}
	_ resource.ResourceWithImportState = &RecordResource{}
)

// NewRecordResource creates a new record resource
func NewRecordResource() resource.Resource {
	return &RecordResource{}
}

// RecordResource defines the resource implementation
type RecordResource struct {
	client *Client
}

// RecordResourceModel describes the resource data model
type RecordResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Zone    types.String `tfsdk:"zone"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	TTL     types.Int64  `tfsdk:"ttl"`
	Class   types.String `tfsdk:"class"`
	Records types.List   `tfsdk:"records"`
	
	// Type-specific fields (for convenience)
	Address    types.String `tfsdk:"address"`     // A, AAAA
	Target     types.String `tfsdk:"target"`      // CNAME, DNAME, NS, PTR
	Priority   types.Int64  `tfsdk:"priority"`    // MX, SRV
	Weight     types.Int64  `tfsdk:"weight"`      // SRV
	Port       types.Int64  `tfsdk:"port"`        // SRV
	Text       types.String `tfsdk:"text"`        // TXT
	Flags      types.Int64  `tfsdk:"flags"`       // CAA
	Tag        types.String `tfsdk:"tag"`         // CAA
	Value      types.String `tfsdk:"value"`       // CAA
}

// Metadata returns the resource type name
func (r *RecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

// Schema defines the schema for the resource
func (r *RecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS record on BIND9 server.",
		MarkdownDescription: `
Manages DNS records on a BIND9 server. Supports all common record types.

## Example Usage

### A Record

` + "```hcl" + `
resource "bind9_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["192.168.1.100"]
}
` + "```" + `

### AAAA Record

` + "```hcl" + `
resource "bind9_record" "www_ipv6" {
  zone    = "example.com"
  name    = "www"
  type    = "AAAA"
  ttl     = 300
  records = ["2001:db8::1"]
}
` + "```" + `

### CNAME Record

` + "```hcl" + `
resource "bind9_record" "alias" {
  zone    = "example.com"
  name    = "blog"
  type    = "CNAME"
  ttl     = 3600
  records = ["www.example.com."]
}
` + "```" + `

### MX Record

` + "```hcl" + `
resource "bind9_record" "mx" {
  zone    = "example.com"
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.example.com.", "20 mail2.example.com."]
}
` + "```" + `

### TXT Record

` + "```hcl" + `
resource "bind9_record" "spf" {
  zone    = "example.com"
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 include:_spf.google.com ~all"]
}
` + "```" + `

### SRV Record

` + "```hcl" + `
resource "bind9_record" "sip" {
  zone    = "example.com"
  name    = "_sip._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["10 60 5060 sip.example.com."]
}
` + "```" + `

### CAA Record

` + "```hcl" + `
resource "bind9_record" "caa" {
  zone    = "example.com"
  name    = "@"
  type    = "CAA"
  ttl     = 3600
  records = ["0 issue \"letsencrypt.org\""]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Record identifier (zone/name/type)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Zone name (e.g., example.com)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Record name (e.g., www, @, _sip._tcp)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Record type (A, AAAA, CNAME, MX, TXT, NS, PTR, SRV, CAA, etc.)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"A", "AAAA", "CNAME", "MX", "TXT", "NS", "PTR", "SOA",
						"SRV", "CAA", "NAPTR", "HTTPS", "SVCB", "TLSA", "SSHFP",
						"DNSKEY", "DS", "LOC", "HINFO", "RP", "DNAME", "URI",
					),
				},
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"class": schema.StringAttribute{
				Description: "Record class (IN, CH, HS)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("IN"),
			},
			"records": schema.ListAttribute{
				Description: "Record data values",
				Required:    true,
				ElementType: types.StringType,
			},
			// Convenience attributes for common record types
			"address": schema.StringAttribute{
				Description: "IP address for A/AAAA records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"target": schema.StringAttribute{
				Description: "Target for CNAME/NS/PTR/DNAME records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "Priority for MX/SRV records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Weight for SRV records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port for SRV records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"text": schema.StringAttribute{
				Description: "Text content for TXT records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"flags": schema.Int64Attribute{
				Description: "Flags for CAA records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag for CAA records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value for CAA records (convenience attribute)",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource
func (r *RecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating record", map[string]any{
		"zone": plan.Zone.ValueString(),
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	// Get records from list
	var records []string
	diags = plan.Records.ElementsAs(ctx, &records, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create each record
	for _, rdata := range records {
		createReq := &RecordCreateRequest{
			RecordType:  plan.Type.ValueString(),
			Name:        plan.Name.ValueString(),
			TTL:         int(plan.TTL.ValueInt64()),
			RecordClass: plan.Class.ValueString(),
			Data:        r.buildRecordData(plan.Type.ValueString(), rdata),
		}

		_, err := r.client.CreateRecord(ctx, plan.Zone.ValueString(), createReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Record",
				fmt.Sprintf("Could not create record %s %s: %s", plan.Name.ValueString(), plan.Type.ValueString(), err.Error()),
			)
			return
		}
	}

	// Set ID
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.Zone.ValueString(), plan.Name.ValueString(), plan.Type.ValueString()))

	// Set computed convenience attributes based on record type and data
	r.setComputedAttributes(&plan, records)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// setComputedAttributes sets the computed convenience attributes based on record type
func (r *RecordResource) setComputedAttributes(model *RecordResourceModel, records []string) {
	// Set all computed attributes to empty/zero values (not null, which stays "unknown")
	model.Address = types.StringValue("")
	model.Target = types.StringValue("")
	model.Priority = types.Int64Value(0)
	model.Weight = types.Int64Value(0)
	model.Port = types.Int64Value(0)
	model.Text = types.StringValue("")
	model.Flags = types.Int64Value(0)
	model.Tag = types.StringValue("")
	model.Value = types.StringValue("")

	if len(records) == 0 {
		return
	}

	rdata := records[0]
	recordType := model.Type.ValueString()

	switch recordType {
	case "A", "AAAA":
		model.Address = types.StringValue(rdata)
	case "CNAME", "DNAME", "NS", "PTR":
		model.Target = types.StringValue(rdata)
	case "TXT":
		model.Text = types.StringValue(strings.Trim(rdata, "\""))
	case "MX":
		parts := strings.SplitN(rdata, " ", 2)
		if len(parts) == 2 {
			if priority, err := parseInt64(parts[0]); err == nil {
				model.Priority = types.Int64Value(priority)
			}
			model.Target = types.StringValue(parts[1])
		}
	case "SRV":
		parts := strings.SplitN(rdata, " ", 4)
		if len(parts) == 4 {
			if priority, err := parseInt64(parts[0]); err == nil {
				model.Priority = types.Int64Value(priority)
			}
			if weight, err := parseInt64(parts[1]); err == nil {
				model.Weight = types.Int64Value(weight)
			}
			if port, err := parseInt64(parts[2]); err == nil {
				model.Port = types.Int64Value(port)
			}
			model.Target = types.StringValue(parts[3])
		}
	case "CAA":
		parts := strings.SplitN(rdata, " ", 3)
		if len(parts) == 3 {
			if flags, err := parseInt64(parts[0]); err == nil {
				model.Flags = types.Int64Value(flags)
			}
			model.Tag = types.StringValue(parts[1])
			model.Value = types.StringValue(strings.Trim(parts[2], "\""))
		}
	}
}

// parseInt64 helper
func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// buildRecordData constructs the data map for creating a record
func (r *RecordResource) buildRecordData(recordType, rdata string) map[string]interface{} {
	data := make(map[string]interface{})

	switch recordType {
	case "A", "AAAA":
		data["address"] = rdata
	case "CNAME", "DNAME":
		data["target"] = rdata
	case "NS":
		data["nameserver"] = rdata
	case "PTR":
		data["ptrdname"] = rdata
	case "MX":
		// Parse "priority exchange" format
		parts := strings.SplitN(rdata, " ", 2)
		if len(parts) == 2 {
			data["preference"] = parts[0]
			data["exchange"] = parts[1]
		}
	case "TXT":
		data["text"] = strings.Trim(rdata, "\"")
	case "SRV":
		// Parse "priority weight port target" format
		parts := strings.SplitN(rdata, " ", 4)
		if len(parts) == 4 {
			data["priority"] = parts[0]
			data["weight"] = parts[1]
			data["port"] = parts[2]
			data["target"] = parts[3]
		}
	case "CAA":
		// Parse "flags tag value" format
		parts := strings.SplitN(rdata, " ", 3)
		if len(parts) == 3 {
			data["flags"] = parts[0]
			data["tag"] = parts[1]
			data["value"] = strings.Trim(parts[2], "\"")
		}
	default:
		data["rdata"] = rdata
	}

	return data
}

// Read refreshes the Terraform state
func (r *RecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading record", map[string]any{
		"zone": state.Zone.ValueString(),
		"name": state.Name.ValueString(),
		"type": state.Type.ValueString(),
	})

	records, err := r.client.GetRecords(ctx, state.Zone.ValueString(), state.Type.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Record",
			"Could not read record: "+err.Error(),
		)
		return
	}

	if len(records) == 0 {
		// API couldn't find the record. For dynamic zones, records may be in the journal
		// and not visible via the zone file parser. Don't remove from state - the record
		// likely still exists. Trust the Create operation succeeded.
		tflog.Warn(ctx, "API returned no records, but record may exist in zone journal. Keeping state.", map[string]any{
			"zone": state.Zone.ValueString(),
			"name": state.Name.ValueString(),
			"type": state.Type.ValueString(),
		})
		// Keep existing state - don't remove or modify
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update state with records
	var recordValues []string
	for _, rec := range records {
		recordValues = append(recordValues, rec.RData)
	}

	recordsList, diags := types.ListValueFrom(ctx, types.StringType, recordValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Records = recordsList
	state.TTL = types.Int64Value(int64(records[0].TTL))

	// Set computed convenience attributes
	r.setComputedAttributes(&state, recordValues)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource
func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RecordResourceModel
	var state RecordResourceModel
	
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating record", map[string]any{
		"zone": plan.Zone.ValueString(),
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	// Get old and new records
	var oldRecords, newRecords []string
	diags = state.Records.ElementsAs(ctx, &oldRecords, false)
	resp.Diagnostics.Append(diags...)
	diags = plan.Records.ElementsAs(ctx, &newRecords, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete old records that are no longer present
	for _, oldRdata := range oldRecords {
		found := false
		for _, newRdata := range newRecords {
			if oldRdata == newRdata {
				found = true
				break
			}
		}
		if !found {
			err := r.client.DeleteRecord(ctx, plan.Zone.ValueString(), plan.Name.ValueString(), plan.Type.ValueString(), oldRdata)
			if err != nil {
				tflog.Warn(ctx, "Could not delete old record", map[string]any{"error": err.Error()})
			}
		}
	}

	// Add new records that don't exist
	for _, newRdata := range newRecords {
		found := false
		for _, oldRdata := range oldRecords {
			if oldRdata == newRdata {
				found = true
				break
			}
		}
		if !found {
			createReq := &RecordCreateRequest{
				RecordType:  plan.Type.ValueString(),
				Name:        plan.Name.ValueString(),
				TTL:         int(plan.TTL.ValueInt64()),
				RecordClass: plan.Class.ValueString(),
				Data:        r.buildRecordData(plan.Type.ValueString(), newRdata),
			}
			_, err := r.client.CreateRecord(ctx, plan.Zone.ValueString(), createReq)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Record",
					fmt.Sprintf("Could not create record: %s", err.Error()),
				)
				return
			}
		}
	}

	// Set computed convenience attributes
	r.setComputedAttributes(&plan, newRecords)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *RecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting record", map[string]any{
		"zone": state.Zone.ValueString(),
		"name": state.Name.ValueString(),
		"type": state.Type.ValueString(),
	})

	// Get records to delete
	var records []string
	diags = state.Records.ElementsAs(ctx, &records, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete each record
	for _, rdata := range records {
		err := r.client.DeleteRecord(ctx, state.Zone.ValueString(), state.Name.ValueString(), state.Type.ValueString(), rdata)
		if err != nil {
			if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
				resp.Diagnostics.AddError(
					"Error Deleting Record",
					fmt.Sprintf("Could not delete record: %s", err.Error()),
				)
				return
			}
		}
	}
}

// ImportState imports an existing resource
func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone/name/type
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: zone/name/type (e.g., example.com/www/A)",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[2])...)
}

