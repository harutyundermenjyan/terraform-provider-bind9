// Record Data Source

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
var _ datasource.DataSource = &RecordDataSource{}

// NewRecordDataSource creates a new record data source
func NewRecordDataSource() datasource.DataSource {
	return &RecordDataSource{}
}

// RecordDataSource defines the data source implementation
type RecordDataSource struct {
	client *Client
}

// RecordDataSourceModel describes the data source data model
type RecordDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Zone    types.String `tfsdk:"zone"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	TTL     types.Int64  `tfsdk:"ttl"`
	Records types.List   `tfsdk:"records"`
}

// Metadata returns the data source type name
func (d *RecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

// Schema defines the schema for the data source
func (d *RecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves DNS record(s) by name and type.",
		MarkdownDescription: `
Retrieves DNS records from a zone by name and type.

## Example Usage

` + "```hcl" + `
data "bind9_record" "www" {
  zone = "example.com"
  name = "www"
  type = "A"
}

output "www_ips" {
  value = data.bind9_record.www.records
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Record identifier",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Zone name",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Record name",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Record type",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Record TTL",
				Computed:    true,
			},
			"records": schema.ListAttribute{
				Description: "Record values",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *RecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *RecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RecordDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading record data", map[string]any{
		"zone": config.Zone.ValueString(),
		"name": config.Name.ValueString(),
		"type": config.Type.ValueString(),
	})

	records, err := d.client.GetRecords(ctx, config.Zone.ValueString(), config.Type.ValueString(), config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Record",
			"Could not read record: "+err.Error(),
		)
		return
	}

	if len(records) == 0 {
		resp.Diagnostics.AddError(
			"Record Not Found",
			fmt.Sprintf("No records found for %s %s in zone %s", config.Name.ValueString(), config.Type.ValueString(), config.Zone.ValueString()),
		)
		return
	}

	config.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", config.Zone.ValueString(), config.Name.ValueString(), config.Type.ValueString()))
	config.TTL = types.Int64Value(int64(records[0].TTL))

	var recordValues []string
	for _, r := range records {
		recordValues = append(recordValues, r.RData)
	}

	recordsList, diags := types.ListValueFrom(ctx, types.StringType, recordValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Records = recordsList

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}


// ============================================================================
// Records (plural) Data Source
// ============================================================================

var _ datasource.DataSource = &RecordsDataSource{}

// NewRecordsDataSource creates a new records data source
func NewRecordsDataSource() datasource.DataSource {
	return &RecordsDataSource{}
}

// RecordsDataSource defines the data source implementation
type RecordsDataSource struct {
	client *Client
}

// RecordsListModel describes a single record
type RecordsListModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	TTL   types.Int64  `tfsdk:"ttl"`
	RData types.String `tfsdk:"rdata"`
}

// RecordsDataSourceModel describes the data source data model
type RecordsDataSourceModel struct {
	ID      types.String       `tfsdk:"id"`
	Zone    types.String       `tfsdk:"zone"`
	Type    types.String       `tfsdk:"type"`
	Name    types.String       `tfsdk:"name"`
	Records []RecordsListModel `tfsdk:"records"`
}

// Metadata returns the data source type name
func (d *RecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

// Schema defines the schema for the data source
func (d *RecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all DNS records in a zone.",
		MarkdownDescription: `
Retrieves all DNS records from a zone with optional filtering.

## Example Usage

` + "```hcl" + `
# Get all records in a zone
data "bind9_records" "all" {
  zone = "example.com"
}

# Get all A records
data "bind9_records" "a_records" {
  zone = "example.com"
  type = "A"
}

# Get records for specific name
data "bind9_records" "www" {
  zone = "example.com"
  name = "www"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Data source identifier",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Zone name",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Filter by record type",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Filter by record name",
				Optional:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "List of records",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Record name",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Record type",
							Computed:    true,
						},
						"ttl": schema.Int64Attribute{
							Description: "Record TTL",
							Computed:    true,
						},
						"rdata": schema.StringAttribute{
							Description: "Record data",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *RecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *RecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RecordsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading records data", map[string]any{
		"zone": config.Zone.ValueString(),
	})

	recordType := ""
	if !config.Type.IsNull() {
		recordType = config.Type.ValueString()
	}

	name := ""
	if !config.Name.IsNull() {
		name = config.Name.ValueString()
	}

	records, err := d.client.GetRecords(ctx, config.Zone.ValueString(), recordType, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Records",
			"Could not read records: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(config.Zone.ValueString())
	config.Records = []RecordsListModel{}

	for _, r := range records {
		config.Records = append(config.Records, RecordsListModel{
			Name:  types.StringValue(r.Name),
			Type:  types.StringValue(r.Type),
			TTL:   types.Int64Value(int64(r.TTL)),
			RData: types.StringValue(r.RData),
		})
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

