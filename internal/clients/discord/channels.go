package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Channel struct {
	ID       string  `json:"id"`
	Type     int     `json:"type"`
	Name     string  `json:"name"`
	Position int     `json:"position"`
	ParentID *string `json:"parent_id"`
	GuildID  string  `json:"guild_id"`
}

func (c *DiscordClient) FetchGuildChannels(id string) (*[]Channel, error) {
	url := fmt.Sprintf("%s/guilds/%s/channels", API_URL, id)
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

	var res []Channel
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &res, nil
}
