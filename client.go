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

// Team represents an Atlassian Team
type Team struct {
	TeamID          string           `json:"teamId,omitempty"`
	DisplayName     string           `json:"displayName"`
	Description     string           `json:"description,omitempty"`
	TeamType        string           `json:"teamType,omitempty"`
	OrganizationId  string           `json:"organizationId,omitempty"`
	CreatorId       string           `json:"creatorId,omitempty"`
	State           string           `json:"state,omitempty"`
	Members         []TeamMember     `json:"members,omitempty"`
	UserPermissions *UserPermissions `json:"userPermissions,omitempty"`
}

// UserPermissions represents team permissions
type UserPermissions struct {
	AddMembers    bool `json:"ADD_MEMBERS"`
	DeleteTeam    bool `json:"DELETE_TEAM"`
	RemoveMembers bool `json:"REMOVE_MEMBERS"`
	UpdateTeam    bool `json:"UPDATE_TEAM"`
}

// TeamMember represents a team member
type TeamMember struct {
	AccountID string `json:"accountId"`
}

// CreateTeamRequest represents the request to create a team
type CreateTeamRequest struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	TeamType    string `json:"teamType"`
	SiteId      string `json:"siteId,omitempty"`
}

// UpdateTeamRequest represents the request to update a team
type UpdateTeamRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
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
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

// CreateTeam creates a new team in Atlassian
func (c *AtlassianClient) CreateTeam(team *CreateTeamRequest) (*Team, error) {
	resp, err := c.makeRequest("POST", c.getTeamAPIPath("/teams/"), team)
	if err != nil {
		return nil, fmt.Errorf("error creating team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error creating team: %s - %s", resp.Status, string(body))
	}

	var createdTeam Team
	if err := json.NewDecoder(resp.Body).Decode(&createdTeam); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &createdTeam, nil
}

// GetTeam retrieves a team by ID
func (c *AtlassianClient) GetTeam(teamID string) (*Team, error) {
	resp, err := c.makeRequest("GET", c.getTeamAPIPath("/teams/"+teamID), nil)
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

	var team Team
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func (c *AtlassianClient) UpdateTeam(teamID string, updateReq *UpdateTeamRequest) (*Team, error) {
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

	var updatedTeam Team
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
func (c *AtlassianClient) getTeamAPIPath(endpoint string) string {
	orgIdentifier := c.OrgId
	if orgIdentifier == "" {
		orgIdentifier = c.Organization
	}
	return fmt.Sprintf("/public/teams/v1/org/%s%s", orgIdentifier, endpoint)
}
