package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type bootstrapper struct {
	cognito      sharedaws.CognitoIdentityProviderAPI
	dynamo       sharedaws.DynamoDBAPI
	userPoolID   string
	authTable    string
	now          func() time.Time
	membershipID func() (auth.MembershipID, error)
}

type membershipRecord struct {
	PK             string `dynamodbav:"PK"`
	SK             string `dynamodbav:"SK"`
	MembershipID   string `dynamodbav:"MembershipID"`
	Subject        string `dynamodbav:"Subject"`
	TenantID       string `dynamodbav:"TenantID"`
	Status         string `dynamodbav:"Status"`
	Role           string `dynamodbav:"Role"`
	AuthValidAfter int64  `dynamodbav:"AuthValidAfter"`
	Version        int64  `dynamodbav:"Version"`
	CreatedAt      string `dynamodbav:"CreatedAt"`
	UpdatedAt      string `dynamodbav:"UpdatedAt"`
}

func (b bootstrapper) bootstrap(ctx context.Context, email string) (auth.Subject, error) {
	return b.bootstrapWithEvents(ctx, email, nil)
}

func (b bootstrapper) bootstrapWithEvents(ctx context.Context, email string, emit func(auth.SecurityEvent, auth.Subject)) (auth.Subject, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return "", err
	}

	user, err := b.findUser(ctx, normalizedEmail)
	if err != nil {
		return "", err
	}
	if user == nil {
		if _, err := b.cognito.AdminCreateUser(ctx, &sharedaws.CognitoAdminCreateUserInput{
			UserPoolId:             sharedaws.String(b.userPoolID),
			Username:               sharedaws.String(normalizedEmail),
			MessageAction:          sharedaws.CognitoMessageActionSuppress,
			DesiredDeliveryMediums: []sharedaws.CognitoDeliveryMedium{sharedaws.CognitoDeliveryMediumEmail},
			UserAttributes:         []sharedaws.CognitoAttribute{{Name: sharedaws.String("email"), Value: sharedaws.String(normalizedEmail)}},
		}); err != nil {
			var usernameExists *types.UsernameExistsException
			if !errors.As(err, &usernameExists) {
				return "", fmt.Errorf("create cognito user: %w", err)
			}
		}
		user, err = b.getUser(ctx, normalizedEmail)
		if err != nil {
			return "", err
		}
	}
	if cognitoEmail(*user) != normalizedEmail {
		return "", fmt.Errorf("conflicting cognito identity")
	}

	subject, err := cognitoSubject(*user)
	if err != nil {
		return "", err
	}
	membershipCreated, err := b.ensureMembership(ctx, subject)
	if err != nil {
		return subject, err
	}
	if membershipCreated && emit != nil {
		emit(auth.EventMembershipStatusChanged, subject)
		emit(auth.EventAuthValidAfterAdvanced, subject)
	}
	if membershipCreated && user.UserStatus == sharedaws.CognitoUserStatusForceChangePassword {
		if _, err := b.cognito.AdminCreateUser(ctx, &sharedaws.CognitoAdminCreateUserInput{
			UserPoolId:    sharedaws.String(b.userPoolID),
			Username:      user.Username,
			MessageAction: sharedaws.CognitoMessageActionResend,
		}); err != nil {
			return subject, fmt.Errorf("send cognito invitation: %w", err)
		}
	}
	return subject, nil
}

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", fmt.Errorf("invalid email address")
	}
	return normalized, nil
}

func (b bootstrapper) findUser(ctx context.Context, email string) (*sharedaws.CognitoUser, error) {
	var match *sharedaws.CognitoUser
	var nextToken *string
	for {
		out, err := b.cognito.ListUsers(ctx, &sharedaws.CognitoListUsersInput{
			UserPoolId:      sharedaws.String(b.userPoolID),
			Filter:          sharedaws.String(fmt.Sprintf("email = %q", email)),
			PaginationToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("find cognito user: %w", err)
		}
		for _, user := range out.Users {
			if cognitoEmail(user) != email {
				continue
			}
			if match != nil {
				return nil, fmt.Errorf("ambiguous cognito users for email")
			}
			copy := user
			match = &copy
		}
		if out.PaginationToken == nil || *out.PaginationToken == "" {
			return match, nil
		}
		nextToken = out.PaginationToken
	}
}

func (b bootstrapper) getUser(ctx context.Context, username string) (*sharedaws.CognitoUser, error) {
	out, err := b.cognito.AdminGetUser(ctx, &sharedaws.CognitoAdminGetUserInput{UserPoolId: sharedaws.String(b.userPoolID), Username: sharedaws.String(username)})
	if err != nil {
		return nil, fmt.Errorf("get cognito user: %w", err)
	}
	return &sharedaws.CognitoUser{Username: out.Username, UserStatus: out.UserStatus, Attributes: out.UserAttributes}, nil
}

func (b bootstrapper) ensureMembership(ctx context.Context, subject auth.Subject) (bool, error) {
	key := map[string]sharedaws.AttributeValue{
		"PK": &sharedaws.AttributeValueMemberS{Value: auth.MembershipPK(subject)},
		"SK": &sharedaws.AttributeValueMemberS{Value: auth.MembershipSK},
	}
	existing, err := b.dynamo.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(b.authTable), ConsistentRead: sharedaws.Bool(true), Key: key})
	if err != nil {
		return false, fmt.Errorf("read membership: %w", err)
	}
	if len(existing.Item) != 0 {
		var record membershipRecord
		if err := sharedaws.UnmarshalMap(existing.Item, &record); err != nil || !validMembershipRecord(record, subject) {
			return false, fmt.Errorf("conflicting membership")
		}
		return false, nil
	}

	now := b.now().UTC().Truncate(time.Second)
	membershipID, err := b.membershipID()
	if err != nil {
		return false, fmt.Errorf("generate membership id: %w", err)
	}
	record := membershipRecord{
		PK: auth.MembershipPK(subject), SK: auth.MembershipSK, MembershipID: string(membershipID), Subject: string(subject),
		TenantID: string(auth.DefaultTenantID), Status: string(auth.MembershipStatusActive), Role: string(auth.RoleAdmin),
		AuthValidAfter: now.Unix(), Version: 1, CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339),
	}
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		return false, fmt.Errorf("marshal membership: %w", err)
	}
	if _, err := b.dynamo.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{
		TableName: sharedaws.String(b.authTable), Item: item,
		ConditionExpression: sharedaws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}); err != nil {
		var conditionFailed *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &conditionFailed) {
			existing, readErr := b.dynamo.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{TableName: sharedaws.String(b.authTable), ConsistentRead: sharedaws.Bool(true), Key: key})
			if readErr != nil {
				return false, fmt.Errorf("read concurrent membership: %w", readErr)
			}
			var existingRecord membershipRecord
			if len(existing.Item) != 0 && sharedaws.UnmarshalMap(existing.Item, &existingRecord) == nil && validMembershipRecord(existingRecord, subject) {
				return false, nil
			}
			return false, fmt.Errorf("conflicting membership")
		}
		return false, fmt.Errorf("create membership: %w", err)
	}
	return true, nil
}

func validMembershipRecord(record membershipRecord, subject auth.Subject) bool {
	createdAt, createdErr := time.Parse(time.RFC3339, record.CreatedAt)
	updatedAt, updatedErr := time.Parse(time.RFC3339, record.UpdatedAt)
	return createdErr == nil && updatedErr == nil && record.PK == auth.MembershipPK(subject) && record.SK == auth.MembershipSK && record.Subject == string(subject) && auth.ValidateMembership(auth.Membership{
		MembershipID: auth.MembershipID(record.MembershipID), Subject: auth.Subject(record.Subject), TenantID: auth.TenantID(record.TenantID),
		Status: auth.MembershipStatus(record.Status), Role: auth.Role(record.Role), AuthValidAfter: record.AuthValidAfter, Version: record.Version,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}) == nil
}

func cognitoEmail(user sharedaws.CognitoUser) string {
	for _, attribute := range user.Attributes {
		if sharedaws.ToString(attribute.Name) == "email" {
			email, err := normalizeEmail(sharedaws.ToString(attribute.Value))
			if err == nil {
				return email
			}
		}
	}
	return ""
}

func cognitoSubject(user sharedaws.CognitoUser) (auth.Subject, error) {
	for _, attribute := range user.Attributes {
		if sharedaws.ToString(attribute.Name) == "sub" && strings.TrimSpace(sharedaws.ToString(attribute.Value)) != "" {
			return auth.Subject(sharedaws.ToString(attribute.Value)), nil
		}
	}
	return "", fmt.Errorf("cognito user has no subject")
}

func newMembershipID() (auth.MembershipID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return auth.MembershipID("MEM_" + hex.EncodeToString(value[:])), nil
}
