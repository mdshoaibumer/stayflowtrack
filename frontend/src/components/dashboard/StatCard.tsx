"use client";

import React from "react";
import { motion } from "framer-motion";
import { AnimatedCounter } from "./AnimatedCounter";
import Link from "next/link";

type GlowColor = "green" | "blue" | "purple" | "orange" | "teal" | "red";

interface StatCardProps {
  title: string;
  value: number;
  prefix?: string;
  suffix?: string;
  subtitle?: string;
  trend?: { value: number; label: string };
  icon: React.ReactNode;
  color: GlowColor;
  href?: string;
  sparkline?: number[];
  index?: number;
}

const colorMap: Record<GlowColor, { bg: string; icon: string; glow: string; ring: string; indicator: string }> = {
  green: {
    bg: "bg-emerald-50",
    icon: "text-emerald-600",
    glow: "stat-glow-green",
    ring: "ring-emerald-500/20",
    indicator: "bg-emerald-500",
  },
  blue: {
    bg: "bg-blue-50",
    icon: "text-blue-600",
    glow: "stat-glow-blue",
    ring: "ring-blue-500/20",
    indicator: "bg-blue-500",
  },
  purple: {
    bg: "bg-violet-50",
    icon: "text-violet-600",
    glow: "stat-glow-purple",
    ring: "ring-violet-500/20",
    indicator: "bg-violet-500",
  },
  orange: {
    bg: "bg-amber-50",
    icon: "text-amber-600",
    glow: "stat-glow-orange",
    ring: "ring-amber-500/20",
    indicator: "bg-amber-500",
  },
  teal: {
    bg: "bg-teal-50",
    icon: "text-teal-600",
    glow: "stat-glow-teal",
    ring: "ring-teal-500/20",
    indicator: "bg-teal-500",
  },
  red: {
    bg: "bg-red-50",
    icon: "text-red-600",
    glow: "stat-glow-orange",
    ring: "ring-red-500/20",
    indicator: "bg-red-500",
  },
};

function MiniSparkline({ data, color }: { data: number[]; color: string }) {
  if (!data || data.length < 2) return null;
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const h = 24;
  const w = 64;
  const step = w / (data.length - 1);

  const points = data.map((v, i) => `${i * step},${h - ((v - min) / range) * h}`).join(" ");

  return (
    <svg width={w} height={h} className="overflow-visible">
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function StatCard({
  title,
  value,
  prefix,
  suffix,
  subtitle,
  trend,
  icon,
  color,
  href,
  sparkline,
  index = 0,
}: StatCardProps) {
  const colors = colorMap[color];

  const content = (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.1, ease: [0.25, 0.46, 0.45, 0.94] }}
      whileHover={{ y: -2, transition: { duration: 0.2 } }}
      className={`
        relative overflow-hidden rounded-xl border border-gray-100
        bg-white p-5 shadow-sm
        hover:shadow-lg hover:${colors.glow}
        transition-shadow duration-300
        cursor-pointer group
      `}
    >
      {/* Subtle gradient background */}
      <div className={`absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 ${colors.bg}`} />

      {/* Content */}
      <div className="relative z-10">
        {/* Header */}
        <div className="flex items-center justify-between mb-3">
          <div className={`flex items-center justify-center w-9 h-9 rounded-lg ${colors.bg} ${colors.icon}`}>
            {icon}
          </div>
          {sparkline && sparkline.length > 0 && (
            <MiniSparkline data={sparkline} color={color === "green" ? "#10b981" : color === "blue" ? "#3b82f6" : color === "purple" ? "#8b5cf6" : color === "orange" ? "#f59e0b" : color === "teal" ? "#14b8a6" : "#ef4444"} />
          )}
        </div>

        {/* Value */}
        <div className="mb-1">
          <AnimatedCounter
            value={value}
            prefix={prefix}
            suffix={suffix}
            className="text-2xl font-bold text-gray-900 tracking-tight"
          />
        </div>

        {/* Title */}
        <p className="text-sm font-medium text-gray-500 mb-2">{title}</p>

        {/* Footer */}
        <div className="flex items-center justify-between">
          {subtitle && (
            <span className="text-xs text-gray-400">{subtitle}</span>
          )}
          {trend && (
            <span className={`inline-flex items-center gap-0.5 text-xs font-medium ${trend.value >= 0 ? "text-emerald-600" : "text-red-500"}`}>
              <svg
                className={`w-3 h-3 ${trend.value < 0 ? "rotate-180" : ""}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2.5}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" />
              </svg>
              {Math.abs(trend.value)}% {trend.label}
            </span>
          )}
        </div>
      </div>

      {/* Active indicator */}
      <div className={`absolute bottom-0 left-0 right-0 h-0.5 ${colors.indicator} opacity-0 group-hover:opacity-100 transition-opacity duration-300`} />
    </motion.div>
  );

  if (href) {
    return <Link href={href} className="block">{content}</Link>;
  }

  return content;
}
