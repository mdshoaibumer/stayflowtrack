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

      {/* Quick Actions */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <QuickAction href="/reservations?action=new" label="New Reservation" color="bg-blue-600" icon={<PlusIcon />} />
        <QuickAction href="/operations?action=walkin" label="Walk-In" color="bg-green-600" icon={<WalkIcon />} />
        <QuickAction href="/operations?tab=checkin" label="Check In" color="bg-purple-600" icon={<ArrowInIcon />} />
        <QuickAction href="/operations?tab=checkout" label="Check Out" color="bg-orange-600" icon={<ArrowOutIcon />} />
        <QuickAction href="/billing?action=charge" label="Quick Charge" color="bg-red-600" icon={<ChargeIcon />} />
        <QuickAction href="/laundry?action=new" label="Add Laundry" color="bg-cyan-600" icon={<LaundryIcon />} />
      </div>

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
