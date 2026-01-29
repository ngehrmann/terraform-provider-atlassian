package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AtlassianClient represents the client for interacting with Atlassian APIs
type AtlassianClient struct {
	APIToken     string
	Email        string
	Organization string
	SiteId       string
	OrgId        string
	BaseURL      string
	HTTPClient   *http.Client
}

// Team represents an Atlassian Team (matches PublicApiTeam schema)
type Team struct {
	TeamID         string `json:"teamId"`
	DisplayName    string `json:"displayName"`
	Description    string `json:"description"`
	TeamType       string `json:"teamType"` // OPEN, MEMBER_INVITE, EXTERNAL, ORG_ADMIN_MANAGED
	OrganizationId string `json:"organizationId"`
	CreatorId      string `json:"creatorId,omitempty"`
	State          string `json:"state"` // ACTIVE, ARCHIVED
}

// TeamResponse represents the response from team operations (matches PublicApiTeamResponse)
type TeamResponse struct {
	TeamID          string           `json:"teamId"`
	DisplayName     string           `json:"displayName"`
	Description     string           `json:"description"`
	TeamType        string           `json:"teamType"`
	OrganizationId  string           `json:"organizationId"`
	CreatorId       string           `json:"creatorId,omitempty"`
	State           string           `json:"state"`
	UserPermissions *UserPermissions `json:"userPermissions"`
}

// TeamResponseWithMembers represents team with members (matches PublicApiTeamResponseWithMembers)
type TeamResponseWithMembers struct {
	TeamID          string           `json:"teamId"`
	DisplayName     string           `json:"displayName"`
	Description     string           `json:"description"`
	TeamType        string           `json:"teamType"`
	OrganizationId  string           `json:"organizationId"`
	CreatorId       string           `json:"creatorId,omitempty"`
	State           string           `json:"state"`
	Members         []TeamMember     `json:"members"`
	UserPermissions *UserPermissions `json:"userPermissions"`
}

// UserPermissions represents team permissions
type UserPermissions struct {
	AddMembers    bool `json:"ADD_MEMBERS"`
	DeleteTeam    bool `json:"DELETE_TEAM"`
	RemoveMembers bool `json:"REMOVE_MEMBERS"`
	UpdateTeam    bool `json:"UPDATE_TEAM"`
}

// TeamMember represents a team member (matches PublicApiMembership)
type TeamMember struct {
	AccountID string `json:"accountId"`
}

// CreateTeamRequest represents the request to create a team (matches PublicApiTeamCreationPayload)
type CreateTeamRequest struct {
	DisplayName string `json:"displayName"`      // maxLength: 250, minLength: 1
	Description string `json:"description"`      // maxLength: 360, minLength: 0
	TeamType    string `json:"teamType"`         // OPEN, MEMBER_INVITE, EXTERNAL, ORG_ADMIN_MANAGED
	SiteId      string `json:"siteId,omitempty"` // maxLength: 255, minLength: 1
}

// UpdateTeamRequest represents the request to update a team (matches PublicApiTeamUpdatePayload)
type UpdateTeamRequest struct {
	DisplayName string `json:"displayName,omitempty"` // maxLength: 250, minLength: 1, pattern: .*\\S+.*
	Description string `json:"description,omitempty"` // maxLength: 360, minLength: 0
}

// NewAtlassianClient creates a new Atlassian API client
func NewAtlassianClient(apiToken, email, organization, siteId, orgId, baseURL string) (*AtlassianClient, error) {
	return &AtlassianClient{
		APIToken:     apiToken,
		Email:        email,
		Organization: organization,
		SiteId:       siteId,
		OrgId:        orgId,
		BaseURL:      baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// makeRequest makes an HTTP request to the Atlassian API
func (c *AtlassianClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	return c.makeRequestWithHeaders(method, path, body, nil)
}

// makeRequestWithHeaders makes an HTTP request with custom headers
func (c *AtlassianClient) makeRequestWithHeaders(method, path string, body interface{}, customHeaders map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	fullURL := c.BaseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set authentication headers - use Bearer token for Teams API
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	// Set default Accept header unless custom header provided
	if customHeaders == nil || customHeaders["Accept"] == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Set custom headers
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

// CreateTeam creates a new team in Atlassian
func (c *AtlassianClient) CreateTeam(team *CreateTeamRequest) (*TeamResponseWithMembers, error) {
	resp, err := c.makeRequest("POST", c.getTeamAPIPath("/teams/"), team)
	if err != nil {
		return nil, fmt.Errorf("error creating team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error creating team: %s - %s", resp.Status, string(body))
	}

	var createdTeam TeamResponseWithMembers
	if err := json.NewDecoder(resp.Body).Decode(&createdTeam); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &createdTeam, nil
}

// GetTeam retrieves a team by ID
func (c *AtlassianClient) GetTeam(teamID string) (*TeamResponse, error) {
	path := c.getTeamAPIPathWithQuery("/teams/"+teamID, c.SiteId)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error getting team: %s - %s", resp.Status, string(body))
	}

	var team TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func (c *AtlassianClient) UpdateTeam(teamID string, updateReq *UpdateTeamRequest) (*TeamResponse, error) {
	resp, err := c.makeRequest("PATCH", c.getTeamAPIPath("/teams/"+teamID), updateReq)
	if err != nil {
		return nil, fmt.Errorf("error updating team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error updating team: %s - %s", resp.Status, string(body))
	}

	var updatedTeam TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&updatedTeam); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &updatedTeam, nil
}

// DeleteTeam deletes a team
func (c *AtlassianClient) DeleteTeam(teamID string) error {
	resp, err := c.makeRequest("DELETE", c.getTeamAPIPath("/teams/"+teamID), nil)
	if err != nil {
		return fmt.Errorf("error deleting team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Team already doesn't exist, consider it successful
		return nil
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error deleting team: %s - %s", resp.Status, string(body))
	}

	return nil
}

// getTeamAPIPath returns the appropriate API path for team operations
// This uses the organization ID for public team APIs
// getTeamAPIPath returns the appropriate API path for team operations
// This uses the organization ID for public team APIs
func (c *AtlassianClient) getTeamAPIPath(endpoint string) string {
	orgIdentifier := c.OrgId
	if orgIdentifier == "" {
		orgIdentifier = c.Organization
	}
	return fmt.Sprintf("/public/teams/v1/org/%s%s", orgIdentifier, endpoint)
}

// getTeamAPIPathWithQuery returns the API path with optional query parameters
func (c *AtlassianClient) getTeamAPIPathWithQuery(endpoint, siteId string) string {
	basePath := c.getTeamAPIPath(endpoint)
	if siteId != "" {
		basePath += "?siteId=" + siteId
	}
	return basePath
}
