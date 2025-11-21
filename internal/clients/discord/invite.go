package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type InviteCode struct {
	Code      string `json:"code"`
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
}

func (c *DiscordClient) FetchInvite(code string) (*InviteCode, error) {
	url := fmt.Sprintf("%s/invites/%s", API_URL, code)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		if strings.Contains(err.Error(), "10006") {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var res InviteCode
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &res, nil
}

// maxAge - time the invite is available, 0 is no limit;
// maxUsages - amount of uses available, 0 is no limit;
// temp - temporary member or not;
func (c *DiscordClient) CreateInvite(channel string, maxAge int, maxUsages int, temp bool) (*InviteCode, error) {
	url := fmt.Sprintf("%s/channels/%s/invites", "https://discord.com/api/v10", channel)

	bodyPayload := map[string]interface{}{
		"max_age":   maxAge,
		"max_uses":  maxUsages,
		"temporary": temp,
	}

	bodyBytes, err := json.Marshal(bodyPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var res InviteCode
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &res, nil
}
