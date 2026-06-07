"use client";

import React, { useState, useEffect } from "react";
import { useApi } from "@/hooks/useApi";
import { useAuth } from "@/contexts/AuthContext";

interface WalkInFormProps {
  onSuccess: (result: any) => void;
  onCancel: () => void;
}

interface AvailableUnit {
  id: string;
  unit_number: string;
  unit_type_name: string;
  rate_per_night: number;
}

export default function WalkInForm({ onSuccess, onCancel }: WalkInFormProps) {
  const api = useApi();
  const { user } = useAuth();
  const propertyId = user?.property_id || "";

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [units, setUnits] = useState<AvailableUnit[]>([]);

  const [form, setForm] = useState({
    guest_first_name: "",
    guest_last_name: "",
    guest_phone: "",
    guest_email: "",
    unit_id: "",
    check_out_date: "",
    num_guests: 1,
    rate_per_night: 0,
    deposit_amount: 2000,
    deposit_method: "cash",
    deposit_reference: "",
    id_document_type: "aadhaar",
    id_document_number: "",
    notes: "",
  });

  // Load available units
  useEffect(() => {
    if (!propertyId) return;
    api.get<any>(`/api/v1/properties/${propertyId}/units/search`, { status: "available" })
      .then((data) => {
        const list = Array.isArray(data) ? data : data?.data || [];
        setUnits(list);
      })
      .catch(() => {});
  }, [api, propertyId]);

  // Auto-fill rate when unit is selected
  const handleUnitChange = (unitId: string) => {
    setForm({ ...form, unit_id: unitId });
    const unit = units.find((u) => u.id === unitId);
    if (unit && unit.rate_per_night) {
      setForm((prev) => ({ ...prev, unit_id: unitId, rate_per_night: unit.rate_per_night }));
    }
  };

  // Calculate total
  const nights = form.check_out_date
    ? Math.max(1, Math.ceil((new Date(form.check_out_date).getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : 0;
  const total = nights * form.rate_per_night;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.guest_first_name || !form.guest_phone || !form.unit_id || !form.check_out_date || !form.id_document_number) {
      setError("Please fill all required fields");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await api.post("/api/v1/operations/walk-in", {
        ...form,
        property_id: propertyId,
        rate_per_night: form.rate_per_night,
        deposit_amount: form.deposit_amount,
      });
      onSuccess(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Walk-in failed");
    } finally {
      setLoading(false);
    }
  };

  // Set default check-out to tomorrow
  useEffect(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    setForm((prev) => ({ ...prev, check_out_date: tomorrow.toISOString().split("T")[0] }));
  }, []);

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[95vh] overflow-y-auto p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Walk-In Guest</h2>
            <p className="text-sm text-gray-500 mt-0.5">One form — guest info, room, deposit, ID. Done.</p>
          </div>
          <button onClick={onCancel} className="text-gray-400 hover:text-gray-600">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        {error && <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">{error}</div>}

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Guest Info */}
          <fieldset className="border rounded-lg p-4">
            <legend className="text-sm font-medium text-gray-700 px-2">Guest Details</legend>
            <div className="grid grid-cols-2 gap-3 mt-2">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">First Name *</label>
                <input type="text" required value={form.guest_first_name} onChange={(e) => setForm({ ...form, guest_first_name: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Raj" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Last Name <span className="text-gray-400">(optional)</span></label>
                <input type="text" value={form.guest_last_name} onChange={(e) => setForm({ ...form, guest_last_name: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Sharma" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Phone *</label>
                <input type="tel" required value={form.guest_phone} onChange={(e) => setForm({ ...form, guest_phone: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="9876543210" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Email</label>
                <input type="email" value={form.guest_email} onChange={(e) => setForm({ ...form, guest_email: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="guest@email.com" />
              </div>
            </div>
          </fieldset>

          {/* Stay Details */}
          <fieldset className="border rounded-lg p-4">
            <legend className="text-sm font-medium text-gray-700 px-2">Stay Details</legend>
            <div className="grid grid-cols-2 gap-3 mt-2">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Unit / Room *</label>
                <select required value={form.unit_id} onChange={(e) => handleUnitChange(e.target.value)} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                  <option value="">Select available unit...</option>
                  {units.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.unit_number} {u.unit_type_name ? `(${u.unit_type_name})` : ""} {u.rate_per_night ? `- ₹${u.rate_per_night}/night` : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Check-Out Date *</label>
                <input type="date" required value={form.check_out_date} onChange={(e) => setForm({ ...form, check_out_date: e.target.value })} min={new Date().toISOString().split("T")[0]} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Rate per Night (₹) *</label>
                <input type="number" required min={1} value={form.rate_per_night || ""} onChange={(e) => setForm({ ...form, rate_per_night: Number(e.target.value) })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Number of Guests</label>
                <input type="number" min={1} max={20} value={form.num_guests} onChange={(e) => setForm({ ...form, num_guests: Number(e.target.value) })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
              </div>
            </div>
            {nights > 0 && form.rate_per_night > 0 && (
              <div className="mt-3 p-2 bg-blue-50 rounded text-sm">
                <span className="font-medium">{nights} night(s)</span> × ₹{form.rate_per_night.toLocaleString()} = <span className="font-bold">₹{total.toLocaleString()}</span>
              </div>
            )}
          </fieldset>

          {/* ID & Deposit */}
          <fieldset className="border rounded-lg p-4">
            <legend className="text-sm font-medium text-gray-700 px-2">ID Verification & Deposit</legend>
            <div className="grid grid-cols-2 gap-3 mt-2">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">ID Document Type *</label>
                <select required value={form.id_document_type} onChange={(e) => setForm({ ...form, id_document_type: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                  <option value="aadhaar">Aadhaar Card</option>
                  <option value="passport">Passport</option>
                  <option value="driving_license">Driving License</option>
                  <option value="voter_id">Voter ID</option>
                  <option value="pan_card">PAN Card</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">ID Number *</label>
                <input type="text" required value={form.id_document_number} onChange={(e) => setForm({ ...form, id_document_number: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="XXXX XXXX XXXX" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Deposit Amount (₹) <span className="text-gray-400">(0 for corporate)</span></label>
                <input type="number" required min={0} value={form.deposit_amount} onChange={(e) => setForm({ ...form, deposit_amount: Number(e.target.value) })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
                {form.rate_per_night > 0 && (
                  <p className="text-xs text-blue-600 mt-1">
                    Suggested: ₹{form.rate_per_night.toLocaleString()} (1 night).
                    <button type="button" onClick={() => setForm({ ...form, deposit_amount: form.rate_per_night })} className="ml-1 underline hover:no-underline">Use this</button>
                  </p>
                )}
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Payment Method *</label>
                <select required value={form.deposit_method} onChange={(e) => setForm({ ...form, deposit_method: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                  <option value="cash">Cash</option>
                  <option value="upi">UPI</option>
                  <option value="card">Card</option>
                  <option value="bank_transfer">Bank Transfer</option>
                  <option value="cheque">Cheque</option>
                </select>
              </div>
            </div>
            {form.deposit_method !== "cash" && (
              <div className="mt-3">
                <label className="block text-xs font-medium text-gray-600 mb-1">Transaction Reference</label>
                <input type="text" value={form.deposit_reference} onChange={(e) => setForm({ ...form, deposit_reference: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="UPI ref / Card last 4 digits" />
              </div>
            )}
          </fieldset>

          {/* Notes */}
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Notes (optional)</label>
            <input type="text" value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Any special requests..." />
          </div>

          {/* Submit */}
          <div className="flex gap-3 justify-end pt-2 border-t">
            <button type="button" onClick={onCancel} className="px-5 py-2.5 text-sm border rounded-md hover:bg-gray-50">Cancel</button>
            <button type="submit" disabled={loading} className="px-5 py-2.5 text-sm bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 font-medium">
              {loading ? "Processing..." : "Check In Guest"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
