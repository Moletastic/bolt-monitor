package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type fakeCognito struct {
	sharedaws.CognitoIdentityProviderAPI
	mu         sync.Mutex
	users      []sharedaws.CognitoUser
	created    []*sharedaws.CognitoAdminCreateUserInput
	events     *[]string
	getUser    sharedaws.CognitoUser
	listErr    error
	createErr  error
	createErrs []error
	getUserErr error
}

func (f *fakeCognito) ListUsers(_ context.Context, _ *sharedaws.CognitoListUsersInput) (*sharedaws.CognitoListUsersOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &sharedaws.CognitoListUsersOutput{Users: f.users}, nil
}

func (f *fakeCognito) AdminCreateUser(_ context.Context, input *sharedaws.CognitoAdminCreateUserInput) (*sharedaws.CognitoAdminCreateUserOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, input)
	if f.events != nil {
		*f.events = append(*f.events, "cognito:"+string(input.MessageAction))
	}
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
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getUserErr != nil {
		return nil, f.getUserErr
	}
	return &sharedaws.CognitoAdminGetUserOutput{Username: f.getUser.Username, UserStatus: f.getUser.UserStatus, UserAttributes: f.getUser.Attributes}, nil
}

type fakeDynamo struct {
	sharedaws.DynamoDBAPI
	mu      sync.Mutex
	item    map[string]sharedaws.AttributeValue
	put     *sharedaws.DynamoDBPutItemInput
	events  *[]string
	persist bool
	getErr  error
	putErr  error
}

func (f *fakeDynamo) GetItem(_ context.Context, _ *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &sharedaws.DynamoDBGetItemOutput{Item: f.item}, nil
}

func (f *fakeDynamo) PutItem(_ context.Context, input *sharedaws.DynamoDBPutItemInput) (*sharedaws.DynamoDBPutItemOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.put = input
	if f.events != nil {
		*f.events = append(*f.events, "dynamo:put")
	}
	if f.putErr != nil {
		return nil, f.putErr
	}
	if f.persist && len(f.item) != 0 {
		return nil, &ddbtypes.ConditionalCheckFailedException{}
	}
	if f.persist {
		f.item = input.Item
	}
	return &sharedaws.DynamoDBPutItemOutput{}, nil
}

func TestBootstrapFirstCreationCreatesCompleteMembershipBeforeInvitation(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	events := []string{}
	cognito := &fakeCognito{getUser: user, events: &events}
	dynamo := &fakeDynamo{events: &events}
	bootstrap := testBootstrapper(cognito, dynamo)

	if _, err := bootstrap.bootstrap(context.Background(), " OPERATOR@EXAMPLE.COM "); err != nil {
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
	if got := strings.Join(events, ","); got != "cognito:SUPPRESS,dynamo:put,cognito:RESEND" {
		t.Fatalf("operation order = %q, want suppressed creation, membership, invitation", got)
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

func TestBootstrapCompleteRetryLeavesEstablishedIdentityUntouched(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", "CONFIRMED")
	record := membershipRecord{PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 1, Version: 1, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:01Z"}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatal(err)
	}
	cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
	dynamo := &fakeDynamo{item: item}

	if _, err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 0 || dynamo.put != nil {
		t.Fatalf("established identity changed: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapPreservesImmutableMembershipAndPendingCredentials(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	record := membershipRecord{PK: "MEMBER#subject-1", SK: auth.MembershipSK, MembershipID: "MEM_existing", Subject: "subject-1", TenantID: "DEFAULT", Status: "ACTIVE", Role: "ADMIN", AuthValidAfter: 200, Version: 3, CreatedAt: "1970-01-01T00:00:01Z", UpdatedAt: "1970-01-01T00:00:02Z"}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatal(err)
	}
	cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
	dynamo := &fakeDynamo{item: item}

	if _, err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 0 || dynamo.put != nil {
		t.Fatalf("retry changed credentials or membership: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapRecoversCognitoOnlyPartialCreation(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	cognito := &fakeCognito{
		getUser:    user,
		createErrs: []error{&cognitotypes.UsernameExistsException{}},
	}
	dynamo := &fakeDynamo{}

	if _, err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err != nil {
		t.Fatalf("bootstrap returned error: %v", err)
	}
	if len(cognito.created) != 2 || dynamo.put == nil {
		t.Fatalf("concurrent creation did not reconcile membership and invitation: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
	}
}

func TestBootstrapReconcilesConcurrentInvocations(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	cognito := &fakeCognito{users: []sharedaws.CognitoUser{user}}
	dynamo := &fakeDynamo{persist: true}
	bootstrap := testBootstrapper(cognito, dynamo)

	var group sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := bootstrap.bootstrap(context.Background(), "operator@example.com")
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent bootstrap returned error: %v", err)
		}
	}
	if len(cognito.created) != 1 || cognito.created[0].MessageAction != sharedaws.CognitoMessageActionResend {
		t.Fatalf("invitation calls = %+v, want exactly one resend", cognito.created)
	}
	if dynamo.item == nil {
		t.Fatal("concurrent bootstrap did not persist membership")
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

			if _, err := testBootstrapper(cognito, dynamo).bootstrap(context.Background(), "operator@example.com"); err == nil {
				t.Fatal("bootstrap accepted conflicting membership")
			}
			if len(cognito.created) != 0 || dynamo.put != nil {
				t.Fatalf("bootstrap overwrote conflicting membership: create calls=%d membershipPut=%v", len(cognito.created), dynamo.put != nil)
			}
		})
	}
}

func TestBootstrapRejectsAmbiguousUsers(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", "CONFIRMED")
	_, err := testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user, user}}, &fakeDynamo{}).bootstrap(context.Background(), "operator@example.com")
	if err == nil {
		t.Fatal("bootstrap accepted ambiguous users")
	}
}

func TestBootstrapRejectsDeniedAWSOperations(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	for name, bootstrap := range map[string]bootstrapper{
		"list users":        testBootstrapper(&fakeCognito{listErr: errors.New("denied")}, &fakeDynamo{}),
		"create user":       testBootstrapper(&fakeCognito{createErr: errors.New("denied")}, &fakeDynamo{}),
		"get created user":  testBootstrapper(&fakeCognito{getUserErr: errors.New("denied")}, &fakeDynamo{}),
		"read membership":   testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user}}, &fakeDynamo{getErr: errors.New("denied")}),
		"create membership": testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user}}, &fakeDynamo{putErr: errors.New("denied")}),
		"send invitation":   testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user}, createErrs: []error{errors.New("denied")}}, &fakeDynamo{}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := bootstrap.bootstrap(context.Background(), "operator@example.com"); err == nil {
				t.Fatal("bootstrap accepted denied AWS operation")
			}
		})
	}
}

func TestBootstrapOutcomeIncludesRequiredSafeContext(t *testing.T) {
	outcome := newBootstrapOutcome("staging", "arn:aws:iam::123456789012:role/operator", "subject-1", "correlation-1", nil)
	var output bytes.Buffer
	emitOutcomeOrFatal(&output, outcome)

	var decoded bootstrapOutcome
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("decode outcome: %v", err)
	}
	if decoded.Event != auth.EventBootstrapReconciled || decoded.Outcome != "success" || decoded.Stage != "staging" || decoded.ActingAWSPrincipal != "arn:aws:iam::123456789012:role/operator" || decoded.TargetSubject != "subject-1" || decoded.Correlation.ID != "correlation-1" {
		t.Fatalf("outcome = %+v, want structured bootstrap context", decoded)
	}
	if decoded.DesiredAuthority.TenantID != "DEFAULT" || decoded.DesiredAuthority.Status != "ACTIVE" || decoded.DesiredAuthority.Role != "ADMIN" {
		t.Fatalf("desired authority = %+v, want DEFAULT ACTIVE ADMIN", decoded.DesiredAuthority)
	}
}

func TestBootstrapOutcomeExcludesFailureAndCredentialValues(t *testing.T) {
	secret := "AKIAIOSFODNN7EXAMPLE"
	outcome := newBootstrapOutcome("staging", "", "subject-1", "correlation-1", errors.New(secret))
	encoded, err := json.Marshal(outcome)
	if err != nil {
		t.Fatalf("marshal outcome: %v", err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Fatalf("outcome exposed credential-like error value: %s", encoded)
	}
	if strings.Contains(string(encoded), "operator@example.com") {
		t.Fatalf("outcome exposed target email: %s", encoded)
	}
	if outcome.Outcome != "failure" {
		t.Fatalf("outcome = %q, want failure", outcome.Outcome)
	}
	if outcome.Component != "admin-bootstrap" || outcome.Operation != "bootstrap" || outcome.Events != 1 || outcome.EMF == nil {
		t.Fatalf("outcome = %+v, want bounded bootstrap failure metric", outcome)
	}
	metricSet := outcome.EMF.CloudWatchMetrics[0]
	if metricSet.Namespace != "BoltMonitor/Auth" || !reflect.DeepEqual(metricSet.Dimensions, [][]string{{"stage", "component", "operation", "outcome"}}) || !reflect.DeepEqual(metricSet.Metrics, []metric{{Name: "AuthenticationEvents", Unit: "Count"}}) {
		t.Fatalf("metric = %#v, want fixed auth dimensions and counter", metricSet)
	}
}

func TestBootstrapEmitsAuthorityChangesOnlyWhenCreatingMembership(t *testing.T) {
	user := cognitoUser("operator@example.com", "subject-1", sharedaws.CognitoUserStatusForceChangePassword)
	bootstrap := testBootstrapper(&fakeCognito{users: []sharedaws.CognitoUser{user}}, &fakeDynamo{})
	var events []auth.SecurityEvent
	if _, err := bootstrap.bootstrapWithEvents(context.Background(), "operator@example.com", func(event auth.SecurityEvent, _ auth.Subject) {
		events = append(events, event)
	}); err != nil {
		t.Fatalf("bootstrapWithEvents() error = %v", err)
	}
	if got, want := events, []auth.SecurityEvent{auth.EventMembershipStatusChanged, auth.EventAuthValidAfterAdvanced}; !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestNormalizeStageRejectsUnsafeValues(t *testing.T) {
	if _, err := normalizeStage("staging"); err != nil {
		t.Fatalf("normalize stage: %v", err)
	}
	if _, err := normalizeStage("secret value"); err == nil {
		t.Fatal("normalize stage accepted unsafe value")
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
