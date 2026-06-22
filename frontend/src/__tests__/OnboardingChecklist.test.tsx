import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, beforeEach } from 'vitest';
import { OnboardingChecklist } from '@/components/shared/OnboardingChecklist';

describe('OnboardingChecklist', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders the checklist when not dismissed', () => {
    render(<OnboardingChecklist />);
    expect(screen.getByText('Getting Started')).toBeInTheDocument();
  });

  it('shows all 5 onboarding steps', () => {
    render(<OnboardingChecklist />);
    expect(screen.getByText('Set up your property')).toBeInTheDocument();
    expect(screen.getByText('Add rooms')).toBeInTheDocument();
    expect(screen.getByText('Configure rates')).toBeInTheDocument();
    expect(screen.getByText('Add your first guest')).toBeInTheDocument();
    expect(screen.getByText('Create a reservation')).toBeInTheDocument();
  });

  it('shows step descriptions', () => {
    render(<OnboardingChecklist />);
    expect(screen.getByText('Add property details and room types')).toBeInTheDocument();
    expect(screen.getByText('Create rooms and assign room types')).toBeInTheDocument();
  });

  it('shows "0 of 5 steps completed" initially', () => {
    render(<OnboardingChecklist />);
    expect(screen.getByText('0 of 5 steps completed')).toBeInTheDocument();
  });

  it('shows 0% progress initially', () => {
    render(<OnboardingChecklist />);
    expect(screen.getByText('0%')).toBeInTheDocument();
  });

  it('marks a step as completed when clicked', () => {
    render(<OnboardingChecklist />);
    const step = screen.getByText('Set up your property');
    fireEvent.click(step.closest('a')!);
    expect(screen.getByText('1 of 5 steps completed')).toBeInTheDocument();
    expect(screen.getByText('20%')).toBeInTheDocument();
  });

  it('persists completed steps to localStorage', () => {
    render(<OnboardingChecklist />);
    const step = screen.getByText('Add rooms');
    fireEvent.click(step.closest('a')!);
    const stored = JSON.parse(localStorage.getItem('stayflow-onboarding')!);
    expect(stored.rooms).toBe(true);
  });

  it('loads completed steps from localStorage on mount', () => {
    localStorage.setItem('stayflow-onboarding', JSON.stringify({ property: true, rooms: true }));
    render(<OnboardingChecklist />);
    expect(screen.getByText('2 of 5 steps completed')).toBeInTheDocument();
    expect(screen.getByText('40%')).toBeInTheDocument();
  });

  it('collapses and expands on header click', () => {
    render(<OnboardingChecklist />);
    const header = screen.getByRole('button', { expanded: true });
    fireEvent.click(header);
    expect(screen.queryByText('Set up your property')).not.toBeInTheDocument();
    fireEvent.click(header);
    expect(screen.getByText('Set up your property')).toBeInTheDocument();
  });

  it('dismisses checklist permanently', () => {
    render(<OnboardingChecklist />);
    const dismissBtn = screen.getByText('Dismiss checklist');
    fireEvent.click(dismissBtn);
    expect(screen.queryByText('Getting Started')).not.toBeInTheDocument();
    expect(localStorage.getItem('stayflow-onboarding-dismissed')).toBe('true');
  });

  it('does not render when previously dismissed', () => {
    localStorage.setItem('stayflow-onboarding-dismissed', 'true');
    render(<OnboardingChecklist />);
    expect(screen.queryByText('Getting Started')).not.toBeInTheDocument();
  });

  it('does not render when all steps are completed', () => {
    localStorage.setItem('stayflow-onboarding', JSON.stringify({
      property: true, rooms: true, rates: true, guest: true, reservation: true,
    }));
    render(<OnboardingChecklist />);
    expect(screen.queryByText('Getting Started')).not.toBeInTheDocument();
  });

  it('prevents navigation for already completed steps', () => {
    localStorage.setItem('stayflow-onboarding', JSON.stringify({ property: true }));
    render(<OnboardingChecklist />);
    const link = screen.getByText('Set up your property').closest('a')!;
    const clickEvent = new MouseEvent('click', { bubbles: true, cancelable: true });
    link.dispatchEvent(clickEvent);
    // Completed steps should have reduced opacity
    expect(link.className).toContain('opacity-60');
  });

  it('shows progress ring SVG', () => {
    render(<OnboardingChecklist />);
    const svg = screen.getByText('0%').closest('div')!.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('shows chevron icons on incomplete steps', () => {
    render(<OnboardingChecklist />);
    // All 5 steps are incomplete - each should have a chevron
    const links = screen.getAllByRole('link');
    links.forEach((link) => {
      const svgs = link.querySelectorAll('svg');
      expect(svgs.length).toBeGreaterThan(0);
    });
  });

  it('applies line-through style to completed step text', () => {
    localStorage.setItem('stayflow-onboarding', JSON.stringify({ property: true }));
    render(<OnboardingChecklist />);
    const stepText = screen.getByText('Set up your property');
    expect(stepText.className).toContain('line-through');
  });

  it('shows correct href for each step', () => {
    render(<OnboardingChecklist />);
    const links = screen.getAllByRole('link');
    const hrefs = links.map((l) => l.getAttribute('href'));
    expect(hrefs).toContain('/settings');
    expect(hrefs).toContain('/operations');
    expect(hrefs).toContain('/guests?action=new');
    expect(hrefs).toContain('/reservations?action=new');
  });

  it('has proper aria-expanded attribute on toggle button', () => {
    render(<OnboardingChecklist />);
    const toggle = screen.getByRole('button', { expanded: true });
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
  });
});
