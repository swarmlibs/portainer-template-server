package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type PortainerAppTemplateScheme struct {
	Version   string           `json:"version"`
	Templates []map[string]any `json:"templates"`
}

type PortainerAppTemplate struct {
	Url    string
	Scheme PortainerAppTemplateScheme
}

func (t *PortainerAppTemplate) FetchTemplate() error {
	// fetch template from URL
	resp, err := http.Get(t.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch template")
	}

	if err := json.NewDecoder(resp.Body).Decode(&t.Scheme); err != nil {
		return err
	}

	return nil
}
