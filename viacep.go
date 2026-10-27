package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errZipcodeNotFound = errors.New("zipcode not found")

type location struct {
	City  string
	State string
}

type viaCEPClient struct {
	baseURL string
	http    *http.Client
}

func (c *viaCEPClient) Lookup(ctx context.Context, cep string) (location, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ws/"+cep+"/json/", nil)
	if err != nil {
		return location{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return location{}, fmt.Errorf("viacep request: %w", err)
	}
	defer resp.Body.Close()

	// A well-formed but unknown CEP is answered with 200 and {"erro": true}.
	if resp.StatusCode != http.StatusOK {
		return location{}, fmt.Errorf("viacep status %d", resp.StatusCode)
	}

	var body struct {
		Erro       any    `json:"erro"`
		Localidade string `json:"localidade"`
		UF         string `json:"uf"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return location{}, fmt.Errorf("viacep decode: %w", err)
	}

	// "erro" arrives as the bool true or the string "true", depending on the API version.
	flagged := body.Erro != nil && body.Erro != false
	if flagged || body.Localidade == "" {
		return location{}, errZipcodeNotFound
	}
	return location{City: body.Localidade, State: body.UF}, nil
}
