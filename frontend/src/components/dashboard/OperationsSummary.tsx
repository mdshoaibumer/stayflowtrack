"use client";

import React from "react";
import Link from "next/link";
import { motion } from "framer-motion";

interface OperationsSummaryProps {
  operations: {
    check_ins_today: number;
    check_outs_today: number;
    expected_arrivals: number;
    expected_departures: number;
  };
  housekeeping: { counts: Record<string, number>; total: number };
  laundry: { counts: Record<string, number>; total: number };
}

function ProgressBar({ value, max, color }: { value: number; max: number; color: string }) {
  const pct = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  return (
    <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
      <motion.div
        initial={{ width: 0 }}
        animate={{ width: `${pct}%` }}
        transition={{ duration: 0.8, ease: [0.25, 0.46, 0.45, 0.94] }}
        className={`h-full rounded-full ${color}`}
      />
    </div>
  );
}

function SectionCard({
  title,
  icon,
  href,
  children,
  delay = 0,
}: {
  title: string;
  icon: React.ReactNode;
  href: string;
  children: React.ReactNode;
  delay?: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay }}
      className="rounded-xl border border-gray-100 bg-white p-5 shadow-sm hover:shadow-md transition-shadow duration-200"
    >
      <div className="flex items-center gap-2 mb-4">
        <div className="w-7 h-7 rounded-lg bg-gray-50 flex items-center justify-center text-gray-500">
          {icon}
        </div>
        <h3 className="text-sm font-semibold text-gray-900">{title}</h3>
      </div>
      {children}
      <Link href={href} className="mt-4 inline-flex items-center gap-1 text-xs font-medium text-teal-600 hover:text-teal-700 transition-colors">
        View all
        <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </Link>
    </motion.div>
  );
}

export function OperationsSummary({ operations, housekeeping, laundry }: OperationsSummaryProps) {
  const hkTotal = Object.values(housekeeping.counts || {}).reduce((a, b) => a + b, 0) || 1;
  const laundryTotal = Object.values(laundry.counts || {}).reduce((a, b) => a + b, 0) || 1;

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {/* Operations */}
      <SectionCard
        title="Today's Operations"
        href="/operations"
        delay={0.4}
        icon={
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
          </svg>
        }
      >
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-600">Check-ins</span>
            <span className="text-sm font-semibold text-emerald-600">{operations.check_ins_today}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-600">Check-outs</span>
            <span className="text-sm font-semibold text-orange-600">{operations.check_outs_today}</span>
          </div>
          <div className="h-px bg-gray-100 my-2" />
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500">Expected Arrivals</span>
            <span className="text-sm font-medium text-gray-700">{operations.expected_arrivals}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500">Expected Departures</span>
            <span className="text-sm font-medium text-gray-700">{operations.expected_departures}</span>
          </div>
        </div>
      </SectionCard>

      {/* Housekeeping */}
      <SectionCard
        title="Housekeeping"
        href="/housekeeping"
        delay={0.5}
        icon={
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
        }
      >
        <div className="space-y-3">
          {[
            { key: "dirty", label: "Dirty", color: "bg-rose-500" },
            { key: "cleaning", label: "Cleaning", color: "bg-amber-500" },
            { key: "inspection", label: "Inspection", color: "bg-blue-500" },
            { key: "ready", label: "Ready", color: "bg-emerald-500" },
          ].map(({ key, label, color }) => {
            const count = housekeeping.counts?.[key] || 0;
            return (
              <div key={key}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-600">{label}</span>
                  <span className="text-sm font-medium text-gray-900">{count}</span>
                </div>
                <ProgressBar value={count} max={hkTotal} color={color} />
              </div>
            );
          })}
        </div>
      </SectionCard>

      {/* Laundry */}
      <SectionCard
        title="Laundry"
        href="/laundry"
        delay={0.6}
        icon={
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        }
      >
        <div className="space-y-3">
          {[
            { key: "received", label: "Received", color: "bg-gray-500" },
            { key: "washing", label: "Washing", color: "bg-blue-500" },
            { key: "drying", label: "Drying", color: "bg-amber-500" },
            { key: "ready", label: "Ready", color: "bg-emerald-500" },
          ].map(({ key, label, color }) => {
            const count = laundry.counts?.[key] || 0;
            return (
              <div key={key}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-600">{label}</span>
                  <span className="text-sm font-medium text-gray-900">{count}</span>
                </div>
                <ProgressBar value={count} max={laundryTotal} color={color} />
              </div>
            );
          })}
        </div>
      </SectionCard>
    </div>
  );
}
