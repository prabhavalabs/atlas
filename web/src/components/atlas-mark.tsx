import { cn } from "@/lib/utils"

export function AtlasMark({ className }: { className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "grid size-8 place-items-center rounded-lg bg-primary text-sm font-semibold text-primary-foreground",
        className
      )}
    >
      A
    </span>
  )
}
