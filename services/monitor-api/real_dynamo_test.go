//go:build real_dynamo
// +build real_dynamo

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/monitorconfig"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func TestCreateMonitorRealDynamo(t *testing.T) {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	repo := newDynamoMonitorRepository(sharedaws.NewDynamoDB(dynamodb.NewFromConfig(awsCfg)), "bolt-monitor-staging-AppTableTable-bdmchtew")

	// First verify we can get the service
	svc, found, err := repo.GetService(ctx, "DEFAULT", "delay-test")
	if err != nil {
		t.Fatalf("GetService error: %v", err)
	}
	if !found {
		t.Fatal("service not found")
	}
	fmt.Printf("Service found: %+v\n", svc)

	// Now try to create a monitor
	monitor := monitorconfig.Monitor{
		TenantID:        "DEFAULT",
		ServiceID:       "delay-test",
		MonitorID:       "investigate-check",
		Name:            "Investigate Check",
		Type:            monitorconfig.MonitorTypeHTTP,
		IntervalSeconds: 60,
		Enabled:         false,
		HTTP: &monitorconfig.HTTPConfiguration{
			Target:    "https://example.com",
			Method:    "GET",
			TimeoutMs: 5000,
		},
	}

	created, err := repo.CreateMonitor(ctx, monitor)
	if err != nil {
		t.Fatalf("CreateMonitor error: %v", err)
	}
	fmt.Printf("Created monitor: %+v\n", created)
}

func TestGetServiceRealDynamo(t *testing.T) {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	repo := newDynamoMonitorRepository(sharedaws.NewDynamoDB(dynamodb.NewFromConfig(awsCfg)), "bolt-monitor-staging-AppTableTable-bdmchtew")

	serviceID := "delay-test"
	service, found, err := repo.GetService(ctx, "DEFAULT", serviceID)
	if err != nil {
		t.Fatalf("GetService error: %v", err)
	}
	if !found {
		t.Fatalf("service %s not found", serviceID)
	}
	fmt.Printf("Service: %+v\n", service)

	monitors, err := repo.ListMonitors(ctx, "DEFAULT", serviceID)
	if err != nil {
		t.Fatalf("ListMonitors error: %v", err)
	}
	fmt.Printf("Monitors: %d\n", len(monitors))
}
