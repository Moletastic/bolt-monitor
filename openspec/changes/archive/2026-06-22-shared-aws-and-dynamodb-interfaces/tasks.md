## 1. Create shared DynamoDB interface

- [x] 1.1 Create `shared/dynamodb/` directory
- [x] 1.2 Create `shared/dynamodb/client.go`
- [x] 1.3 Define `DynamoDBAPI` interface with `GetItem`, `Query`, `TransactWriteItems`

## 2. Create AWS client factories

- [x] 2.1 Create `shared/aws/` directory
- [x] 2.2 Create `shared/aws/clients.go`
- [x] 2.3 Add `NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error)`
- [x] 2.4 Add `NewSQSClient(ctx context.Context) (*sqs.Client, error)`

## 3. Update monitor-api

- [x] 3.1 Update `services/monitor-api/main.go` to use `aws.NewDynamoDBClient()`
- [x] 3.2 Import `bolt-monitor/shared/dynamodb` in `repository.go`
- [x] 3.3 Replace local `dynamoAPI` interface with import from `shared/dynamodb`
- [x] 3.4 Verify `go build ./services/monitor-api/...` passes

## 4. Update check-runtime

- [x] 4.1 Update `services/check-runtime/main.go` to use `aws.NewDynamoDBClient()` and `aws.NewSQSClient()`
- [x] 4.2 Import `bolt-monitor/shared/dynamodb` in `repository.go`
- [x] 4.3 Replace local `dynamoAPI` interface with import from `shared/dynamodb`
- [x] 4.4 Verify `go build ./services/check-runtime/...` passes

## 5. Update notify-runtime

- [x] 5.1 Update `services/notify-runtime/main.go` to use `aws.NewDynamoDBClient()`
- [x] 5.2 Verify `go build ./services/notify-runtime/...` passes

## 6. Verify

- [x] 6.1 Run `go build ./services/... ./shared/...`
- [x] 6.2 Run `go test ./services/... ./shared/...`
- [x] 6.3 Run `make lint-go`
