## Context

Lambda `main.go` files currently do inline AWS client creation:

```go
// monitor-api/main.go (current)
awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
dynamodb.NewFromConfig(awsCfg)

// check-runtime/main.go (current)
awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
dynamodb.NewFromConfig(awsCfg)
sqs.NewFromConfig(awsCfg)

// notify-runtime/main.go (current)
awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
dynamodb.NewFromConfig(awsCfg)
```

Each service also defines its own `dynamoAPI` interface variant locally.

## Goals / Non-Goals

**Goals:**
- Single `DynamoDBAPI` interface definition in shared package
- Reusable AWS client factory functions
- Simplified main.go files

**Non-Goals:**
- No DI container (wire/fx) — constructor injection is sufficient
- No handler interface extraction
- No repository factory functions (they already exist as constructors)
- No testability changes — factories are already testable via constructor injection

## Decision 1: DynamoDBAPI Interface Scope

**Choice:** Interface covers methods used by all three services

```go
// shared/dynamodb/client.go
type DynamoDBAPI interface {
    GetItem(ctx context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
    Query(ctx context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
    TransactWriteItems(ctx context.Context, *dynamodb.TransactWriteItemsInput, ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}
```

**Rationale:** These3 methods cover all current repository needs. `PutItem`, `DeleteItem`, `UpdateItem`, `Scan` exist only in `notify-runtime`'s `DynamoDBAPI` variant — not needed by monitor-api or check-runtime.

**Alternatives considered:**
- Include all DynamoDB operations: Overkill, unused methods clutter the interface
- Separate interfaces per service: Defeats the purpose of sharing

## Decision 2: File Location

**Choice:** `shared/dynamodb/client.go` for interface, `shared/aws/clients.go` for factories

```
shared/
  dynamodb/
    client.go    # DynamoDBAPI interface
  aws/
    clients.go   # NewDynamoDBClient, NewSQSClient
```

**Rationale:** Follows existing pattern in `shared/dynamodbschema/` and `shared/notifications/`. AWS-specific concerns in `aws/`, DynamoDB interface in `dynamodb/`.

## Decision 3: Factory Function Signatures

**Choice:**

```go
// shared/aws/clients.go
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error)
func NewSQSClient(ctx context.Context) (*sqs.Client, error)
```

**Rationale:** Returns concrete client (not interface). Callers pass to repository constructors which accept the interface. Error returned for `LoadDefaultConfig` failures.

## Decision 4: AWS Config Loading

**Choice:** Each factory calls `awsconfig.LoadDefaultConfig(ctx)` internally

**Rationale:** Consistent with current behavior. Each Lambda gets its own AWS config. No global state.

## Decision 5: Main.go Changes

**monitor-api/main.go — After:**
```go
dynamoClient, err := aws.NewDynamoDBClient(ctx)
repo := newDynamoMonitorRepository(dynamoClient, os.Getenv("TABLE_NAME"))
```

**check-runtime/main.go — After:**
```go
dynamoClient, err := aws.NewDynamoDBClient(ctx)
sqsClient, err := aws.NewSQSClient(ctx)
repo := newDynamoRuntimeRepository(dynamoClient, os.Getenv("TABLE_NAME"))
```

**notify-runtime/main.go — After:**
```go
dynamoClient, err := aws.NewDynamoDBClient(ctx)
channelRepo := notifications.NewDynamoNotificationChannelRepository(dynamoClient, tableName)
```

## Interface Compatibility

| Service | Old Interface | New Shared Interface |
|---------|---------------|---------------------|
| monitor-api | `dynamoAPI` (local) | `DynamoDBAPI` |
| check-runtime | `dynamoAPI` (local) | `DynamoDBAPI` |
| notify-runtime | `DynamoDBAPI` (in notifications) | `DynamoDBAPI` |

All existing repository constructors accept `dynamoAPI`/`DynamoDBAPI` — no signature changes needed.

## Migration Plan

1. Create `shared/dynamodb/client.go` with `DynamoDBAPI` interface
2. Create `shared/aws/clients.go` with `NewDynamoDBClient`, `NewSQSClient`
3. Update `services/monitor-api/main.go` to use factories, remove local `dynamoAPI`
4. Update `services/check-runtime/main.go` to use factories, remove local `dynamoAPI`
5. Update `services/notify-runtime/main.go` to use factories
6. Update `services/monitor-api/repository.go` to import `DynamoDBAPI` from shared
7. Update `services/check-runtime/repository.go` to import `DynamoDBAPI` from shared
8. Delete local `dynamoAPI` interface from both repository files
9. Run `go build` and `go test` to verify

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking existing tests that mock dynamoAPI | Interface method signatures unchanged |
| notify-runtime uses wider DynamoDBAPI variant | Already compatible — notify-runtime uses shared, not vice versa |
| Factory error handling changes | Same pattern, just moved to factory |
