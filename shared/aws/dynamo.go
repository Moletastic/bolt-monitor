package aws

import (
	"context"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *DynamoDBGetItemInput) (*DynamoDBGetItemOutput, error)
	PutItem(ctx context.Context, params *DynamoDBPutItemInput) (*DynamoDBPutItemOutput, error)
	UpdateItem(ctx context.Context, params *DynamoDBUpdateItemInput) (*DynamoDBUpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *DynamoDBDeleteItemInput) (*DynamoDBDeleteItemOutput, error)
	Query(ctx context.Context, params *DynamoDBQueryInput) (*DynamoDBQueryOutput, error)
	Scan(ctx context.Context, params *DynamoDBScanInput) (*DynamoDBScanOutput, error)
	TransactWriteItems(ctx context.Context, params *DynamoDBTransactWriteItemsInput) (*DynamoDBTransactWriteItemsOutput, error)
}

type DynamoDBGetItemInput = dynamodb.GetItemInput
type DynamoDBGetItemOutput = dynamodb.GetItemOutput
type DynamoDBPutItemInput = dynamodb.PutItemInput
type DynamoDBPutItemOutput = dynamodb.PutItemOutput
type DynamoDBUpdateItemInput = dynamodb.UpdateItemInput
type DynamoDBUpdateItemOutput = dynamodb.UpdateItemOutput
type DynamoDBDeleteItemInput = dynamodb.DeleteItemInput
type DynamoDBDeleteItemOutput = dynamodb.DeleteItemOutput
type DynamoDBQueryInput = dynamodb.QueryInput
type DynamoDBQueryOutput = dynamodb.QueryOutput
type DynamoDBScanInput = dynamodb.ScanInput
type DynamoDBScanOutput = dynamodb.ScanOutput
type DynamoDBTransactWriteItemsInput = dynamodb.TransactWriteItemsInput
type DynamoDBTransactWriteItemsOutput = dynamodb.TransactWriteItemsOutput
type AttributeValue = ddbtypes.AttributeValue
type TransactWriteItem = ddbtypes.TransactWriteItem
type Put = ddbtypes.Put
type Delete = ddbtypes.Delete
type AttributeValueMemberS = ddbtypes.AttributeValueMemberS

func String(value string) *string { return sdkaws.String(value) }
func Int32(value int32) *int32    { return sdkaws.Int32(value) }
func Bool(value bool) *bool       { return sdkaws.Bool(value) }
func ToString(value *string) string {
	return sdkaws.ToString(value)
}

func MarshalMap(in any) (map[string]AttributeValue, error) {
	return attributevalue.MarshalMap(in)
}

func UnmarshalMap(in map[string]AttributeValue, out any) error {
	return attributevalue.UnmarshalMap(in, out)
}

type dynamoDB struct {
	client *dynamodb.Client
}

func NewDynamoDB(client *dynamodb.Client) DynamoDBAPI {
	return &dynamoDB{client: client}
}

func (d *dynamoDB) GetItem(ctx context.Context, params *DynamoDBGetItemInput) (*DynamoDBGetItemOutput, error) {
	return d.client.GetItem(ctx, params)
}

func (d *dynamoDB) PutItem(ctx context.Context, params *DynamoDBPutItemInput) (*DynamoDBPutItemOutput, error) {
	return d.client.PutItem(ctx, params)
}

func (d *dynamoDB) UpdateItem(ctx context.Context, params *DynamoDBUpdateItemInput) (*DynamoDBUpdateItemOutput, error) {
	return d.client.UpdateItem(ctx, params)
}

func (d *dynamoDB) DeleteItem(ctx context.Context, params *DynamoDBDeleteItemInput) (*DynamoDBDeleteItemOutput, error) {
	return d.client.DeleteItem(ctx, params)
}

func (d *dynamoDB) Query(ctx context.Context, params *DynamoDBQueryInput) (*DynamoDBQueryOutput, error) {
	return d.client.Query(ctx, params)
}

func (d *dynamoDB) Scan(ctx context.Context, params *DynamoDBScanInput) (*DynamoDBScanOutput, error) {
	return d.client.Scan(ctx, params)
}

func (d *dynamoDB) TransactWriteItems(ctx context.Context, params *DynamoDBTransactWriteItemsInput) (*DynamoDBTransactWriteItemsOutput, error) {
	return d.client.TransactWriteItems(ctx, params)
}
