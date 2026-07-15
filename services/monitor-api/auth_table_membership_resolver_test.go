package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	sharederrors "bolt-monitor/shared/errors"
)

type fakeAuthTableDynamoClient struct {
	sharedaws.DynamoDBAPI
	getItemInput  *sharedaws.DynamoDBGetItemInput
	getItemOutput *sharedaws.DynamoDBGetItemOutput
	getItemErr    error
}

func (f *fakeAuthTableDynamoClient) GetItem(_ context.Context, input *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	f.getItemInput = input
	return f.getItemOutput, f.getItemErr
}

func TestAuthTableMembershipResolverStronglyReadsAndNormalizesPrincipal(t *testing.T) {
	client := &fakeAuthTableDynamoClient{getItemOutput: membershipGetItemOutput(t, validAuthMembershipRecord())}
	resolver := newAuthTableMembershipResolver(client, "authoritative-auth-table")

	principal, err := resolver.Resolve(context.Background(), auth.AuthenticatedIdentity{Subject: "operator-subject", AuthTime: 11})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if want := (auth.Principal{Subject: "operator-subject", TenantID: auth.DefaultTenantID, Role: auth.RoleAdmin}); principal != want {
		t.Fatalf("Resolve() principal = %#v, want %#v", principal, want)
	}
	if client.getItemInput == nil || sharedaws.ToString(client.getItemInput.TableName) != "authoritative-auth-table" {
		t.Fatalf("GetItem() table = %v, want authoritative AuthTable", client.getItemInput)
	}
	if client.getItemInput.ConsistentRead == nil || !*client.getItemInput.ConsistentRead {
		t.Fatalf("GetItem() ConsistentRead = %v, want true", client.getItemInput.ConsistentRead)
	}
	if got := client.getItemInput.Key["PK"].(*sharedaws.AttributeValueMemberS).Value; got != "MEMBER#operator-subject" {
		t.Fatalf("GetItem() PK = %q, want membership key", got)
	}
	if got := client.getItemInput.Key["SK"].(*sharedaws.AttributeValueMemberS).Value; got != auth.MembershipSK {
		t.Fatalf("GetItem() SK = %q, want %q", got, auth.MembershipSK)
	}
}

func TestAuthTableMembershipResolverDeniesInvalidAuthorityAndRevokedCeremonies(t *testing.T) {
	tests := []struct {
		name     string
		identity auth.AuthenticatedIdentity
		mutate   func(*authMembershipRecord)
	}{
		{name: "missing membership", mutate: func(record *authMembershipRecord) { record.MembershipID = "" }},
		{name: "conflicting subject", mutate: func(record *authMembershipRecord) { record.Subject = "other-subject" }},
		{name: "wrong tenant", mutate: func(record *authMembershipRecord) { record.TenantID = "OTHER" }},
		{name: "inactive status", mutate: func(record *authMembershipRecord) { record.Status = "DISABLED" }},
		{name: "unsupported role", mutate: func(record *authMembershipRecord) { record.Role = "VIEWER" }},
		{name: "invalid version", mutate: func(record *authMembershipRecord) { record.Version = 0 }},
		{name: "missing timestamp", mutate: func(record *authMembershipRecord) { record.UpdatedAt = "" }},
		{name: "auth time equals boundary", identity: auth.AuthenticatedIdentity{Subject: "operator-subject", AuthTime: 10}},
		{name: "auth time predates boundary", identity: auth.AuthenticatedIdentity{Subject: "operator-subject", AuthTime: 9}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := validAuthMembershipRecord()
			if test.mutate != nil {
				test.mutate(&record)
			}
			identity := test.identity
			if identity.Subject == "" {
				identity = auth.AuthenticatedIdentity{Subject: "operator-subject", AuthTime: 11}
			}
			client := &fakeAuthTableDynamoClient{getItemOutput: membershipGetItemOutput(t, record)}
			_, err := newAuthTableMembershipResolver(client, "auth-table").Resolve(context.Background(), identity)
			assertAuthorizationDenied(t, err)
		})
	}
}

func TestAuthTableMembershipResolverPreservesDynamoFailures(t *testing.T) {
	want := errors.New("dynamodb unavailable")
	client := &fakeAuthTableDynamoClient{getItemErr: want}

	_, err := newAuthTableMembershipResolver(client, "auth-table").Resolve(context.Background(), auth.AuthenticatedIdentity{Subject: "operator-subject", AuthTime: 11})
	if !errors.Is(err, want) {
		t.Fatalf("Resolve() error = %v, want DynamoDB error", err)
	}
}

func validAuthMembershipRecord() authMembershipRecord {
	createdAt := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	return authMembershipRecord{
		PK:             "MEMBER#operator-subject",
		SK:             auth.MembershipSK,
		MembershipID:   "membership-1",
		Subject:        "operator-subject",
		TenantID:       string(auth.DefaultTenantID),
		Status:         string(auth.MembershipStatusActive),
		Role:           string(auth.RoleAdmin),
		AuthValidAfter: 10,
		Version:        1,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
}

func membershipGetItemOutput(t *testing.T, record authMembershipRecord) *sharedaws.DynamoDBGetItemOutput {
	t.Helper()
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap() error = %v", err)
	}
	return &sharedaws.DynamoDBGetItemOutput{Item: item}
}

func assertAuthorizationDenied(t *testing.T, err error) {
	t.Helper()
	typed, ok := sharederrors.As(err)
	if !ok || typed.Code != sharederrors.CodeAuthorizationDenied {
		t.Fatalf("Resolve() error = %v, want AUTHORIZATION_DENIED", err)
	}
	if typed.Details != nil {
		t.Fatalf("Resolve() details = %#v, want no membership details", typed.Details)
	}
}
