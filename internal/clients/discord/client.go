package discord

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const API_URL = "https://discord.com/api/v10"

type DiscordClient struct {
	headers      map[string]string
	client       *http.Client
	globalMu     sync.Mutex
	globalWait   time.Time
	globalLimit  int
	globalTokens int
	maxRetries   int
}

func NewDiscordClient(token string) *DiscordClient {
	return &DiscordClient{
		headers: map[string]string{
			"Authorization": token,
			"User-Agent":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"Content-Type":  "application/json",
		},
		client:       &http.Client{Timeout: 5 * time.Second},
		globalLimit:  45,
		globalTokens: 45,
		maxRetries:   3,
	}
}

func (c *DiscordClient) doRequest(req *http.Request) (*http.Response, error) {
	for retry := range c.maxRetries {
		c.globalMu.Lock()
		if c.globalTokens <= 0 && time.Now().Before(c.globalWait) {
			wait := time.Until(c.globalWait)
			c.globalMu.Unlock()
			time.Sleep(wait)
			continue
		}
		c.globalTokens--
		if c.globalTokens == c.globalLimit-1 {
			c.globalWait = time.Now().Add(time.Second)
		}
		c.globalMu.Unlock()

		resp, err := c.client.Do(req)
		if err != nil {
			c.globalMu.Lock()
			c.globalTokens++
			c.globalMu.Unlock()
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter, err := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
			if err != nil || retryAfter == float64(0) {
				retryAfter = 1
			}
			resp.Body.Close()

			c.globalMu.Lock()
			c.globalTokens++
			waitDuration := time.Duration(retryAfter * float64(time.Second) * float64(retry))
			if time.Now().Add(waitDuration).After(c.globalWait) {
				c.globalWait = time.Now().Add(waitDuration)
			}
			c.globalMu.Unlock()

			time.Sleep(waitDuration)
			continue
		}

		if remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining")); err == nil {
			c.globalMu.Lock()
			if remaining > c.globalTokens {
				c.globalTokens = remaining
			}
			c.globalMu.Unlock()
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}

		return resp, nil
	}

	return nil, fmt.Errorf("exceeded maximum retries (%d) due to rate limits", c.maxRetries)
}
