export default function MonitoringTemplate({ children }: { children: React.ReactNode }) {
  return <div className="motion-safe:animate-segment-fade-in">{children}</div>
}
