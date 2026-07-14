package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	sharedaws "bolt-monitor/shared/aws"
)

type historyCursor struct {
	Resource string            `json:"resource"`
	Key      map[string]string `json:"key"`
}

func encodeHistoryCursor(resource string, key map[string]sharedaws.AttributeValue) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	values := make(map[string]string, len(key))
	for name, value := range key {
		stringValue, ok := value.(*sharedaws.AttributeValueMemberS)
		if !ok {
			return "", fmt.Errorf("cursor key %q is not a string", name)
		}
		values[name] = stringValue.Value
	}
	encoded, err := json.Marshal(historyCursor{Resource: resource, Key: values})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeHistoryCursor(raw, resource, resourceKeyName string) (map[string]sharedaws.AttributeValue, error) {
	if raw == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var cursor historyCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, fmt.Errorf("parse cursor: %w", err)
	}
	if cursor.Resource != resource || cursor.Key[resourceKeyName] != resource || len(cursor.Key) == 0 {
		return nil, fmt.Errorf("cursor does not match requested resource")
	}
	key := make(map[string]sharedaws.AttributeValue, len(cursor.Key))
	for name, value := range cursor.Key {
		if value == "" {
			return nil, fmt.Errorf("cursor key %q is empty", name)
		}
		key[name] = &sharedaws.AttributeValueMemberS{Value: value}
	}
	return key, nil
}
