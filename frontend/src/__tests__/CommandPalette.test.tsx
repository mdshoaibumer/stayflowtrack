import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CommandPalette from '@/components/shared/CommandPalette';

const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush, replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
}));

describe('CommandPalette', () => {
  beforeEach(() => {
    mockPush.mockClear();
  });

  it('does not render when closed', () => {
    render(<CommandPalette />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('opens with Ctrl+K keyboard shortcut', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('opens with Meta+K (Mac) keyboard shortcut', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true }));
    });
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('toggles closed when pressing Ctrl+K again', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('closes when Escape is pressed', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    fireEvent.keyDown(input, { key: 'Escape' });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('closes when backdrop is clicked', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const backdrop = screen.getByRole('dialog').querySelector('[aria-hidden="true"]');
    fireEvent.click(backdrop!);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders search input with placeholder', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
  });

  it('shows all commands initially', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByText('Go to Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Go to Reservations')).toBeInTheDocument();
    expect(screen.getByText('Go to Guests')).toBeInTheDocument();
    expect(screen.getByText('New Reservation')).toBeInTheDocument();
  });

  it('shows Navigation and Actions sections', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByText('Navigation')).toBeInTheDocument();
    expect(screen.getByText('Actions')).toBeInTheDocument();
  });

  it('filters commands based on search query', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await userEvent.type(input, 'billing');
    expect(screen.getByText('Go to Billing')).toBeInTheDocument();
    expect(screen.queryByText('Go to Dashboard')).not.toBeInTheDocument();
  });

  it('shows empty state when no results match', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await userEvent.type(input, 'xyznonexistent');
    expect(screen.getByText(/No results found/)).toBeInTheDocument();
  });

  it('navigates items with ArrowDown', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await act(async () => {
      fireEvent.keyDown(input, { key: 'ArrowDown' });
    });
    const items = screen.getAllByRole('option');
    expect(items[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('navigates items with ArrowUp', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await act(async () => {
      fireEvent.keyDown(input, { key: 'ArrowDown' });
      fireEvent.keyDown(input, { key: 'ArrowDown' });
      fireEvent.keyDown(input, { key: 'ArrowUp' });
    });
    const items = screen.getAllByRole('option');
    expect(items[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('does not go below 0 index with ArrowUp at start', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    const items = screen.getAllByRole('option');
    expect(items[0]).toHaveAttribute('aria-selected', 'true');
  });

  it('executes action on Enter key', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(mockPush).toHaveBeenCalledWith('/dashboard');
  });

  it('executes action on item click', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    fireEvent.click(screen.getByText('Go to Guests'));
    expect(mockPush).toHaveBeenCalledWith('/guests');
  });

  it('closes after executing an action', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    fireEvent.click(screen.getByText('Go to Guests'));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('updates selected index on mouse enter', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const items = screen.getAllByRole('option');
    await act(async () => {
      fireEvent.mouseEnter(items[3]);
    });
    expect(items[3]).toHaveAttribute('aria-selected', 'true');
  });

  it('resets query and selected index when reopened', () => {
    render(<CommandPalette />);
    // Open and type something
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    fireEvent.change(input, { target: { value: 'billing' } });
    // Close
    fireEvent.keyDown(input, { key: 'Escape' });
    // Reopen
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const newInput = screen.getByLabelText('Search commands');
    expect(newInput).toHaveValue('');
  });

  it('resets selected index when query changes', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await act(async () => {
      fireEvent.keyDown(input, { key: 'ArrowDown' });
      fireEvent.keyDown(input, { key: 'ArrowDown' });
    });
    // Now type to filter - should reset
    await userEvent.type(input, 'guest');
    const items = screen.getAllByRole('option');
    expect(items[0]).toHaveAttribute('aria-selected', 'true');
  });

  it('displays keyboard shortcut hints in footer', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByText('navigate')).toBeInTheDocument();
    expect(screen.getByText('select')).toBeInTheDocument();
    expect(screen.getByText('close')).toBeInTheDocument();
  });

  it('has proper aria attributes on dialog', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-label', 'Command palette');
  });

  it('renders listbox role on results container', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('shows shortcut badge on actions', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByText('⌘N')).toBeInTheDocument();
  });

  it('shows description text for action items', () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    expect(screen.getByText('Create a new booking')).toBeInTheDocument();
  });

  it('filters by description text', async () => {
    render(<CommandPalette />);
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }));
    });
    const input = screen.getByLabelText('Search commands');
    await userEvent.type(input, 'walk-in');
    expect(screen.getByText('Walk-In Guest')).toBeInTheDocument();
  });

  it('cleans up event listener on unmount', () => {
    const removeEventSpy = vi.spyOn(document, 'removeEventListener');
    const { unmount } = render(<CommandPalette />);
    unmount();
    expect(removeEventSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    removeEventSpy.mockRestore();
  });
});
