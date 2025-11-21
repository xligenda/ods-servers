package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *DiscordClient) FetchGuildRoles(guildID string) (map[string]string, error) {
	rolesURL := fmt.Sprintf("%s/guilds/%s/roles", API_URL, guildID)
	req, err := http.NewRequest("GET", rolesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rolesData []map[string]any
	if err := json.Unmarshal(body, &rolesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	nameToID := make(map[string]string)
	for _, role := range rolesData {
		id, idOk := role["id"].(string)
		name, nameOk := role["name"].(string)
		if idOk && nameOk {
			nameToID[name] = id
		}
	}

	return nameToID, nil
}

func (c *DiscordClient) FetchMemberRoles(guildID, userID string) ([]string, error) {
	url := fmt.Sprintf("%s/guilds/%s/members/%s", API_URL, guildID, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var memberData map[string]any
	if err := json.Unmarshal(body, &memberData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var roleIDs []string
	if roles, ok := memberData["roles"].([]any); ok {
		for _, roleID := range roles {
			if idStr, ok := roleID.(string); ok {
				roleIDs = append(roleIDs, idStr)
			}
		}
	}

	return roleIDs, nil
}
