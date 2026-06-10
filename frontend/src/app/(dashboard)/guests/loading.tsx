import { SkeletonTable } from '@/components/ui/skeleton';
import { Skeleton } from '@/components/ui/skeleton';

export default function GuestsLoading() {
  return (
    <div className="page-container animate-in fade-in duration-300">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-7 w-24" />
          <Skeleton className="h-4 w-40" />
        </div>
        <Skeleton className="h-9 w-28 rounded-lg" />
      </div>
      <Skeleton className="h-10 w-full rounded-lg" />
      <SkeletonTable rows={6} cols={4} />
    </div>
  );
}
