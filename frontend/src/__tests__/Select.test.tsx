import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';

function TestSelect({ value, onChange }: { value?: string; onChange?: (v: string) => void }) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger>
        <SelectValue placeholder="Pick one" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="apple">Apple</SelectItem>
        <SelectItem value="banana">Banana</SelectItem>
        <SelectItem value="cherry">Cherry</SelectItem>
      </SelectContent>
    </Select>
  );
}

describe('Select Component', () => {
  it('renders trigger with placeholder', () => {
    render(<TestSelect />);
    expect(screen.getByText('Pick one')).toBeInTheDocument();
  });

  it('has combobox role on trigger', () => {
    render(<TestSelect />);
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('has aria-expanded=false initially', () => {
    render(<TestSelect />);
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-expanded', 'false');
  });

  it('has aria-haspopup=listbox', () => {
    render(<TestSelect />);
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-haspopup', 'listbox');
  });

  it('opens dropdown on click', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    expect(screen.getByRole('listbox')).toBeInTheDocument();
    expect(screen.getByText('Apple')).toBeInTheDocument();
  });

  it('sets aria-expanded=true when open', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-expanded', 'true');
  });

  it('opens with Enter key', () => {
    render(<TestSelect />);
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('opens with Space key', () => {
    render(<TestSelect />);
    fireEvent.keyDown(screen.getByRole('combobox'), { key: ' ' });
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('closes with Escape key on trigger', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    expect(screen.getByRole('listbox')).toBeInTheDocument();
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Escape' });
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('closes when clicking outside (backdrop)', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    expect(screen.getByRole('listbox')).toBeInTheDocument();
    // Click the backdrop (fixed overlay)
    const backdrop = document.querySelector('.fixed.inset-0.z-40');
    fireEvent.click(backdrop!);
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('selects an item on click', () => {
    const onChange = vi.fn();
    render(<TestSelect onChange={onChange} />);
    fireEvent.click(screen.getByRole('combobox'));
    fireEvent.click(screen.getByText('Banana'));
    expect(onChange).toHaveBeenCalledWith('banana');
  });

  it('closes dropdown after selection', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    fireEvent.click(screen.getByText('Apple'));
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('displays selected value', () => {
    render(<TestSelect value="cherry" />);
    expect(screen.getByText('cherry')).toBeInTheDocument();
  });

  it('items have role=option', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    expect(options).toHaveLength(3);
  });

  it('selected item has aria-selected=true', () => {
    render(<TestSelect value="banana" />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    const bananaOption = options.find((opt) => opt.textContent === 'Banana');
    expect(bananaOption).toHaveAttribute('aria-selected', 'true');
  });

  it('non-selected items have aria-selected=false', () => {
    render(<TestSelect value="banana" />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    const appleOption = options.find((opt) => opt.textContent === 'Apple');
    expect(appleOption).toHaveAttribute('aria-selected', 'false');
  });

  it('items are keyboard-selectable with Enter', () => {
    const onChange = vi.fn();
    render(<TestSelect onChange={onChange} />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    fireEvent.keyDown(options[2], { key: 'Enter' });
    expect(onChange).toHaveBeenCalledWith('cherry');
  });

  it('items are keyboard-selectable with Space', () => {
    const onChange = vi.fn();
    render(<TestSelect onChange={onChange} />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    fireEvent.keyDown(options[0], { key: ' ' });
    expect(onChange).toHaveBeenCalledWith('apple');
  });

  it('items have tabIndex=0 for focusability', () => {
    render(<TestSelect />);
    fireEvent.click(screen.getByRole('combobox'));
    const options = screen.getAllByRole('option');
    options.forEach((opt) => {
      expect(opt).toHaveAttribute('tabindex', '0');
    });
  });

  it('toggles open/close on repeated clicks', () => {
    render(<TestSelect />);
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);
    expect(screen.getByRole('listbox')).toBeInTheDocument();
    fireEvent.click(trigger);
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('trigger has type=button to prevent form submission', () => {
    render(<TestSelect />);
    expect(screen.getByRole('combobox')).toHaveAttribute('type', 'button');
  });
});
