import { PollingProvider } from '@/components/polling-provider'
import { ToastWatcher } from '@/components/toast-watcher'
import { requireDashboardSession } from '@/lib/auth/session-guard'

export default async function MonitoringLayout({ children }: { children: React.ReactNode }) {
  await requireDashboardSession()

  return (
    <>
      <PollingProvider intervalMs={5000} />
      <ToastWatcher />
      {children}
    </>
  )
}
