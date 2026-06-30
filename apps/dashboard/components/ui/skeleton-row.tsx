import { Skeleton } from '@/components/ui/skeleton'
import { TableCell, TableRow } from '@/components/ui/table'

const cellWidths = ['w-32', 'w-40', 'w-28', 'w-24', 'w-36', 'w-20']

export function TableCellSkeleton({ index = 0 }: { index?: number }) {
  const width = cellWidths[index % cellWidths.length]
  return (
    <TableCell>
      <Skeleton className={width} height="0.875rem" />
    </TableCell>
  )
}

export function TableRowSkeleton({
  columns,
  rows = 1,
}: {
  columns: number
  rows?: number
}) {
  return (
    <>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <TableRow key={rowIndex} aria-hidden="true">
          {Array.from({ length: columns }).map((_, colIndex) => (
            <TableCellSkeleton index={colIndex + rowIndex} key={colIndex} />
          ))}
        </TableRow>
      ))}
    </>
  )
}
