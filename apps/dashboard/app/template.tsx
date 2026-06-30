export default function Template({ children }: { children: React.ReactNode }) {
  return <div className="motion-safe:animate-segment-fade-in">{children}</div>
}
