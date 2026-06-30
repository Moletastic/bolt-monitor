package probelocationcatalog

import (
	"fmt"
	"sort"
	"strings"
)

type Location struct {
	LocationID      string `json:"locationId"`
	DisplayName     string `json:"displayName"`
	ExecutionTarget string `json:"executionTarget"`
	Enabled         bool   `json:"enabled"`
}

type Catalog struct {
	Locations []Location `json:"locations"`
}

func (l Location) Validate() error {
	if strings.TrimSpace(l.LocationID) == "" {
		return fmt.Errorf("locationId is required")
	}
	if strings.TrimSpace(l.DisplayName) == "" {
		return fmt.Errorf("displayName is required")
	}
	if strings.TrimSpace(l.ExecutionTarget) == "" {
		return fmt.Errorf("executionTarget is required")
	}
	return nil
}

func (c Catalog) Validate() error {
	if len(c.Locations) == 0 {
		return fmt.Errorf("catalog must contain at least one location")
	}
	seen := make(map[string]struct{}, len(c.Locations))
	for _, location := range c.Locations {
		if err := location.Validate(); err != nil {
			return err
		}
		if _, exists := seen[location.LocationID]; exists {
			return fmt.Errorf("duplicate locationId %q", location.LocationID)
		}
		seen[location.LocationID] = struct{}{}
	}
	return nil
}

func (c Catalog) IsSelectableLocation(locationID string) bool {
	for _, location := range c.Locations {
		if location.LocationID == locationID && location.Enabled {
			return true
		}
	}
	return false
}

func (c Catalog) IDs() []string {
	ids := make([]string, 0, len(c.Locations))
	for _, location := range c.Locations {
		ids = append(ids, location.LocationID)
	}
	sort.Strings(ids)
	return ids
}
