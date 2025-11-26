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
var _ datasource.DataSource = &PermissionsDataSource{}

// NewPermissionsDataSource creates a new permissions data source.
func NewPermissionsDataSource() datasource.DataSource {
	return &PermissionsDataSource{}
}

// PermissionsDataSource defines the data source implementation.
type PermissionsDataSource struct {
	client *Client
}

func (d *PermissionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (d *PermissionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Fetches the permissions for a Google Wallet Issuer.

Use this data source to read the current permissions for an issuer.

## Example Usage

` + "```hcl" + `
data "googlewallet_permissions" "example" {
  issuer_id = "1234567890123456789"
}

output "permission_count" {
  value = length(data.googlewallet_permissions.example.permissions)
}

output "owners" {
  value = [for p in data.googlewallet_permissions.example.permissions : p.email_address if p.role == "OWNER"]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"issuer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the issuer to fetch permissions for.",
				Required:            true,
			},
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The list of permissions for this issuer.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email_address": schema.StringAttribute{
							MarkdownDescription: "The email address of the user, group, or service account.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role granted (OWNER, WRITER, or READER).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PermissionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the issuer ID
	issuerID, err := strconv.ParseInt(data.IssuerID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Issuer ID",
			fmt.Sprintf("Could not parse issuer ID %q: %s", data.IssuerID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Reading permissions data source", map[string]interface{}{
		"issuer_id": data.IssuerID.ValueString(),
	})

	// Get permissions from API
	permissions, err := d.client.GetPermissions(ctx, issuerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Permissions",
			fmt.Sprintf("Could not read permissions for issuer %s: %s", data.IssuerID.ValueString(), err.Error()),
		)
		return
	}

	// Map response to Terraform state
	data.IssuerID = types.StringValue(strconv.FormatInt(permissions.IssuerId, 10))
	data.Permissions = make([]PermissionModel, len(permissions.Permissions))
	for i, perm := range permissions.Permissions {
		data.Permissions[i] = PermissionModel{
			EmailAddress: types.StringValue(perm.EmailAddress),
			Role:         types.StringValue(perm.Role),
		}
	}

	tflog.Info(ctx, "Read permissions data source", map[string]interface{}{
		"issuer_id":         data.IssuerID.ValueString(),
		"permissions_count": len(data.Permissions),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
