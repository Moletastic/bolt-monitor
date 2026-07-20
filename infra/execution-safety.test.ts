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
