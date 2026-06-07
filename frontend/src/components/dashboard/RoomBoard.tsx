"use client";

import React from "react";
import { motion } from "framer-motion";

interface UnitStatus {
  id: string;
  unit_number: string;
  status: string;
  unit_type_name?: string;
  guest_name?: string;
}

interface RoomBoardProps {
  units: UnitStatus[];
}

const statusConfig: Record<string, { bg: string; border: string; text: string; dot: string; label: string }> = {
  available: {
    bg: "bg-emerald-50 hover:bg-emerald-100",
    border: "border-emerald-200",
    text: "text-emerald-700",
    dot: "bg-emerald-500",
    label: "Available",
  },
  occupied: {
    bg: "bg-blue-50 hover:bg-blue-100",
    border: "border-blue-200",
    text: "text-blue-700",
    dot: "bg-blue-500",
    label: "Occupied",
  },
  reserved: {
    bg: "bg-amber-50 hover:bg-amber-100",
    border: "border-amber-200",
    text: "text-amber-700",
    dot: "bg-amber-500",
    label: "Reserved",
  },
  cleaning: {
    bg: "bg-orange-50 hover:bg-orange-100",
    border: "border-orange-200",
    text: "text-orange-700",
    dot: "bg-orange-500",
    label: "Cleaning",
  },
  dirty: {
    bg: "bg-rose-50 hover:bg-rose-100",
    border: "border-rose-200",
    text: "text-rose-700",
    dot: "bg-rose-500",
    label: "Dirty",
  },
  maintenance: {
    bg: "bg-gray-100 hover:bg-gray-200",
    border: "border-gray-300",
    text: "text-gray-600",
    dot: "bg-gray-500",
    label: "Maintenance",
  },
  inspection: {
    bg: "bg-violet-50 hover:bg-violet-100",
    border: "border-violet-200",
    text: "text-violet-700",
    dot: "bg-violet-500",
    label: "Inspection",
  },
};

function getConfig(status: string) {
  return statusConfig[status] || statusConfig.available;
}

function GuestInitials({ name }: { name: string }) {
  const initials = name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  // Generate consistent color from name
  const colors = ["bg-blue-500", "bg-teal-500", "bg-violet-500", "bg-amber-500", "bg-rose-500", "bg-indigo-500"];
  const colorIndex = name.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0) % colors.length;

  return (
    <div className={`w-5 h-5 rounded-full ${colors[colorIndex]} flex items-center justify-center`}>
      <span className="text-[8px] font-bold text-white leading-none">{initials}</span>
    </div>
  );
}

export function RoomBoard({ units }: RoomBoardProps) {
  const statusCounts = units.reduce((acc, unit) => {
    acc[unit.status] = (acc[unit.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: 0.3 }}
      className="rounded-xl border border-gray-100 bg-white p-5 shadow-sm"
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-navy-50 flex items-center justify-center">
            <svg className="w-4 h-4 text-navy-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-gray-900">Room Status Board</h3>
            <p className="text-xs text-gray-400">{units.length} total units</p>
          </div>
        </div>

        {/* Legend pills */}
        <div className="hidden md:flex items-center gap-2 flex-wrap">
          {Object.entries(statusCounts).map(([status, count]) => {
            const cfg = getConfig(status);
            return (
              <span key={status} className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium bg-gray-50 text-gray-600">
                <span className={`w-2 h-2 rounded-full ${cfg.dot}`} />
                {cfg.label} ({count})
              </span>
            );
          })}
        </div>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-8 xl:grid-cols-10 gap-2">
        {units.map((unit, idx) => {
          const cfg = getConfig(unit.status);
          return (
            <motion.div
              key={unit.id}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.2, delay: idx * 0.02 }}
              whileHover={{ scale: 1.08, zIndex: 10 }}
              whileTap={{ scale: 0.95 }}
              className={`
                relative p-2.5 rounded-lg border ${cfg.border} ${cfg.bg}
                cursor-pointer transition-colors duration-150
                flex flex-col items-center justify-center min-h-[60px]
                group
              `}
              title={`${unit.unit_number} — ${cfg.label}${unit.guest_name ? ` (${unit.guest_name})` : ""}`}
            >
              {/* Room number */}
              <span className={`text-sm font-bold ${cfg.text}`}>{unit.unit_number}</span>

              {/* Status label */}
              <span className={`text-[9px] uppercase tracking-wider font-medium ${cfg.text} opacity-70 mt-0.5`}>
                {unit.status.replace("_", " ")}
              </span>

              {/* Guest badge */}
              {unit.guest_name && (
                <div className="absolute -top-1 -right-1">
                  <GuestInitials name={unit.guest_name} />
                </div>
              )}

              {/* Hover tooltip */}
              {unit.guest_name && (
                <div className="absolute -bottom-7 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity duration-150 z-20 pointer-events-none">
                  <span className="whitespace-nowrap text-[10px] bg-gray-900 text-white px-2 py-0.5 rounded shadow-lg">
                    {unit.guest_name}
                  </span>
                </div>
              )}
            </motion.div>
          );
        })}
      </div>
    </motion.div>
  );
}
