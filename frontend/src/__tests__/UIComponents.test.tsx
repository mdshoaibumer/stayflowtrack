import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Skeleton, SkeletonCard, SkeletonTable, SkeletonStats } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { StatusBadge } from '@/components/ui/status-badge';
import { Alert } from '@/components/ui/alert';

describe('Skeleton Components', () => {
  describe('Skeleton', () => {
    it('renders a div with animate-pulse', () => {
      render(<Skeleton data-testid="skel" />);
      const el = screen.getByTestId('skel');
      expect(el.className).toContain('animate-pulse');
      expect(el.className).toContain('rounded-lg');
    });

    it('applies custom className', () => {
      render(<Skeleton data-testid="skel" className="h-8 w-32" />);
      const el = screen.getByTestId('skel');
      expect(el.className).toContain('h-8');
      expect(el.className).toContain('w-32');
    });
  });

  describe('SkeletonCard', () => {
    it('renders card skeleton with proper structure', () => {
      const { container } = render(<SkeletonCard />);
      const card = container.firstElementChild!;
      expect(card.className).toContain('rounded-xl');
      expect(card.className).toContain('border');
    });

    it('contains multiple skeleton lines', () => {
      const { container } = render(<SkeletonCard />);
      const pulseElements = container.querySelectorAll('.animate-pulse');
      expect(pulseElements.length).toBeGreaterThanOrEqual(3);
    });

    it('applies custom className', () => {
      const { container } = render(<SkeletonCard className="mt-4" />);
      expect(container.firstElementChild!.className).toContain('mt-4');
    });
  });

  describe('SkeletonTable', () => {
    it('renders default 5 rows', () => {
      const { container } = render(<SkeletonTable />);
      const rows = container.querySelectorAll('.border-t');
      expect(rows.length).toBe(5);
    });

    it('renders custom row count', () => {
      const { container } = render(<SkeletonTable rows={3} />);
      const rows = container.querySelectorAll('.border-t');
      expect(rows.length).toBe(3);
    });

    it('renders header row', () => {
      const { container } = render(<SkeletonTable />);
      const header = container.querySelector('.bg-gray-50\\/80');
      expect(header).toBeInTheDocument();
    });

    it('renders correct number of columns', () => {
      const { container } = render(<SkeletonTable cols={6} />);
      const headerCols = container.querySelector('.bg-gray-50\\/80')!.querySelectorAll('.animate-pulse');
      expect(headerCols.length).toBe(6);
    });
  });

  describe('SkeletonStats', () => {
    it('renders default 4 stat cards', () => {
      const { container } = render(<SkeletonStats />);
      const cards = container.querySelectorAll('.rounded-xl.border');
      expect(cards.length).toBe(4);
    });

    it('renders custom count', () => {
      const { container } = render(<SkeletonStats count={2} />);
      const cards = container.querySelectorAll('.rounded-xl.border');
      expect(cards.length).toBe(2);
    });

    it('has grid layout', () => {
      const { container } = render(<SkeletonStats />);
      expect(container.firstElementChild!.className).toContain('grid');
    });
  });
});

describe('EmptyState Component', () => {
  it('renders title', () => {
    render(<EmptyState title="No results" />);
    expect(screen.getByText('No results')).toBeInTheDocument();
  });

  it('renders description', () => {
    render(<EmptyState title="No items" description="Try adding some" />);
    expect(screen.getByText('Try adding some')).toBeInTheDocument();
  });

  it('renders action slot', () => {
    render(<EmptyState title="Empty" action={<button>Add Item</button>} />);
    expect(screen.getByRole('button', { name: 'Add Item' })).toBeInTheDocument();
  });

  it('renders custom icon', () => {
    render(<EmptyState title="Empty" icon={<span data-testid="custom-icon">★</span>} />);
    expect(screen.getByTestId('custom-icon')).toBeInTheDocument();
  });

  it('renders default icon when none provided', () => {
    const { container } = render(<EmptyState title="Empty" />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('has centered text layout', () => {
    const { container } = render(<EmptyState title="Empty" />);
    expect(container.firstElementChild!.className).toContain('text-center');
  });

  it('applies custom className', () => {
    const { container } = render(<EmptyState title="Empty" className="mt-8" />);
    expect(container.firstElementChild!.className).toContain('mt-8');
  });
});

describe('StatusBadge Component', () => {
  it('renders with text content', () => {
    render(<StatusBadge>Active</StatusBadge>);
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('auto-resolves status to variant', () => {
    render(<StatusBadge>confirmed</StatusBadge>);
    const badge = screen.getByText('confirmed');
    expect(badge.className).toContain('blue');
  });

  it('maps pending to yellow', () => {
    render(<StatusBadge>pending</StatusBadge>);
    const badge = screen.getByText('pending');
    expect(badge.className).toContain('amber');
  });

  it('maps checked_in to green', () => {
    render(<StatusBadge>checked_in</StatusBadge>);
    const badge = screen.getByText('checked_in');
    expect(badge.className).toContain('emerald');
  });

  it('maps cancelled to red', () => {
    render(<StatusBadge>cancelled</StatusBadge>);
    const badge = screen.getByText('cancelled');
    expect(badge.className).toContain('red');
  });

  it('applies explicit variant over auto-detection', () => {
    render(<StatusBadge variant="purple">anything</StatusBadge>);
    const badge = screen.getByText('anything');
    expect(badge.className).toContain('violet');
  });

  it('renders dot indicator when dot prop is true', () => {
    render(<StatusBadge dot>active</StatusBadge>);
    const badge = screen.getByText('active');
    const dot = badge.querySelector('.rounded-full');
    expect(dot).toBeInTheDocument();
  });

  it('does not render dot when dot prop is false/absent', () => {
    const { container } = render(<StatusBadge>active</StatusBadge>);
    const innerDots = container.querySelectorAll('.h-1\\.5');
    expect(innerDots.length).toBe(0);
  });

  it('has rounded-full class', () => {
    render(<StatusBadge>test</StatusBadge>);
    expect(screen.getByText('test').className).toContain('rounded-full');
  });

  it('has ring-1 ring-inset styling', () => {
    render(<StatusBadge variant="blue">test</StatusBadge>);
    expect(screen.getByText('test').className).toContain('ring-1');
    expect(screen.getByText('test').className).toContain('ring-inset');
  });

  it('applies custom className', () => {
    render(<StatusBadge className="ml-2">custom</StatusBadge>);
    expect(screen.getByText('custom').className).toContain('ml-2');
  });

  it('defaults unknown status to gray', () => {
    render(<StatusBadge>unknown_status_xyz</StatusBadge>);
    const badge = screen.getByText('unknown_status_xyz');
    expect(badge.className).toContain('gray');
  });
});

describe('Alert Component', () => {
  it('renders with role=alert', () => {
    render(<Alert variant="error">Error message</Alert>);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('renders error variant', () => {
    render(<Alert variant="error">Oops</Alert>);
    expect(screen.getByText('Oops')).toBeInTheDocument();
  });

  it('renders success variant', () => {
    render(<Alert variant="success">Done!</Alert>);
    expect(screen.getByText('Done!')).toBeInTheDocument();
  });

  it('renders warning variant', () => {
    render(<Alert variant="warning">Careful</Alert>);
    expect(screen.getByText('Careful')).toBeInTheDocument();
  });

  it('renders info variant', () => {
    render(<Alert variant="info">FYI</Alert>);
    expect(screen.getByText('FYI')).toBeInTheDocument();
  });

  it('has aria-live=polite', () => {
    render(<Alert variant="info">Live</Alert>);
    expect(screen.getByRole('alert')).toHaveAttribute('aria-live', 'polite');
  });

  it('renders dismiss button when onDismiss provided', () => {
    const onDismiss = vi.fn();
    render(<Alert variant="info" onDismiss={onDismiss}>Dismissible</Alert>);
    const dismissBtn = screen.getByLabelText('Dismiss');
    expect(dismissBtn).toBeInTheDocument();
  });

  it('calls onDismiss when dismiss button is clicked', () => {
    const onDismiss = vi.fn();
    render(<Alert variant="error" onDismiss={onDismiss}>Bye</Alert>);
    fireEvent.click(screen.getByLabelText('Dismiss'));
    expect(onDismiss).toHaveBeenCalledTimes(1);
  });

  it('does not render dismiss button when onDismiss is not provided', () => {
    render(<Alert variant="info">No dismiss</Alert>);
    expect(screen.queryByLabelText('Dismiss')).not.toBeInTheDocument();
  });

  it('renders action slot', () => {
    render(<Alert variant="warning" action={<button>Undo</button>}>With action</Alert>);
    expect(screen.getByRole('button', { name: 'Undo' })).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<Alert variant="info" className="mt-4">Custom</Alert>);
    expect(screen.getByRole('alert').className).toContain('mt-4');
  });

  it('renders icon SVG', () => {
    const { container } = render(<Alert variant="error">Has icon</Alert>);
    expect(container.querySelector('svg')).toBeInTheDocument();
  });
});
