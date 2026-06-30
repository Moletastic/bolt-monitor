## Context

Most table and list UIs allow clicking anywhere on a row to navigate to the detail view. Currently, the monitor overview list requires clicking on specific elements (like the monitor name), which is a poor UX pattern.

## Goals / Non-Goals

**Goals:**
- Make entire monitor row clickable
- Show pointer cursor on hover
- Provide visual feedback on hover (background color change)

**Non-Goals:**
- Changing the actual navigation destination
- Modifying the row content or layout

## Implementation

### Option 1: CSS Pointer Events on Row

```css
.monitor-row {
  cursor: pointer;
}

.monitor-row:hover {
  background-color: var(--hover-bg);
}
```

```tsx
<Link href={`/services/${serviceId}/monitors/${monitorId}`} className="monitor-row">
  {/* row content */}
</Link>
```

### Option 2: JavaScript Click Handler

```tsx
<div 
  className="monitor-row" 
  onClick={() => router.push(`/services/${serviceId}/monitors/${monitorId}`)}
  role="button"
  tabIndex={0}
>
  {/* row content */}
</div>
```

### Recommended: Option1 (CSS)

Using CSS pointer events is cleaner and more accessible. The `Link` component already handles keyboard navigation properly.

## Affected Components

| Component | File | Change |
|-----------|------|--------|
| Monitor list item | `apps/dashboard/app/services/**/monitors/*.tsx` | Add row click styling |
| Service monitors list | `apps/dashboard/app/services/[serviceId]/monitors/*.tsx` | Add row click styling |
