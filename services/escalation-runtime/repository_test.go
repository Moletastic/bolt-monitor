package main

import (
	"context"
	"testing"

	"bolt-monitor/shared/escalation"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type fakeDynamoClient struct {
	items         map[string]map[string]ddbtypes.AttributeValue
	put           map[string]ddbtypes.AttributeValue
	transactItems []ddbtypes.TransactWriteItem
}

func (f *fakeDynamoClient) GetItem(_ context.Context, input *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	pk := input.Key["PK"].(*ddbtypes.AttributeValueMemberS).Value
	sk := input.Key["SK"].(*ddbtypes.AttributeValueMemberS).Value
	if item, ok := f.items[pk+"|"+sk]; ok {
		return &dynamodb.GetItemOutput{Item: item}, nil
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (f *fakeDynamoClient) PutItem(_ context.Context, input *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	f.put = input.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamoClient) TransactWriteItems(_ context.Context, input *dynamodb.TransactWriteItemsInput, _ ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	f.transactItems = input.TransactItems
	return &dynamodb.TransactWriteItemsOutput{}, nil
}

func (f *fakeDynamoClient) Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, nil
}

func (f *fakeDynamoClient) DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
}

func (f *fakeDynamoClient) Scan(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{}, nil
}

func TestPutAndGetEscalationState(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]ddbtypes.AttributeValue{}}
	repo := newDynamoEscalationRepository(client, "table-name")
	state := escalation.EscalationState{TenantID: "DEFAULT", IncidentID: "INC_1", PolicyID: "POL_1", ServiceID: "auth", MonitorID: "public-http", CurrentStep: 1, SelectedPath: pathBusinessHours, Status: escalation.EscalationStatusActive, CreatedAt: "2026-06-16T00:00:00Z", UpdatedAt: "2026-06-16T00:00:00Z"}
	if err := repo.PutEscalationState(context.Background(), state); err != nil {
		t.Fatalf("PutEscalationState returned error: %v", err)
	}
	pk := client.put["PK"].(*ddbtypes.AttributeValueMemberS).Value
	sk := client.put["SK"].(*ddbtypes.AttributeValueMemberS).Value
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
	serviceItem, _ := attributevalue.MarshalMap(serviceRecord{TenantID: "DEFAULT", ServiceID: "auth", EscalationPolicyID: "POL_1", BusinessHours: &escalation.BusinessHoursConfig{Timezone: "UTC", StartHour: 9, EndHour: 17, DaysOfWeek: []int{1}}})
	policyItem, _ := attributevalue.MarshalMap(escalationPolicyRecord{TenantID: "DEFAULT", PolicyID: "POL_1", Name: "Primary"})
	client := &fakeDynamoClient{items: map[string]map[string]ddbtypes.AttributeValue{"SERVICE#DEFAULT#AUTH|META": serviceItem, "TENANT#DEFAULT|ESCALATION_POLICY#POL_1": policyItem}}
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
	client := &fakeDynamoClient{items: map[string]map[string]ddbtypes.AttributeValue{}}
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
		pk := put["PK"].(*ddbtypes.AttributeValueMemberS).Value
		sk := put["SK"].(*ddbtypes.AttributeValueMemberS).Value
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
