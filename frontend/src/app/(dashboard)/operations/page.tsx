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
    <Suspense fallback={<div className="flex justify-center py-12"><div className="w-6 h-6 rounded-full border-[2.5px] border-muted border-t-primary animate-spin" role="status" aria-label="Loading" /></div>}>
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
  const [searchQuery, setSearchQuery] = useState("");
  const [lastCheckoutReservationId, setLastCheckoutReservationId] = useState<string | null>(null);
  const [generatingInvoice, setGeneratingInvoice] = useState(false);

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
    setLastCheckoutReservationId(data.reservation_id);
    setActiveCheckout(null);
    setSuccessMsg("Check-out completed successfully! Generate invoice below.");
    fetchOperations();
  };

  const handleGenerateInvoice = async () => {
    if (!lastCheckoutReservationId) return;
    setGeneratingInvoice(true);
    try {
      // Get the folio for this reservation to find the invoice
      const folio = await api.get<any>(`/api/v1/billing/folios/reservation/${lastCheckoutReservationId}`);
      if (folio?.id) {
        // Generate PDF
        const result = await api.post<any>(`/api/v1/billing/invoices/${folio.id}/pdf`);
        const url = result?.url || result?.pdf_url || result?.download_url;
        if (url) {
          window.open(url, "_blank");
        } else {
          const pdfData = await api.get<any>(`/api/v1/billing/invoices/${folio.id}/pdf`);
          const pdfUrl = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
          if (pdfUrl) window.open(pdfUrl, "_blank");
        }
      }
    } catch { /* silent */ }
    finally {
      setGeneratingInvoice(false);
      setLastCheckoutReservationId(null);
      setSuccessMsg(null);
    }
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

  const handleNoShow = async (reservationId: string) => {
    if (!confirm("Mark this guest as No-Show? This will cancel the reservation.")) return;
    try {
      await api.post("/api/v1/operations/no-show", { reservation_id: reservationId });
      setSuccessMsg("Marked as No-Show. Reservation cancelled.");
      setTimeout(() => setSuccessMsg(null), 3000);
      fetchOperations();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to mark no-show");
    }
  };

  // Filter function for search
  const filterBySearch = (items: TodayOperation[]) => {
    if (!searchQuery) return items;
    const q = searchQuery.toLowerCase();
    return items.filter((item) => item.guest_name.toLowerCase().includes(q) || item.unit_number.toLowerCase().includes(q));
  };

  const tabs = [
    { id: "arrivals", label: "Expected Today", count: arrivals.length },
    { id: "departures", label: "Leaving Today", count: departures.length },
    { id: "inhouse", label: "In-House Guests", count: inHouse.length },
    { id: "checkin", label: "Check In" },
    { id: "checkout", label: "Check Out" },
  ];

  return (
    <div className="space-y-5 max-w-[1600px] mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <h1 className="text-xl lg:text-2xl font-bold text-gray-900 tracking-tight">Operations</h1>
        <button
          onClick={() => setShowWalkIn(true)}
          className="inline-flex items-center justify-center gap-2 px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 shadow-sm transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" /></svg>
          Walk-In Guest
        </button>
      </div>

      {/* Success message */}
      {successMsg && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700">
          <div className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
            {successMsg}
          </div>
          {lastCheckoutReservationId && (
            <div className="flex gap-2 mt-2">
              <button
                onClick={handleGenerateInvoice}
                disabled={generatingInvoice}
                className="px-3 py-1.5 text-xs bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
              >
                {generatingInvoice ? "Generating..." : "📄 Generate & Print Invoice"}
              </button>
              <button
                onClick={() => { setLastCheckoutReservationId(null); setSuccessMsg(null); }}
                className="px-3 py-1.5 text-xs border rounded-md hover:bg-white"
              >
                Dismiss
              </button>
            </div>
          )}
        </div>
      )}

      {/* Search */}
      <div className="relative">
        <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          type="text"
          placeholder="Search by guest name or unit number..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-input rounded-lg text-sm bg-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors"
        />
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
              tab === t.id ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300"
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
        <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700" role="alert">
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
          {filterBySearch(arrivals).length === 0 ? (
            <EmptyState message={searchQuery ? "No matching arrivals" : "No arrivals expected today"} />
          ) : (
            filterBySearch(arrivals).map((a) => (
              <OperationCard
                key={a.reservation_id}
                operation={a}
                actionLabel="Check In"
                actionColor="bg-green-600 hover:bg-green-700"
                onAction={() => setActiveCheckin(a)}
                showNoShow
                onNoShow={() => handleNoShow(a.reservation_id)}
              />
            ))
          )}
        </div>
      )}

      {/* Departures Tab */}
      {!loading && tab === "departures" && (
        <div className="space-y-3">
          {filterBySearch(departures).length === 0 ? (
            <EmptyState message={searchQuery ? "No matching departures" : "No departures today"} />
          ) : (
            filterBySearch(departures).map((d) => (
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
          {filterBySearch(inHouse).length === 0 ? (
            <EmptyState message={searchQuery ? "No matching guests" : "No guests currently in-house"} />
          ) : (
            filterBySearch(inHouse).map((g) => (
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

function OperationCard({ operation, actionLabel, actionColor, onAction, showExtend, onExtend, showNoShow, onNoShow }: {
  operation: TodayOperation;
  actionLabel: string;
  actionColor: string;
  onAction: () => void;
  showExtend?: boolean;
  onExtend?: (newDate: string) => void;
  showNoShow?: boolean;
  onNoShow?: () => void;
}) {
  const api = useApi();
  const { user } = useAuth();
  const [showExtendForm, setShowExtendForm] = useState(false);
  const [newDate, setNewDate] = useState("");
  const [conflictWarning, setConflictWarning] = useState<string | null>(null);
  const [checkingConflict, setCheckingConflict] = useState(false);

  // Calculate additional nights and cost
  const additionalNights = newDate ? Math.max(0, Math.round((new Date(newDate).getTime() - new Date(operation.check_out_date).getTime()) / 86400000)) : 0;

  // Check for conflicts when date changes
  const checkConflict = async (date: string) => {
    setNewDate(date);
    setConflictWarning(null);
    if (!date || !user?.property_id) return;
    setCheckingConflict(true);
    try {
      const data = await api.get<any>("/api/v1/reservations", { property_id: user.property_id, per_page: "100" });
      const reservations = Array.isArray(data) ? data : data?.data || [];
      // Check if any other reservation uses the same unit during the extended period
      const conflict = reservations.find((r: any) =>
        r.unit_id === operation.unit_id &&
        r.reservation_id !== operation.reservation_id &&
        r.id !== operation.reservation_id &&
        r.status !== "cancelled" &&
        r.status !== "checked_out" &&
        r.check_in_date < date &&
        r.check_out_date > operation.check_out_date
      );
      if (conflict) {
        setConflictWarning(`⚠️ Conflict: ${conflict.guest_name} has this unit booked from ${conflict.check_in_date}`);
      }
    } catch { /* silent */ }
    finally { setCheckingConflict(false); }
  };

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
          {showNoShow && onNoShow && (
            <button onClick={onNoShow} className="px-3 py-1.5 text-xs border border-red-300 text-red-600 rounded-md hover:bg-red-50 font-medium">
              No-Show
            </button>
          )}
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
            <input type="date" value={newDate} onChange={(e) => checkConflict(e.target.value)} min={operation.check_out_date} className="flex-1 rounded-md border border-gray-300 px-3 py-1.5 text-sm" />
            <button onClick={() => { if (newDate && onExtend && !conflictWarning) { onExtend(newDate); setShowExtendForm(false); } }} disabled={!newDate || !!conflictWarning || checkingConflict} className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
              Confirm
            </button>
          </div>
          {checkingConflict && <p className="text-xs text-gray-500">Checking availability...</p>}
          {conflictWarning && (
            <p className="text-xs text-red-600 bg-red-50 p-2 rounded border border-red-200">
              {conflictWarning}
            </p>
          )}
          {additionalNights > 0 && !conflictWarning && (
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
