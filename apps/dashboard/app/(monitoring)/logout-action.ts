'use server'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { requireDashboardCsrf } from '@/lib/auth/csrf'
import type { DashboardSessionReference } from '@/lib/auth/contracts'
import { requireDashboardSession } from '@/lib/auth/session-guard'
import {
  DASHBOARD_SESSION_COOKIE,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'

export async function logoutAction() {
  await requireDashboardCsrf()
  await requireDashboardSession()

  const cookieStore = await cookies()
  const reference = cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value
  if (reference) {
    await createDynamoDashboardSessionStoreFromEnv().invalidate(
      reference as DashboardSessionReference
    )
  }
  cookieStore.set(DASHBOARD_SESSION_COOKIE.name, '', {
    httpOnly: DASHBOARD_SESSION_COOKIE.httpOnly,
    secure: DASHBOARD_SESSION_COOKIE.secure,
    sameSite: DASHBOARD_SESSION_COOKIE.sameSite,
    path: DASHBOARD_SESSION_COOKIE.path,
    maxAge: 0,
  })
  redirect('/login')
}
