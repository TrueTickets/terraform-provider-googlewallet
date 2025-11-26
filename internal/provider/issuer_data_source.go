// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IssuerDataSource{}

// NewIssuerDataSource creates a new issuer data source.
func NewIssuerDataSource() datasource.DataSource {
	return &IssuerDataSource{}
}

// IssuerDataSource defines the data source implementation.
type IssuerDataSource struct {
	client *Client
}

func (d *IssuerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issuer"
}

func (d *IssuerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Fetches a single Google Wallet Issuer by ID.

Use this data source to look up an existing issuer by its unique identifier.

## Example Usage

` + "```hcl" + `
data "googlewallet_issuer" "example" {
  id = "1234567890123456789"
}

output "issuer_name" {
  value = data.googlewallet_issuer.example.name
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the issuer. This is a string representation of the int64 issuer ID.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The account name of the issuer.",
				Computed:            true,
			},
			"homepage_url": schema.StringAttribute{
				MarkdownDescription: "URL for the issuer's home page.",
				Computed:            true,
			},
			"contact_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Contact information for the issuer.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Primary contact name.",
						Computed:            true,
					},
					"phone": schema.StringAttribute{
						MarkdownDescription: "Primary contact phone number.",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Primary contact email address.",
						Computed:            true,
					},
					"alerts_emails": schema.ListAttribute{
						MarkdownDescription: "Email addresses that receive alerts.",
						Computed:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

func (d *IssuerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *IssuerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IssuerModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	issuerID, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Issuer ID",
			fmt.Sprintf("Could not parse issuer ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Reading issuer data source", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get the issuer from API
	issuer, err := d.client.GetIssuer(ctx, issuerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Issuer",
			fmt.Sprintf("Could not read issuer %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map response to Terraform state
	data.ID = types.StringValue(strconv.FormatInt(issuer.IssuerId, 10))
	data.Name = types.StringValue(issuer.Name)

	if issuer.HomepageUrl != "" {
		data.HomepageURL = types.StringValue(issuer.HomepageUrl)
	} else {
		data.HomepageURL = types.StringNull()
	}

	if issuer.ContactInfo != nil {
		contactInfoObj, diags := contactInfoToObject(ctx, issuer.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ContactInfo = contactInfoObj
	} else {
		data.ContactInfo = types.ObjectNull(contactInfoAttrTypes())
	}

	tflog.Info(ctx, "Read issuer data source", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
