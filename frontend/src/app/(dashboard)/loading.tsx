import { Skeleton, SkeletonStats } from '@/components/ui/skeleton';

export default function DashboardLoading() {
  return (
    <div className="space-y-5 max-w-[1600px] mx-auto animate-in fade-in duration-300">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-7 w-40" />
          <Skeleton className="h-4 w-56" />
        </div>
        <Skeleton className="h-7 w-24 rounded-full" />
      </div>

      {/* Morning Brief */}
      <Skeleton className="h-20 w-full rounded-xl" />

      {/* Stats */}
      <SkeletonStats count={4} />

      {/* Room Board */}
      <div className="rounded-xl border bg-card p-6 space-y-4">
        <Skeleton className="h-5 w-32" />
        <div className="grid grid-cols-4 lg:grid-cols-8 gap-2">
          {Array.from({ length: 16 }).map((_, i) => (
            <Skeleton key={i} className="h-14 rounded-lg" />
          ))}
        </div>
      </div>
    </div>
  );
}
