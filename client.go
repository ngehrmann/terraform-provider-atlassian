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
	BaseURL      string
	HTTPClient   *http.Client
}

// Team represents an Atlassian Team
type Team struct {
	ID           string       `json:"id,omitempty"`
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Type         string       `json:"type"`
	Organization string       `json:"organization"`
	Members      []TeamMember `json:"members,omitempty"`
	CreatedAt    string       `json:"createdAt,omitempty"`
	UpdatedAt    string       `json:"updatedAt,omitempty"`
}

// TeamMember represents a team member
type TeamMember struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
}

// CreateTeamRequest represents the request to create a team
type CreateTeamRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type"`
	Organization string `json:"organization"`
}

// UpdateTeamRequest represents the request to update a team
type UpdateTeamRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// NewAtlassianClient creates a new Atlassian API client
func NewAtlassianClient(apiToken, email, organization, baseURL string) (*AtlassianClient, error) {
	return &AtlassianClient{
		APIToken:     apiToken,
		Email:        email,
		Organization: organization,
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

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set authentication headers
	req.SetBasicAuth(c.Email, c.APIToken)
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
	resp, err := c.makeRequest("POST", "/admin/v1/orgs/"+c.Organization+"/teams", team)
	if err != nil {
		return nil, fmt.Errorf("error creating team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
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
	resp, err := c.makeRequest("GET", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID, nil)
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
	resp, err := c.makeRequest("PUT", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID, updateReq)
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
	resp, err := c.makeRequest("DELETE", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID, nil)
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

// AddTeamMember adds a member to a team
func (c *AtlassianClient) AddTeamMember(teamID string, member *TeamMember) error {
	resp, err := c.makeRequest("POST", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID+"/members", member)
	if err != nil {
		return fmt.Errorf("error adding team member: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error adding team member: %s - %s", resp.Status, string(body))
	}

	return nil
}

// RemoveTeamMember removes a member from a team
func (c *AtlassianClient) RemoveTeamMember(teamID, accountID string) error {
	resp, err := c.makeRequest("DELETE", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID+"/members/"+accountID, nil)
	if err != nil {
		return fmt.Errorf("error removing team member: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Member already doesn't exist, consider it successful
		return nil
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error removing team member: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetTeamMembers retrieves all members of a team
func (c *AtlassianClient) GetTeamMembers(teamID string) ([]TeamMember, error) {
	resp, err := c.makeRequest("GET", "/admin/v1/orgs/"+c.Organization+"/teams/"+teamID+"/members", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting team members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error getting team members: %s - %s", resp.Status, string(body))
	}

	var members []TeamMember
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return members, nil
}
