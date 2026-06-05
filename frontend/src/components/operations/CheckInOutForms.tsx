"use client";

import React, { useState } from "react";

interface CheckInFormProps {
  reservationId: string;
  unitId: string;
  onSubmit: (data: CheckInData) => Promise<void>;
  onCancel: () => void;
}

interface CheckInData {
  reservation_id: string;
  assigned_unit_id: string;
  deposit_amount: number;
  deposit_method: string;
  deposit_reference: string;
  id_document_type: string;
  notes: string;
}

export function CheckInForm({ reservationId, unitId, onSubmit, onCancel }: CheckInFormProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<CheckInData>({
    reservation_id: reservationId,
    assigned_unit_id: unitId,
    deposit_amount: 0,
    deposit_method: "",
    deposit_reference: "",
    id_document_type: "",
    notes: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await onSubmit(form);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Check-in failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h3 className="text-lg font-semibold">Check-In</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Deposit Amount</label>
          <input
            type="number"
            min="0"
            step="0.01"
            value={form.deposit_amount}
            onChange={(e) => setForm({ ...form, deposit_amount: parseFloat(e.target.value) || 0 })}
            className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Payment Method</label>
          <select
            value={form.deposit_method}
            onChange={(e) => setForm({ ...form, deposit_method: e.target.value })}
            className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
          >
            <option value="">Select...</option>
            <option value="cash">Cash</option>
            <option value="upi">UPI</option>
            <option value="card">Card</option>
            <option value="bank_transfer">Bank Transfer</option>
          </select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">Reference Number</label>
        <input
          type="text"
          value={form.deposit_reference}
          onChange={(e) => setForm({ ...form, deposit_reference: e.target.value })}
          placeholder="Transaction ID / Receipt no."
          className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">ID Document Type</label>
        <select
          value={form.id_document_type}
          onChange={(e) => setForm({ ...form, id_document_type: e.target.value })}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
        >
          <option value="">Select...</option>
          <option value="aadhaar">Aadhaar</option>
          <option value="passport">Passport</option>
          <option value="driving_license">Driving License</option>
          <option value="voter_id">Voter ID</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">Notes</label>
        <textarea
          value={form.notes}
          onChange={(e) => setForm({ ...form, notes: e.target.value })}
          rows={2}
          className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
        />
      </div>

      <div className="flex gap-2 justify-end">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-sm border rounded hover:bg-gray-100"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
        >
          {loading ? "Processing..." : "Complete Check-In"}
        </button>
      </div>
    </form>
  );
}

interface CheckOutFormProps {
  reservationId: string;
  guestName: string;
  onSubmit: (data: { reservation_id: string; notes: string }) => Promise<void>;
  onCancel: () => void;
}

export function CheckOutForm({ reservationId, guestName, onSubmit, onCancel }: CheckOutFormProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notes, setNotes] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await onSubmit({ reservation_id: reservationId, notes });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Check-out failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h3 className="text-lg font-semibold">Check-Out</h3>
      <p className="text-sm text-gray-600">
        Checking out <strong>{guestName}</strong>
      </p>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
          {error}
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-gray-700">Notes</label>
        <textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          rows={3}
          placeholder="Any remarks about the check-out..."
          className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
        />
      </div>

      <div className="flex gap-2 justify-end">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-sm border rounded hover:bg-gray-100"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 text-sm bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
        >
          {loading ? "Processing..." : "Complete Check-Out"}
        </button>
      </div>
    </form>
  );
}
