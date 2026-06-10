import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect } from 'vitest';
import { Button } from '@/components/ui/button';

describe('Button Component', () => {
  it('renders with default variant', () => {
    render(<Button>Click me</Button>);
    const btn = screen.getByRole('button', { name: 'Click me' });
    expect(btn).toBeInTheDocument();
    expect(btn.className).toContain('bg-primary');
  });

  it('renders destructive variant', () => {
    render(<Button variant="destructive">Delete</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('bg-destructive');
  });

  it('renders outline variant', () => {
    render(<Button variant="outline">Outline</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('border');
    expect(btn.className).toContain('bg-white');
  });

  it('renders secondary variant', () => {
    render(<Button variant="secondary">Secondary</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('bg-secondary');
  });

  it('renders ghost variant', () => {
    render(<Button variant="ghost">Ghost</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('hover:bg-gray-100');
  });

  it('renders link variant', () => {
    render(<Button variant="link">Link</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('underline-offset-4');
  });

  it('renders success variant', () => {
    render(<Button variant="success">Success</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('bg-emerald-600');
  });

  it('renders warning variant', () => {
    render(<Button variant="warning">Warning</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('bg-orange-600');
  });

  it('renders default size', () => {
    render(<Button>Sized</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('h-9');
    expect(btn.className).toContain('px-4');
  });

  it('renders sm size', () => {
    render(<Button size="sm">Small</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('h-8');
  });

  it('renders lg size', () => {
    render(<Button size="lg">Large</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('h-11');
  });

  it('renders icon size', () => {
    render(<Button size="icon">I</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('h-9');
    expect(btn.className).toContain('w-9');
  });

  it('renders xs size', () => {
    render(<Button size="xs">Tiny</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('h-7');
  });

  it('shows loading spinner when loading', () => {
    render(<Button loading>Loading</Button>);
    const btn = screen.getByRole('button');
    const spinner = btn.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('is disabled when loading', () => {
    render(<Button loading>Loading</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('is disabled when disabled prop is set', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('applies custom className', () => {
    render(<Button className="custom-class">Custom</Button>);
    expect(screen.getByRole('button').className).toContain('custom-class');
  });

  it('handles click events', async () => {
    const user = userEvent.setup();
    let clicked = false;
    render(<Button onClick={() => { clicked = true; }}>Click</Button>);
    await user.click(screen.getByRole('button'));
    expect(clicked).toBe(true);
  });

  it('does not fire click when disabled', async () => {
    const user = userEvent.setup();
    let clicked = false;
    render(<Button disabled onClick={() => { clicked = true; }}>Click</Button>);
    await user.click(screen.getByRole('button'));
    expect(clicked).toBe(false);
  });

  it('has focus-visible ring classes', () => {
    render(<Button>Focus</Button>);
    const btn = screen.getByRole('button');
    expect(btn.className).toContain('focus-visible:ring-2');
  });

  it('has rounded-lg border radius', () => {
    render(<Button>Round</Button>);
    expect(screen.getByRole('button').className).toContain('rounded-lg');
  });

  it('renders children correctly', () => {
    render(<Button><span data-testid="child">Inner</span></Button>);
    expect(screen.getByTestId('child')).toBeInTheDocument();
  });

  it('forwards ref', () => {
    const ref = { current: null as HTMLButtonElement | null };
    render(<Button ref={ref}>Ref</Button>);
    expect(ref.current).toBeInstanceOf(HTMLButtonElement);
  });

  it('has transition-colors class', () => {
    render(<Button>Trans</Button>);
    expect(screen.getByRole('button').className).toContain('transition-colors');
  });
});
