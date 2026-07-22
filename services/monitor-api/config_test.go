package main

import "testing"

func TestNewMonitorAPIConfig(t *testing.T) {
	getenv := func(key string) string {
		return map[string]string{
			"TABLE_NAME":         "app-table",
			"AUTH_TABLE_NAME":    "auth-table",
			"COGNITO_CLIENT_IDS": "first, second ",
			"SST_STAGE":          "staging",
		}[key]
	}
	config, err := newMonitorAPIConfig(getenv)
	if err != nil {
		t.Fatalf("newMonitorAPIConfig: %v", err)
	}
	if config.TableName != "app-table" || config.AuthTableName != "auth-table" || len(config.CognitoClientIDs) != 2 || config.CognitoClientIDs[1] != "second" {
		t.Fatalf("config = %+v", config)
	}
}

func TestNewMonitorAPIConfigRequiresAuthInputs(t *testing.T) {
	for _, key := range []string{"TABLE_NAME", "AUTH_TABLE_NAME", "COGNITO_CLIENT_IDS"} {
		t.Run(key, func(t *testing.T) {
			getenv := func(name string) string {
				if name == key {
					return ""
				}
				return map[string]string{"TABLE_NAME": "app-table", "AUTH_TABLE_NAME": "auth-table", "COGNITO_CLIENT_IDS": "client-id"}[name]
			}
			if _, err := newMonitorAPIConfig(getenv); err == nil {
				t.Fatal("newMonitorAPIConfig error = nil")
			}
		})
	}
}

func TestNewProductionMonitorHandlerAssemblesDependenciesWithoutAWS(t *testing.T) {
	handler := newProductionMonitorHandler(&fakeDynamoClient{}, monitorAPIConfig{
		TableName: "app-table", AuthTableName: "auth-table", CognitoClientIDs: []string{"client-id"}, SecurityEventStage: "staging",
	})
	if handler.operations.ServiceStore == nil || handler.principalResolver == nil || handler.membershipResolver == nil || handler.senders == nil || handler.executor == nil || handler.validateDestination == nil {
		t.Fatal("production monitor handler has missing dependency")
	}
	if event := handler.newSecurityEvent("authorization.denied", "failure", "subject", "request-id"); event.Stage != "staging" {
		t.Fatalf("security event stage = %q, want staging", event.Stage)
	}
}
