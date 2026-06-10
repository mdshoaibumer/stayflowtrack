'use client';

import React, { useState, useEffect } from 'react';

interface OnboardingStep {
  id: string;
  title: string;
  description: string;
  href?: string;
  completed: boolean;
}

const STORAGE_KEY = 'stayflow-onboarding';

function getStoredSteps(): Record<string, boolean> {
  if (typeof window === 'undefined') return {};
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
  } catch {
    return {};
  }
}

function setStoredStep(id: string, value: boolean) {
  const stored = getStoredSteps();
  stored[id] = value;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
}

export function useOnboarding() {
  const [steps, setSteps] = useState<OnboardingStep[]>([]);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    const stored = getStoredSteps();
    const isDismissed = localStorage.getItem(`${STORAGE_KEY}-dismissed`) === 'true';
    setDismissed(isDismissed);

    setSteps([
      { id: 'property', title: 'Set up your property', description: 'Add property details and room types', href: '/settings', completed: !!stored['property'] },
      { id: 'rooms', title: 'Add rooms', description: 'Create rooms and assign room types', href: '/operations', completed: !!stored['rooms'] },
      { id: 'rates', title: 'Configure rates', description: 'Set up pricing and rate plans', href: '/settings', completed: !!stored['rates'] },
      { id: 'guest', title: 'Add your first guest', description: 'Register a guest profile', href: '/guests?action=new', completed: !!stored['guest'] },
      { id: 'reservation', title: 'Create a reservation', description: 'Book a room for a guest', href: '/reservations?action=new', completed: !!stored['reservation'] },
    ]);
  }, []);

  const completeStep = (id: string) => {
    setStoredStep(id, true);
    setSteps((prev) => prev.map((s) => s.id === id ? { ...s, completed: true } : s));
  };

  const dismiss = () => {
    localStorage.setItem(`${STORAGE_KEY}-dismissed`, 'true');
    setDismissed(true);
  };

  const reset = () => {
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(`${STORAGE_KEY}-dismissed`);
    setDismissed(false);
    setSteps((prev) => prev.map((s) => ({ ...s, completed: false })));
  };

  const completedCount = steps.filter((s) => s.completed).length;
  const progress = steps.length > 0 ? Math.round((completedCount / steps.length) * 100) : 0;

  return { steps, completeStep, dismiss, reset, dismissed, progress, completedCount, total: steps.length };
}

export function OnboardingChecklist() {
  const { steps, completeStep, dismiss, dismissed, progress, completedCount, total } = useOnboarding();
  const [expanded, setExpanded] = useState(true);

  if (dismissed || total === 0) return null;
  if (completedCount === total) return null;

  return (
    <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
      {/* Header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-5 py-4 flex items-center gap-4 text-left hover:bg-muted/30 transition-colors"
        aria-expanded={expanded}
      >
        {/* Progress ring */}
        <div className="relative shrink-0">
          <svg className="w-10 h-10 -rotate-90" viewBox="0 0 36 36">
            <circle cx="18" cy="18" r="15" fill="none" stroke="currentColor" strokeWidth="2" className="text-muted/40" />
            <circle
              cx="18" cy="18" r="15" fill="none" stroke="currentColor" strokeWidth="2.5"
              className="text-primary transition-all duration-500"
              strokeDasharray={`${progress * 0.94} 100`}
              strokeLinecap="round"
            />
          </svg>
          <span className="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-primary">
            {progress}%
          </span>
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-foreground">Getting Started</h3>
          <p className="text-xs text-muted-foreground">{completedCount} of {total} steps completed</p>
        </div>
        <svg
          className={`w-4 h-4 text-muted-foreground transition-transform ${expanded ? 'rotate-180' : ''}`}
          fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Steps */}
      {expanded && (
        <div className="px-5 pb-4 space-y-1">
          {steps.map((step) => (
            <a
              key={step.id}
              href={step.href}
              onClick={(e) => {
                if (step.completed) {
                  e.preventDefault();
                  return;
                }
                completeStep(step.id);
              }}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                step.completed
                  ? 'opacity-60'
                  : 'hover:bg-muted/50 cursor-pointer'
              }`}
            >
              {/* Checkbox */}
              <div className={`shrink-0 w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors ${
                step.completed
                  ? 'bg-primary border-primary'
                  : 'border-gray-300'
              }`}>
                {step.completed && (
                  <svg className="w-3 h-3 text-primary-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <span className={`font-medium ${step.completed ? 'line-through text-muted-foreground' : 'text-foreground'}`}>
                  {step.title}
                </span>
                <p className="text-xs text-muted-foreground truncate">{step.description}</p>
              </div>
              {!step.completed && (
                <svg className="w-4 h-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                </svg>
              )}
            </a>
          ))}

          {/* Dismiss */}
          <div className="pt-2 border-t mt-3">
            <button
              onClick={dismiss}
              className="text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              Dismiss checklist
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
