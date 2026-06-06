"use client";

import React, { useState, useEffect, useCallback, Suspense } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { useSearchParams } from "next/navigation";
import { CheckInForm, CheckOutForm } from "@/components/operations/CheckInOutForms";
import WalkInForm from "@/components/operations/WalkInForm";

interface TodayOperation {
  reservation_id: string;
  guest_name: string;
  unit_number: string;
  unit_id: string;
  check_in_date: string;
  check_out_date: string;
  status: string;
}

export default function OperationsPage() {
  return (
    <Suspense fallback={<div className="flex justify-center py-8"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>}>
      <OperationsContent />
    </Suspense>
  );
}

function OperationsContent() {
  const { user } = useAuth();
  const api = useApi();
  const searchParams = useSearchParams();
  const initialTab = searchParams.get("tab") || "arrivals";
  const action = searchParams.get("action");

  const [tab, setTab] = useState(initialTab);
  const [arrivals, setArrivals] = useState<TodayOperation[]>([]);
  const [departures, setDepartures] = useState<TodayOperation[]>([]);
  const [inHouse, setInHouse] = useState<TodayOperation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeCheckin, setActiveCheckin] = useState<TodayOperation | null>(null);
  const [activeCheckout, setActiveCheckout] = useState<TodayOperation | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [showWalkIn, setShowWalkIn] = useState(action === "walkin");

  const propertyId = user?.property_id || "";

  const fetchOperations = useCallback(async () => {
    if (!propertyId) return;
    setLoading(true);
    try {
      const today = new Date().toISOString().split("T")[0];
      const data = await api.get<any>("/api/v1/reservations", {
        property_id: propertyId,
        per_page: "100",
      });
      const reservations = Array.isArray(data) ? data : data?.data || [];
      setArrivals(reservations.filter((r: any) => r.check_in_date === today && (r.status === "confirmed" || r.status === "pending")));
      setDepartures(reservations.filter((r: any) => r.check_out_date === today && r.status === "checked_in"));
      setInHouse(reservations.filter((r: any) => r.status === "checked_in"));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load operations");
    } finally {
      setLoading(false);
    }
  }, [api, propertyId]);

  useEffect(() => {
    fetchOperations();
  }, [fetchOperations]);

  const handleCheckIn = async (data: any) => {
    await api.post("/api/v1/operations/check-in", data);
    setActiveCheckin(null);
    setSuccessMsg("Check-in completed successfully!");
    setTimeout(() => setSuccessMsg(null), 3000);
    fetchOperations();
  };

  const handleCheckOut = async (data: any) => {
    await api.post("/api/v1/operations/check-out", data);
    setActiveCheckout(null);
    setSuccessMsg("Check-out completed successfully!");
    setTimeout(() => setSuccessMsg(null), 3000);
    fetchOperations();
  };

  const handleExtendStay = async (reservationId: string, newCheckout: string) => {
    await api.post("/api/v1/operations/extend-stay", {
      reservation_id: reservationId,
      new_check_out_date: newCheckout,
    });
    setSuccessMsg("Stay extended successfully!");
    setTimeout(() => setSuccessMsg(null), 3000);
    fetchOperations();
  };

  const tabs = [
    { id: "arrivals", label: "Expected Today", count: arrivals.length },
    { id: "departures", label: "Leaving Today", count: departures.length },
    { id: "inhouse", label: "In-House Guests", count: inHouse.length },
    { id: "checkin", label: "Check In" },
    { id: "checkout", label: "Check Out" },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Operations</h1>
        <button
          onClick={() => setShowWalkIn(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" /></svg>
          Walk-In Guest
        </button>
      </div>

      {/* Success message */}
      {successMsg && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700 flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
          {successMsg}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
              tab === t.id ? "border-blue-600 text-blue-600" : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {t.label}
            {t.count !== undefined && (
              <span className="ml-1.5 px-1.5 py-0.5 text-xs rounded-full bg-gray-100">{t.count}</span>
            )}
          </button>
        ))}
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
          {error}
          <button onClick={fetchOperations} className="ml-2 underline">Retry</button>
        </div>
      )}

      {loading && (
        <div className="flex justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      )}

      {/* Arrivals Tab */}
      {!loading && tab === "arrivals" && (
        <div className="space-y-3">
          {arrivals.length === 0 ? (
            <EmptyState message="No arrivals expected today" />
          ) : (
            arrivals.map((a) => (
              <OperationCard
                key={a.reservation_id}
                operation={a}
                actionLabel="Check In"
                actionColor="bg-green-600 hover:bg-green-700"
                onAction={() => setActiveCheckin(a)}
              />
            ))
          )}
        </div>
      )}

      {/* Departures Tab */}
      {!loading && tab === "departures" && (
        <div className="space-y-3">
          {departures.length === 0 ? (
            <EmptyState message="No departures today" />
          ) : (
            departures.map((d) => (
              <OperationCard
                key={d.reservation_id}
                operation={d}
                actionLabel="Check Out"
                actionColor="bg-orange-600 hover:bg-orange-700"
                onAction={() => setActiveCheckout(d)}
                showExtend
                onExtend={(newDate) => handleExtendStay(d.reservation_id, newDate)}
              />
            ))
          )}
        </div>
      )}

      {/* In-House Guests Tab - extend stay from any day */}
      {!loading && tab === "inhouse" && (
        <div className="space-y-3">
          {inHouse.length === 0 ? (
            <EmptyState message="No guests currently in-house" />
          ) : (
            inHouse.map((g) => (
              <OperationCard
                key={g.reservation_id}
                operation={g}
                actionLabel="Check Out"
                actionColor="bg-orange-600 hover:bg-orange-700"
                onAction={() => setActiveCheckout(g)}
                showExtend
                onExtend={(newDate) => handleExtendStay(g.reservation_id, newDate)}
              />
            ))
          )}
        </div>
      )}

      {/* Check In Form Tab */}
      {!loading && tab === "checkin" && (
        <div className="max-w-lg">
          {activeCheckin ? (
            <CheckInForm
              reservationId={activeCheckin.reservation_id}
              unitId={activeCheckin.unit_id}
              onSubmit={handleCheckIn}
              onCancel={() => setActiveCheckin(null)}
            />
          ) : (
            <div>
              <p className="text-sm text-gray-600 mb-4">Select a reservation to check in:</p>
              {arrivals.length === 0 ? (
                <EmptyState message="No pending check-ins" />
              ) : (
                <div className="space-y-2">
                  {arrivals.map((a) => (
                    <button key={a.reservation_id} onClick={() => setActiveCheckin(a)} className="w-full text-left border rounded-lg p-3 hover:bg-gray-50">
                      <span className="font-medium">{a.guest_name}</span>
                      <span className="text-sm text-gray-500 ml-2">• {a.unit_number}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Check Out Form Tab */}
      {!loading && tab === "checkout" && (
        <div className="max-w-lg">
          {activeCheckout ? (
            <CheckOutForm
              reservationId={activeCheckout.reservation_id}
              guestName={activeCheckout.guest_name}
              onSubmit={handleCheckOut}
              onCancel={() => setActiveCheckout(null)}
            />
          ) : (
            <div>
              <p className="text-sm text-gray-600 mb-4">Select a reservation to check out:</p>
              {departures.length === 0 ? (
                <EmptyState message="No pending check-outs" />
              ) : (
                <div className="space-y-2">
                  {departures.map((d) => (
                    <button key={d.reservation_id} onClick={() => setActiveCheckout(d)} className="w-full text-left border rounded-lg p-3 hover:bg-gray-50">
                      <span className="font-medium">{d.guest_name}</span>
                      <span className="text-sm text-gray-500 ml-2">• {d.unit_number}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Check-in modal over current tab */}
      {activeCheckin && tab !== "checkin" && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
            <CheckInForm
              reservationId={activeCheckin.reservation_id}
              unitId={activeCheckin.unit_id}
              onSubmit={handleCheckIn}
              onCancel={() => setActiveCheckin(null)}
            />
          </div>
        </div>
      )}
      {activeCheckout && tab !== "checkout" && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full p-6">
            <CheckOutForm
              reservationId={activeCheckout.reservation_id}
              guestName={activeCheckout.guest_name}
              onSubmit={handleCheckOut}
              onCancel={() => setActiveCheckout(null)}
            />
          </div>
        </div>
      )}

      {/* Walk-In Form Modal */}
      {showWalkIn && (
        <WalkInForm
          onSuccess={(result) => {
            setShowWalkIn(false);
            setSuccessMsg(`Walk-in complete! ${result.guest_name} checked into ${result.unit_number}`);
            setTimeout(() => setSuccessMsg(null), 5000);
            fetchOperations();
          }}
          onCancel={() => setShowWalkIn(false)}
        />
      )}
    </div>
  );
}

function OperationCard({ operation, actionLabel, actionColor, onAction, showExtend, onExtend }: {
  operation: TodayOperation;
  actionLabel: string;
  actionColor: string;
  onAction: () => void;
  showExtend?: boolean;
  onExtend?: (newDate: string) => void;
}) {
  const [showExtendForm, setShowExtendForm] = useState(false);
  const [newDate, setNewDate] = useState("");

  // Calculate additional nights and cost
  const additionalNights = newDate ? Math.max(0, Math.round((new Date(newDate).getTime() - new Date(operation.check_out_date).getTime()) / 86400000)) : 0;

  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <div className="flex items-center justify-between">
        <div>
          <p className="font-medium">{operation.guest_name}</p>
          <p className="text-sm text-gray-500">
            {operation.unit_number} • {operation.check_in_date} → {operation.check_out_date}
          </p>
        </div>
        <div className="flex gap-2">
          {showExtend && onExtend && (
            <button onClick={() => setShowExtendForm(!showExtendForm)} className="px-3 py-1.5 text-xs border border-blue-300 text-blue-600 rounded-md hover:bg-blue-50 font-medium">
              Extend Stay
            </button>
          )}
          <button onClick={onAction} className={`px-3 py-1.5 text-xs text-white rounded-md ${actionColor}`}>
            {actionLabel}
          </button>
        </div>
      </div>
      {showExtendForm && (
        <div className="mt-3 pt-3 border-t space-y-2">
          <div className="flex items-center gap-2">
            <label className="text-xs text-gray-600 whitespace-nowrap">New checkout:</label>
            <input type="date" value={newDate} onChange={(e) => setNewDate(e.target.value)} min={operation.check_out_date} className="flex-1 rounded-md border border-gray-300 px-3 py-1.5 text-sm" />
            <button onClick={() => { if (newDate && onExtend) { onExtend(newDate); setShowExtendForm(false); } }} disabled={!newDate} className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
              Confirm
            </button>
          </div>
          {additionalNights > 0 && (
            <p className="text-xs text-blue-700 bg-blue-50 p-2 rounded">
              +{additionalNights} additional night{additionalNights > 1 ? "s" : ""} will be charged to guest&apos;s bill at the current room rate.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="text-center py-12 text-gray-500">
      <svg className="w-10 h-10 mx-auto mb-2 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p className="text-sm">{message}</p>
    </div>
  );
}
