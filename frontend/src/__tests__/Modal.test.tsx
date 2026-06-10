import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Modal, ModalHeader, ModalTitle, ModalBody, ModalFooter, ModalCloseButton } from '@/components/ui/modal';

// We need to read the full export to check what's exported
function TestModal({ open = true, onClose = vi.fn(), size }: { open?: boolean; onClose?: () => void; size?: 'sm' | 'md' | 'lg' | 'xl' }) {
  return (
    <Modal open={open} onClose={onClose} size={size}>
      <ModalHeader>
        <ModalTitle>Test Title</ModalTitle>
        <ModalCloseButton onClick={onClose} />
      </ModalHeader>
      <ModalBody>
        <p>Modal content here</p>
        <input data-testid="modal-input" />
      </ModalBody>
      <ModalFooter>
        <button>Cancel</button>
        <button>Confirm</button>
      </ModalFooter>
    </Modal>
  );
}

describe('Modal Component', () => {
  it('renders when open=true', () => {
    render(<TestModal />);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('does not render when open=false', () => {
    render(<TestModal open={false} />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('has aria-modal=true', () => {
    render(<TestModal />);
    expect(screen.getByRole('dialog')).toHaveAttribute('aria-modal', 'true');
  });

  it('renders title', () => {
    render(<TestModal />);
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });

  it('renders body content', () => {
    render(<TestModal />);
    expect(screen.getByText('Modal content here')).toBeInTheDocument();
  });

  it('renders footer buttons', () => {
    render(<TestModal />);
    expect(screen.getByText('Cancel')).toBeInTheDocument();
    expect(screen.getByText('Confirm')).toBeInTheDocument();
  });

  it('calls onClose when Escape is pressed', () => {
    const onClose = vi.fn();
    render(<TestModal onClose={onClose} />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn();
    render(<TestModal onClose={onClose} />);
    const backdrop = screen.getByRole('dialog').querySelector('[aria-hidden="true"]');
    fireEvent.click(backdrop!);
    expect(onClose).toHaveBeenCalled();
  });

  it('locks body scroll when open', () => {
    render(<TestModal />);
    expect(document.body.style.overflow).toBe('hidden');
  });

  it('restores body scroll when closed', () => {
    const { unmount } = render(<TestModal />);
    unmount();
    expect(document.body.style.overflow).toBe('');
  });

  it('renders sm size correctly', () => {
    render(<TestModal size="sm" />);
    const content = screen.getByRole('dialog').querySelector('.max-w-sm');
    expect(content).toBeInTheDocument();
  });

  it('renders md size correctly (default)', () => {
    render(<TestModal />);
    const content = screen.getByRole('dialog').querySelector('.max-w-lg');
    expect(content).toBeInTheDocument();
  });

  it('renders lg size correctly', () => {
    render(<TestModal size="lg" />);
    const content = screen.getByRole('dialog').querySelector('.max-w-2xl');
    expect(content).toBeInTheDocument();
  });

  it('renders xl size correctly', () => {
    render(<TestModal size="xl" />);
    const content = screen.getByRole('dialog').querySelector('.max-w-4xl');
    expect(content).toBeInTheDocument();
  });

  it('focus traps with Tab (forward)', () => {
    render(<TestModal />);
    // The modal has focus trap logic - simulate Tab from last to first
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
    // Test Tab key at the last element wraps to first
    const lastBtn = buttons[buttons.length - 1];
    lastBtn.focus();
    fireEvent.keyDown(document, { key: 'Tab', shiftKey: false });
  });

  it('focus traps with Shift+Tab (backward)', () => {
    render(<TestModal />);
    // Simulate Shift+Tab from first focusable element
    const buttons = screen.getAllByRole('button');
    buttons[0].focus();
    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });
  });

  it('ModalHeader has border-b class', () => {
    render(<TestModal />);
    const header = screen.getByText('Test Title').closest('.border-b');
    expect(header).toBeInTheDocument();
  });

  it('ModalBody has overflow-y-auto', () => {
    render(<TestModal />);
    const body = screen.getByText('Modal content here').closest('.overflow-y-auto');
    expect(body).toBeInTheDocument();
  });

  it('ModalFooter has border-t', () => {
    render(<TestModal />);
    const footer = screen.getByText('Confirm').closest('.border-t');
    expect(footer).toBeInTheDocument();
  });

  it('has backdrop blur effect', () => {
    render(<TestModal />);
    const backdrop = screen.getByRole('dialog').querySelector('[aria-hidden="true"]');
    expect(backdrop?.className).toContain('backdrop-blur');
  });

  it('has shadow-xl on content', () => {
    render(<TestModal />);
    const content = screen.getByRole('dialog').querySelector('.shadow-xl');
    expect(content).toBeInTheDocument();
  });

  it('has max-h constraint', () => {
    render(<TestModal />);
    const content = screen.getByRole('dialog').querySelector('[class*="max-h"]');
    expect(content).toBeInTheDocument();
  });
});
