"use client";

import React, { useState, useEffect, useCallback, Suspense } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { useSearchParams } from "next/navigation";
import ReservationCalendar from "@/components/calendar/ReservationCalendar";

interface Reservation {
  id: string;
  guest_name: string;
  guest_id: string;
  unit_number: string;
  unit_id: string;
  check_in_date: string;
  check_out_date: string;
  status: string;
  nights: number;
  total_amount: number;
  created_at: string;
}

interface AvailableUnit {
  id: string;
  unit_number: string;
  unit_type_name: string;
  floor: number;
  rate: number;
}

interface GuestOption {
  id: string;
  full_name: string;
  phone: string;
}

const statusColors: Record<string, string> = {
  confirmed: "bg-blue-100 text-blue-700",
  checked_in: "bg-green-100 text-green-700",
  checked_out: "bg-gray-100 text-gray-700",
  cancelled: "bg-red-100 text-red-700",
  pending: "bg-yellow-100 text-yellow-700",
};

export default function ReservationsPage() {
  return (
    <Suspense fallback={<div className="flex justify-center py-8"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>}>
      <ReservationsContent />
    </Suspense>
  );
}

function ReservationsContent() {
  const { user } = useAuth();
  const api = useApi();
  const searchParams = useSearchParams();
  const action = searchParams.get("action");

  const [view, setView] = useState<"calendar" | "list">("calendar");
  const [showNewBooking, setShowNewBooking] = useState(action === "new");
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedReservation, setSelectedReservation] = useState<Reservation | null>(null);

  const propertyId = user?.property_id || "";

  const fetchReservations = useCallback(async () => {
    if (!propertyId) return;
    setLoading(true);
    try {
      const data = await api.get<any>("/api/v1/reservations", { property_id: propertyId, per_page: "50" });
      setReservations(Array.isArray(data) ? data : data?.data || []);
    } catch {
      // Silent — calendar is primary view
    } finally {
      setLoading(false);
    }
  }, [api, propertyId]);

  useEffect(() => {
    if (view === "list") fetchReservations();
  }, [view, fetchReservations]);

  const handleMoveBooking = async (reservationId: string, newUnitId: string, newCheckIn: string, newCheckOut: string) => {
    await api.post("/api/v1/calendar/move", {
      reservation_id: reservationId,
      new_unit_id: newUnitId,
      new_check_in: newCheckIn,
      new_check_out: newCheckOut,
    });
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Reservations</h1>
        <div className="flex items-center gap-2">
          {/* View Toggle */}
          <div className="flex border rounded-md overflow-hidden">
            <button
              onClick={() => setView("calendar")}
              className={`px-3 py-1.5 text-xs font-medium ${view === "calendar" ? "bg-blue-600 text-white" : "bg-white text-gray-600 hover:bg-gray-50"}`}
            >
              Calendar
            </button>
            <button
              onClick={() => setView("list")}
              className={`px-3 py-1.5 text-xs font-medium ${view === "list" ? "bg-blue-600 text-white" : "bg-white text-gray-600 hover:bg-gray-50"}`}
            >
              List
            </button>
          </div>
          <button
            onClick={() => setShowNewBooking(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New Booking
          </button>
        </div>
      </div>

      {/* Calendar View */}
      {view === "calendar" && propertyId && (
        <ReservationCalendar
          propertyId={propertyId}
          onMoveBooking={handleMoveBooking}
          onCellClick={(unitId, date) => {
            setShowNewBooking(true);
          }}
          onEntryClick={(entry) => {
            setSelectedReservation(entry as any);
          }}
        />
      )}

      {/* List View */}
      {view === "list" && (
        <div>
          {loading && (
            <div className="flex justify-center py-8">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            </div>
          )}
          {!loading && reservations.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              <p className="font-medium">No reservations found</p>
              <p className="text-sm mt-1">Create a new booking to get started</p>
            </div>
          )}
          {!loading && reservations.length > 0 && (
            <div className="border rounded-lg overflow-hidden bg-white">
              {/* Mobile cards */}
              <div className="lg:hidden space-y-3 p-3">
                {reservations.map((r) => (
                  <div key={r.id} onClick={() => setSelectedReservation(r)} className="border rounded-lg p-3 cursor-pointer hover:bg-gray-50">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{r.guest_name}</span>
                      <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[r.status] || "bg-gray-100"}`}>
                        {r.status?.replace("_", " ")}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      {r.unit_number} • {r.check_in_date} → {r.check_out_date}
                    </p>
                  </div>
                ))}
              </div>
              {/* Desktop table */}
              <table className="w-full text-sm hidden lg:table">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Guest</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Unit</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Check-in</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Check-out</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
                    <th className="text-left px-4 py-3 font-medium text-gray-600">Amount</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {reservations.map((r) => (
                    <tr key={r.id} onClick={() => setSelectedReservation(r)} className="hover:bg-gray-50 cursor-pointer">
                      <td className="px-4 py-3 font-medium">{r.guest_name}</td>
                      <td className="px-4 py-3">{r.unit_number}</td>
                      <td className="px-4 py-3">{r.check_in_date}</td>
                      <td className="px-4 py-3">{r.check_out_date}</td>
                      <td className="px-4 py-3">
                        <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[r.status] || "bg-gray-100"}`}>
                          {r.status?.replace("_", " ")}
                        </span>
                      </td>
                      <td className="px-4 py-3">₹{r.total_amount?.toLocaleString() || "0"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* New Booking Modal */}
      {showNewBooking && (
        <NewBookingModal
          propertyId={propertyId}
          onClose={() => setShowNewBooking(false)}
          onCreated={() => { setShowNewBooking(false); fetchReservations(); }}
        />
      )}

      {/* Reservation Detail */}
      {selectedReservation && (
        <ReservationDetailDrawer
          reservation={selectedReservation}
          onClose={() => setSelectedReservation(null)}
          onUpdated={fetchReservations}
        />
      )}
    </div>
  );
}

function NewBookingModal({ propertyId, onClose, onCreated }: { propertyId: string; onClose: () => void; onCreated: () => void }) {
  const api = useApi();
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Step 1: Dates & Availability
  const [checkIn, setCheckIn] = useState("");
  const [checkOut, setCheckOut] = useState("");
  const [availableUnits, setAvailableUnits] = useState<AvailableUnit[]>([]);
  const [selectedUnit, setSelectedUnit] = useState<AvailableUnit | null>(null);

  // Step 2: Guest
  const [guestSearch, setGuestSearch] = useState("");
  const [guestResults, setGuestResults] = useState<GuestOption[]>([]);
  const [selectedGuest, setSelectedGuest] = useState<GuestOption | null>(null);

  // Step 3: Rate & Advance
  const [ratePerNight, setRatePerNight] = useState(0);
  const [notes, setNotes] = useState("");
  const [bookingSource, setBookingSource] = useState("phone");
  const [advanceAmount, setAdvanceAmount] = useState(0);
  const [advanceMethod, setAdvanceMethod] = useState("upi");
  const [advanceReference, setAdvanceReference] = useState("");

  const searchAvailability = async () => {
    if (!checkIn || !checkOut) { setError("Select check-in and check-out dates"); return; }
    setLoading(true);
    setError(null);
    try {
      const data = await api.get<any>("/api/v1/reservations/availability", {
        property_id: propertyId,
        check_in: checkIn,
        check_out: checkOut,
      });
      const units = Array.isArray(data) ? data : data?.units || data?.data || [];
      setAvailableUnits(units);
      if (units.length === 0) setError("No units available for selected dates");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to check availability");
    } finally {
      setLoading(false);
    }
  };

  const searchGuests = async (query: string) => {
    setGuestSearch(query);
    if (query.length < 2) { setGuestResults([]); return; }
    try {
      const data = await api.get<any>("/api/v1/guests/search", { search: query });
      setGuestResults(Array.isArray(data) ? data : data?.data || []);
    } catch { /* silent */ }
  };

  const handleBook = async () => {
    if (!selectedUnit || !selectedGuest) return;
    setLoading(true);
    setError(null);
    try {
      await api.post("/api/v1/reservations", {
        property_id: propertyId,
        guest_id: selectedGuest.id,
        unit_id: selectedUnit.id,
        check_in_date: checkIn,
        check_out_date: checkOut,
        rate_per_night: ratePerNight,
        num_guests: 1,
        booking_source: bookingSource,
        notes,
        advance_amount: advanceAmount > 0 ? advanceAmount : undefined,
        advance_method: advanceAmount > 0 ? advanceMethod : undefined,
        advance_reference: advanceAmount > 0 ? advanceReference : undefined,
      });
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create reservation");
    } finally {
      setLoading(false);
    }
  };

  const nights = checkIn && checkOut ? Math.max(1, Math.round((new Date(checkOut).getTime() - new Date(checkIn).getTime()) / 86400000)) : 0;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">New Reservation</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          {/* Step Indicator */}
          <div className="flex items-center gap-2 mb-6">
            {[1, 2, 3].map((s) => (
              <div key={s} className={`flex-1 h-1 rounded-full ${step >= s ? "bg-blue-600" : "bg-gray-200"}`} />
            ))}
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>
          )}

          {/* Step 1: Dates & Unit */}
          {step === 1 && (
            <div className="space-y-4">
              <p className="text-sm font-medium text-gray-700">Step 1: Select Dates & Unit</p>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-gray-600">Check-in</label>
                  <input type="date" value={checkIn} onChange={(e) => { setCheckIn(e.target.value); setAvailableUnits([]); setSelectedUnit(null); }} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
                </div>
                <div>
                  <label className="block text-sm text-gray-600">Check-out</label>
                  <input type="date" value={checkOut} onChange={(e) => { setCheckOut(e.target.value); setAvailableUnits([]); setSelectedUnit(null); }} min={checkIn} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
                </div>
              </div>
              {nights > 0 && <p className="text-xs text-gray-500">{nights} night{nights > 1 ? "s" : ""}</p>}
              <button onClick={searchAvailability} disabled={loading || !checkIn || !checkOut} className="w-full py-2 text-sm bg-gray-900 text-white rounded-md hover:bg-gray-800 disabled:opacity-50">
                {loading ? "Checking..." : "Check Availability"}
              </button>

              {availableUnits.length > 0 && (
                <div className="space-y-2 mt-4">
                  <p className="text-sm font-medium text-gray-700">{availableUnits.length} units available</p>
                  {availableUnits.map((unit) => (
                    <label key={unit.id} className={`flex items-center justify-between p-3 border rounded-lg cursor-pointer ${selectedUnit?.id === unit.id ? "border-blue-500 bg-blue-50" : "hover:bg-gray-50"}`}>
                      <div className="flex items-center gap-3">
                        <input type="radio" name="unit" checked={selectedUnit?.id === unit.id} onChange={() => { setSelectedUnit(unit); setRatePerNight(unit.rate || 0); }} className="text-blue-600" />
                        <div>
                          <span className="font-medium text-sm">{unit.unit_number}</span>
                          <span className="text-xs text-gray-500 ml-2">{unit.unit_type_name}</span>
                        </div>
                      </div>
                      {unit.rate > 0 && <span className="text-sm font-medium">₹{unit.rate}/night</span>}
                    </label>
                  ))}
                </div>
              )}

              <div className="flex justify-end pt-2">
                <button onClick={() => setStep(2)} disabled={!selectedUnit} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
                  Next
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Guest */}
          {step === 2 && (
            <GuestSelectionStep
              api={api}
              guestSearch={guestSearch}
              setGuestSearch={setGuestSearch}
              guestResults={guestResults}
              searchGuests={searchGuests}
              selectedGuest={selectedGuest}
              setSelectedGuest={setSelectedGuest}
              onBack={() => setStep(1)}
              onNext={() => setStep(3)}
            />
          )}

          {/* Step 3: Confirm */}
          {step === 3 && (
            <div className="space-y-4">
              <p className="text-sm font-medium text-gray-700">Step 3: Confirm & Rate</p>
              <div className="bg-gray-50 rounded-lg p-4 space-y-2 text-sm">
                <div className="flex justify-between"><span className="text-gray-500">Guest</span><span className="font-medium">{selectedGuest?.full_name}</span></div>
                <div className="flex justify-between"><span className="text-gray-500">Phone</span><span className="font-medium">{selectedGuest?.phone}</span></div>
                <div className="flex justify-between"><span className="text-gray-500">Unit</span><span className="font-medium">{selectedUnit?.unit_number} ({selectedUnit?.unit_type_name})</span></div>
                <div className="flex justify-between"><span className="text-gray-500">Dates</span><span className="font-medium">{checkIn} → {checkOut}</span></div>
                <div className="flex justify-between"><span className="text-gray-500">Nights</span><span className="font-medium">{nights}</span></div>
              </div>
              <div>
                <label className="block text-sm text-gray-600">Booking Source</label>
                <select value={bookingSource} onChange={(e) => setBookingSource(e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                  <option value="phone">Phone Call</option>
                  <option value="walk_in">Walk-In</option>
                  <option value="email">Email</option>
                  <option value="website">Website</option>
                  <option value="ota_booking_com">Booking.com</option>
                  <option value="ota_makemytrip">MakeMyTrip</option>
                  <option value="ota_goibibo">Goibibo</option>
                  <option value="ota_airbnb">Airbnb</option>
                  <option value="ota_other">Other OTA</option>
                  <option value="referral">Referral</option>
                  <option value="corporate">Corporate</option>
                  <option value="repeat">Repeat Guest</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600">Rate per Night (₹)</label>
                <input type="number" min="0" value={ratePerNight} onChange={(e) => setRatePerNight(Number(e.target.value))} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
              </div>
              {ratePerNight > 0 && (
                <div className="p-3 bg-blue-50 rounded-lg">
                  <div className="flex justify-between text-sm"><span>Room Charges</span><span>₹{(ratePerNight * nights).toLocaleString()}</span></div>
                  <div className="flex justify-between text-sm text-gray-500"><span>GST (12%)</span><span>₹{Math.round(ratePerNight * nights * 0.12).toLocaleString()}</span></div>
                  <div className="flex justify-between text-sm font-bold border-t pt-1 mt-1"><span>Estimated Total</span><span>₹{Math.round(ratePerNight * nights * 1.12).toLocaleString()}</span></div>
                </div>
              )}
              <div>
                <label className="block text-sm text-gray-600">Notes</label>
                <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Optional notes..." />
              </div>

              {/* Advance Payment (optional) */}
              <div className="border rounded-lg p-3 bg-green-50">
                <p className="text-sm font-medium text-green-800 mb-2">Advance Payment (optional)</p>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs text-gray-600">Amount (₹)</label>
                    <input type="number" min="0" value={advanceAmount || ""} onChange={(e) => setAdvanceAmount(Number(e.target.value))} placeholder="0" className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-600">Method</label>
                    <select value={advanceMethod} onChange={(e) => setAdvanceMethod(e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                      <option value="cash">Cash</option>
                      <option value="upi">UPI</option>
                      <option value="card">Card</option>
                      <option value="bank_transfer">Bank Transfer</option>
                    </select>
                  </div>
                </div>
                {advanceAmount > 0 && advanceMethod !== "cash" && (
                  <div className="mt-2">
                    <label className="block text-xs text-gray-600">Reference / Transaction ID</label>
                    <input type="text" value={advanceReference} onChange={(e) => setAdvanceReference(e.target.value)} placeholder="UPI ref / Card last 4" className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
                  </div>
                )}
              </div>
              <div className="flex justify-between pt-2">
                <button onClick={() => setStep(2)} className="px-4 py-2 text-sm border rounded-md hover:bg-gray-50">Back</button>
                <button onClick={handleBook} disabled={loading} className="px-4 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50">
                  {loading ? "Creating..." : "Confirm Booking"}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ReservationDetailDrawer({ reservation, onClose, onUpdated }: { reservation: Reservation; onClose: () => void; onUpdated: () => void }) {
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [showCancel, setShowCancel] = useState(false);

  const handleCancel = async () => {
    setLoading(true);
    try {
      await api.post(`/api/v1/reservations/${reservation.id}/cancel`, { reason: "Cancelled by staff" });
      onUpdated();
      onClose();
    } catch { /* silent */ }
    finally { setLoading(false); }
  };

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await api.post(`/api/v1/reservations/${reservation.id}/confirm`);
      onUpdated();
      onClose();
    } catch { /* silent */ }
    finally { setLoading(false); }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex justify-end" onClick={onClose}>
      <div className="bg-white w-full max-w-md h-full overflow-y-auto shadow-xl" onClick={(e) => e.stopPropagation()}>
        <div className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Reservation</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          <div className="space-y-3 text-sm">
            <div className="flex justify-between"><span className="text-gray-500">Guest</span><span className="font-medium">{reservation.guest_name}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">Unit</span><span className="font-medium">{reservation.unit_number}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">Check-in</span><span>{reservation.check_in_date}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">Check-out</span><span>{reservation.check_out_date}</span></div>
            <div className="flex justify-between"><span className="text-gray-500">Status</span><span className={`px-2 py-0.5 rounded-full text-xs ${statusColors[reservation.status]}`}>{reservation.status?.replace("_", " ")}</span></div>
            {reservation.total_amount > 0 && (
              <div className="flex justify-between"><span className="text-gray-500">Amount</span><span className="font-medium">₹{reservation.total_amount?.toLocaleString()}</span></div>
            )}
          </div>

          {/* Actions */}
          <div className="border-t pt-4 space-y-2">
            {reservation.status === "pending" && (
              <button onClick={handleConfirm} disabled={loading} className="w-full py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
                {loading ? "Confirming..." : "✓ Confirm Reservation"}
              </button>
            )}
            {(reservation.status === "confirmed" || reservation.status === "pending") && (
              <>
                <a href={`/operations?tab=checkin&reservation=${reservation.id}`} className="block w-full text-center py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700">
                  Check In
                </a>
                <button onClick={() => setShowCancel(true)} className="w-full py-2 text-sm border border-red-300 text-red-600 rounded-md hover:bg-red-50">
                  Cancel Reservation
                </button>
              </>
            )}
            {reservation.status === "checked_in" && (
              <a href={`/operations?tab=checkout&reservation=${reservation.id}`} className="block w-full text-center py-2 text-sm bg-orange-600 text-white rounded-md hover:bg-orange-700">
                Check Out
              </a>
            )}
          </div>

          {/* Cancel Confirmation */}
          {showCancel && (
            <div className="border border-red-200 rounded-lg p-4 bg-red-50">
              <p className="text-sm text-red-700 font-medium mb-3">Are you sure you want to cancel this reservation?</p>
              <div className="flex gap-2">
                <button onClick={() => setShowCancel(false)} className="flex-1 py-2 text-sm border rounded-md hover:bg-white">No, Keep</button>
                <button onClick={handleCancel} disabled={loading} className="flex-1 py-2 text-sm bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50">
                  {loading ? "Cancelling..." : "Yes, Cancel"}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function GuestSelectionStep({ api, guestSearch, setGuestSearch, guestResults, searchGuests, selectedGuest, setSelectedGuest, onBack, onNext }: {
  api: any;
  guestSearch: string;
  setGuestSearch: (s: string) => void;
  guestResults: GuestOption[];
  searchGuests: (q: string) => void;
  selectedGuest: GuestOption | null;
  setSelectedGuest: (g: GuestOption | null) => void;
  onBack: () => void;
  onNext: () => void;
}) {
  const [showCreateGuest, setShowCreateGuest] = useState(false);
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [newGuest, setNewGuest] = useState({ first_name: "", last_name: "", phone: "", email: "" });

  const handleCreateGuest = async () => {
    if (!newGuest.first_name || !newGuest.phone) {
      setCreateError("Name and phone are required");
      return;
    }
    setCreating(true);
    setCreateError(null);
    try {
      const result = await api.post("/api/v1/guests", {
        first_name: newGuest.first_name,
        last_name: newGuest.last_name || "",
        phone: newGuest.phone,
        email: newGuest.email || undefined,
      });
      setSelectedGuest({ id: result.id, full_name: `${newGuest.first_name} ${newGuest.last_name}`.trim(), phone: newGuest.phone });
      setShowCreateGuest(false);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "Failed to create guest");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-4">
      <p className="text-sm font-medium text-gray-700">Step 2: Select Guest</p>
      <input
        type="text"
        placeholder="Search guest by name or phone..."
        value={guestSearch}
        onChange={(e) => searchGuests(e.target.value)}
        className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
      />
      {guestResults.length > 0 && (
        <div className="space-y-2 max-h-48 overflow-y-auto">
          {guestResults.map((g) => (
            <label key={g.id} className={`flex items-center gap-3 p-3 border rounded-lg cursor-pointer ${selectedGuest?.id === g.id ? "border-blue-500 bg-blue-50" : "hover:bg-gray-50"}`}>
              <input type="radio" name="guest" checked={selectedGuest?.id === g.id} onChange={() => setSelectedGuest(g)} className="text-blue-600" />
              <div>
                <span className="font-medium text-sm">{g.full_name}</span>
                <span className="text-xs text-gray-500 ml-2">{g.phone}</span>
              </div>
            </label>
          ))}
        </div>
      )}

      {/* Create New Guest inline */}
      {!showCreateGuest ? (
        <button
          type="button"
          onClick={() => setShowCreateGuest(true)}
          className="w-full py-2 text-sm border-2 border-dashed border-blue-300 text-blue-600 rounded-lg hover:bg-blue-50 hover:border-blue-400"
        >
          + Create New Guest
        </button>
      ) : (
        <div className="border rounded-lg p-4 bg-blue-50 space-y-3">
          <h4 className="text-sm font-medium text-blue-800">New Guest</h4>
          {createError && <p className="text-xs text-red-600">{createError}</p>}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-gray-600">Name *</label>
              <input type="text" required placeholder="Full name" value={newGuest.first_name} onChange={(e) => setNewGuest({ ...newGuest, first_name: e.target.value })} className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="block text-xs text-gray-600">Last Name</label>
              <input type="text" placeholder="Optional" value={newGuest.last_name} onChange={(e) => setNewGuest({ ...newGuest, last_name: e.target.value })} className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="block text-xs text-gray-600">Phone *</label>
              <input type="tel" required placeholder="9876543210" value={newGuest.phone} onChange={(e) => setNewGuest({ ...newGuest, phone: e.target.value })} className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label className="block text-xs text-gray-600">Email</label>
              <input type="email" placeholder="Optional" value={newGuest.email} onChange={(e) => setNewGuest({ ...newGuest, email: e.target.value })} className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
          </div>
          <div className="flex gap-2 justify-end">
            <button type="button" onClick={() => setShowCreateGuest(false)} className="px-3 py-1.5 text-xs border rounded-md">Cancel</button>
            <button type="button" onClick={handleCreateGuest} disabled={creating} className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
              {creating ? "Creating..." : "Create & Select"}
            </button>
          </div>
        </div>
      )}

      {selectedGuest && (
        <p className="text-sm text-green-700 bg-green-50 p-2 rounded">
          Selected: <strong>{selectedGuest.full_name}</strong> ({selectedGuest.phone})
        </p>
      )}
      <div className="flex justify-between pt-2">
        <button onClick={onBack} className="px-4 py-2 text-sm border rounded-md hover:bg-gray-50">Back</button>
        <button onClick={onNext} disabled={!selectedGuest} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">Next</button>
      </div>
    </div>
  );
}
