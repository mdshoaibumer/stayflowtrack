"use client";

import React from "react";
import Link from "next/link";
import { motion } from "framer-motion";

interface QuickActionItem {
  href: string;
  label: string;
  icon: React.ReactNode;
  color: string;
  hoverColor: string;
}

const actions: QuickActionItem[] = [
  {
    href: "/reservations?action=new",
    label: "New Booking",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />,
    color: "bg-blue-600",
    hoverColor: "hover:bg-blue-700",
  },
  {
    href: "/operations?action=walkin",
    label: "Walk-In",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />,
    color: "bg-emerald-600",
    hoverColor: "hover:bg-emerald-700",
  },
  {
    href: "/operations?tab=checkin",
    label: "Check In",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />,
    color: "bg-violet-600",
    hoverColor: "hover:bg-violet-700",
  },
  {
    href: "/operations?tab=checkout",
    label: "Check Out",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />,
    color: "bg-orange-600",
    hoverColor: "hover:bg-orange-700",
  },
  {
    href: "/operations?tab=inhouse",
    label: "Extend Stay",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />,
    color: "bg-indigo-600",
    hoverColor: "hover:bg-indigo-700",
  },
  {
    href: "/billing?action=charge",
    label: "Quick Charge",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />,
    color: "bg-rose-600",
    hoverColor: "hover:bg-rose-700",
  },
  {
    href: "/laundry?action=new",
    label: "Add Laundry",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />,
    color: "bg-cyan-600",
    hoverColor: "hover:bg-cyan-700",
  },
  {
    href: "/reports?type=end_of_day",
    label: "Close Day",
    icon: <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />,
    color: "bg-gray-800",
    hoverColor: "hover:bg-gray-900",
  },
];

export function QuickActions() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: 0.2 }}
      className="grid grid-cols-4 sm:grid-cols-4 lg:grid-cols-8 gap-2"
    >
      {actions.map((action, idx) => (
        <Link key={action.href} href={action.href}>
          <motion.div
            whileHover={{ scale: 1.05, y: -2 }}
            whileTap={{ scale: 0.95 }}
            transition={{ type: "spring", stiffness: 400, damping: 17 }}
            className={`
              flex flex-col items-center gap-1.5 p-3 rounded-xl text-white
              ${action.color} ${action.hoverColor}
              shadow-sm hover:shadow-md transition-shadow duration-200
            `}
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              {action.icon}
            </svg>
            <span className="text-[10px] sm:text-xs font-medium text-center leading-tight">{action.label}</span>
          </motion.div>
        </Link>
      ))}
    </motion.div>
  );
}
