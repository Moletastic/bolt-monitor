package probelocationcatalog

import "testing"

func TestCatalogValidateRejectsDuplicateIDs(t *testing.T) {
	catalog := Catalog{Locations: []Location{
		{LocationID: "iad", DisplayName: "US East", ExecutionTarget: "worker-us-east", Enabled: true},
		{LocationID: "iad", DisplayName: "Duplicate", ExecutionTarget: "worker-copy", Enabled: true},
	}}

	if err := catalog.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want duplicate locationId failure")
	}
}

func TestCatalogSelectableLocationRequiresEnabled(t *testing.T) {
	catalog := Catalog{Locations: []Location{
		{LocationID: "iad", DisplayName: "US East", ExecutionTarget: "worker-us-east", Enabled: true},
		{LocationID: "dub", DisplayName: "Dublin", ExecutionTarget: "worker-dub", Enabled: false},
	}}

	if !catalog.IsSelectableLocation("iad") {
		t.Fatal("iad should be selectable")
	}
	if catalog.IsSelectableLocation("dub") {
		t.Fatal("dub should not be selectable when disabled")
	}
}
