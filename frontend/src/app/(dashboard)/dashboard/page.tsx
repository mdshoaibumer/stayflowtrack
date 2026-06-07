"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import Link from "next/link";

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

  useEffect(() => {
    fetchDashboard();
    const interval = setInterval(fetchDashboard, 30000);
    return () => clearInterval(interval);
  }, [fetchDashboard]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <p className="text-red-700">{error}</p>
        <button onClick={fetchDashboard} className="mt-2 text-sm text-red-600 underline">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-sm text-gray-500">
            {metrics?.date || new Date().toLocaleDateString()} • Auto-refreshes every 30s
          </p>
        </div>
      </div>

      {/* Morning Brief */}
      {metrics && (
        <div className="p-4 bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-100 rounded-lg">
          <p className="text-sm text-gray-700">
            <span className="font-semibold text-blue-900">Today:</span>{" "}
            {metrics.operations.expected_arrivals > 0 ? `${metrics.operations.expected_arrivals} arrival${metrics.operations.expected_arrivals > 1 ? "s" : ""}` : "No arrivals"}
            {", "}
            {metrics.operations.expected_departures > 0 ? `${metrics.operations.expected_departures} departure${metrics.operations.expected_departures > 1 ? "s" : ""}` : "no departures"}
            {". "}
            {metrics.occupancy.occupied_units}/{metrics.occupancy.total_units} units occupied ({metrics.occupancy.occupancy_rate.toFixed(0)}%).
            {metrics.pending_payments.pending_amount > 0 && (
              <Link href="/reports?type=outstanding" className="text-red-600 font-medium hover:underline"> ₹{metrics.pending_payments.pending_amount.toLocaleString()} outstanding →</Link>
            )}
          </p>
        </div>
      )}

      {/* Quick Actions */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
        <QuickAction href="/reservations?action=new" label="New Reservation" color="bg-blue-600" icon={<PlusIcon />} />
        <QuickAction href="/operations?action=walkin" label="Walk-In" color="bg-green-600" icon={<WalkIcon />} />
        <QuickAction href="/operations?tab=checkin" label="Check In" color="bg-purple-600" icon={<ArrowInIcon />} />
        <QuickAction href="/operations?tab=checkout" label="Check Out" color="bg-orange-600" icon={<ArrowOutIcon />} />
        <QuickAction href="/operations?tab=inhouse" label="Extend Stay" color="bg-indigo-600" icon={<ExtendIcon />} />
        <QuickAction href="/billing?action=charge" label="Quick Charge" color="bg-red-600" icon={<ChargeIcon />} />
        <QuickAction href="/laundry?action=new" label="Add Laundry" color="bg-cyan-600" icon={<LaundryIcon />} />
        <QuickAction href="/reports?type=end_of_day" label="Close Day" color="bg-gray-800" icon={<CloseDayIcon />} />
      </div>

      {/* Alerts: Overdue Checkouts & Tomorrow's Arrivals */}
      <DashboardAlerts propertyId={propertyId} metrics={metrics} />

      {/* Room Status Board */}
      {units.length > 0 && (
        <div className="border rounded-lg p-4 bg-white shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold text-gray-700 text-sm">Room Status Board</h3>
            <div className="flex gap-3 text-xs">
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-green-500"></span> Available</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-blue-500"></span> Occupied</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-yellow-500"></span> Reserved</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-orange-500"></span> Cleaning</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-red-500"></span> Maintenance</span>
            </div>
          </div>
          <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 gap-2">
            {units.map((unit) => (
              <div
                key={unit.id}
                className={`p-3 rounded-lg border-2 text-center cursor-pointer transition-all hover:shadow-md ${getUnitStatusStyle(unit.status)}`}
                title={unit.guest_name ? `Guest: ${unit.guest_name}` : unit.status}
              >
                <div className="font-bold text-sm">{unit.unit_number}</div>
                <div className="text-[10px] uppercase tracking-wide mt-0.5 opacity-75">{unit.status.replace('_', ' ')}</div>
                {unit.guest_name && <div className="text-[10px] mt-0.5 truncate font-medium">{unit.guest_name}</div>}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* KPI Cards */}
      {metrics && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <Link href="/reservations" className="block">
            <KPICard
              title="Occupancy"
              value={`${metrics.occupancy.occupancy_rate.toFixed(0)}%`}
              subtitle={`${metrics.occupancy.occupied_units}/${metrics.occupancy.total_units} units`}
              color="border-l-green-500"
            />
          </Link>
          <Link href="/billing" className="block">
            <KPICard
              title="Revenue Today"
              value={formatCurrency(metrics.revenue.today)}
              subtitle={`Month: ${formatCurrency(metrics.revenue.this_month)}`}
              color="border-l-blue-500"
            />
          </Link>
          <Link href="/operations?tab=checkin" className="block">
            <KPICard
              title="Arrivals Today"
              value={String(metrics.operations.expected_arrivals)}
              subtitle={`Checked in: ${metrics.operations.check_ins_today}`}
              color="border-l-purple-500"
            />
          </Link>
          <Link href="/reports?type=outstanding" className="block">
            <KPICard
              title="Pending Payments"
              value={formatCurrency(metrics.pending_payments.pending_amount)}
              subtitle={`${metrics.pending_payments.pending_count} folios • ${metrics.pending_payments.overdue_count} overdue`}
              color="border-l-orange-500"
            />
          </Link>
        </div>
      )}

      {/* Quick Reports Row */}
      {metrics && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Link href="/reports" className="block border rounded-lg p-4 bg-white shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-sm font-medium text-gray-700">Daily Collection</h4>
                <p className="text-xs text-gray-500 mt-0.5">Cash / UPI / Card breakdown for today</p>
              </div>
              <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
            </div>
          </Link>
          <Link href="/reports" className="block border rounded-lg p-4 bg-white shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-sm font-medium text-gray-700">Outstanding Dues</h4>
                <p className="text-xs text-gray-500 mt-0.5">Guests with pending balance</p>
              </div>
              <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
            </div>
          </Link>
        </div>
      )}

      {/* Operations Summary */}
      {metrics && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Today's Operations */}
          <div className="border rounded-lg p-4 bg-white shadow-sm">
            <h3 className="font-semibold text-gray-700 mb-3 text-sm">Today&apos;s Operations</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Check-ins</span>
                <span className="font-medium text-green-600">{metrics.operations.check_ins_today}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Check-outs</span>
                <span className="font-medium text-orange-600">{metrics.operations.check_outs_today}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Expected Arrivals</span>
                <span className="font-medium">{metrics.operations.expected_arrivals}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Expected Departures</span>
                <span className="font-medium">{metrics.operations.expected_departures}</span>
              </div>
            </div>
            <Link href="/operations" className="mt-3 block text-xs text-blue-600 hover:underline">
              View all operations →
            </Link>
          </div>

          {/* Housekeeping */}
          <div className="border rounded-lg p-4 bg-white shadow-sm">
            <h3 className="font-semibold text-gray-700 mb-3 text-sm">Housekeeping</h3>
            <div className="space-y-2 text-sm">
              <StatusRow label="Dirty" count={metrics.housekeeping.counts?.dirty || 0} color="bg-red-500" />
              <StatusRow label="Cleaning" count={metrics.housekeeping.counts?.cleaning || 0} color="bg-yellow-500" />
              <StatusRow label="Inspection" count={metrics.housekeeping.counts?.inspection || 0} color="bg-blue-500" />
              <StatusRow label="Ready" count={metrics.housekeeping.counts?.ready || 0} color="bg-green-500" />
            </div>
            <Link href="/housekeeping" className="mt-3 block text-xs text-blue-600 hover:underline">
              View board →
            </Link>
          </div>

          {/* Laundry */}
          <div className="border rounded-lg p-4 bg-white shadow-sm">
            <h3 className="font-semibold text-gray-700 mb-3 text-sm">Laundry</h3>
            <div className="space-y-2 text-sm">
              <StatusRow label="Received" count={metrics.laundry.counts?.received || 0} color="bg-gray-500" />
              <StatusRow label="Washing" count={metrics.laundry.counts?.washing || 0} color="bg-blue-500" />
              <StatusRow label="Ready" count={metrics.laundry.counts?.ready || 0} color="bg-green-500" />
            </div>
            <Link href="/laundry" className="mt-3 block text-xs text-blue-600 hover:underline">
              View tracker →
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}

function KPICard({ title, value, subtitle, color }: { title: string; value: string; subtitle: string; color: string }) {
  return (
    <div className={`border rounded-lg p-4 bg-white shadow-sm border-l-4 ${color} hover:shadow-md transition-shadow`}>
      <span className="text-xs text-gray-500 uppercase font-medium">{title}</span>
      <div className="text-xl lg:text-2xl font-bold mt-1">{value}</div>
      <div className="text-xs text-gray-500 mt-1">{subtitle}</div>
    </div>
  );
}

function StatusRow({ label, count, color }: { label: string; count: number; color: string }) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${color}`}></div>
        <span className="text-gray-600">{label}</span>
      </div>
      <span className="font-medium">{count}</span>
    </div>
  );
}

function QuickAction({ href, label, color, icon }: { href: string; label: string; color: string; icon: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={`flex flex-col items-center gap-1.5 p-3 rounded-lg text-white ${color} hover:opacity-90 transition-opacity shadow-sm`}
    >
      {icon}
      <span className="text-xs font-medium text-center">{label}</span>
    </Link>
  );
}

function PlusIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>;
}
function WalkIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>;
}
function ArrowInIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" /></svg>;
}
function ArrowOutIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>;
}
function LaundryIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>;
}
function ChargeIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
}
function ExtendIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>;
}
function CloseDayIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
}

function DashboardAlerts({ propertyId, metrics }: { propertyId: string; metrics: DashboardMetrics | null }) {
  const api = useApi();
  const [overdueCheckouts, setOverdueCheckouts] = useState<any[]>([]);
  const [tomorrowArrivals, setTomorrowArrivals] = useState<any[]>([]);

  useEffect(() => {
    if (!propertyId) return;
    const fetchAlerts = async () => {
      try {
        const data = await api.get<any>("/api/v1/reservations", { property_id: propertyId, per_page: "100" });
        const reservations = Array.isArray(data) ? data : data?.data || [];
        const today = new Date().toISOString().split("T")[0];
        const tomorrow = new Date(Date.now() + 86400000).toISOString().split("T")[0];

        // Overdue: checkout date is today or earlier but still checked_in
        const overdue = reservations.filter((r: any) => r.status === "checked_in" && r.check_out_date <= today);
        setOverdueCheckouts(overdue);

        // Tomorrow arrivals
        const arriving = reservations.filter((r: any) => r.check_in_date === tomorrow && (r.status === "confirmed" || r.status === "pending"));
        setTomorrowArrivals(arriving);
      } catch { /* silent */ }
    };
    fetchAlerts();
  }, [api, propertyId]);

  if (overdueCheckouts.length === 0 && tomorrowArrivals.length === 0) return null;

  return (
    <div className="space-y-3">
      {overdueCheckouts.length > 0 && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <svg className="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
            <h3 className="text-sm font-bold text-red-800">Overdue Checkouts ({overdueCheckouts.length})</h3>
          </div>
          <div className="space-y-1">
            {overdueCheckouts.map((r: any) => (
              <div key={r.reservation_id || r.id} className="flex items-center justify-between text-sm">
                <span className="text-red-700">
                  <strong>{r.unit_number}</strong> — {r.guest_name} (was due: {r.check_out_date})
                </span>
                <Link href="/operations?tab=checkout" className="text-xs bg-red-600 text-white px-2 py-1 rounded hover:bg-red-700">
                  Check Out
                </Link>
              </div>
            ))}
          </div>
        </div>
      )}
      {tomorrowArrivals.length > 0 && (
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <svg className="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <h3 className="text-sm font-bold text-blue-800">Tomorrow&apos;s Arrivals ({tomorrowArrivals.length})</h3>
          </div>
          <div className="space-y-1">
            {tomorrowArrivals.map((r: any) => (
              <div key={r.reservation_id || r.id} className="text-sm text-blue-700">
                <strong>{r.unit_number}</strong> — {r.guest_name} (until {r.check_out_date})
              </div>
            ))}
          </div>
          <p className="text-xs text-blue-600 mt-2">Ensure rooms are ready for tomorrow&apos;s guests.</p>
        </div>
      )}
    </div>
  );
}

function getUnitStatusStyle(status: string): string {
  switch (status) {
    case "available":
      return "border-green-300 bg-green-50 text-green-800";
    case "occupied":
      return "border-blue-300 bg-blue-50 text-blue-800";
    case "reserved":
      return "border-yellow-300 bg-yellow-50 text-yellow-800";
    case "cleaning":
      return "border-orange-300 bg-orange-50 text-orange-800";
    case "maintenance":
      return "border-red-300 bg-red-50 text-red-800";
    default:
      return "border-gray-300 bg-gray-50 text-gray-800";
  }
}
