package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type fakeCognito struct {
	sharedaws.CognitoIdentityProviderAPI
	users      []sharedaws.CognitoUser
	created    []*sharedaws.CognitoAdminCreateUserInput
	getUser    sharedaws.CognitoUser
	listErr    error
	createErr  error
	createErrs []error
	getUserErr error
}

func (f *fakeCognito) ListUsers(_ context.Context, _ *sharedaws.CognitoListUsersInput) (*sharedaws.CognitoListUsersOutput, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &sharedaws.CognitoListUsersOutput{Users: f.users}, nil
}

func (f *fakeCognito) AdminCreateUser(_ context.Context, input *sharedaws.CognitoAdminCreateUserInput) (*sharedaws.CognitoAdminCreateUserOutput, error) {
	f.created = append(f.created, input)
	if len(f.createErrs) >= len(f.created) {
		if err := f.createErrs[len(f.created)-1]; err != nil {
			return nil, err
		}
	}
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &sharedaws.CognitoAdminCreateUserOutput{}, nil
}

func (f *fakeCognito) AdminGetUser(_ context.Context, _ *sharedaws.CognitoAdminGetUserInput) (*sharedaws.CognitoAdminGetUserOutput, error) {
	if f.getUserErr != nil {
		return nil, f.getUserErr
	}
	return &sharedaws.CognitoAdminGetUserOutput{Username: f.getUser.Username, UserStatus: f.getUser.UserStatus, UserAttributes: f.getUser.Attributes}, nil
}

type fakeDynamo struct {
	sharedaws.DynamoDBAPI
	item   map[string]sharedaws.AttributeValue
	put    *sharedaws.DynamoDBPutItemInput
	getErr error
	putErr error
}

func (f *fakeDynamo) GetItem(_ context.Context, _ *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &sharedaws.DynamoDBGetItemOutput{Item: f.item}, nil
}

func (f *fakeDynamo) PutItem(_ context.Context, input *sharedaws.DynamoDBPutItemInput) (*sharedaws.DynamoDBPutItemOutput, error) {
	f.put = input
	if f.putErr != nil {
		return nil, f.putErr
	}
	return &sharedaws.DynamoDBPutItemOutput{}, nil
}

func TestBootstrapCreatesMembershipBeforeSendingInvitation(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	cognito := &fakeCognito{getUser: user}
	dynamo := &fakeDynamo{}
	bootstrap := testBootstrapper(cognito, dynamo)

	if err := bootstrap.bootstrap(context.Background(), " OPERATOR@EXAMPLE.COM "); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 2 {
		t.Fatalf("AdminCreateUser calls = %d, want create then invitation resend", len(cognito.created))
	}
	if got := cognito.created[0].MessageAction; got != sharedaws.CognitoMessageActionSuppress {
		t.Fatalf("initial message action = %q, want SUPPRESS", got)
	}
	if got := cognito.created[0].Username; got == nil || *got != "operator@example.com" {
		t.Fatalf("initial username = %v, want normalized email", got)
	}
	if dynamo.put == nil {
		t.Fatal("membership was not created before invitation")
	}
	if got := cognito.created[1].MessageAction; got != sharedaws.CognitoMessageActionResend {
		t.Fatalf("invitation message action = %q, want RESEND", got)
	}
	var record membershipRecord
	if err := sharedaws.UnmarshalMap(dynamo.put.Item, &record); err != nil {
		t.Fatalf("unmarshal membership: %v", err)
	}
	if record.PK != "MEMBER#subject-1" || record.SK != auth.MembershipSK || record.TenantID != "DEFAULT" || record.Status != "ACTIVE" || record.Role != "ADMIN" || record.Version != 1 {
		t.Fatalf("membership record = %+v, want complete active default admin authority", record)
	}
	if record.MembershipID != "MEM_test" || record.AuthValidAfter != 100 || record.CreatedAt != "1970-01-01T00:01:40Z" || record.UpdatedAt != record.CreatedAt {
		t.Fatalf("membership immutable/versioned fields = %+v", record)
	}
}

func TestBootstrapDoesNotInviteEstablishedUser(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", "CONFIRMED")
	record := membershipRecord{PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatal(err)
	}
	cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
	dynamo := &fakeDynamo{item: item}

	if err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 0 || dynamo.put != nil {
		t.Fatalf("established identity changed: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapRetryPreservesPendingCredentialsAndMembership(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	record := membershipRecord{PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 200, Version: 3, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:02Z"}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatal(err)
	}
	cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
	dynamo := &fakeDynamo{item: item}

	if err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 0 || dynamo.put != nil {
		t.Fatalf("retry changed credentials or membership: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapRecoversConcurrentCognitoCreation(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	cognito := &fakeCognito{
		getUser:    user,
		createErrs: []error{&cognitotypes.UsernameExistsException{}},
	}
	dynamo := &fakeDynamo{}

	if err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 2 || dynamo.put == nil {
		t.Fatalf("concurrent creation did not reconcile membership and invitation: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapRejectsConflictingMembership(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", "CONFIRMED")
	for name, record := range map[string]membershipRecord{
		"tenant": {PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "OTHER", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"},
		"status": {PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "DISABLED", Role: "ADMIN", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"},
		"role":   {PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "VIEWER", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"},
		"shape":  {PK: "MEMBER#subject-1", SK: auth.MembershipSK, Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"},
	} {
		t.Run(name, func(t *testing.T) {
			item, err := sharedaws.MarshalMap(record)
			if err != nil {
				t.Fatal(err)
			}
			cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
			dynamo := &fakeDynamo{item: item}

			if err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err == nil {
				t.Fatal("bootstrap accepted conflicting membership")
			}
			if len(cognito.created) != 0 || dynamo.put != nil {
				t.Fatalf("bootstrap overwrote conflicting membership: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
			}
		})
	}
}

func TestBootstrapRejectsAmbiguousUsersAndStorageFailures(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", "CONFIRMED")
	err := testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user, user}}, &fakeDynamo{}).bootstrap(context.Background(), "operator@example.com")
	if err == nil {
		t.Fatal("bootstrap accepted ambiguous users")
	}

	err = testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user}}, &fakeDynamo{getErr: errors.New("denied")}).bootstrap(context.Background(), "operator@example.com")
	if err == nil {
		t.Fatal("bootstrap accepted denied membership read")
	}
}

func testBootstrapper(cognito sharedaws.CognitoIdentityProviderAPI, dynamo sharedaws.DynamoDBAPI) bootstrapper {
	return bootstrapper{
		cognito: cognito, dynamo: dynamo, userPoolID: "pool", authTable: "auth", now: func() time.Time { return time.Unix(100, 0) },
		membershipID: func() (auth.MembershipID, error) { return "MEM_test", nil },
	}
}

func cognitoUser(email, subject string, status sharedaws.CognitoUserStatus) sharedaws.CognitoUser {
	return sharedaws.CognitoUser{Username: sharedaws.String(email), UserStatus: status, Attributes: []sharedaws.CognitoAttribute{
		{Name: sharedaws.String("email"), Value: sharedaws.String(email)},
		{Name: sharedaws.String("sub"), Value: sharedaws.String(subject)},
	}}
}
