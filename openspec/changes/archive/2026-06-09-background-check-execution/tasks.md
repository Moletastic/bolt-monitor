## 1. Infrastructure (infra/stacks/bootstrap.ts)

- [x] 1.1 Add SQS queue `execution-queue` (standard, 1 hour retention)
- [x] 1.2 Add SQS queue `execution-queue-dlq` (standard, 24 hour retention)
- [x] 1.3 Configure execution-queue to send to DLQ after maxReceiveCount=3
- [x] 1.4 Add EventBridge Schedule `SchedulerSchedule` (rate 30 seconds)
- [x] 1.5 Add scheduler Lambda function (RUNTIME_MODE=scheduler, linked to table, EXECUTION_QUEUE_URL env)
- [x] 1.6 Add worker Lambda function (RUNTIME_MODE=worker, linked to table)
- [x] 1.7 Configure SQS trigger on worker Lambda (batch size 1)
- [x] 1.8 Connect EventBridge Schedule to scheduler Lambda

## 2. Code: SQS Client (services/check-runtime/sqs.go) — NEW FILE

- [x] 2.1 Define `sqsClient` interface with `SendMessage(ctx, queueURL, body) error`
- [x] 2.2 Implement `awsSQSClient` wrapping AWS SDK SQS client
- [x] 2.3 Add AWS SDK SQS dependency to go.mod

## 3. Code: Runtime Handler Updates (services/check-runtime/runtime.go)

- [x] 3.1 Add `sqsClient` and `queueURL` fields to `runtimeHandler` struct
- [x] 3.2 Update `newRuntimeHandler` to accept SQS client and queue URL
- [x] 3.3 Modify `runScheduler` to send to SQS after building ExecutionRequests
- [x] 3.4 Add `handleSQSEvent()` method for worker mode
- [x] 3.5 Parse SQS message body as ExecutionRequest JSON
- [x] 3.6 Call `ExecuteHTTP()` with parsed request
- [x] 3.7 Call `RecordExecutionResult()` with result
- [x] 3.8 Return appropriate summary (processed count)

## 4. Code: Main Entry Point (services/check-runtime/main.go)

- [x] 4.1 Initialize SQS client
- [x] 4.2 Pass SQS client and queue URL to `newRuntimeHandler`
- [x] 4.3 Add conditional handler based on RUNTIME_MODE:
  - scheduler → `handleCloudWatchEvent`
  - worker → `handleSQSEvent`

## 5. Code: Repository Interface (services/check-runtime/repository.go)

- [x] 5.1 Add `RecordExecutionResult(ctx, monitor, runID, probeLocationID, trigger, result)` method to `runtimeRepository` interface
- [x] 5.2 Implement in `dynamoRuntimeRepository` (reuse logic from manual run recording)

## 6. Verify and Test

- [x] 6.1 Run `make lint-go` to check for linting issues
- [x] 6.2 Run `make test-go-all` to run all Go tests
- [x] 6.3 Build the Lambda: `make build-go`
- [x] 6.4 Deploy to staging: `make deploy-infra`
- [x] 6.5 Test scheduler manually:
  - Enable recurring in scheduler config ✓
  - Create enabled monitor ✓ (example service already had one)
  - Verify SQS message appears in execution-queue ✓ (recurring runs appearing)
  - Verify RUN_REQUEST# item in DynamoDB ✓ (runs recorded)
- [x] 6.6 Test worker:
  - Trigger manual run (should still work via monitor-api) ✓
  - Verify SQS message is processed ✓ (sent direct SQS message, run appeared)
  - Verify RUN# item written to DynamoDB ✓
  - Verify STATUS item updated ✓
- [x] 6.8 Test DLQ:
  - Manually poison a message (send to DLQ) ✓
  - Verify message appears in execution-queue-dlq ✓ (1 message appeared after3 failures)

(End of file - total 65 lines)