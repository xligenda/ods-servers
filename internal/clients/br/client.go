package br

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
}

func NewCLient() *Client {
	return &Client{}
}

const API_URL = "https://blackrussia.online/api"

func (c *Client) Gameservers() ([]*Server, error) {
	resp, err := http.Get(API_URL + "/gameservers/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var servers []*Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, err
	}

	return servers, nil
}

func (c *Client) Techwork() (*Techwork, error) {
	resp, err := http.Get(API_URL + "/techwork/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var techwork *Techwork
	if err := json.NewDecoder(resp.Body).Decode(&techwork); err != nil {
		return nil, err
	}

	return techwork, nil
}

func (c *Client) Highlights() ([]*Highlight, error) {
	resp, err := http.Get(API_URL + "/highlights/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var highlights []*Highlight
	if err := json.NewDecoder(resp.Body).Decode(&highlights); err != nil {
		return nil, err
	}

	return highlights, nil
}

func (c *Client) News() ([]*News, error) {
	resp, err := http.Get(API_URL + "/news/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var news []*News
	if err := json.NewDecoder(resp.Body).Decode(&news); err != nil {
		return nil, err
	}

	return news, nil
}

func (c *Client) Categories() ([]*Category, error) {
	resp, err := http.Get(API_URL + "/categories/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var categories []*Category
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// https://blackrussia.online/api/donate/methods/
// https://blackrussia.online/api/donate/transaction/
// https://blackrussia.online/api/donate/
// POST https://blackrussia.online/api/register

// POST https://wiki.blackrussia.online/api/users/auth
