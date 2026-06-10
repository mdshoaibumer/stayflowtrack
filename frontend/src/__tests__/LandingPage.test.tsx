import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock useAuth
const mockUseAuth = vi.fn();
vi.mock('@/contexts/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}));

// Must import AFTER mocks
import Home from '@/app/page';

const mockReplace = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: mockReplace, back: vi.fn(), prefetch: vi.fn() }),
  usePathname: () => '/',
  useSearchParams: () => new URLSearchParams(),
}));

describe('Landing Page (Home)', () => {
  beforeEach(() => {
    mockReplace.mockClear();
  });

  describe('when loading', () => {
    it('shows loading spinner', () => {
      mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: true });
      render(<Home />);
      const spinner = document.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });

    it('does not show landing page content', () => {
      mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: true });
      render(<Home />);
      expect(screen.queryByText(/Property operations,/)).not.toBeInTheDocument();
    });
  });

  describe('when authenticated', () => {
    it('redirects to /dashboard', () => {
      mockUseAuth.mockReturnValue({ isAuthenticated: true, isLoading: false });
      render(<Home />);
      expect(mockReplace).toHaveBeenCalledWith('/dashboard');
    });

    it('does not show landing page', () => {
      mockUseAuth.mockReturnValue({ isAuthenticated: true, isLoading: false });
      render(<Home />);
      expect(screen.queryByText(/Property operations,/)).not.toBeInTheDocument();
    });
  });

  describe('when not authenticated (landing page)', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: false });
    });

    it('renders hero heading', () => {
      render(<Home />);
      expect(screen.getByText(/Property operations,/)).toBeInTheDocument();
      expect(screen.getByText(/finally simple/)).toBeInTheDocument();
    });

    it('renders navigation bar with logo', () => {
      render(<Home />);
      // StayFlow appears multiple times (nav + footer), just check it exists
      const stayflows = screen.getAllByText('StayFlow');
      expect(stayflows.length).toBeGreaterThanOrEqual(1);
    });

    it('renders Sign in link', () => {
      render(<Home />);
      const signIns = screen.getAllByText('Sign in');
      expect(signIns.length).toBeGreaterThanOrEqual(1);
    });

    it('renders Start Free Trial button in nav', () => {
      render(<Home />);
      const ctas = screen.getAllByText('Start Free Trial');
      expect(ctas.length).toBeGreaterThanOrEqual(1);
    });

    it('renders hero subtitle', () => {
      render(<Home />);
      expect(screen.getByText(/Reservations, billing, housekeeping, laundry/)).toBeInTheDocument();
    });

    it('renders social proof metrics', () => {
      render(<Home />);
      expect(screen.getAllByText('50+').length).toBeGreaterThanOrEqual(1);
    });

    it('renders Problem → Solution section', () => {
      render(<Home />);
      expect(screen.getByText('Without StayFlow')).toBeInTheDocument();
      expect(screen.getByText('With StayFlow')).toBeInTheDocument();
    });

    it('renders problem items', () => {
      render(<Home />);
      expect(screen.getByText('Bookings scattered across channels')).toBeInTheDocument();
      expect(screen.getByText('Manual billing with frequent errors')).toBeInTheDocument();
    });

    it('renders solution items', () => {
      render(<Home />);
      expect(screen.getByText('Unified reservation calendar')).toBeInTheDocument();
      expect(screen.getByText(/Automated GST-compliant invoicing/)).toBeInTheDocument();
    });

    it('renders Features section', () => {
      render(<Home />);
      expect(screen.getByText('Everything you need to run your property')).toBeInTheDocument();
    });

    it('renders all 6 features', () => {
      render(<Home />);
      expect(screen.getByText('Reservations')).toBeInTheDocument();
      expect(screen.getByText('Billing & Invoicing')).toBeInTheDocument();
      expect(screen.getByText('Operations')).toBeInTheDocument();
      expect(screen.getByText('Housekeeping')).toBeInTheDocument();
      expect(screen.getByText('Laundry')).toBeInTheDocument();
      expect(screen.getByText('Reports & Analytics')).toBeInTheDocument();
    });

    it('renders Pricing section with 3 plans', () => {
      render(<Home />);
      expect(screen.getByText('Starter')).toBeInTheDocument();
      expect(screen.getByText('Professional')).toBeInTheDocument();
      expect(screen.getByText('Enterprise')).toBeInTheDocument();
    });

    it('renders pricing amounts', () => {
      render(<Home />);
      expect(screen.getByText('₹2,999')).toBeInTheDocument();
      expect(screen.getByText('₹5,999')).toBeInTheDocument();
      expect(screen.getByText('Custom')).toBeInTheDocument();
    });

    it('highlights Professional as Most Popular', () => {
      render(<Home />);
      expect(screen.getByText('Most Popular')).toBeInTheDocument();
    });

    it('renders FAQ section', () => {
      render(<Home />);
      expect(screen.getByText('Frequently asked questions')).toBeInTheDocument();
    });

    it('renders FAQ questions', () => {
      render(<Home />);
      expect(screen.getByText('Is there a free trial?')).toBeInTheDocument();
      expect(screen.getByText('Can I import my existing data?')).toBeInTheDocument();
      expect(screen.getByText('Is it GST compliant?')).toBeInTheDocument();
      expect(screen.getByText('Does it work on mobile?')).toBeInTheDocument();
      expect(screen.getByText('Can I manage multiple properties?')).toBeInTheDocument();
    });

    it('FAQ items expand on click', () => {
      render(<Home />);
      const faqBtn = screen.getByText('Is there a free trial?');
      fireEvent.click(faqBtn);
      expect(screen.getByText(/14-day free trial/)).toBeInTheDocument();
    });

    it('FAQ items collapse on second click (via max-h-0)', () => {
      render(<Home />);
      const faqBtn = screen.getByText('Is there a free trial?');
      fireEvent.click(faqBtn);
      // Content is in the DOM but hidden via max-h-0 after second click
      fireEvent.click(faqBtn);
      const answerContainer = screen.getByText(/14-day free trial/).closest('[aria-hidden]');
      expect(answerContainer).toHaveAttribute('aria-hidden', 'true');
    });

    it('FAQ button has aria-expanded attribute', () => {
      render(<Home />);
      const faqBtn = screen.getByText('Is there a free trial?').closest('button')!;
      expect(faqBtn).toHaveAttribute('aria-expanded', 'false');
      fireEvent.click(faqBtn);
      expect(faqBtn).toHaveAttribute('aria-expanded', 'true');
    });

    it('renders final CTA section', () => {
      render(<Home />);
      expect(screen.getByText('Ready to simplify your operations?')).toBeInTheDocument();
    });

    it('renders footer with copyright', () => {
      render(<Home />);
      expect(screen.getByText(/© \d{4} StayFlow/)).toBeInTheDocument();
    });

    it('renders navigation links', () => {
      render(<Home />);
      // Features appears in both nav and footer
      const featuresLinks = screen.getAllByText('Features');
      expect(featuresLinks.length).toBeGreaterThanOrEqual(1);
      const pricingLinks = screen.getAllByText('Pricing');
      expect(pricingLinks.length).toBeGreaterThanOrEqual(1);
    });

    it('nav links have correct href anchors', () => {
      render(<Home />);
      const featuresLinks = screen.getAllByText('Features');
      const featuresLink = featuresLinks[0].closest('a');
      expect(featuresLink).toHaveAttribute('href', '#features');
    });

    it('pricing cards list features', () => {
      render(<Home />);
      expect(screen.getByText('Up to 20 rooms')).toBeInTheDocument();
      expect(screen.getByText('Up to 50 rooms')).toBeInTheDocument();
      expect(screen.getByText('Unlimited rooms')).toBeInTheDocument();
    });

    it('renders Contact Sales for enterprise plan', () => {
      render(<Home />);
      expect(screen.getByText('Contact Sales')).toBeInTheDocument();
    });

    it('renders badge text', () => {
      render(<Home />);
      expect(screen.getByText(/Built for service apartments/)).toBeInTheDocument();
    });

    it('has navigation role on nav element', () => {
      render(<Home />);
      expect(screen.getByRole('navigation', { name: 'Main navigation' })).toBeInTheDocument();
    });

    it('renders testimonials section', () => {
      render(<Home />);
      expect(screen.getByText(/Trusted by property managers across India/)).toBeInTheDocument();
    });

    it('renders product showcase section', () => {
      render(<Home />);
      expect(screen.getByText('One dashboard. Complete visibility.')).toBeInTheDocument();
    });

    it('renders metrics strip', () => {
      render(<Home />);
      expect(screen.getByText('Properties Managed')).toBeInTheDocument();
      expect(screen.getByText('99.9%')).toBeInTheDocument();
    });
  });
});
