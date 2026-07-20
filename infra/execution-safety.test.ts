import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const stackSource = readFileSync(
  fileURLToPath(new URL('./stacks/bootstrap.ts', import.meta.url)),
  'utf8'
)

test('execution worker uses bounded retry-safe settings', () => {
  const workerStart = stackSource.indexOf('executionQueue.subscribe(')
  assert.notEqual(workerStart, -1)
  const worker = stackSource.slice(workerStart, stackSource.indexOf('\n  )', workerStart))

  assert.match(worker, /timeout: '45 seconds'/)
  assert.match(worker, /memory: '512 MB'/)
  assert.match(worker, /WORKER_LAMBDA_TIMEOUT_SECONDS: '45'/)
  assert.match(worker, /WORK_LEASE_DURATION_SECONDS: '60'/)
  assert.match(worker, /EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY: '5'/)
  assert.match(worker, /size: 10/)
  assert.match(worker, /partialResponses: true/)
  assert.doesNotMatch(worker, /ESCALATION_QUEUE_URL|sqs:SendMessage/)
})

test('notification transition stream uses insert filtering and partial failures', () => {
  assert.match(stackSource, /stream: 'new-image'/)
  const streamStart = stackSource.indexOf('table.subscribe(')
  assert.notEqual(streamStart, -1)
  const stream = stackSource.slice(streamStart)
  assert.match(stream, /EscalationRuntimeTransitionStream/)
  assert.match(stream, /eventName: \['INSERT'\]/)
  assert.match(stream, /EntityType: \{ S: \['TransitionOutbox'\] \}/)
  assert.match(stream, /functionResponseTypes = \['ReportBatchItemFailures'\]/)
  assert.match(stream, /maximumRetryAttempts = 3/)
})

test('notification queue preserves poison messages with partial responses', () => {
  const queueStart = stackSource.indexOf('notificationQueue.subscribe(')
  assert.notEqual(queueStart, -1)
  const queue = stackSource.slice(queueStart, stackSource.indexOf('\n  )', queueStart))
  assert.match(queue, /partialResponses: true/)
})

test('transition dispatch failures alarm on notification DLQ depth', () => {
  assert.match(stackSource, /EscalationTransitionDispatchAlarm/)
  assert.match(stackSource, /QueueName: notificationQueueDLQ\.nodes\.queue\.name/)
  assert.match(stackSource, /ApproximateNumberOfMessagesVisible/)
})

test('escalation schedules use AWS Scheduler with managed group + delete action', () => {
  assert.match(stackSource, /EscalationScheduleGroup/)
  assert.match(stackSource, /EscalationScheduleExecutionRole/)
  assert.match(stackSource, /SCHEDULE_GROUP_NAME: escalationScheduleGroup\.name/)
  assert.match(stackSource, /SCHEDULE_EXECUTION_ROLE_ARN: escalationScheduleRole\.arn/)
  assert.doesNotMatch(
    stackSource,
    /EventBridge|PutRule|PutTargets|AddPermission|events\.amazonaws\.com|lambda:InvokeFunction/
  )
})

test('scheduler runtime permissions are scoped to managed group + dedicated role', () => {
  assert.match(stackSource, /scheduler:CreateSchedule/)
  assert.match(stackSource, /scheduler:DeleteSchedule/)
  assert.match(stackSource, /iam:PassRole/)
  assert.match(stackSource, /schedule\/\$\{escalationScheduleGroup\.name\}\/\*/)
})

test('scheduler execution policy resolves queue outputs before IAM serialization', () => {
  const policyStart = stackSource.indexOf("'EscalationScheduleExecutionPolicy'")
  assert.notEqual(policyStart, -1)
  const policy = stackSource.slice(policyStart, stackSource.indexOf('\n  )', policyStart))

  assert.match(policy, /notificationQueue\.arn\.apply\(\(notificationQueueArn\) =>/)
  assert.match(policy, /notificationQueueDLQ\.arn\.apply\(\(notificationQueueDLQArn\) =>/)
  assert.match(policy, /Resource: \[notificationQueueArn, notificationQueueDLQArn\]/)
  assert.doesNotMatch(policy, /Resource: \[notificationQueue\.arn, notificationQueueDLQ\.arn\]/)
})
