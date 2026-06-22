"use client";

import React, { useEffect, useState, useCallback } from "react";

interface DashboardMetrics {
  date: string;
  occupancy: { total_units: number; occupied_units: number; available_units: number; occupancy_rate: number };
  revenue: { today: number; this_week: number; this_month: number; currency: string };
  operations: { check_ins_today: number; check_outs_today: number; expected_arrivals: number; expected_departures: number };
  housekeeping: { counts: Record<string, number>; total: number };
  laundry: { counts: Record<string, number>; total: number };
  pending_payments: { pending_count: number; pending_amount: number; overdue_count: number };
}

interface DashboardProps {
  propertyId: string;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", minimumFractionDigits: 0 }).format(amount);
}

function getToken(): string {
  if (typeof window !== "undefined") return sessionStorage.getItem("sf_at") || "";
  return "";
}

export default function OperationsDashboard({ propertyId }: DashboardProps) {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchDashboard = useCallback(async () => {
    try {
      const resp = await fetch(`/api/v1/dashboard/${propertyId}`, {
        headers: { Authorization: `Bearer ${getToken()}` },
      });
      if (resp.ok) {
        const data = await resp.json();
        setMetrics(data.data);
      }
    } finally {
      setLoading(false);
    }
  }, [propertyId]);

  useEffect(() => {
    fetchDashboard();
    const interval = setInterval(fetchDashboard, 30000); // Auto-refresh every 30s
    return () => clearInterval(interval);
  }, [fetchDashboard]);

  if (loading) {
    return <div className="flex items-center justify-center h-64"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div></div>;
  }
  if (!metrics) return <div className="text-red-500">Failed to load dashboard</div>;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Operations Dashboard</h1>
      <p className="text-sm text-gray-500">{metrics.date} • Auto-refreshes every 30s</p>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <KPICard
          title="Occupancy"
          value={`${metrics.occupancy.occupancy_rate.toFixed(0)}%`}
          subtitle={`${metrics.occupancy.occupied_units}/${metrics.occupancy.total_units} units`}
          color="bg-green-500"
        />
        <KPICard
          title="Revenue Today"
          value={formatCurrency(metrics.revenue.today)}
          subtitle={`Month: ${formatCurrency(metrics.revenue.this_month)}`}
          color="bg-blue-500"
        />
        <KPICard
          title="Check-ins Today"
          value={String(metrics.operations.check_ins_today)}
          subtitle={`Expected: ${metrics.operations.expected_arrivals}`}
          color="bg-purple-500"
        />
        <KPICard
          title="Pending Payments"
          value={formatCurrency(metrics.pending_payments.pending_amount)}
          subtitle={`${metrics.pending_payments.pending_count} folios`}
          color="bg-orange-500"
        />
      </div>

      {/* Operations Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Operations */}
        <div className="border rounded-lg p-4 bg-white shadow-sm">
          <h3 className="font-semibold text-gray-700 mb-3">Today&apos;s Operations</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between"><span>Check-ins</span><span className="font-medium text-green-600">{metrics.operations.check_ins_today}</span></div>
            <div className="flex justify-between"><span>Check-outs</span><span className="font-medium text-orange-600">{metrics.operations.check_outs_today}</span></div>
            <div className="flex justify-between"><span>Expected Arrivals</span><span className="font-medium">{metrics.operations.expected_arrivals}</span></div>
            <div className="flex justify-between"><span>Expected Departures</span><span className="font-medium">{metrics.operations.expected_departures}</span></div>
          </div>
        </div>

        {/* Housekeeping */}
        <div className="border rounded-lg p-4 bg-white shadow-sm">
          <h3 className="font-semibold text-gray-700 mb-3">Housekeeping</h3>
          <div className="space-y-2 text-sm">
            <StatusRow label="Dirty" count={metrics.housekeeping.counts?.dirty || 0} color="bg-red-500" />
            <StatusRow label="Cleaning" count={metrics.housekeeping.counts?.cleaning || 0} color="bg-yellow-500" />
            <StatusRow label="Inspection" count={metrics.housekeeping.counts?.inspection || 0} color="bg-blue-500" />
            <StatusRow label="Ready" count={metrics.housekeeping.counts?.ready || 0} color="bg-green-500" />
          </div>
        </div>

        {/* Laundry */}
        <div className="border rounded-lg p-4 bg-white shadow-sm">
          <h3 className="font-semibold text-gray-700 mb-3">Laundry</h3>
          <div className="space-y-2 text-sm">
            <StatusRow label="Received" count={metrics.laundry.counts?.received || 0} color="bg-gray-500" />
            <StatusRow label="Washing" count={metrics.laundry.counts?.washing || 0} color="bg-blue-500" />
            <StatusRow label="Ready" count={metrics.laundry.counts?.ready || 0} color="bg-green-500" />
          </div>
        </div>
      </div>
    </div>
  );
}

function KPICard({ title, value, subtitle, color }: { title: string; value: string; subtitle: string; color: string }) {
  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <div className="flex items-center gap-2 mb-1">
        <div className={`w-2 h-2 rounded-full ${color}`}></div>
        <span className="text-xs text-gray-500 uppercase">{title}</span>
      </div>
      <div className="text-2xl font-bold">{value}</div>
      <div className="text-xs text-gray-500 mt-1">{subtitle}</div>
    </div>
  );
}

function StatusRow({ label, count, color }: { label: string; count: number; color: string }) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${color}`}></div>
        <span>{label}</span>
      </div>
      <span className="font-medium">{count}</span>
    </div>
  );
}
