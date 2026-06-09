"use client";

import React, { useState, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { GuestCard, GuestDetailPanel } from "@/components/dashboard/GuestCard";

interface Guest {
  id: string;
  full_name: string;
  email: string;
  phone: string;
  id_type: string;
  id_number: string;
  nationality: string;
  total_stays: number;
  last_stay_date: string;
  created_at: string;
}

export default function GuestsPage() {
  const { user } = useAuth();
  const api = useApi();
  const [guests, setGuests] = useState<Guest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [selectedGuest, setSelectedGuest] = useState<Guest | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const fetchGuests = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = { page: String(page), per_page: "20" };
      if (search) params.search = search;

      const data = await api.get<{ data: Guest[]; meta?: { total_pages: number } }>(
        search ? "/api/v1/guests/search" : "/api/v1/guests",
        params
      );
      const rawGuests = Array.isArray(data) ? data : (data as any)?.data || (data as any) || [];
      // Normalize: API returns first_name/last_name, UI uses full_name
      const normalized = rawGuests.map((g: any) => ({
        ...g,
        full_name: g.full_name || [g.first_name, g.last_name].filter(Boolean).join(" ") || "Unknown",
      }));
      setGuests(normalized);
      if ((data as any)?.meta?.total_pages) setTotalPages((data as any).meta.total_pages);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load guests");
    } finally {
      setLoading(false);
    }
  }, [api, page, search]);

  useEffect(() => {
    const debounce = setTimeout(fetchGuests, search ? 300 : 0);
    return () => clearTimeout(debounce);
  }, [fetchGuests]);

  return (
    <div className="space-y-5 max-w-[1600px] mx-auto">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
      >
        <div>
          <h1 className="text-xl lg:text-2xl font-bold text-gray-900 tracking-tight">Guests</h1>
          <p className="text-sm text-gray-400">{guests.length > 0 ? `${guests.length} guests found` : "Manage your guest records"}</p>
        </div>
        <motion.button
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-navy-900 text-white text-sm font-medium rounded-xl hover:bg-navy-800 shadow-sm transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Guest
        </motion.button>
      </motion.div>

      {/* Search */}
      <div className="relative">
        <svg className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          type="text"
          placeholder="Search by name, email, or phone..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          className="w-full pl-11 pr-4 py-3 border border-gray-200 rounded-xl text-sm bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500/20 focus:border-teal-400 transition-all duration-200"
        />
      </div>

      {/* Error */}
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
          {error}
          <button onClick={fetchGuests} className="ml-2 underline">Retry</button>
        </div>
      )}

      {/* Loading */}
      {loading && (
        <div className="flex justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      )}

      {/* Guest List */}
      {!loading && guests.length === 0 && (
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center py-16"
        >
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-50 flex items-center justify-center">
            <svg className="w-8 h-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </div>
          <p className="text-base font-semibold text-gray-900">No guests found</p>
          <p className="text-sm text-gray-500 mt-1">
            {search ? "Try a different search term" : "Add your first guest to get started"}
          </p>
        </motion.div>
      )}

      {!loading && guests.length > 0 && (
        <div className="flex gap-6">
          {/* Guest list */}
          <div className="flex-1 min-w-0">
            <div className="space-y-2">
              <AnimatePresence mode="popLayout">
                {guests.map((guest, idx) => (
                  <GuestCard
                    key={guest.id}
                    guest={guest}
                    index={idx}
                    selected={selectedGuest?.id === guest.id}
                    onSelect={setSelectedGuest}
                  />
                ))}
              </AnimatePresence>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-3 mt-6">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 text-sm font-medium border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Previous
                </button>
                <span className="text-sm text-gray-500 font-medium">Page {page} of {totalPages}</span>
                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 text-sm font-medium border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </div>

          {/* Detail panel (desktop) */}
          <div className="hidden lg:block w-80 flex-shrink-0">
            <GuestDetailPanel
              guest={selectedGuest}
              onClose={() => setSelectedGuest(null)}
            />
          </div>
        </div>
      )}

      {/* Create Guest Modal */}
      {showCreate && (
        <CreateGuestModal
          onClose={() => setShowCreate(false)}
          onCreated={() => { setShowCreate(false); fetchGuests(); }}
        />
      )}

      {/* Guest Detail Drawer */}
      {selectedGuest && (
        <GuestDetailDrawer
          guest={selectedGuest}
          onClose={() => setSelectedGuest(null)}
          onUpdated={fetchGuests}
        />
      )}
    </div>
  );
}

function CreateGuestModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState({
    full_name: "",
    email: "",
    phone: "",
    id_type: "",
    id_number: "",
    nationality: "Indian",
    company_name: "",
    gstin: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.full_name || !form.phone) {
      setError("Name and phone are required");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await api.post("/api/v1/guests", form);
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create guest");
    } finally {
      setLoading(false);
    }
  };

  const update = (field: string, value: string) => setForm((f) => ({ ...f, [field]: value }));

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">Add Guest</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Full Name *</label>
              <input type="text" required value={form.full_name} onChange={(e) => update("full_name", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Guest name" />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Phone *</label>
                <input type="tel" required value={form.phone} onChange={(e) => update("phone", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="+91 98765 43210" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Email</label>
                <input type="email" value={form.email} onChange={(e) => update("email", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="guest@email.com" />
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">ID Type</label>
                <select value={form.id_type} onChange={(e) => update("id_type", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500">
                  <option value="">Select...</option>
                  <option value="aadhaar">Aadhaar</option>
                  <option value="passport">Passport</option>
                  <option value="driving_license">Driving License</option>
                  <option value="voter_id">Voter ID</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">ID Number</label>
                <input type="text" value={form.id_number} onChange={(e) => update("id_number", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Document number" />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Nationality</label>
              <input type="text" value={form.nationality} onChange={(e) => update("nationality", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
            </div>

            {/* B2B / GST Fields */}
            <div className="border-t pt-4">
              <p className="text-xs font-medium text-gray-500 uppercase mb-3">Corporate / GST (for B2B invoicing)</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Company Name</label>
                  <input type="text" value={form.company_name} onChange={(e) => update("company_name", e.target.value)} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="ABC Pvt Ltd" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">GSTIN</label>
                  <input type="text" value={form.gstin} onChange={(e) => update("gstin", e.target.value.toUpperCase())} maxLength={15} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="22AAAAA0000A1Z5" />
                  {form.gstin && form.gstin.length !== 15 && (
                    <p className="text-xs text-red-500 mt-1">GSTIN must be 15 characters</p>
                  )}
                </div>
              </div>
            </div>
            <div className="flex gap-3 justify-end pt-2">
              <button type="button" onClick={onClose} className="px-4 py-2 text-sm border rounded-md hover:bg-gray-50">Cancel</button>
              <button type="submit" disabled={loading} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
                {loading ? "Saving..." : "Add Guest"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

function GuestDetailDrawer({ guest, onClose, onUpdated }: { guest: Guest; onClose: () => void; onUpdated: () => void }) {
  const api = useApi();
  const [history, setHistory] = useState<any[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(true);

  useEffect(() => {
    api.get(`/api/v1/guests/${guest.id}/history`).then((data: any) => {
      setHistory(Array.isArray(data) ? data : data?.data || []);
    }).catch(() => {}).finally(() => setLoadingHistory(false));
  }, [api, guest.id]);

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex justify-end" onClick={onClose}>
      <div
        className="bg-white w-full max-w-md h-full overflow-y-auto shadow-xl animate-in slide-in-from-right"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold">Guest Details</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          {/* Guest Info */}
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center text-lg font-semibold text-blue-700">
                {guest.full_name.charAt(0).toUpperCase()}
              </div>
              <div>
                <p className="font-medium text-gray-900">{guest.full_name}</p>
                <p className="text-sm text-gray-500">{guest.total_stays || 0} stays</p>
              </div>
            </div>

            <div className="border-t pt-4 space-y-3">
              <InfoRow label="Phone" value={guest.phone} />
              <InfoRow label="Email" value={guest.email} />
              <InfoRow label="ID Type" value={guest.id_type?.replace("_", " ")} />
              <InfoRow label="ID Number" value={guest.id_number} />
              <InfoRow label="Nationality" value={guest.nationality} />
              <InfoRow label="Added" value={guest.created_at ? new Date(guest.created_at).toLocaleDateString() : ""} />
            </div>

            {/* Stay History */}
            <div className="border-t pt-4">
              <h3 className="font-medium text-sm text-gray-700 mb-3">Stay History</h3>
              {loadingHistory && <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600 mx-auto"></div>}
              {!loadingHistory && history.length === 0 && (
                <p className="text-sm text-gray-400">No previous stays</p>
              )}
              {!loadingHistory && history.map((stay: any, i: number) => (
                <div key={i} className="border rounded p-3 mb-2 text-sm">
                  <div className="flex justify-between">
                    <span className="font-medium">{stay.unit_number || "Unit"}</span>
                    <span className="text-xs text-gray-500">{stay.status}</span>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    {stay.check_in_date} → {stay.check_out_date}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="flex justify-between text-sm">
      <span className="text-gray-500">{label}</span>
      <span className="text-gray-900 capitalize">{value || "—"}</span>
    </div>
  );
}
