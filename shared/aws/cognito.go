package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CognitoIdentityProviderAPI keeps credentialed administration behind a testable facade.
type CognitoIdentityProviderAPI interface {
	ListUsers(ctx context.Context, params *CognitoListUsersInput) (*CognitoListUsersOutput, error)
	AdminGetUser(ctx context.Context, params *CognitoAdminGetUserInput) (*CognitoAdminGetUserOutput, error)
	AdminCreateUser(ctx context.Context, params *CognitoAdminCreateUserInput) (*CognitoAdminCreateUserOutput, error)
	AdminResetUserPassword(ctx context.Context, params *CognitoAdminResetUserPasswordInput) (*CognitoAdminResetUserPasswordOutput, error)
}

type CognitoListUsersInput = cognitoidentityprovider.ListUsersInput
type CognitoListUsersOutput = cognitoidentityprovider.ListUsersOutput
type CognitoAdminGetUserInput = cognitoidentityprovider.AdminGetUserInput
type CognitoAdminGetUserOutput = cognitoidentityprovider.AdminGetUserOutput
type CognitoAdminCreateUserInput = cognitoidentityprovider.AdminCreateUserInput
type CognitoAdminCreateUserOutput = cognitoidentityprovider.AdminCreateUserOutput
type CognitoAdminResetUserPasswordInput = cognitoidentityprovider.AdminResetUserPasswordInput
type CognitoAdminResetUserPasswordOutput = cognitoidentityprovider.AdminResetUserPasswordOutput

type cognitoIdentityProvider struct {
	client *cognitoidentityprovider.Client
}

func NewCognitoIdentityProvider(client *cognitoidentityprovider.Client) CognitoIdentityProviderAPI {
	return &cognitoIdentityProvider{client: client}
}

func (c *cognitoIdentityProvider) ListUsers(ctx context.Context, params *CognitoListUsersInput) (*CognitoListUsersOutput, error) {
	return c.client.ListUsers(ctx, params)
}

func (c *cognitoIdentityProvider) AdminGetUser(ctx context.Context, params *CognitoAdminGetUserInput) (*CognitoAdminGetUserOutput, error) {
	return c.client.AdminGetUser(ctx, params)
}

func (c *cognitoIdentityProvider) AdminCreateUser(ctx context.Context, params *CognitoAdminCreateUserInput) (*CognitoAdminCreateUserOutput, error) {
	return c.client.AdminCreateUser(ctx, params)
}

func (c *cognitoIdentityProvider) AdminResetUserPassword(ctx context.Context, params *CognitoAdminResetUserPasswordInput) (*CognitoAdminResetUserPasswordOutput, error) {
	return c.client.AdminResetUserPassword(ctx, params)
}
