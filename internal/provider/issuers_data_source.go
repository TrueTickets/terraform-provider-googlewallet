// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

Use this data source to list all issuers you have access to. By default, archived issuers (names starting with "[ARCHIVED] ") and test issuers (names starting with "[TESTING] ") are excluded.

## Example Usage

` + "```hcl" + `
# List only active issuers (default behavior - excludes archived and testing)
data "googlewallet_issuers" "active" {}

# List all issuers including archived ones
data "googlewallet_issuers" "with_archived" {
  include_archived = true
}

# List all issuers including test issuers (useful for test cleanup)
data "googlewallet_issuers" "with_testing" {
  include_testing = true
}

output "active_issuer_count" {
  value = length(data.googlewallet_issuers.active.issuers)
}

output "issuer_names" {
  value = [for issuer in data.googlewallet_issuers.active.issuers : issuer.name]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"include_archived": schema.BoolAttribute{
				MarkdownDescription: "Whether to include archived issuers (those with names starting with \"[ARCHIVED] \"). Defaults to false.",
				Optional:            true,
			},
			"include_testing": schema.BoolAttribute{
				MarkdownDescription: "Whether to include test issuers (those with names starting with \"[TESTING] \"). Defaults to false. Useful for cleaning up after acceptance tests.",
				Optional:            true,
			},
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

	// Determine whether to include archived issuers (default: false)
	includeArchived := false
	if !data.IncludeArchived.IsNull() {
		includeArchived = data.IncludeArchived.ValueBool()
	}

	// Determine whether to include testing issuers (default: false)
	includeTesting := false
	if !data.IncludeTesting.IsNull() {
		includeTesting = data.IncludeTesting.ValueBool()
	}

	tflog.Debug(ctx, "Listing issuers", map[string]interface{}{
		"include_archived": includeArchived,
		"include_testing":  includeTesting,
	})

	// Get all issuers from API
	issuers, err := d.client.ListIssuers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Issuers",
			fmt.Sprintf("Could not list issuers: %s", err.Error()),
		)
		return
	}

	// Filter and map response to Terraform state
	data.Issuers = make([]IssuerModel, 0, len(issuers))
	archivedSkipped := 0
	testingSkipped := 0

	for _, issuer := range issuers {
		// Filter out archived issuers unless explicitly included
		if !includeArchived && strings.HasPrefix(issuer.Name, ArchivedPrefix) {
			archivedSkipped++
			continue
		}

		// Filter out testing issuers unless explicitly included
		if !includeTesting && strings.HasPrefix(issuer.Name, TestingPrefix) {
			testingSkipped++
			continue
		}

		issuerModel := IssuerModel{
			ID:   types.StringValue(strconv.FormatInt(issuer.IssuerId, 10)),
			Name: types.StringValue(issuer.Name),
		}

		if issuer.HomepageUrl != "" {
			issuerModel.HomepageURL = types.StringValue(issuer.HomepageUrl)
		} else {
			issuerModel.HomepageURL = types.StringNull()
		}

		if issuer.ContactInfo != nil {
			contactInfoObj, diags := contactInfoToObject(ctx, issuer.ContactInfo)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			issuerModel.ContactInfo = contactInfoObj
		} else {
			issuerModel.ContactInfo = types.ObjectNull(contactInfoAttrTypes())
		}

		data.Issuers = append(data.Issuers, issuerModel)
	}

	tflog.Info(ctx, "Listed issuers", map[string]interface{}{
		"total":            len(issuers),
		"returned":         len(data.Issuers),
		"archived_skipped": archivedSkipped,
		"testing_skipped":  testingSkipped,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
