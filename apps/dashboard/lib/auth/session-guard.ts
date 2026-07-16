import 'server-only'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import type {
  DashboardSession,
  DashboardSessionReference,
  DashboardSessionStore,
} from './contracts'
import {
  DASHBOARD_SESSION_COOKIE,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'

/**
 * Validates the opaque browser session against the authoritative AuthTable.
 * Layout checks protect renders; actions and route handlers must call this too.
 */
export async function requireDashboardSession(): Promise<DashboardSession> {
  const cookieStore = await cookies()
  const reference = cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value
  if (!reference) redirect('/login')

  const session = await readDashboardSession(
    reference as DashboardSessionReference,
    createDynamoDashboardSessionStoreFromEnv()
  )
  if (!session) redirect('/login')
  return session
}

/** Redirects established operators away from public authentication pages. */
export async function redirectIfDashboardSession(): Promise<void> {
  const cookieStore = await cookies()
  const reference = cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value
  if (!reference) return

  const session = await readDashboardSession(
    reference as DashboardSessionReference,
    createDynamoDashboardSessionStoreFromEnv()
  )
  if (session) redirect('/')
}

export async function readDashboardSession(
  reference: DashboardSessionReference,
  store: Pick<DashboardSessionStore, 'read'>
): Promise<DashboardSession | null> {
  const result = await store.read(reference)
  return result.ok ? result.value : null
}
