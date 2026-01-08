// DNSSEC Key Resource

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var _ resource.Resource = &DNSSECKeyResource{}

// NewDNSSECKeyResource creates a new DNSSEC key resource
func NewDNSSECKeyResource() resource.Resource {
	return &DNSSECKeyResource{}
}

// DNSSECKeyResource defines the resource implementation
type DNSSECKeyResource struct {
	client *Client
}

// DNSSECKeyResourceModel describes the resource data model
type DNSSECKeyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Zone       types.String `tfsdk:"zone"`
	KeyType    types.String `tfsdk:"key_type"`
	Algorithm  types.Int64  `tfsdk:"algorithm"`
	Bits       types.Int64  `tfsdk:"bits"`
	TTL        types.Int64  `tfsdk:"ttl"`
	KeyTag     types.Int64  `tfsdk:"key_tag"`
	State      types.String `tfsdk:"state"`
	Flags      types.Int64  `tfsdk:"flags"`
	PublicKey  types.String `tfsdk:"public_key"`
	DSRecords  types.List   `tfsdk:"ds_records"`
	SignZone   types.Bool   `tfsdk:"sign_zone"`
}

// Metadata returns the resource type name
func (r *DNSSECKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_key"
}

// Schema defines the schema for the resource
func (r *DNSSECKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNSSEC key for a zone.",
		MarkdownDescription: `
Manages DNSSEC keys for DNS zones.

## Example Usage

### Generate KSK (Key Signing Key)

` + "```hcl" + `
resource "bind9_dnssec_key" "ksk" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}
` + "```" + `

### Generate ZSK (Zone Signing Key)

` + "```hcl" + `
resource "bind9_dnssec_key" "zsk" {
  zone      = "example.com"
  key_type  = "ZSK"
  algorithm = 13  # ECDSAP256SHA256
}
` + "```" + `

## Algorithm Reference

| Value | Algorithm |
|-------|-----------|
| 8 | RSASHA256 |
| 10 | RSASHA512 |
| 13 | ECDSAP256SHA256 (recommended) |
| 14 | ECDSAP384SHA384 |
| 15 | ED25519 |
| 16 | ED448 |
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Key identifier (zone/key_tag)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Zone name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_type": schema.StringAttribute{
				Description: "Key type: KSK, ZSK, or CSK",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("KSK", "ZSK", "CSK"),
				},
			},
			"algorithm": schema.Int64Attribute{
				Description: "DNSSEC algorithm number (8=RSASHA256, 13=ECDSAP256SHA256, 14=ECDSAP384SHA384, 15=ED25519)",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(13),
				PlanModifiers: []planmodifier.Int64{
					// Requires replace since algorithm can't be changed
				},
				Validators: []validator.Int64{
					int64validator.OneOf(8, 10, 13, 14, 15, 16),
				},
			},
			"bits": schema.Int64Attribute{
				Description: "Key size in bits (only for RSA algorithms)",
				Optional:    true,
				Computed:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "DNSKEY record TTL",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"key_tag": schema.Int64Attribute{
				Description: "Key tag (computed)",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Key state",
				Computed:    true,
			},
			"flags": schema.Int64Attribute{
				Description: "DNSKEY flags",
				Computed:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "Base64-encoded public key",
				Computed:    true,
			},
			"ds_records": schema.ListAttribute{
				Description: "DS records for registrar",
				Computed:    true,
				ElementType: types.StringType,
			},
			"sign_zone": schema.BoolAttribute{
				Description: "Sign zone after key creation",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *DNSSECKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DNSSECKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DNSSECKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DNSSEC key", map[string]any{
		"zone":     plan.Zone.ValueString(),
		"key_type": plan.KeyType.ValueString(),
	})

	createReq := &DNSSECKeyCreateRequest{
		KeyType:   plan.KeyType.ValueString(),
		Algorithm: int(plan.Algorithm.ValueInt64()),
		TTL:       int(plan.TTL.ValueInt64()),
	}

	if !plan.Bits.IsNull() && plan.Bits.ValueInt64() > 0 {
		createReq.Bits = int(plan.Bits.ValueInt64())
	}

	key, err := r.client.CreateDNSSECKey(ctx, plan.Zone.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNSSEC Key",
			"Could not create DNSSEC key: "+err.Error(),
		)
		return
	}

	// Sign zone if requested
	if !plan.SignZone.IsNull() && plan.SignZone.ValueBool() {
		if err := r.client.SignZone(ctx, plan.Zone.ValueString()); err != nil {
			tflog.Warn(ctx, "Could not sign zone", map[string]any{"error": err.Error()})
		}
	}

	// Set state
	plan.ID = types.StringValue(fmt.Sprintf("%s/%d", plan.Zone.ValueString(), key.KeyTag))
	plan.KeyTag = types.Int64Value(int64(key.KeyTag))
	plan.State = types.StringValue(key.State)
	plan.Flags = types.Int64Value(int64(key.Flags))
	plan.Bits = types.Int64Value(int64(key.Bits))
	
	if key.PublicKey != "" {
		plan.PublicKey = types.StringValue(key.PublicKey)
	}

	// DS records
	if len(key.DSRecords) > 0 {
		dsRecords, diags := types.ListValueFrom(ctx, types.StringType, key.DSRecords)
		resp.Diagnostics.Append(diags...)
		plan.DSRecords = dsRecords
	} else {
		plan.DSRecords = types.ListNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state
func (r *DNSSECKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DNSSECKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading DNSSEC key", map[string]any{
		"zone":    state.Zone.ValueString(),
		"key_tag": state.KeyTag.ValueInt64(),
	})

	keys, err := r.client.ListDNSSECKeys(ctx, state.Zone.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading DNSSEC Key",
			"Could not read DNSSEC keys: "+err.Error(),
		)
		return
	}

	// Find the key by tag
	var foundKey *DNSSECKey
	for _, k := range keys {
		if int64(k.KeyTag) == state.KeyTag.ValueInt64() {
			foundKey = &k
			break
		}
	}

	if foundKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state
	state.State = types.StringValue(foundKey.State)
	state.Flags = types.Int64Value(int64(foundKey.Flags))
	state.Bits = types.Int64Value(int64(foundKey.Bits))
	state.Algorithm = types.Int64Value(int64(foundKey.Algorithm))
	
	if foundKey.PublicKey != "" {
		state.PublicKey = types.StringValue(foundKey.PublicKey)
	}

	if len(foundKey.DSRecords) > 0 {
		dsRecords, diags := types.ListValueFrom(ctx, types.StringType, foundKey.DSRecords)
		resp.Diagnostics.Append(diags...)
		state.DSRecords = dsRecords
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource
func (r *DNSSECKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// DNSSEC keys are immutable - no update needed
	var plan DNSSECKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sign zone if requested
	if !plan.SignZone.IsNull() && plan.SignZone.ValueBool() {
		if err := r.client.SignZone(ctx, plan.Zone.ValueString()); err != nil {
			tflog.Warn(ctx, "Could not sign zone", map[string]any{"error": err.Error()})
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *DNSSECKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DNSSECKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting DNSSEC key", map[string]any{
		"zone":    state.Zone.ValueString(),
		"key_tag": state.KeyTag.ValueInt64(),
	})

	err := r.client.DeleteDNSSECKey(ctx, state.Zone.ValueString(), int(state.KeyTag.ValueInt64()))
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error Deleting DNSSEC Key",
				"Could not delete DNSSEC key: "+err.Error(),
			)
			return
		}
	}
}

