import { PollingProvider } from '@/components/polling-provider'
import { ToastWatcher } from '@/components/toast-watcher'

export default function MonitoringLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <PollingProvider intervalMs={5000} />
      <ToastWatcher />
      {children}
    </>
  )
}
