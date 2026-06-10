import * as React from 'react';

function cn(...classes: (string | undefined | false)[]) {
  return classes.filter(Boolean).join(' ');
}

type BadgeVariant = 'blue' | 'green' | 'yellow' | 'red' | 'gray' | 'purple' | 'orange' | 'teal';

interface StatusBadgeProps {
  variant?: BadgeVariant;
  children: React.ReactNode;
  className?: string;
  dot?: boolean;
}

const variantClasses: Record<BadgeVariant, string> = {
  blue: 'bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20',
  green: 'bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/20',
  yellow: 'bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-600/20',
  red: 'bg-red-50 text-red-700 ring-1 ring-inset ring-red-600/20',
  gray: 'bg-gray-100 text-gray-700 ring-1 ring-inset ring-gray-500/20',
  purple: 'bg-violet-50 text-violet-700 ring-1 ring-inset ring-violet-600/20',
  orange: 'bg-orange-50 text-orange-700 ring-1 ring-inset ring-orange-600/20',
  teal: 'bg-teal-50 text-teal-700 ring-1 ring-inset ring-teal-600/20',
};

const dotClasses: Record<BadgeVariant, string> = {
  blue: 'bg-blue-500',
  green: 'bg-emerald-500',
  yellow: 'bg-amber-500',
  red: 'bg-red-500',
  gray: 'bg-gray-500',
  purple: 'bg-violet-500',
  orange: 'bg-orange-500',
  teal: 'bg-teal-500',
};

// Map common status strings to variants
const statusMap: Record<string, BadgeVariant> = {
  confirmed: 'blue',
  pending: 'yellow',
  checked_in: 'green',
  checked_out: 'gray',
  cancelled: 'red',
  open: 'blue',
  settled: 'green',
  partial: 'yellow',
  overdue: 'red',
  active: 'green',
  inactive: 'gray',
  completed: 'green',
  in_progress: 'blue',
  urgent: 'red',
  normal: 'gray',
  express: 'purple',
  regular: 'blue',
};

function getVariantFromStatus(status: string): BadgeVariant {
  return statusMap[status.toLowerCase()] || 'gray';
}

function StatusBadge({ variant, children, className, dot }: StatusBadgeProps) {
  const resolvedVariant = variant || (typeof children === 'string' ? getVariantFromStatus(children) : 'gray');
  
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium',
        variantClasses[resolvedVariant],
        className
      )}
    >
      {dot && <span className={cn('h-1.5 w-1.5 rounded-full', dotClasses[resolvedVariant])} />}
      {children}
    </span>
  );
}

export { StatusBadge, getVariantFromStatus };
export type { BadgeVariant };
