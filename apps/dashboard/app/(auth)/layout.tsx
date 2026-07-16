import { redirectIfDashboardSession } from '@/lib/auth/session-guard'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  await redirectIfDashboardSession()
  return children
}
