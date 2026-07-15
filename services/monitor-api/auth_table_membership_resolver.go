package main

import (
	"context"
	"time"

	"bolt-monitor/shared/auth"
	sharedaws "bolt-monitor/shared/aws"
	sharederrors "bolt-monitor/shared/errors"
)

// MembershipResolver authorizes a validated identity against current AuthTable state.
type MembershipResolver interface {
	Resolve(context.Context, auth.AuthenticatedIdentity) (auth.Principal, error)
}

type authTableMembershipResolver struct {
	client    sharedaws.DynamoDBAPI
	tableName string
}

type authMembershipRecord struct {
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

func newAuthTableMembershipResolver(client sharedaws.DynamoDBAPI, tableName string) MembershipResolver {
	return authTableMembershipResolver{client: client, tableName: tableName}
}

func (r authTableMembershipResolver) Resolve(ctx context.Context, identity auth.AuthenticatedIdentity) (auth.Principal, error) {
	out, err := r.client.GetItem(ctx, &sharedaws.DynamoDBGetItemInput{
		TableName:      sharedaws.String(r.tableName),
		ConsistentRead: sharedaws.Bool(true),
		Key: map[string]sharedaws.AttributeValue{
			"PK": &sharedaws.AttributeValueMemberS{Value: auth.MembershipPK(identity.Subject)},
			"SK": &sharedaws.AttributeValueMemberS{Value: auth.MembershipSK},
		},
	})
	if err != nil {
		return auth.Principal{}, err
	}
	if out == nil || len(out.Item) == 0 {
		return auth.Principal{}, authorizationDenied()
	}

	var record authMembershipRecord
	if err := sharedaws.UnmarshalMap(out.Item, &record); err != nil {
		return auth.Principal{}, authorizationDenied()
	}
	if record.PK != auth.MembershipPK(identity.Subject) || record.SK != auth.MembershipSK || record.Subject != string(identity.Subject) {
		return auth.Principal{}, authorizationDenied()
	}

	membership, ok := membershipFromRecord(record)
	if !ok || auth.ValidateMembership(membership) != nil || !auth.IsAuthorizedAuthTime(identity.AuthTime, membership.AuthValidAfter) {
		return auth.Principal{}, authorizationDenied()
	}
	return auth.Principal{Subject: membership.Subject, TenantID: membership.TenantID, Role: membership.Role}, nil
}

func membershipFromRecord(record authMembershipRecord) (auth.Membership, bool) {
	createdAt, err := time.Parse(time.RFC3339, record.CreatedAt)
	if err != nil {
		return auth.Membership{}, false
	}
	updatedAt, err := time.Parse(time.RFC3339, record.UpdatedAt)
	if err != nil {
		return auth.Membership{}, false
	}
	return auth.Membership{
		MembershipID:   auth.MembershipID(record.MembershipID),
		Subject:        auth.Subject(record.Subject),
		TenantID:       auth.TenantID(record.TenantID),
		Status:         auth.MembershipStatus(record.Status),
		Role:           auth.Role(record.Role),
		AuthValidAfter: record.AuthValidAfter,
		Version:        record.Version,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, true
}

func authorizationDenied() error {
	return sharederrors.New(sharederrors.CodeAuthorizationDenied, nil)
}
