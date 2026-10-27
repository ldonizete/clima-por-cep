package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var errCityNotFound = errors.New("city not found")

// weatherAPI error code for "No matching location found".
const weatherNoLocationCode = 1006

type weatherClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// CurrentC returns the current temperature in Celsius for the given city.
func (c *weatherClient) CurrentC(ctx context.Context, loc location) (float64, error) {
	q := url.Values{}
	q.Set("key", c.apiKey)
	q.Set("q", loc.City+", "+loc.State+", Brazil")
	q.Set("aqi", "no")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/current.json?"+q.Encode(), nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("weatherapi request: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Current struct {
			TempC *float64 `json:"temp_c"`
		} `json:"current"`
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&body)

	if resp.StatusCode == http.StatusBadRequest && body.Error.Code == weatherNoLocationCode {
		return 0, errCityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi status %d (code %d)", resp.StatusCode, body.Error.Code)
	}
	if decodeErr != nil {
		return 0, fmt.Errorf("weatherapi decode: %w", decodeErr)
	}
	if body.Current.TempC == nil {
		return 0, errors.New("weatherapi: response without temp_c")
	}
	return *body.Current.TempC, nil
}
