## Why

Each Lambda's `main.go` directly instantiates AWS clients and duplicates the `dynamoAPI` interface. This scatter causes:
- Duplicated interface definitions across services
- Inline AWS client creation that can't be reused or mocked in tests
- No canonical place for AWS client configuration

## What Changes

- Define `DynamoDBAPI` interface in `shared/dynamodb/client.go`
- Add `NewDynamoDBClient()` and `NewSQSClient()` factory functions in `shared/aws/clients.go`
- Update each Lambda's `main.go` to use the shared factories

## Capabilities

### New Capabilities

- `aws-client-factories`: Reusable AWS client constructors in `shared/aws/`
- `dynamodb-interface`: Shared DynamoDB interface in `shared/dynamodb/`

### Modified Capabilities

- `lambda-main`: Simplified main.go using shared factories

## Impact

- New files: `shared/dynamodb/client.go`, `shared/aws/clients.go`
- Updated: `services/monitor-api/main.go`, `services/check-runtime/main.go`, `services/notify-runtime/main.go`
- No behavioral changes — pure infrastructure reuse
