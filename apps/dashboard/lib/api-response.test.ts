import { describe, expect, it } from 'vitest'

import { Status, err, isError, isSuccess, ok, okPaginated, type ApiResponse } from './api-response'

describe('api response envelope', () => {
  it('ok() emits a success envelope without reason/message', () => {
    const envelope = ok({ name: 'ok' })
    expect(envelope.status).toBe(Status.Success)
    expect(envelope.data?.name).toBe('ok')
    expect(envelope.reason).toBeUndefined()
    expect(envelope.message).toBeUndefined()
    expect(isSuccess(envelope)).toBe(true)
  })

  it('ok() preserves an explicit message', () => {
    const envelope = ok({ id: 'svc' }, 'created')
    expect(envelope.message).toBe('created')
  })

  it('err() emits an error envelope with reason code and details', () => {
    const envelope = err<never>('NOT_FOUND', { id: 'svc' })
    expect(envelope.status).toBe(Status.Error)
    expect(envelope.reason?.code).toBe('NOT_FOUND')
    expect(envelope.reason?.details?.id).toBe('svc')
    expect(isError(envelope)).toBe(true)
  })

  it('okPaginated() sets pagination metadata and items', () => {
    const envelope = okPaginated([{ id: 1 }, { id: 2 }], 1, 2, 2)
    expect(envelope.pagination).toBeDefined()
    expect(envelope.pagination?.page).toBe(1)
    expect(envelope.pagination?.size).toBe(2)
    expect(envelope.pagination?.total).toBe(2)
    expect(envelope.pagination?.items).toHaveLength(2)
  })

  it('isSuccess narrows the data field', () => {
    const envelope: ApiResponse<{ name: string }> = ok({ name: 'x' })
    expect(isSuccess(envelope)).toBe(true)
    if (!isSuccess(envelope)) throw new Error('isSuccess should match')
    const data: { name: string } = envelope.data
    expect(data.name).toBe('x')
  })

  it('isError narrows the reason field', () => {
    const envelope = err('BOOM')
    expect(isError(envelope)).toBe(true)
    if (!isError(envelope)) throw new Error('isError should match')
    expect(envelope.reason.code).toBe('BOOM')
  })
})
