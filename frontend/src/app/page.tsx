"use client";

import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function Home() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const [showLanding, setShowLanding] = useState(false);

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        router.replace("/dashboard");
      } else {
        setShowLanding(true);
      }
    }
  }, [isAuthenticated, isLoading, router]);

  if (!showLanding) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="w-8 h-8 rounded-full border-[3px] border-muted border-t-primary animate-spin" role="status" aria-label="Loading" />
      </div>
    );
  }

  return <LandingPage />;
}

/* ─── Landing Page ─────────────────────────────────── */

function LandingPage() {
  return (
    <div className="min-h-screen bg-background text-foreground overflow-x-hidden">
      {/* Navbar */}
      <nav className="sticky top-0 z-50 backdrop-blur-xl bg-background/70 border-b border-border/30 shadow-sm" role="navigation" aria-label="Main navigation">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 flex items-center justify-center" aria-hidden="true">
              <img src="/logo.png" alt="StayFlow Logo" className="w-full h-full object-contain" />
            </div>
            <span className="text-lg font-bold tracking-tight">StayFlow</span>
          </div>
          <div className="hidden md:flex items-center gap-8 text-sm text-muted-foreground">
            <a href="#features" className="hover:text-foreground transition-colors">Features</a>
            <a href="#how-it-works" className="hover:text-foreground transition-colors">How It Works</a>
            <a href="#pricing" className="hover:text-foreground transition-colors">Pricing</a>
            <a href="#faq" className="hover:text-foreground transition-colors">FAQ</a>
          </div>
          <div className="flex items-center gap-3">
            <Link href="/login" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors hidden sm:block">
              Sign in
            </Link>
            <Link href="/register" className="inline-flex items-center px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 shadow-sm transition-all hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2">
              Start Free Trial
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative overflow-hidden" aria-labelledby="hero-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16 lg:pt-32 lg:pb-28">
          <div className="max-w-3xl mx-auto text-center">
            {/* Badge */}
            <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full text-xs font-medium bg-accent/5 text-accent border border-accent/15 mb-8 shadow-sm animate-fade-in-up">
              <span className="w-1.5 h-1.5 rounded-full bg-accent animate-pulse-subtle" aria-hidden="true" />
              Built for service apartments &amp; boutique hotels
            </div>

            <h1 id="hero-heading" className="text-4xl sm:text-5xl lg:text-[3.75rem] font-bold tracking-tight leading-[1.08] animate-blur-in">
              Property operations,{" "}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary via-primary/80 to-accent">
                finally simple
              </span>
            </h1>

            <p className="mt-6 text-lg lg:text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
              Reservations, billing, housekeeping, laundry, and guest management in one cohesive platform.
              Stop managing chaos — start managing properties.
            </p>

            <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-3">
              <Link href="/register" className="group inline-flex items-center gap-2 px-7 py-3.5 text-sm font-semibold text-primary-foreground bg-primary rounded-xl hover:bg-primary/90 shadow-lg hover:shadow-xl transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 active:scale-[0.98]">
                Start Free Trial
                <svg className="w-4 h-4 transition-transform duration-200 group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </Link>
              <a href="#how-it-works" className="group inline-flex items-center gap-2 px-6 py-3 text-sm font-medium text-foreground bg-card border border-border rounded-xl hover:bg-muted hover:border-primary/20 shadow-sm transition-all duration-200 active:scale-[0.98]">
                <svg className="w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                See how it works
              </a>
            </div>

            {/* Social proof */}
            <div className="mt-14 flex flex-col items-center gap-4">
              <div className="flex -space-x-2.5" aria-hidden="true">
                {['bg-blue-100 text-blue-600', 'bg-emerald-100 text-emerald-600', 'bg-violet-100 text-violet-600', 'bg-amber-100 text-amber-600', 'bg-rose-100 text-rose-600'].map((cls, i) => (
                  <div key={i} className={`w-9 h-9 rounded-full ${cls} border-2 border-background flex items-center justify-center text-xs font-semibold`}>
                    {['AK', 'RS', 'PJ', 'VM', 'NK'][i]}
                  </div>
                ))}
              </div>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <span className="font-semibold text-foreground">50+</span> properties
                </span>
                <span className="w-1 h-1 rounded-full bg-border" aria-hidden="true" />
                <span className="flex items-center gap-1">
                  <svg className="w-4 h-4 text-amber-400" fill="currentColor" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" /></svg>
                  <span className="font-semibold text-foreground">4.9</span> average rating
                </span>
                <span className="w-1 h-1 rounded-full bg-border" aria-hidden="true" />
                <span>No credit card required</span>
              </div>
            </div>
          </div>
        </div>

        {/* Background gradient orbs — premium depth */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-[1400px] h-[800px] opacity-[0.12] pointer-events-none" aria-hidden="true">
          <div className="absolute top-16 left-[15%] w-96 h-96 rounded-full bg-primary blur-[140px] animate-float" />
          <div className="absolute top-40 right-[15%] w-80 h-80 rounded-full bg-accent blur-[130px] animate-float" style={{ animationDelay: "1.5s" }} />
          <div className="absolute top-64 left-[45%] w-64 h-64 rounded-full bg-violet-500 blur-[120px] animate-float" style={{ animationDelay: "3s" }} />
        </div>
      </section>

      {/* Metrics strip */}
      <section className="border-y bg-muted/30" aria-label="Key metrics">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {[
              { value: '50+', label: 'Properties Managed' },
              { value: '10K+', label: 'Reservations Processed' },
              { value: '99.9%', label: 'Uptime Guarantee' },
              { value: '2min', label: 'Average Check-in Time' },
            ].map((stat, i) => (
              <div key={i} className="text-center">
                <div className="text-2xl sm:text-3xl font-bold text-foreground">{stat.value}</div>
                <div className="text-sm text-muted-foreground mt-1">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Problem → Solution */}
      <section id="how-it-works" className="py-20 lg:py-28" aria-labelledby="problem-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 id="problem-heading" className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight">
              Replace spreadsheets, WhatsApp groups,<br className="hidden sm:block" /> and paper registers
            </h2>
            <p className="mt-4 text-lg text-muted-foreground max-w-2xl mx-auto">
              Most property managers waste hours daily on manual operations. StayFlow brings everything into one system of record.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-center">
            {/* Before */}
            <div className="rounded-2xl border border-destructive/20 bg-destructive/[0.02] p-6 space-y-4">
              <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-medium bg-destructive/10 text-destructive">
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                Without StayFlow
              </div>
              <ul className="space-y-3" aria-label="Problems without StayFlow">
                {['Bookings scattered across channels', 'Manual billing with frequent errors', 'Housekeeping coordinated via phone calls', 'No real-time revenue visibility', 'Guest data in multiple systems'].map((item, j) => (
                  <li key={j} className="flex items-start gap-2.5 text-sm text-muted-foreground">
                    <svg className="w-4 h-4 shrink-0 mt-0.5 text-destructive/60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                    {item}
                  </li>
                ))}
              </ul>
            </div>

            {/* Arrow */}
            <div className="hidden lg:flex items-center justify-center">
              <div className="flex flex-col items-center gap-3">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                  <svg className="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </div>
                <span className="text-xs font-medium text-muted-foreground">Switch to</span>
              </div>
            </div>

            {/* After */}
            <div className="rounded-2xl border border-emerald-200/60 bg-emerald-50/30 p-6 space-y-4">
              <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700">
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>
                With StayFlow
              </div>
              <ul className="space-y-3" aria-label="Benefits with StayFlow">
                {['Unified reservation calendar', 'Automated GST-compliant invoicing', 'Real-time task assignment & tracking', 'Live dashboard with key metrics', 'Single source of truth for all data'].map((item, j) => (
                  <li key={j} className="flex items-start gap-2.5 text-sm text-foreground">
                    <svg className="w-4 h-4 shrink-0 mt-0.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="py-20 lg:py-28 bg-muted/30 border-y" aria-labelledby="features-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 id="features-heading" className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight">
              Everything you need to run your property
            </h2>
            <p className="mt-4 text-lg text-muted-foreground max-w-2xl mx-auto">
              Purpose-built modules for Indian service apartments and boutique hotels.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[
              { icon: "M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z", title: "Reservations", desc: "Visual calendar with drag-and-drop. Walk-ins, OTA bookings, and conflict detection built in.", color: "text-blue-600 bg-blue-50" },
              { icon: "M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z", title: "Billing & Invoicing", desc: "GST-compliant invoices, split payments, advance deposits, and night audit in one click.", color: "text-emerald-600 bg-emerald-50" },
              { icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4", title: "Operations", desc: "Streamlined check-in/out workflows, ID verification, and real-time status tracking.", color: "text-violet-600 bg-violet-50" },
              { icon: "M19 21l-7-5-7 5V5a2 2 0 012-2h10a2 2 0 012 2z", title: "Housekeeping", desc: "Task assignment, priority management, real-time status updates, and inspection checklists.", color: "text-amber-600 bg-amber-50" },
              { icon: "M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15", title: "Laundry", desc: "Order tracking, rate cards per item, express processing, and delivery management.", color: "text-teal-600 bg-teal-50" },
              { icon: "M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z", title: "Reports & Analytics", desc: "Occupancy trends, revenue insights, housekeeping efficiency, and payment tracking.", color: "text-rose-600 bg-rose-50" },
            ].map((feat, i) => (
              <div key={i} className="group relative rounded-2xl border bg-card p-6 hover:shadow-xl hover:border-primary/15 transition-all duration-300 hover:-translate-y-1">
                {/* Subtle gradient overlay on hover */}
                <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-transparent to-primary/[0.02] opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                <div className="relative">
                  <div className={`w-11 h-11 rounded-xl flex items-center justify-center mb-4 ${feat.color} transition-transform duration-300 group-hover:scale-110 group-hover:shadow-sm`}>
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d={feat.icon} />
                    </svg>
                  </div>
                  <h3 className="text-base font-semibold text-foreground mb-2">{feat.title}</h3>
                  <p className="text-sm text-muted-foreground leading-relaxed">{feat.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Product showcase / Dashboard preview */}
      <section className="py-20 lg:py-28" aria-labelledby="showcase-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 id="showcase-heading" className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight">
              One dashboard. Complete visibility.
            </h2>
            <p className="mt-4 text-lg text-muted-foreground max-w-2xl mx-auto">
              See occupancy, revenue, operations, and alerts at a glance. Make faster decisions with real-time data.
            </p>
          </div>
          {/* Mock dashboard */}
          <div className="rounded-2xl border bg-card shadow-2xl overflow-hidden ring-1 ring-black/[0.03]">
            <div className="border-b bg-muted/30 px-4 py-3 flex items-center gap-2">
              <div className="flex gap-1.5" aria-hidden="true">
                <div className="w-3 h-3 rounded-full bg-red-400/80" />
                <div className="w-3 h-3 rounded-full bg-amber-400/80" />
                <div className="w-3 h-3 rounded-full bg-emerald-400/80" />
              </div>
              <div className="flex-1 flex justify-center">
                <div className="px-4 py-1 rounded-md bg-muted text-xs text-muted-foreground font-medium">dashboard.stayflow.app</div>
              </div>
            </div>
            <div className="p-6 lg:p-8 space-y-6">
              {/* Stats row */}
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                {[
                  { label: 'Occupancy', value: '78%', change: '+5%', color: 'text-emerald-600' },
                  { label: "Today's Revenue", value: '₹1,24,500', change: '+12%', color: 'text-emerald-600' },
                  { label: 'Check-ins Today', value: '6', change: '', color: '' },
                  { label: 'Pending Tasks', value: '3', change: '-2', color: 'text-emerald-600' },
                ].map((stat, i) => (
                  <div key={i} className="rounded-xl border bg-background p-4">
                    <div className="text-xs text-muted-foreground">{stat.label}</div>
                    <div className="mt-1 text-xl font-bold text-foreground">{stat.value}</div>
                    {stat.change && <div className={`text-xs ${stat.color} mt-0.5`}>{stat.change} from yesterday</div>}
                  </div>
                ))}
              </div>
              {/* Room grid mock */}
              <div className="rounded-lg border p-4">
                <div className="text-sm font-medium text-foreground mb-3">Room Status</div>
                <div className="grid grid-cols-4 sm:grid-cols-8 gap-2">
                  {['101','102','103','104','201','202','203','204','301','302','303','304','401','402','403','404'].map((room, i) => (
                    <div key={i} className={`rounded-md px-2 py-2 text-center text-xs font-medium ${
                      i < 6 ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' :
                      i < 8 ? 'bg-amber-50 text-amber-700 border border-amber-200' :
                      i < 12 ? 'bg-blue-50 text-blue-700 border border-blue-200' :
                      'bg-muted text-muted-foreground border'
                    }`}>
                      {room}
                    </div>
                  ))}
                </div>
                <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-sm bg-emerald-100 border border-emerald-200" aria-hidden="true" />Occupied</span>
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-sm bg-amber-100 border border-amber-200" aria-hidden="true" />Checkout Today</span>
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-sm bg-blue-100 border border-blue-200" aria-hidden="true" />Reserved</span>
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-sm bg-muted border" aria-hidden="true" />Available</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="py-20 lg:py-28 bg-muted/30 border-y" aria-labelledby="testimonials-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 id="testimonials-heading" className="text-2xl sm:text-3xl font-bold tracking-tight text-center mb-12">
            Trusted by property managers across India
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[
              { quote: "StayFlow cut our check-in time from 8 minutes to under 2. The guests love the speed, and my team can focus on hospitality instead of paperwork.", name: "Arjun Kumar", role: "Operations Manager", property: "Maple Suites, Bangalore" },
              { quote: "The billing module alone saved us from 3-4 hours of manual invoicing every week. GST compliance is automatic now — we just click and send.", name: "Priya Sharma", role: "Owner", property: "Haven Serviced Apartments, Pune" },
              { quote: "Before StayFlow, housekeeping coordination was chaos. Now every room status is visible in real-time. Turnaround improved by 40%.", name: "Vikram Mehta", role: "General Manager", property: "The Residency, Mumbai" },
            ].map((t, i) => (
              <div key={i} className="rounded-xl border bg-card p-6 flex flex-col">
                <div className="flex gap-1 mb-4" aria-label={`5 star rating`}>
                  {[...Array(5)].map((_, j) => (
                    <svg key={j} className="w-4 h-4 text-amber-400" fill="currentColor" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" /></svg>
                  ))}
                </div>
                <blockquote className="text-sm text-foreground leading-relaxed flex-1">
                  &ldquo;{t.quote}&rdquo;
                </blockquote>
                <div className="mt-5 pt-4 border-t">
                  <div className="text-sm font-medium text-foreground">{t.name}</div>
                  <div className="text-xs text-muted-foreground">{t.role} — {t.property}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="py-20 lg:py-28" aria-labelledby="pricing-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 id="pricing-heading" className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight">
              Simple, transparent pricing
            </h2>
            <p className="mt-4 text-lg text-muted-foreground">Start free. Upgrade when you grow. No hidden fees.</p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl mx-auto">
            {[
              { name: "Starter", price: "₹2,999", period: "/month", desc: "For single properties up to 20 rooms", features: ["Up to 20 rooms", "Reservations & billing", "Guest management", "Basic reports", "Email support"], cta: "Start Free Trial" },
              { name: "Professional", price: "₹5,999", period: "/month", desc: "For growing businesses up to 50 rooms", features: ["Up to 50 rooms", "Everything in Starter", "Housekeeping module", "Laundry module", "Priority support", "API access"], popular: true, cta: "Start Free Trial" },
              { name: "Enterprise", price: "Custom", period: "", desc: "For hotel chains with multiple properties", features: ["Unlimited rooms", "Multi-property dashboard", "Dedicated account manager", "Custom integrations", "SLA guarantee", "On-premise option"], cta: "Contact Sales" },
            ].map((plan, i) => (
              <div key={i} className={`group rounded-2xl border p-8 flex flex-col transition-all duration-300 hover:-translate-y-1 ${plan.popular ? 'border-primary/40 shadow-xl ring-1 ring-primary/10 relative bg-card hover:shadow-2xl' : 'bg-card shadow-sm hover:shadow-lg hover:border-primary/15'}`}>
                {plan.popular && (
                  <div className="absolute -top-3.5 left-1/2 -translate-x-1/2">
                    <span className="px-3 py-1 text-xs font-semibold text-primary-foreground bg-primary rounded-full shadow-sm">Most Popular</span>
                  </div>
                )}
                <div>
                  <h3 className="text-lg font-semibold text-foreground">{plan.name}</h3>
                  <p className="text-sm text-muted-foreground mt-1">{plan.desc}</p>
                </div>
                <div className="mt-5 mb-6">
                  <span className="text-4xl font-bold text-foreground">{plan.price}</span>
                  <span className="text-sm text-muted-foreground">{plan.period}</span>
                </div>
                <ul className="space-y-3 flex-1" aria-label={`${plan.name} features`}>
                  {plan.features.map((feat, j) => (
                    <li key={j} className="flex items-center gap-2.5 text-sm text-muted-foreground">
                      <svg className="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {feat}
                    </li>
                  ))}
                </ul>
                <Link
                  href={plan.cta === "Contact Sales" ? "/login" : "/register"}
                  className={`mt-8 block text-center px-4 py-2.5 text-sm font-medium rounded-lg transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 ${
                    plan.popular
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm'
                      : 'border border-border text-foreground hover:bg-muted'
                  }`}
                >
                  {plan.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ */}
      <section id="faq" className="py-20 lg:py-28 bg-muted/30 border-t" aria-labelledby="faq-heading">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 id="faq-heading" className="text-2xl sm:text-3xl font-bold tracking-tight">
              Frequently asked questions
            </h2>
            <p className="mt-3 text-muted-foreground">Everything you need to know about StayFlow.</p>
          </div>
          <div className="space-y-3">
            {[
              { q: "Is there a free trial?", a: "Yes! Every plan includes a 14-day free trial with full access to all features. No credit card required to get started." },
              { q: "Can I import my existing data?", a: "Absolutely. We support CSV imports for guests, reservations, and room data. Our team provides free migration assistance for Professional and Enterprise plans." },
              { q: "Is it GST compliant?", a: "Yes. StayFlow generates fully GST-compliant invoices with proper GSTIN, HSN codes (9963 for hospitality), and correct tax breakdowns at 5%, 12%, 18%, or 28% based on room tariff." },
              { q: "Does it work on mobile?", a: "StayFlow is fully responsive and works beautifully on tablets and phones. Your front desk and housekeeping staff can use it on any device." },
              { q: "Can I manage multiple properties?", a: "Yes! The Enterprise plan includes multi-property management with a unified dashboard, shared guest database, and property-level reporting." },
              { q: "What kind of support do you offer?", a: "Starter plans get email support with 24-hour response time. Professional includes priority support with 4-hour response. Enterprise gets a dedicated account manager and phone support." },
            ].map((item, i) => (
              <FAQItem key={i} question={item.q} answer={item.a} />
            ))}
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section className="py-20 lg:py-28" aria-labelledby="cta-heading">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="rounded-3xl bg-gradient-to-br from-primary via-primary/95 to-accent/70 p-10 sm:p-16 text-center relative overflow-hidden shadow-2xl">
            <div className="relative z-10">
              <h2 id="cta-heading" className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight text-primary-foreground">
                Ready to simplify your operations?
              </h2>
              <p className="mt-4 text-primary-foreground/80 max-w-xl mx-auto text-lg">
                Join 50+ properties already saving hours every day with StayFlow.
              </p>
              <div className="mt-8 flex flex-col sm:flex-row items-center justify-center gap-4">
                <Link href="/register" className="group inline-flex items-center gap-2 px-7 py-3.5 text-sm font-semibold bg-white text-primary rounded-xl hover:bg-white/95 shadow-lg hover:shadow-xl transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-primary active:scale-[0.98]">
                  Start Free Trial
                  <svg className="w-4 h-4 transition-transform duration-200 group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </Link>
                <Link href="/login" className="text-sm font-medium text-primary-foreground/90 hover:text-primary-foreground transition-colors underline-offset-4 hover:underline">
                  Already have an account? Sign in
                </Link>
              </div>
            </div>
            {/* Decorative elements */}
            <div className="absolute inset-0 pointer-events-none" aria-hidden="true">
              <div className="absolute top-0 right-0 w-64 h-64 rounded-full bg-white/5 -translate-y-1/3 translate-x-1/3" />
              <div className="absolute bottom-0 left-0 w-48 h-48 rounded-full bg-white/5 translate-y-1/3 -translate-x-1/3" />
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-10" role="contentinfo">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-6">
            <div className="flex items-center gap-2.5">
              <div className="w-7 h-7 flex items-center justify-center" aria-hidden="true">
                <img src="/logo.png" alt="StayFlow Logo" className="w-full h-full object-contain" />
              </div>
              <span className="text-sm font-semibold">StayFlow</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-muted-foreground">
              <a href="#features" className="hover:text-foreground transition-colors">Features</a>
              <a href="#pricing" className="hover:text-foreground transition-colors">Pricing</a>
              <a href="#faq" className="hover:text-foreground transition-colors">FAQ</a>
              <Link href="/login" className="hover:text-foreground transition-colors">Sign in</Link>
            </div>
            <p className="text-xs text-muted-foreground">© {new Date().getFullYear()} StayFlow. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  );
}

/* ─── FAQ Accordion Item ───────────────────────────── */

function FAQItem({ question, answer }: { question: string; answer: string }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-xl border bg-card overflow-hidden">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-5 py-4 text-left text-sm font-medium text-foreground hover:bg-muted/30 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
        aria-expanded={open}
      >
        {question}
        <svg
          className={`w-4 h-4 text-muted-foreground transition-transform duration-200 shrink-0 ml-4 ${open ? 'rotate-180' : ''}`}
          fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}
          aria-hidden="true"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      <div
        className={`overflow-hidden transition-all duration-200 ${open ? 'max-h-40' : 'max-h-0'}`}
        aria-hidden={!open}
      >
        <div className="px-5 pb-4 text-sm text-muted-foreground leading-relaxed">
          {answer}
        </div>
      </div>
    </div>
  );
}
