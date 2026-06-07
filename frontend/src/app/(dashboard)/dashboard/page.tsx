"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import Link from "next/link";
import { motion } from "framer-motion";
import { StatCard } from "@/components/dashboard/StatCard";
import { RoomBoard } from "@/components/dashboard/RoomBoard";
import { QuickActions } from "@/components/dashboard/QuickActions";
import { DashboardAlerts } from "@/components/dashboard/DashboardAlerts";
import { OperationsSummary } from "@/components/dashboard/OperationsSummary";
import { MorningBrief } from "@/components/dashboard/MorningBrief";

interface DashboardMetrics {
  date: string;
  occupancy: { total_units: number; occupied_units: number; available_units: number; occupancy_rate: number };
  revenue: { today: number; this_week: number; this_month: number; currency: string };
  operations: { check_ins_today: number; check_outs_today: number; expected_arrivals: number; expected_departures: number };
  housekeeping: { counts: Record<string, number>; total: number };
  laundry: { counts: Record<string, number>; total: number };
  pending_payments: { pending_count: number; pending_amount: number; overdue_count: number };
}

interface UnitStatus {
  id: string;
  unit_number: string;
  status: string;
  unit_type_name?: string;
  guest_name?: string;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", minimumFractionDigits: 0 }).format(amount);
}

export default function DashboardPage() {
  const { user } = useAuth();
  const api = useApi();
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [units, setUnits] = useState<UnitStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [alerts, setAlerts] = useState<any[]>([]);

  const propertyId = user?.property_id || "";

  const fetchDashboard = useCallback(async () => {
    if (!propertyId) {
      setLoading(false);
      return;
    }
    try {
      const [data, unitData] = await Promise.all([
        api.get<DashboardMetrics>(`/api/v1/dashboard/${propertyId}`),
        api.get<any>(`/api/v1/properties/${propertyId}/units`),
      ]);
      setMetrics(data);
      setUnits(Array.isArray(unitData) ? unitData : unitData?.data || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load dashboard");
    } finally {
      setLoading(false);
    }
  }, [api, propertyId]);

  // Fetch alerts
  useEffect(() => {
    if (!propertyId) return;
    const fetchAlerts = async () => {
      try {
        const data = await api.get<any>("/api/v1/reservations", { property_id: propertyId, per_page: "100" });
        const reservations = Array.isArray(data) ? data : data?.data || [];
        const today = new Date().toISOString().split("T")[0];
        const tomorrow = new Date(Date.now() + 86400000).toISOString().split("T")[0];

        const overdueCheckouts = reservations.filter((r: any) => r.status === "checked_in" && r.check_out_date <= today);
        const tomorrowArrivals = reservations.filter((r: any) => r.check_in_date === tomorrow && (r.status === "confirmed" || r.status === "pending"));

        const alertItems: any[] = [];
        if (overdueCheckouts.length > 0) {
          alertItems.push({
            id: "overdue",
            type: "overdue",
            title: `Overdue Checkouts (${overdueCheckouts.length})`,
            message: overdueCheckouts.map((r: any) => `${r.unit_number} — ${r.guest_name}`).join(", "),
            actionLabel: "Check Out",
            actionHref: "/operations?tab=checkout",
          });
        }
        if (tomorrowArrivals.length > 0) {
          alertItems.push({
            id: "arriving",
            type: "arriving",
            title: `Tomorrow's Arrivals (${tomorrowArrivals.length})`,
            message: tomorrowArrivals.map((r: any) => `${r.unit_number} — ${r.guest_name}`).join(", "),
            actionLabel: "Prepare",
            actionHref: "/housekeeping",
          });
        }
        setAlerts(alertItems);
      } catch { /* silent */ }
    };
    fetchAlerts();
  }, [api, propertyId]);

  useEffect(() => {
    fetchDashboard();
    const interval = setInterval(fetchDashboard, 30000);
    return () => clearInterval(interval);
  }, [fetchDashboard]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex flex-col items-center gap-3">
          <div className="relative">
            <div className="w-10 h-10 rounded-full border-[3px] border-gray-200 border-t-teal-600 animate-spin" />
          </div>
          <p className="text-sm text-gray-400 font-medium">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="p-6 bg-red-50 border border-red-200 rounded-xl text-center"
      >
        <div className="w-12 h-12 mx-auto mb-3 rounded-full bg-red-100 flex items-center justify-center">
          <svg className="w-6 h-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
        </div>
        <p className="text-red-800 font-medium mb-1">Failed to load dashboard</p>
        <p className="text-red-600 text-sm mb-3">{error}</p>
        <button
          onClick={fetchDashboard}
          className="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 transition-colors"
        >
          Try Again
        </button>
      </motion.div>
    );
  }

  return (
    <div className="space-y-5 max-w-[1600px] mx-auto">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
      >
        <div>
          <h1 className="text-xl lg:text-2xl font-bold text-gray-900 tracking-tight">Dashboard</h1>
          <p className="text-sm text-gray-400">
            {metrics?.date || new Date().toLocaleDateString("en-IN", { weekday: "long", day: "numeric", month: "long", year: "numeric" })}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-100">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse-subtle" />
            Live
          </span>
          <span className="text-xs text-gray-400">Auto-refreshes every 30s</span>
        </div>
      </motion.div>

      {/* Morning Brief */}
      {metrics && <MorningBrief metrics={metrics} />}

      {/* Alerts */}
      <DashboardAlerts alerts={alerts} />

      {/* Quick Actions */}
      <QuickActions />

      {/* KPI Stats */}
      {metrics && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            title="Occupancy"
            value={metrics.occupancy.occupancy_rate}
            suffix="%"
            subtitle={`${metrics.occupancy.occupied_units}/${metrics.occupancy.total_units} units`}
            icon={
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            }
            color="green"
            href="/reservations"
            index={0}
          />
          <StatCard
            title="Revenue Today"
            value={metrics.revenue.today}
            prefix="₹"
            subtitle={`Month: ${formatCurrency(metrics.revenue.this_month)}`}
            icon={
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
            color="blue"
            href="/billing"
            index={1}
          />
          <StatCard
            title="Arrivals Today"
            value={metrics.operations.expected_arrivals}
            subtitle={`Checked in: ${metrics.operations.check_ins_today}`}
            icon={
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
              </svg>
            }
            color="purple"
            href="/operations?tab=checkin"
            index={2}
          />
          <StatCard
            title="Pending Payments"
            value={metrics.pending_payments.pending_amount}
            prefix="₹"
            subtitle={`${metrics.pending_payments.pending_count} folios • ${metrics.pending_payments.overdue_count} overdue`}
            icon={
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
            color={metrics.pending_payments.overdue_count > 0 ? "red" : "orange"}
            href="/reports?type=outstanding"
            index={3}
          />
        </div>
      )}

      {/* Room Status Board */}
      {units.length > 0 && <RoomBoard units={units} />}

      {/* Quick Reports */}
      {metrics && (
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.35 }}
          className="grid grid-cols-1 sm:grid-cols-2 gap-3"
        >
          <Link href="/reports" className="group block rounded-xl border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md hover:border-gray-200 transition-all duration-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-emerald-50 flex items-center justify-center">
                  <svg className="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">Daily Collection</h4>
                  <p className="text-xs text-gray-500">Cash / UPI / Card breakdown</p>
                </div>
              </div>
              <svg className="w-4 h-4 text-gray-300 group-hover:text-gray-500 group-hover:translate-x-0.5 transition-all duration-200" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </Link>
          <Link href="/reports?type=outstanding" className="group block rounded-xl border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md hover:border-gray-200 transition-all duration-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-amber-50 flex items-center justify-center">
                  <svg className="w-4 h-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">Outstanding Dues</h4>
                  <p className="text-xs text-gray-500">Guests with pending balance</p>
                </div>
              </div>
              <svg className="w-4 h-4 text-gray-300 group-hover:text-gray-500 group-hover:translate-x-0.5 transition-all duration-200" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </Link>
        </motion.div>
      )}

      {/* Operations Summary */}
      {metrics && (
        <OperationsSummary
          operations={metrics.operations}
          housekeeping={metrics.housekeeping}
          laundry={metrics.laundry}
        />
      )}
    </div>
  );
}
