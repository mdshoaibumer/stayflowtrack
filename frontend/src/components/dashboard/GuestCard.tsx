"use client";

import React from "react";
import { motion, AnimatePresence } from "framer-motion";

interface Guest {
  id: string;
  full_name: string;
  email: string;
  phone: string;
  id_type: string;
  id_number: string;
  nationality: string;
  total_stays: number;
  last_stay_date: string;
  created_at: string;
}

interface GuestCardProps {
  guest: Guest;
  index?: number;
  onSelect?: (guest: Guest) => void;
  selected?: boolean;
}

const avatarColors = [
  "from-blue-500 to-blue-600",
  "from-teal-500 to-teal-600",
  "from-violet-500 to-violet-600",
  "from-amber-500 to-amber-600",
  "from-rose-500 to-rose-600",
  "from-indigo-500 to-indigo-600",
  "from-emerald-500 to-emerald-600",
  "from-pink-500 to-pink-600",
];

function getAvatarColor(name: string): string {
  const idx = (name || "").split("").reduce((acc, c) => acc + c.charCodeAt(0), 0) % avatarColors.length;
  return avatarColors[idx];
}

function getInitials(name: string): string {
  return (name || "?")
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

function formatDate(dateStr: string): string {
  if (!dateStr) return "—";
  const d = new Date(dateStr);
  return d.toLocaleDateString("en-IN", { day: "numeric", month: "short", year: "numeric" });
}

function getLoyaltyTier(stays: number): { label: string; color: string } {
  if (stays >= 20) return { label: "Platinum", color: "bg-violet-100 text-violet-700" };
  if (stays >= 10) return { label: "Gold", color: "bg-amber-100 text-amber-700" };
  if (stays >= 5) return { label: "Silver", color: "bg-gray-100 text-gray-700" };
  return { label: "New", color: "bg-blue-100 text-blue-700" };
}

export function GuestCard({ guest, index = 0, onSelect, selected }: GuestCardProps) {
  const loyalty = getLoyaltyTier(guest.total_stays);

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -12 }}
      transition={{ duration: 0.3, delay: index * 0.05 }}
      whileHover={{ y: -1 }}
      onClick={() => onSelect?.(guest)}
      className={`
        relative rounded-xl border bg-white p-4 cursor-pointer
        transition-all duration-200 group
        ${selected
          ? "border-teal-300 shadow-md ring-2 ring-teal-500/20"
          : "border-gray-100 shadow-sm hover:shadow-md hover:border-gray-200"
        }
      `}
    >
      <div className="flex items-start gap-3.5">
        {/* Avatar */}
        <div className={`
          relative flex-shrink-0 w-11 h-11 rounded-full bg-gradient-to-br ${getAvatarColor(guest.full_name)}
          flex items-center justify-center shadow-sm
        `}>
          <span className="text-sm font-bold text-white">{getInitials(guest.full_name)}</span>
          {/* Online indicator for recent guests */}
          {guest.last_stay_date && new Date(guest.last_stay_date) > new Date(Date.now() - 7 * 86400000) && (
            <span className="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-emerald-400 border-2 border-white" />
          )}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-0.5">
            <h4 className="text-sm font-semibold text-gray-900 truncate">{guest.full_name}</h4>
            <span className={`flex-shrink-0 px-1.5 py-0.5 rounded text-[10px] font-semibold ${loyalty.color}`}>
              {loyalty.label}
            </span>
          </div>

          <div className="flex items-center gap-3 text-xs text-gray-500 mb-2">
            {guest.email && (
              <span className="truncate max-w-[150px]">{guest.email}</span>
            )}
            {guest.phone && (
              <span className="flex items-center gap-1">
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
                {guest.phone}
              </span>
            )}
          </div>

          {/* Stats row */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1.5">
              <svg className="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
              <span className="text-xs font-medium text-gray-700">{guest.total_stays} stays</span>
            </div>
            {guest.nationality && (
              <div className="flex items-center gap-1.5">
                <svg className="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="text-xs text-gray-600">{guest.nationality}</span>
              </div>
            )}
            {guest.last_stay_date && (
              <span className="text-xs text-gray-400">Last: {formatDate(guest.last_stay_date)}</span>
            )}
          </div>
        </div>

        {/* Arrow */}
        <div className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
          <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </div>
      </div>
    </motion.div>
  );
}

interface GuestDetailPanelProps {
  guest: Guest | null;
  onClose: () => void;
}

export function GuestDetailPanel({ guest, onClose }: GuestDetailPanelProps) {
  const loyalty = guest ? getLoyaltyTier(guest.total_stays) : { label: "", color: "" };

  return (
    <AnimatePresence>
      {guest && (
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: 20 }}
          transition={{ duration: 0.25, ease: [0.25, 0.46, 0.45, 0.94] }}
          className="rounded-xl border border-gray-100 bg-white shadow-lg p-6 sticky top-6"
        >
          {/* Close */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-1.5 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          {/* Profile header */}
          <div className="text-center mb-6">
            <div className={`
              mx-auto w-16 h-16 rounded-full bg-gradient-to-br ${getAvatarColor(guest.full_name)}
              flex items-center justify-center shadow-lg mb-3
            `}>
              <span className="text-xl font-bold text-white">{getInitials(guest.full_name)}</span>
            </div>
            <h3 className="text-lg font-bold text-gray-900">{guest.full_name}</h3>
            <p className="text-sm text-gray-500">{guest.email}</p>
            <span className={`inline-block mt-2 px-2.5 py-1 rounded-full text-xs font-semibold ${loyalty.color}`}>
              {loyalty.label} Member
            </span>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-3 mb-6">
            <div className="text-center p-3 rounded-lg bg-gray-50">
              <p className="text-lg font-bold text-gray-900">{guest.total_stays}</p>
              <p className="text-[10px] uppercase tracking-wider font-medium text-gray-500">Stays</p>
            </div>
            <div className="text-center p-3 rounded-lg bg-gray-50">
              <p className="text-lg font-bold text-gray-900">{guest.nationality || "—"}</p>
              <p className="text-[10px] uppercase tracking-wider font-medium text-gray-500">Country</p>
            </div>
            <div className="text-center p-3 rounded-lg bg-gray-50">
              <p className="text-lg font-bold text-gray-900">
                {guest.last_stay_date ? new Date(guest.last_stay_date).toLocaleDateString("en-IN", { day: "numeric", month: "short" }) : "—"}
              </p>
              <p className="text-[10px] uppercase tracking-wider font-medium text-gray-500">Last Visit</p>
            </div>
          </div>

          {/* Contact details */}
          <div className="space-y-3 mb-6">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-gray-400">Contact</h4>
            {guest.phone && (
              <div className="flex items-center gap-3 p-2.5 rounded-lg bg-gray-50">
                <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
                <span className="text-sm text-gray-700">{guest.phone}</span>
              </div>
            )}
            {guest.id_type && (
              <div className="flex items-center gap-3 p-2.5 rounded-lg bg-gray-50">
                <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2" />
                </svg>
                <span className="text-sm text-gray-700">{guest.id_type}: {guest.id_number}</span>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="space-y-2">
            <button className="w-full py-2.5 px-4 rounded-lg bg-navy-900 text-white text-sm font-medium hover:bg-navy-800 transition-colors">
              New Reservation
            </button>
            <button className="w-full py-2.5 px-4 rounded-lg border border-gray-200 text-gray-700 text-sm font-medium hover:bg-gray-50 transition-colors">
              View History
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
