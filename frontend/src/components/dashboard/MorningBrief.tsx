"use client";

import React from "react";
import Link from "next/link";
import { motion } from "framer-motion";

interface MorningBriefProps {
  metrics: {
    date: string;
    occupancy: { total_units: number; occupied_units: number; occupancy_rate: number };
    operations: { expected_arrivals: number; expected_departures: number };
    pending_payments: { pending_amount: number };
  };
}

export function MorningBrief({ metrics }: MorningBriefProps) {
  const hour = new Date().getHours();
  const greeting = hour < 12 ? "Good morning" : hour < 17 ? "Good afternoon" : "Good evening";

  return (
    <motion.div
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="relative overflow-hidden rounded-xl border border-navy-100 bg-gradient-to-r from-navy-900 via-navy-800 to-navy-900 p-5 text-white"
    >
      {/* Background pattern */}
      <div className="absolute inset-0 opacity-10">
        <svg className="w-full h-full" viewBox="0 0 400 100" fill="none">
          <circle cx="350" cy="10" r="60" fill="white" />
          <circle cx="380" cy="80" r="30" fill="white" />
        </svg>
      </div>

      <div className="relative z-10">
        <p className="text-navy-200 text-sm font-medium mb-1">{greeting}</p>
        <div className="flex flex-wrap items-baseline gap-x-6 gap-y-1">
          <span className="text-base font-medium text-white/90">
            {metrics.operations.expected_arrivals > 0
              ? `${metrics.operations.expected_arrivals} arrival${metrics.operations.expected_arrivals > 1 ? "s" : ""}`
              : "No arrivals"} today
          </span>
          <span className="text-sm text-white/70">•</span>
          <span className="text-base font-medium text-white/90">
            {metrics.operations.expected_departures} departure{metrics.operations.expected_departures !== 1 ? "s" : ""}
          </span>
          <span className="text-sm text-white/70">•</span>
          <span className="text-base font-medium text-white/90">
            {metrics.occupancy.occupied_units}/{metrics.occupancy.total_units} occupied ({metrics.occupancy.occupancy_rate.toFixed(0)}%)
          </span>
          {metrics.pending_payments.pending_amount > 0 && (
            <>
              <span className="text-sm text-white/70">•</span>
              <Link href="/reports?type=outstanding" className="text-base font-medium text-amber-300 hover:text-amber-200 transition-colors">
                ₹{metrics.pending_payments.pending_amount.toLocaleString("en-IN")} outstanding
              </Link>
            </>
          )}
        </div>
      </div>
    </motion.div>
  );
}
