package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PublicApiMembershipFetchPayload matches OpenAPI spec
type PublicApiMembershipFetchPayload struct {
	After string `json:"after,omitempty"` // Pagination cursor
	First int32  `json:"first,omitempty"` // Maximum 50, default 50
}

// PublicApiPageInfoAccountId matches OpenAPI spec
type PublicApiPageInfoAccountId struct {
	EndCursor   string `json:"endCursor,omitempty"`
	HasNextPage bool   `json:"hasNextPage"`
}

// PublicApiFetchResponsePublicApiMembershipAccountId matches OpenAPI spec
type PublicApiFetchResponsePublicApiMembershipAccountId struct {
	PageInfo PublicApiPageInfoAccountId `json:"pageInfo"`
	Results  []TeamMember               `json:"results"`
}

// PublicApiMembershipAddPayload matches OpenAPI spec
type PublicApiMembershipAddPayload struct {
	Members []TeamMember `json:"members"` // maxItems: 50, minItems: 1, uniqueItems: true
}

// PublicApiMembershipCodedError matches OpenAPI spec
type PublicApiMembershipCodedError struct {
	AccountID string `json:"accountId"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

// PublicApiMembershipAddResponse matches OpenAPI spec
type PublicApiMembershipAddResponse struct {
	Errors  []PublicApiMembershipCodedError `json:"errors"`  // uniqueItems: true
	Members []TeamMember                    `json:"members"` // uniqueItems: true
}

// PublicApiMembershipRemovePayload matches OpenAPI spec
type PublicApiMembershipRemovePayload struct {
	Members []TeamMember `json:"members"` // maxItems: 50, minItems: 1, uniqueItems: true
}

// PublicApiMembershipRemoveResponse matches OpenAPI spec
type PublicApiMembershipRemoveResponse struct {
	Errors []PublicApiMembershipCodedError `json:"errors"` // uniqueItems: true
}

// FetchTeamMembers retrieves team members with optional siteId and pagination
func (c *AtlassianClient) FetchTeamMembers(orgID, teamID, siteId, after string, first int32) (*PublicApiFetchResponsePublicApiMembershipAccountId, error) {
	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/%s/members", orgID, teamID)
	if siteId != "" {
		path += "?siteId=" + siteId
	}

	// Create request body with pagination parameters
	payload := PublicApiMembershipFetchPayload{}
	if after != "" {
		payload.After = after
	}
	if first > 0 && first <= 50 {
		payload.First = first
	}

	// Set Accept header to */* as per OpenAPI spec
	resp, err := c.makeRequestWithHeaders("POST", path, payload, map[string]string{
		"Accept": "*/*",
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching team members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error fetching team members: %s - %s", resp.Status, string(body))
	}

	var membersResponse PublicApiFetchResponsePublicApiMembershipAccountId
	if err := json.NewDecoder(resp.Body).Decode(&membersResponse); err != nil {
		return nil, fmt.Errorf("error decoding team members response: %w", err)
	}

	return &membersResponse, nil
}

// AddTeamMembers adds members to a team
func (c *AtlassianClient) AddTeamMembers(orgID, teamID string, members []TeamMember) (*PublicApiMembershipAddResponse, error) {
	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/%s/members/add", orgID, teamID)

	request := PublicApiMembershipAddPayload{
		Members: members,
	}

	resp, err := c.makeRequestWithHeaders("POST", path, request, map[string]string{
		"Accept": "*/*",
	})
	if err != nil {
		return nil, fmt.Errorf("error adding team members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error adding team members: %s - %s", resp.Status, string(body))
	}

	var addResponse PublicApiMembershipAddResponse
	if err := json.NewDecoder(resp.Body).Decode(&addResponse); err != nil {
		return nil, fmt.Errorf("error decoding add members response: %w", err)
	}

	return &addResponse, nil
}

// RemoveTeamMembers removes members from a team
func (c *AtlassianClient) RemoveTeamMembers(orgID, teamID string, members []TeamMember) (*PublicApiMembershipRemoveResponse, error) {
	path := fmt.Sprintf("/public/teams/v1/org/%s/teams/%s/members/remove", orgID, teamID)

	request := PublicApiMembershipRemovePayload{
		Members: members,
	}

	resp, err := c.makeRequestWithHeaders("POST", path, request, map[string]string{
		"Accept": "*/*",
	})
	if err != nil {
		return nil, fmt.Errorf("error removing team members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error removing team members: %s - %s", resp.Status, string(body))
	}

	var removeResponse PublicApiMembershipRemoveResponse
	if err := json.NewDecoder(resp.Body).Decode(&removeResponse); err != nil {
		return nil, fmt.Errorf("error decoding remove members response: %w", err)
	}

	return &removeResponse, nil
}
