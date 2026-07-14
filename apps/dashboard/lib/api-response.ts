export interface ApiReason {
  code: string
  details: Record<string, unknown>
}

export interface OffsetPagination<T = unknown> {
  page: number
  size: number
  total: number
  items: T[]
}

export interface CursorPagination {
  size: number
  nextCursor?: string
}

export type ApiPagination<T = unknown> = OffsetPagination<T> | CursorPagination

export interface ApiResponse<T = unknown> {
  status: 'success' | 'error'
  data?: T
  reason?: ApiReason
  message?: string
  pagination?: ApiPagination
}

export const Status = {
  Success: 'success',
  Error: 'error',
} as const

export type StatusValue = (typeof Status)[keyof typeof Status]

export function isSuccess<T>(response: ApiResponse<T>): response is ApiResponse<T> & { data: T } {
  return response.status === Status.Success && response.data !== undefined
}

export function isError<T>(
  response: ApiResponse<T>
): response is ApiResponse<T> & { reason: ApiReason } {
  return response.status === Status.Error && response.reason !== undefined
}

export function isCursorPagination(
  pagination: ApiPagination | undefined
): pagination is CursorPagination {
  return pagination !== undefined && !('page' in pagination)
}

export function ok<T>(data: T, message?: string): ApiResponse<T> {
  return message !== undefined
    ? { status: Status.Success, data, message }
    : { status: Status.Success, data }
}

export function err<T = never>(
  code: string,
  details: Record<string, unknown> = {},
  message?: string
): ApiResponse<T> {
  const reason: ApiReason = { code, details }
  return message !== undefined
    ? { status: Status.Error, reason, message }
    : { status: Status.Error, reason }
}

export function okPaginated<T>(data: T, page: number, size: number, total: number): ApiResponse<T> {
  return {
    status: Status.Success,
    data,
    pagination: { page, size, total, items: data as unknown[] },
  }
}

export function okCursorPaginated<T>(data: T, size: number, nextCursor?: string): ApiResponse<T> {
  return {
    status: Status.Success,
    data,
    pagination: { size, ...(nextCursor ? { nextCursor } : {}) },
  }
}
