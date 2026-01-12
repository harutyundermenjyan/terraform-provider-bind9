// ACL Resource

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &ACLResource{}
	_ resource.ResourceWithImportState = &ACLResource{}
)

// NewACLResource creates a new ACL resource
func NewACLResource() resource.Resource {
	return &ACLResource{}
}

// ACLResource defines the resource implementation
type ACLResource struct {
	client *Client
}

// ACLResourceModel describes the resource data model
type ACLResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Entries types.List   `tfsdk:"entries"`
	Comment types.String `tfsdk:"comment"`
}

// ACL API response model
type ACLAPIResponse struct {
	Name    string   `json:"name"`
	Entries []string `json:"entries"`
	Comment string   `json:"comment,omitempty"`
}

// ACL API request model
type ACLAPIRequest struct {
	Name    string   `json:"name"`
	Entries []string `json:"entries"`
	Comment string   `json:"comment,omitempty"`
}

// ACL Update request model
type ACLUpdateRequest struct {
	Entries []string `json:"entries,omitempty"`
	Comment string   `json:"comment,omitempty"`
}

// Metadata returns the resource type name
func (r *ACLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl"
}

// Schema defines the schema for the resource
func (r *ACLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a BIND9 Access Control List (ACL).",
		MarkdownDescription: `
Manages an Access Control List (ACL) on a BIND9 server.

ACLs are named groups of IP addresses, networks, TSIG keys, or other ACLs 
that can be referenced in zone configurations for access control.

## Example Usage

### Basic ACL

` + "```hcl" + `
resource "bind9_acl" "internal" {
  name = "internal"
  entries = [
    "localhost",
    "localnets",
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
  ]
  comment = "Internal networks"
}
` + "```" + `

### ACL with TSIG Key

` + "```hcl" + `
resource "bind9_acl" "ddns_clients" {
  name = "ddns-clients"
  entries = [
    "172.25.44.0/24",
    "key \"ddns-key\"",
  ]
  comment = "Clients allowed to perform dynamic DNS updates"
}
` + "```" + `

### ACL for Zone Transfers

` + "```hcl" + `
resource "bind9_acl" "trusted_secondaries" {
  name = "trusted-secondaries"
  entries = [
    "192.168.1.11",
    "192.168.1.12",
    "key \"transfer-key\"",
  ]
  comment = "Secondary DNS servers"
}
` + "```" + `

## Usage in Zones

Once defined, ACLs can be referenced in zone configurations:

` + "```hcl" + `
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
  
  allow_query    = ["internal"]           # Reference the ACL by name
  allow_transfer = ["trusted-secondaries"]
  allow_update   = ["ddns-clients"]
}
` + "```" + `

## Valid Entry Types

- IP addresses: ` + "`192.168.1.1`" + `
- Networks (CIDR): ` + "`10.0.0.0/8`" + `
- TSIG keys: ` + "`key \"keyname\"`" + `
- Built-in ACLs: ` + "`localhost`" + `, ` + "`localnets`" + `, ` + "`any`" + `, ` + "`none`" + `
- Other ACL names: ` + "`internal`" + `
- Negated entries: ` + "`!192.168.1.100`" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ACL identifier (same as name)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "ACL name (must be unique, cannot be reserved names like 'any', 'none', 'localhost', 'localnets')",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entries": schema.ListAttribute{
				Description: "List of ACL entries (IPs, networks, keys, or ACL references)",
				Required:    true,
				ElementType: types.StringType,
			},
			"comment": schema.StringAttribute{
				Description: "Optional description/comment for the ACL. Server may append a timestamp.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *ACLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new ACL
func (r *ACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert entries from types.List to []string
	var entries []string
	diags = plan.Entries.ElementsAs(ctx, &entries, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	aclReq := ACLAPIRequest{
		Name:    plan.Name.ValueString(),
		Entries: entries,
		Comment: plan.Comment.ValueString(),
	}

	tflog.Debug(ctx, "Creating ACL", map[string]interface{}{
		"name":    aclReq.Name,
		"entries": entries,
	})

	// Create ACL - pass struct directly, doRequest will marshal it
	httpResp, err := r.client.doRequest(ctx, "POST", "/api/v1/acls", aclReq)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not create ACL: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		errBody := make([]byte, 1024)
		n, _ := httpResp.Body.Read(errBody)
		resp.Diagnostics.AddError(
			"Error Creating ACL",
			fmt.Sprintf("Could not create ACL: API error %d: %s", httpResp.StatusCode, string(errBody[:n])),
		)
		return
	}

	// Parse response
	var aclResp ACLAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&aclResp); err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	// Set state
	plan.ID = types.StringValue(aclResp.Name)
	plan.Name = types.StringValue(aclResp.Name)

	entriesList, diags := types.ListValueFrom(ctx, types.StringType, aclResp.Entries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Entries = entriesList

	if aclResp.Comment != "" {
		plan.Comment = types.StringValue(aclResp.Comment)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read reads the ACL state
func (r *ACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading ACL", map[string]interface{}{"name": name})

	// Get ACL from API
	httpResp, err := r.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/acls/%s", name), nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not read ACL: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		// ACL was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		errBody := make([]byte, 1024)
		n, _ := httpResp.Body.Read(errBody)
		resp.Diagnostics.AddError(
			"Error Reading ACL",
			fmt.Sprintf("Could not read ACL: API error %d: %s", httpResp.StatusCode, string(errBody[:n])),
		)
		return
	}

	var aclResp ACLAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&aclResp); err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	// Update state
	state.ID = types.StringValue(aclResp.Name)
	state.Name = types.StringValue(aclResp.Name)

	entriesList, diags := types.ListValueFrom(ctx, types.StringType, aclResp.Entries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Entries = entriesList

	if aclResp.Comment != "" {
		state.Comment = types.StringValue(aclResp.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates an existing ACL
func (r *ACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	// Convert entries from types.List to []string
	var entries []string
	diags = plan.Entries.ElementsAs(ctx, &entries, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := ACLUpdateRequest{
		Entries: entries,
		Comment: plan.Comment.ValueString(),
	}

	tflog.Debug(ctx, "Updating ACL", map[string]interface{}{
		"name":    name,
		"entries": entries,
	})

	// Update ACL - pass struct directly, doRequest will marshal it
	httpResp, err := r.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/acls/%s", name), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not update ACL: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		errBody := make([]byte, 1024)
		n, _ := httpResp.Body.Read(errBody)
		resp.Diagnostics.AddError(
			"Error Updating ACL",
			fmt.Sprintf("Could not update ACL: API error %d: %s", httpResp.StatusCode, string(errBody[:n])),
		)
		return
	}

	// Parse response
	var aclResp ACLAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&aclResp); err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	// Update state
	plan.ID = types.StringValue(aclResp.Name)

	entriesList, diags := types.ListValueFrom(ctx, types.StringType, aclResp.Entries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Entries = entriesList

	if aclResp.Comment != "" {
		plan.Comment = types.StringValue(aclResp.Comment)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes an ACL
func (r *ACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting ACL", map[string]interface{}{"name": name})

	// Delete ACL
	httpResp, err := r.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/acls/%s", name), nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Could not delete ACL: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// 204 No Content or 404 Not Found are both acceptable
	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusNotFound {
		errBody := make([]byte, 1024)
		n, _ := httpResp.Body.Read(errBody)
		resp.Diagnostics.AddError(
			"Error Deleting ACL",
			fmt.Sprintf("Could not delete ACL: API error %d: %s", httpResp.StatusCode, string(errBody[:n])),
		)
		return
	}
}

// ImportState imports an existing ACL into Terraform state
func (r *ACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by name
	name := strings.TrimSpace(req.ID)

	tflog.Debug(ctx, "Importing ACL", map[string]interface{}{"name": name})

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}
