// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/walletobjects/v1"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &PermissionsResource{}
	_ resource.ResourceWithImportState = &PermissionsResource{}
)

// NewPermissionsResource creates a new permissions resource.
func NewPermissionsResource() resource.Resource {
	return &PermissionsResource{}
}

// PermissionsResource defines the resource implementation.
type PermissionsResource struct {
	client *Client
}

// sortPermissions sorts permissions by email address for consistent ordering.
// This is necessary because the Google Wallet API may return permissions in
// a different order than they were set, causing Terraform to detect false changes.
func sortPermissions(perms []PermissionModel) {
	sort.Slice(perms, func(i, j int) bool {
		return perms[i].EmailAddress.ValueString() < perms[j].EmailAddress.ValueString()
	})
}

func (r *PermissionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (r *PermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages permissions for a Google Wallet Issuer.

This resource is **authoritative** - it manages all permissions for an issuer. Any permissions not defined in this resource will be removed when applied.

~> **Warning:** This resource will replace ALL permissions for the issuer. Make sure to include all desired permissions in your configuration.

## Example Usage

` + "```hcl" + `
resource "googlewallet_permissions" "example" {
  issuer_id = googlewallet_issuer.example.id

  permissions = [
    {
      email_address = "admin@example.com"
      role          = "OWNER"
    },
    {
      email_address = "developer@example.com"
      role          = "WRITER"
    },
    {
      email_address = "viewer@example.com"
      role          = "READER"
    }
  ]
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"issuer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the issuer to manage permissions for. This is the unique identifier assigned by Google.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The complete list of permissions for this issuer. This is authoritative - any permissions not included will be removed.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email_address": schema.StringAttribute{
							MarkdownDescription: "The email address of the user, group, or service account.",
							Required:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role to grant. Valid values are `OWNER`, `WRITER`, or `READER`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("OWNER", "WRITER", "READER"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *PermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *PermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PermissionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	tflog.Debug(ctx, "Creating permissions", map[string]interface{}{
		"issuer_id": data.IssuerID.ValueString(),
	})

	// Build the API request
	permissions := &walletobjects.Permissions{
		IssuerId: issuerID,
	}

	for _, perm := range data.Permissions {
		permissions.Permissions = append(permissions.Permissions, &walletobjects.Permission{
			EmailAddress: perm.EmailAddress.ValueString(),
			Role:         perm.Role.ValueString(),
		})
	}

	// Update the permissions (create is the same as update for this resource)
	updated, err := r.client.UpdatePermissions(ctx, issuerID, permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Permissions",
			fmt.Sprintf("Could not create permissions for issuer %s: %s", data.IssuerID.ValueString(), err.Error()),
		)
		return
	}

	// Map response back to Terraform state
	// Note: Keep the issuer_id from the plan - the API doesn't return it in the response
	// Note: The API returns roles in lowercase, but we normalize to uppercase for consistency
	data.Permissions = make([]PermissionModel, len(updated.Permissions))
	for i, perm := range updated.Permissions {
		data.Permissions[i] = PermissionModel{
			EmailAddress: types.StringValue(perm.EmailAddress),
			Role:         types.StringValue(strings.ToUpper(perm.Role)),
		}
	}
	// Sort permissions by email to ensure consistent ordering
	sortPermissions(data.Permissions)

	tflog.Info(ctx, "Created permissions", map[string]interface{}{
		"issuer_id":         data.IssuerID.ValueString(),
		"permissions_count": len(data.Permissions),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PermissionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
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

	tflog.Debug(ctx, "Reading permissions", map[string]interface{}{
		"issuer_id": data.IssuerID.ValueString(),
	})

	// Get permissions from API
	permissions, err := r.client.GetPermissions(ctx, issuerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Permissions",
			fmt.Sprintf("Could not read permissions for issuer %s: %s", data.IssuerID.ValueString(), err.Error()),
		)
		return
	}

	// Map response to Terraform state
	// Note: Keep the issuer_id from the current state - the API doesn't return it in the response
	// Note: The API returns roles in lowercase, but we normalize to uppercase for consistency
	data.Permissions = make([]PermissionModel, len(permissions.Permissions))
	for i, perm := range permissions.Permissions {
		data.Permissions[i] = PermissionModel{
			EmailAddress: types.StringValue(perm.EmailAddress),
			Role:         types.StringValue(strings.ToUpper(perm.Role)),
		}
	}
	// Sort permissions by email to ensure consistent ordering
	sortPermissions(data.Permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PermissionsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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

	tflog.Debug(ctx, "Updating permissions", map[string]interface{}{
		"issuer_id": data.IssuerID.ValueString(),
	})

	// Build the API request
	permissions := &walletobjects.Permissions{
		IssuerId: issuerID,
	}

	for _, perm := range data.Permissions {
		permissions.Permissions = append(permissions.Permissions, &walletobjects.Permission{
			EmailAddress: perm.EmailAddress.ValueString(),
			Role:         perm.Role.ValueString(),
		})
	}

	// Update the permissions
	updated, err := r.client.UpdatePermissions(ctx, issuerID, permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Permissions",
			fmt.Sprintf("Could not update permissions for issuer %s: %s", data.IssuerID.ValueString(), err.Error()),
		)
		return
	}

	// Map response back to Terraform state
	// Note: Keep the issuer_id from the plan - the API doesn't return it in the response
	// Note: The API returns roles in lowercase, but we normalize to uppercase for consistency
	data.Permissions = make([]PermissionModel, len(updated.Permissions))
	for i, perm := range updated.Permissions {
		data.Permissions[i] = PermissionModel{
			EmailAddress: types.StringValue(perm.EmailAddress),
			Role:         types.StringValue(strings.ToUpper(perm.Role)),
		}
	}
	// Sort permissions by email to ensure consistent ordering
	sortPermissions(data.Permissions)

	tflog.Info(ctx, "Updated permissions", map[string]interface{}{
		"issuer_id":         data.IssuerID.ValueString(),
		"permissions_count": len(data.Permissions),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PermissionsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "Deleting permissions resource from Terraform state. Note: The issuer may retain some default permissions.", map[string]interface{}{
		"issuer_id": data.IssuerID.ValueString(),
	})

	// For delete, we could either:
	// 1. Remove all permissions (leaving issuer in potentially unusable state)
	// 2. Just remove from state (current approach)
	//
	// We choose option 2 as it's safer. The issuer will retain whatever
	// permissions it currently has.
}

func (r *PermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by issuer_id
	resource.ImportStatePassthroughID(ctx, path.Root("issuer_id"), req, resp)
}
