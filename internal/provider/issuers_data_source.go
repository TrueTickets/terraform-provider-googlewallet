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
var _ datasource.DataSource = &IssuersDataSource{}

// NewIssuersDataSource creates a new issuers data source.
func NewIssuersDataSource() datasource.DataSource {
	return &IssuersDataSource{}
}

// IssuersDataSource defines the data source implementation for listing issuers.
type IssuersDataSource struct {
	client *Client
}

func (d *IssuersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issuers"
}

func (d *IssuersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Fetches all Google Wallet Issuers accessible to the authenticated service account.

Use this data source to list all issuers you have access to.

## Example Usage

` + "```hcl" + `
data "googlewallet_issuers" "all" {}

output "issuer_count" {
  value = length(data.googlewallet_issuers.all.issuers)
}

output "issuer_names" {
  value = [for issuer in data.googlewallet_issuers.all.issuers : issuer.name]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"issuers": schema.ListNestedAttribute{
				MarkdownDescription: "List of all issuers accessible to the authenticated service account.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier for the issuer.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *IssuersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IssuersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IssuersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Listing all issuers")

	// Get all issuers from API
	issuers, err := d.client.ListIssuers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Issuers",
			fmt.Sprintf("Could not list issuers: %s", err.Error()),
		)
		return
	}

	// Map response to Terraform state
	data.Issuers = make([]IssuerModel, len(issuers))
	for i, issuer := range issuers {
		data.Issuers[i] = IssuerModel{
			ID:   types.StringValue(strconv.FormatInt(issuer.IssuerId, 10)),
			Name: types.StringValue(issuer.Name),
		}

		if issuer.HomepageUrl != "" {
			data.Issuers[i].HomepageURL = types.StringValue(issuer.HomepageUrl)
		} else {
			data.Issuers[i].HomepageURL = types.StringNull()
		}

		if issuer.ContactInfo != nil {
			contactInfoObj, diags := contactInfoToObject(ctx, issuer.ContactInfo)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.Issuers[i].ContactInfo = contactInfoObj
		} else {
			data.Issuers[i].ContactInfo = types.ObjectNull(contactInfoAttrTypes())
		}
	}

	tflog.Info(ctx, "Listed issuers", map[string]interface{}{
		"count": len(issuers),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
