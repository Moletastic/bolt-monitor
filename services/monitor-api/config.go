package main

import (
	"fmt"
	"strings"
)

type monitorAPIConfig struct {
	TableName          string
	AuthTableName      string
	CognitoClientIDs   []string
	SecurityEventStage string
}

func newMonitorAPIConfig(getenv func(string) string) (monitorAPIConfig, error) {
	config := monitorAPIConfig{
		TableName:          strings.TrimSpace(getenv("TABLE_NAME")),
		AuthTableName:      strings.TrimSpace(getenv("AUTH_TABLE_NAME")),
		SecurityEventStage: strings.TrimSpace(getenv("SST_STAGE")),
	}
	for _, clientID := range strings.Split(getenv("COGNITO_CLIENT_IDS"), ",") {
		if clientID = strings.TrimSpace(clientID); clientID != "" {
			config.CognitoClientIDs = append(config.CognitoClientIDs, clientID)
		}
	}
	if config.TableName == "" {
		return monitorAPIConfig{}, fmt.Errorf("TABLE_NAME is required")
	}
	if config.AuthTableName == "" {
		return monitorAPIConfig{}, fmt.Errorf("AUTH_TABLE_NAME is required")
	}
	if len(config.CognitoClientIDs) == 0 {
		return monitorAPIConfig{}, fmt.Errorf("COGNITO_CLIENT_IDS is required")
	}
	return config, nil
}
