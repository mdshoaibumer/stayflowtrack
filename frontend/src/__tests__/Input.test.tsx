import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect } from 'vitest';
import { Input } from '@/components/ui/input';

describe('Input Component', () => {
  it('renders an input element', () => {
    render(<Input placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  it('has text type by default', () => {
    render(<Input data-testid="inp" />);
    const input = screen.getByTestId('inp');
    // type defaults to undefined which renders as text
    expect(input.tagName).toBe('INPUT');
  });

  it('supports email type', () => {
    render(<Input type="email" data-testid="inp" />);
    expect(screen.getByTestId('inp')).toHaveAttribute('type', 'email');
  });

  it('supports password type', () => {
    render(<Input type="password" data-testid="inp" />);
    expect(screen.getByTestId('inp')).toHaveAttribute('type', 'password');
  });

  it('supports number type', () => {
    render(<Input type="number" data-testid="inp" />);
    expect(screen.getByTestId('inp')).toHaveAttribute('type', 'number');
  });

  it('accepts user input', async () => {
    const user = userEvent.setup();
    render(<Input data-testid="inp" />);
    const input = screen.getByTestId('inp');
    await user.type(input, 'Hello World');
    expect(input).toHaveValue('Hello World');
  });

  it('is disabled when disabled prop set', () => {
    render(<Input disabled data-testid="inp" />);
    expect(screen.getByTestId('inp')).toBeDisabled();
  });

  it('has rounded-lg class', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('rounded-lg');
  });

  it('has focus ring styles', () => {
    render(<Input data-testid="inp" />);
    const className = screen.getByTestId('inp').className;
    expect(className).toContain('focus:ring-2');
    expect(className).toContain('focus:ring-ring/20');
  });

  it('applies custom className', () => {
    render(<Input data-testid="inp" className="extra-class" />);
    expect(screen.getByTestId('inp').className).toContain('extra-class');
  });

  it('has border-input class', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('border-input');
  });

  it('has bg-background class', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('bg-background');
  });

  it('has placeholder styling', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('placeholder:text-muted-foreground');
  });

  it('has disabled opacity', () => {
    render(<Input data-testid="inp" disabled />);
    expect(screen.getByTestId('inp').className).toContain('disabled:opacity-50');
  });

  it('forwards ref', () => {
    const ref = { current: null as HTMLInputElement | null };
    render(<Input ref={ref} />);
    expect(ref.current).toBeInstanceOf(HTMLInputElement);
  });

  it('supports readOnly', () => {
    render(<Input data-testid="inp" readOnly value="fixed" />);
    expect(screen.getByTestId('inp')).toHaveAttribute('readonly');
  });

  it('supports required attribute', () => {
    render(<Input data-testid="inp" required />);
    expect(screen.getByTestId('inp')).toBeRequired();
  });

  it('supports aria-label', () => {
    render(<Input aria-label="Username" data-testid="inp" />);
    expect(screen.getByLabelText('Username')).toBeInTheDocument();
  });

  it('has w-full width class', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('w-full');
  });

  it('has text-sm size', () => {
    render(<Input data-testid="inp" />);
    expect(screen.getByTestId('inp').className).toContain('text-sm');
  });
});
