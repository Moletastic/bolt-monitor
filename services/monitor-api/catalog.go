package main

import "bolt-monitor/shared/probelocationcatalog"

func defaultProbeLocationCatalog() probelocationcatalog.Catalog {
	return probelocationcatalog.Catalog{
		Locations: []probelocationcatalog.Location{
			{
				LocationID:      "iad",
				DisplayName:     "US East",
				ExecutionTarget: "worker-us-east",
				Enabled:         true,
			},
		},
	}
}
