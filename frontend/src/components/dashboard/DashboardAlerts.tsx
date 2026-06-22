"use client";

import React from "react";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";

export interface AlertItem {
  id: string;
  type: "overdue" | "arriving" | "warning";
  title: string;
  message: string;
  actionLabel?: string;
  actionHref?: string;
}

interface DashboardAlertsProps {
  alerts: AlertItem[];
}

const alertStyles = {
  overdue: {
    bg: "bg-red-50 border-red-200",
    icon: "text-red-600",
    title: "text-red-800",
    message: "text-red-700",
    button: "bg-red-600 hover:bg-red-700 text-white",
  },
  arriving: {
    bg: "bg-blue-50 border-blue-200",
    icon: "text-blue-600",
    title: "text-blue-800",
    message: "text-blue-700",
    button: "bg-blue-600 hover:bg-blue-700 text-white",
  },
  warning: {
    bg: "bg-amber-50 border-amber-200",
    icon: "text-amber-600",
    title: "text-amber-800",
    message: "text-amber-700",
    button: "bg-amber-600 hover:bg-amber-700 text-white",
  },
};

const alertIcons = {
  overdue: (
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
  ),
  arriving: (
    <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  ),
  warning: (
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  ),
};

export function DashboardAlerts({ alerts }: DashboardAlertsProps) {
  if (alerts.length === 0) return null;

  return (
    <div className="space-y-2">
      <AnimatePresence>
        {alerts.map((alert, idx) => {
          const styles = alertStyles[alert.type];
          return (
            <motion.div
              key={alert.id}
              initial={{ opacity: 0, height: 0, y: -10 }}
              animate={{ opacity: 1, height: "auto", y: 0 }}
              exit={{ opacity: 0, height: 0, y: -10 }}
              transition={{ duration: 0.3, delay: idx * 0.05 }}
              className={`rounded-xl border p-4 ${styles.bg}`}
            >
              <div className="flex items-start gap-3">
                <svg className={`w-5 h-5 flex-shrink-0 mt-0.5 ${styles.icon}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  {alertIcons[alert.type]}
                </svg>
                <div className="flex-1 min-w-0">
                  <h4 className={`text-sm font-semibold ${styles.title}`}>{alert.title}</h4>
                  <p className={`text-xs mt-0.5 ${styles.message}`}>{alert.message}</p>
                </div>
                {alert.actionHref && alert.actionLabel && (
                  <Link
                    href={alert.actionHref}
                    className={`flex-shrink-0 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${styles.button}`}
                  >
                    {alert.actionLabel}
                  </Link>
                )}
              </div>
            </motion.div>
          );
        })}
      </AnimatePresence>
    </div>
  );
}
