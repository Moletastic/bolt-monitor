//go:build inline_channel_migration

package main

import (
	"context"
	"log"
	"os"
	"reflect"
	"strings"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/escalation"
)

// This build-tagged entrypoint keeps legacy data repair out of Lambda query
// paths. Run only with explicit operator intent:
// TABLE_NAME=... MIGRATE_INLINE_CHANNELS=yes go run -tags inline_channel_migration ./services/monitor-api
func main() {
	if os.Getenv("MIGRATE_INLINE_CHANNELS") != "yes" {
		log.Fatal("MIGRATE_INLINE_CHANNELS=yes is required")
	}
	tableName := strings.TrimSpace(os.Getenv("TABLE_NAME"))
	if tableName == "" {
		log.Fatal("TABLE_NAME is required")
	}
	tenantID := strings.TrimSpace(os.Getenv("TENANT_ID"))
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	client, err := sharedaws.NewDynamoDBAPI(context.Background())
	if err != nil {
		log.Fatalf("create dynamodb client: %v", err)
	}
	repository := newDynamoMonitorRepository(client, tableName)
	policies, err := repository.ListEscalationPolicies(context.Background(), tenantID)
	if err != nil {
		log.Fatalf("list escalation policies: %v", err)
	}

	migrated := 0
	for index := range policies {
		before := policies[index]
		if err := repository.MigrateRouteInlineChannels(context.Background(), &policies[index]); err != nil {
			log.Fatalf("migrate escalation policy %s: %v", before.PolicyID, err)
		}
		if !sameRouteChannels(before, policies[index]) {
			migrated++
		}
	}
	log.Printf("inline-channel migration complete: tenant=%s migrated=%d scanned=%d", tenantID, migrated, len(policies))
}

func sameRouteChannels(left, right escalation.EscalationPolicy) bool {
	return reflect.DeepEqual(left.BusinessHoursPath, right.BusinessHoursPath) && reflect.DeepEqual(left.OffHoursPath, right.OffHoursPath)
}
