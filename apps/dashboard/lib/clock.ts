// Native Date is isolated here so application code can depend on an injectable clock.
export function now() {
  return new Date()
}
