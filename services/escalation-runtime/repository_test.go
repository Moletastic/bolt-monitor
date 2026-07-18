package main

import (
	"context"
	"testing"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
)

type fakeDynamoClient struct {
	items         map[string]map[string]sharedaws.AttributeValue
	put           map[string]sharedaws.AttributeValue
	transactItems []sharedaws.TransactWriteItem
}

func (f *fakeDynamoClient) GetItem(_ context.Context, input *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	pk := input.Key["PK"].(*sharedaws.AttributeValueMemberS).Value
	sk := input.Key["SK"].(*sharedaws.AttributeValueMemberS).Value
	if item, ok := f.items[pk+"|"+sk]; ok {
		return &sharedaws.DynamoDBGetItemOutput{Item: item}, nil
	}
	return &sharedaws.DynamoDBGetItemOutput{}, nil
}

func (f *fakeDynamoClient) PutItem(_ context.Context, input *sharedaws.DynamoDBPutItemInput) (*sharedaws.DynamoDBPutItemOutput, error) {
	f.put = input.Item
	return &sharedaws.DynamoDBPutItemOutput{}, nil
}

func (f *fakeDynamoClient) TransactWriteItems(_ context.Context, input *sharedaws.DynamoDBTransactWriteItemsInput) (*sharedaws.DynamoDBTransactWriteItemsOutput, error) {
	f.transactItems = input.TransactItems
	return &sharedaws.DynamoDBTransactWriteItemsOutput{}, nil
}

func (f *fakeDynamoClient) Query(context.Context, *sharedaws.DynamoDBQueryInput) (*sharedaws.DynamoDBQueryOutput, error) {
	return &sharedaws.DynamoDBQueryOutput{}, nil
}

func (f *fakeDynamoClient) DeleteItem(context.Context, *sharedaws.DynamoDBDeleteItemInput) (*sharedaws.DynamoDBDeleteItemOutput, error) {
	return &sharedaws.DynamoDBDeleteItemOutput{}, nil
}

func (f *fakeDynamoClient) Scan(context.Context, *sharedaws.DynamoDBScanInput) (*sharedaws.DynamoDBScanOutput, error) {
	return &sharedaws.DynamoDBScanOutput{}, nil
}

func (f *fakeDynamoClient) UpdateItem(context.Context, *sharedaws.DynamoDBUpdateItemInput) (*sharedaws.DynamoDBUpdateItemOutput, error) {
	return &sharedaws.DynamoDBUpdateItemOutput{}, nil
}

func TestPutAndGetEscalationState(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoEscalationRepository(client, "table-name")
	state := escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 1, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"}
	if err := repo.PutEscalationState(context.Background(), state); err != nil {
		t.Fatalf("PutEscalationState returned error: %v", err)
	}
	pk := client.put["PK"].(*sharedaws.AttributeValueMemberS).Value
	sk := client.put["SK"].(*sharedaws.AttributeValueMemberS).Value
	client.items[pk+"|"+sk] = client.put
	loaded, err := repo.GetEscalationState(context.Background(), "DEFAULT", "INC_1")
	if err != nil {
		t.Fatalf("GetEscalationState returned error: %v", err)
	}
	if loaded == nil || loaded.PolicyID != "POL_1" {
		t.Fatalf("loaded = %+v, want policy POL_1", loaded)
	}
}

func TestGetServiceAndPolicy(t *testing.T) {
	serviceItem, _ := sharedaws.MarshalMap(serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{1}}})
	policyItem, _ := sharedaws.MarshalMap(dynamodbrecord.EscalationPolicyItemRecord{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary"})
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{"SERVICE#DEFAULT#AUTH|META": serviceItem, "TENANT#DEFAULT|ESCALATION_POLICY#POL_1": policyItem}}
	repo := newDynamoEscalationRepository(client, "table-name")
	service, err := repo.GetService(context.Background(), "DEFAULT", "auth")
	if err != nil {
		t.Fatalf("GetService returned error: %v", err)
	}
	if service == nil || service.EscalationPolicyID != "POL_1" {
		t.Fatalf("service = %+v", service)
	}
	policy, err := repo.GetEscalationPolicy(context.Background(), "DEFAULT", "POL_1")
	if err != nil {
		t.Fatalf("GetEscalationPolicy returned error: %v", err)
	}
	if policy == nil || policy.Name != "Primary" {
		t.Fatalf("policy = %+v", policy)
	}
}

func TestCreateAndGetIncident(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoEscalationRepository(client, "table-name")
	incident := incidentRecord{TenantID: "DEFAULT", ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Type: "escalation.exhausted", Summary: "Escalation exhausted", Status: incidentStatusOpen, OpenedAt: "2026-06-16T10:00:00Z", UpdatedAt: "2026-06-16T10:00:00Z", OriginalIncidentID: "INC_ORIG"}
	if err := repo.CreateIncident(context.Background(), incident); err != nil {
		t.Fatalf("CreateIncident returned error: %v", err)
	}
	if len(client.transactItems) != 3 {
		t.Fatalf("transact items = %d, want 3", len(client.transactItems))
	}
	for _, item := range client.transactItems {
		put := item.Put.Item
		pk := put["PK"].(*sharedaws.AttributeValueMemberS).Value
		sk := put["SK"].(*sharedaws.AttributeValueMemberS).Value
		client.items[pk+"|"+sk] = put
	}
	loaded, err := repo.GetIncident(context.Background(), "INC_1")
	if err != nil {
		t.Fatalf("GetIncident returned error: %v", err)
	}
	if loaded == nil || loaded.OriginalIncidentID != "INC_ORIG" {
		t.Fatalf("incident = %+v, want original incident link", loaded)
	}
}
