// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/walletobjects/v1"
)

// Client wraps the Google Wallet API service and provides convenient access
// to issuer and permissions operations.
type Client struct {
	// service is the underlying Google Wallet API service.
	service *walletobjects.Service

	// Issuers provides access to issuer operations.
	Issuers *walletobjects.IssuerService

	// Permissions provides access to permissions operations.
	Permissions *walletobjects.PermissionsService
}

// NewClient creates a new Google Wallet API client.
// If credentialsJSON is provided, it uses explicit credentials.
// If credentialsJSON is empty, it uses Application Default Credentials (ADC).
func NewClient(ctx context.Context, credentialsJSON string) (*Client, error) {
	var opts []option.ClientOption

	// Always set the required scope
	opts = append(opts, option.WithScopes(walletobjects.WalletObjectIssuerScope))

	// If credentials are provided, use them explicitly
	// Otherwise, the Google API client will use Application Default Credentials (ADC)
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	// Create the walletobjects service
	service, err := walletobjects.NewService(ctx, opts...)
	if err != nil {
		if credentialsJSON == "" {
			return nil, fmt.Errorf("failed to create walletobjects service with Application Default Credentials: %w", err)
		}
		return nil, fmt.Errorf("failed to create walletobjects service: %w", err)
	}

	return &Client{
		service:     service,
		Issuers:     service.Issuer,
		Permissions: service.Permissions,
	}, nil
}

// GetIssuer retrieves an issuer by its resource ID.
// Note: Google Wallet API uses int64 for issuer IDs.
func (c *Client) GetIssuer(ctx context.Context, resourceID int64) (*walletobjects.Issuer, error) {
	issuer, err := c.Issuers.Get(resourceID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get issuer %d: %w", resourceID, err)
	}
	return issuer, nil
}

// ListIssuers retrieves all issuers.
func (c *Client) ListIssuers(ctx context.Context) ([]*walletobjects.Issuer, error) {
	resp, err := c.Issuers.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list issuers: %w", err)
	}
	return resp.Resources, nil
}

// CreateIssuer creates a new issuer.
func (c *Client) CreateIssuer(ctx context.Context, issuer *walletobjects.Issuer) (*walletobjects.Issuer, error) {
	created, err := c.Issuers.Insert(issuer).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create issuer: %w", err)
	}
	return created, nil
}

// UpdateIssuer updates an existing issuer using PATCH (partial update).
// Note: Google Wallet API uses int64 for issuer IDs.
func (c *Client) UpdateIssuer(ctx context.Context, resourceID int64, issuer *walletobjects.Issuer) (*walletobjects.Issuer, error) {
	updated, err := c.Issuers.Patch(resourceID, issuer).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update issuer %d: %w", resourceID, err)
	}
	return updated, nil
}

// GetPermissions retrieves permissions for a resource.
// Note: Google Wallet API uses int64 for resource IDs.
func (c *Client) GetPermissions(ctx context.Context, resourceID int64) (*walletobjects.Permissions, error) {
	perms, err := c.Permissions.Get(resourceID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for %d: %w", resourceID, err)
	}
	return perms, nil
}

// UpdatePermissions updates permissions for a resource.
// Note: Google Wallet API uses int64 for resource IDs.
func (c *Client) UpdatePermissions(ctx context.Context, resourceID int64, permissions *walletobjects.Permissions) (*walletobjects.Permissions, error) {
	updated, err := c.Permissions.Update(resourceID, permissions).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update permissions for %d: %w", resourceID, err)
	}
	return updated, nil
}
