package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// PublicApiBulkOperationRequest matches OpenAPI spec
type PublicApiBulkOperationRequest struct {
	TeamIDs []string `json:"teamIds"` // maxItems: 100, minItems: 1
}

// PublicApiBulkTeamOperationError matches OpenAPI spec
type PublicApiBulkTeamOperationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TeamID  string `json:"teamId"`
}

// PublicApiBulkOperationResponse matches OpenAPI spec
type PublicApiBulkOperationResponse struct {
	Errors            []PublicApiBulkTeamOperationError `json:"errors"`
	SuccessfulTeamIds []string                          `json:"successfulTeamIds"`
}

// PublicApiTeamPaginationResult matches OpenAPI spec
type PublicApiTeamPaginationResult struct {
	Cursor   string `json:"cursor,omitempty"` // The cursor for pagination
	Entities []Team `json:"entities"`         // The list of teams
}

// ArchiveTeams archives multiple teams in bulk
func (c *AtlassianClient) ArchiveTeams(orgID string, teamIDs []string) (*PublicApiBulkOperationResponse, error) {
	if len(teamIDs) == 0 || len(teamIDs) > 100 {
		return nil, fmt.Errorf("teamIDs must contain between 1 and 100 items, got %d", len(teamIDs))
	}

	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/archive", orgID)

	request := PublicApiBulkOperationRequest{
		TeamIDs: teamIDs,
	}

	resp, err := c.makeRequest("POST", path, request)
	if err != nil {
		return nil, fmt.Errorf("error archiving teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error archiving teams: %s - %s", resp.Status, string(body))
	}

	var archiveResponse PublicApiBulkOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&archiveResponse); err != nil {
		return nil, fmt.Errorf("error decoding archive response: %w", err)
	}

	return &archiveResponse, nil
}

// UnarchiveTeams unarchives multiple teams in bulk
func (c *AtlassianClient) UnarchiveTeams(orgID string, teamIDs []string) (*PublicApiBulkOperationResponse, error) {
	if len(teamIDs) == 0 || len(teamIDs) > 100 {
		return nil, fmt.Errorf("teamIDs must contain between 1 and 100 items, got %d", len(teamIDs))
	}

	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/unarchive", orgID)

	request := PublicApiBulkOperationRequest{
		TeamIDs: teamIDs,
	}

	resp, err := c.makeRequest("POST", path, request)
	if err != nil {
		return nil, fmt.Errorf("error unarchiving teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error unarchiving teams: %s - %s", resp.Status, string(body))
	}

	var unarchiveResponse PublicApiBulkOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&unarchiveResponse); err != nil {
		return nil, fmt.Errorf("error decoding unarchive response: %w", err)
	}

	return &unarchiveResponse, nil
}

// RestoreTeam restores a single soft-deleted team
func (c *AtlassianClient) RestoreTeam(orgID, teamID string) error {
	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/%s/restore", orgID, teamID)

	resp, err := c.makeRequest("POST", path, nil)
	if err != nil {
		return fmt.Errorf("error restoring team: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("team not found for restore: %s", teamID)
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error restoring team: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetTeams retrieves a list of teams for an organization with optional parameters
func (c *AtlassianClient) GetTeams(orgID, siteId string, size int32, cursor string) (*PublicApiTeamPaginationResult, error) {
	path := fmt.Sprintf("/public/teams/v1/org/%s/teams", orgID)

	// Build query parameters according to OpenAPI spec
	queryParams := make([]string, 0)

	if siteId != "" {
		queryParams = append(queryParams, "siteId="+siteId)
	}

	if size > 0 {
		if size > 300 {
			size = 300 // Maximum allowed by API
		}
		queryParams = append(queryParams, "size="+strconv.FormatInt(int64(size), 10))
	}

	if cursor != "" {
		queryParams = append(queryParams, "cursor="+cursor)
	}

	if len(queryParams) > 0 {
		path += "?" + strings.Join(queryParams, "&")
	}

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting teams list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error getting teams list: %s - %s", resp.Status, string(body))
	}

	var teamsResponse PublicApiTeamPaginationResult
	if err := json.NewDecoder(resp.Body).Decode(&teamsResponse); err != nil {
		return nil, fmt.Errorf("error decoding teams list response: %w", err)
	}

	return &teamsResponse, nil
}
