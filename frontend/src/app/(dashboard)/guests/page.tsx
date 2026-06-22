"use client";

import React, { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { GuestCard, GuestDetailPanel } from "@/components/dashboard/GuestCard";
import { PageHeader } from "@/components/ui/page-header";
import { SearchInput } from "@/components/ui/search-input";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { EmptyState } from "@/components/ui/empty-state";
import { Pagination } from "@/components/ui/pagination";
import { Spinner } from "@/components/ui/spinner";
import { Modal, ModalHeader, ModalTitle, ModalBody, ModalFooter, ModalCloseButton } from "@/components/ui/modal";
import { Input } from "@/components/ui/input";

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
  useAuth(); // ensure authenticated
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
      const rawGuests = Array.isArray(data) ? data : (data as { data: Guest[] })?.data || [];
      // Normalize: API returns first_name/last_name, UI uses full_name
      const normalized = rawGuests.map((g: Guest & { first_name?: string; last_name?: string }) => ({
        ...g,
        full_name: g.full_name || [g.first_name, g.last_name].filter(Boolean).join(" ") || "Unknown",
      }));
      setGuests(normalized);
      if ((data as { meta?: { total_pages: number } })?.meta?.total_pages) setTotalPages((data as { meta: { total_pages: number } }).meta.total_pages);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load guests");
    } finally {
      setLoading(false);
    }
  }, [api, page]); // eslint-disable-line react-hooks/exhaustive-deps -- search triggers via debounced effect below

  useEffect(() => {
    const debounce = setTimeout(fetchGuests, search ? 300 : 0);
    return () => clearTimeout(debounce);
  }, [fetchGuests, search]);

  return (
    <div className="page-container">
      {/* Header */}
      <PageHeader
        title="Guests"
        description={guests.length > 0 ? `${guests.length} guests found` : "Manage your guest records"}
      >
        <Button onClick={() => setShowCreate(true)}>
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Guest
        </Button>
      </PageHeader>

      {/* Search */}
      <SearchInput
        placeholder="Search by name, email, or phone..."
        value={search}
        onChange={(e) => { setSearch(e.target.value); setPage(1); }}
        onClear={() => { setSearch(""); setPage(1); }}
        aria-label="Search guests"
      />

      {/* Error */}
      {error && (
        <Alert variant="error" action={<button onClick={fetchGuests} className="text-xs font-medium underline">Retry</button>}>
          {error}
        </Alert>
      )}

      {/* Loading */}
      {loading && (
        <div className="flex justify-center py-12">
          <Spinner />
        </div>
      )}

      {/* Guest List */}
      {!loading && guests.length === 0 && (
        <EmptyState
          icon={
            <svg className="w-6 h-6 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          }
          title="No guests found"
          description={search ? "Try a different search term" : "Add your first guest to get started"}
          action={!search ? <Button size="sm" onClick={() => setShowCreate(true)}>Add Guest</Button> : undefined}
        />
      )}

      {!loading && guests.length > 0 && (
        <div className="flex gap-6">
          {/* Guest list */}
          <div className="flex-1 min-w-0">
            <div className="space-y-2">
              {guests.map((guest, idx) => (
                <GuestCard
                  key={guest.id}
                  guest={guest}
                  index={idx}
                  selected={selectedGuest?.id === guest.id}
                  onSelect={setSelectedGuest}
                />
              ))}
            </div>

            {/* Pagination */}
            <Pagination
              page={page}
              totalPages={totalPages}
              onPageChange={setPage}
              className="mt-6"
            />
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
      // Backend expects first_name + last_name, split from full_name
      const nameParts = form.full_name.trim().split(/\s+/);
      const payload: Record<string, string> = {
        first_name: nameParts[0] || "",
        last_name: nameParts.slice(1).join(" ") || "",
        email: form.email,
        phone: form.phone,
        nationality: form.nationality,
      };
      if (form.id_type) {
        if (form.id_type === "aadhaar") payload.aadhaar_number = form.id_number;
        else if (form.id_type === "passport") payload.passport_number = form.id_number;
      }
      if (form.company_name) payload.company_name = form.company_name;
      if (form.gstin) payload.gstin = form.gstin;
      await api.post("/api/v1/guests", payload);
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create guest");
    } finally {
      setLoading(false);
    }
  };

  const update = (field: string, value: string) => setForm((f) => ({ ...f, [field]: value }));

  return (
    <Modal open={true} onClose={onClose} size="lg">
      <ModalHeader>
        <ModalTitle>Add Guest</ModalTitle>
        <ModalCloseButton onClick={onClose} />
      </ModalHeader>
      <ModalBody>
        {error && <Alert variant="error" className="mb-4">{error}</Alert>}

        <form id="create-guest-form" onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Full Name <span className="text-destructive">*</span></label>
            <Input type="text" required value={form.full_name} onChange={(e) => update("full_name", e.target.value)} placeholder="Guest name" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">Phone <span className="text-destructive">*</span></label>
              <Input type="tel" required value={form.phone} onChange={(e) => update("phone", e.target.value)} placeholder="+91 98765 43210" />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">Email</label>
              <Input type="email" value={form.email} onChange={(e) => update("email", e.target.value)} placeholder="guest@email.com" />
            </div>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">ID Type</label>
              <select value={form.id_type} onChange={(e) => update("id_type", e.target.value)} className="block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                <option value="">Select...</option>
                <option value="aadhaar">Aadhaar</option>
                <option value="passport">Passport</option>
                <option value="driving_license">Driving License</option>
                <option value="voter_id">Voter ID</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">ID Number</label>
              <Input type="text" value={form.id_number} onChange={(e) => update("id_number", e.target.value)} placeholder="Document number" />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Nationality</label>
            <Input type="text" value={form.nationality} onChange={(e) => update("nationality", e.target.value)} />
          </div>

          {/* B2B / GST Fields */}
          <div className="border-t pt-4">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">Corporate / GST (for B2B invoicing)</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">Company Name</label>
                <Input type="text" value={form.company_name} onChange={(e) => update("company_name", e.target.value)} placeholder="ABC Pvt Ltd" />
                </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">GSTIN</label>
                <Input type="text" value={form.gstin} onChange={(e) => update("gstin", e.target.value.toUpperCase())} maxLength={15} placeholder="22AAAAA0000A1Z5" />
                {form.gstin && form.gstin.length !== 15 && (
                  <p className="text-xs text-destructive mt-1">GSTIN must be 15 characters</p>
                )}
              </div>
            </div>
          </div>
        </form>
      </ModalBody>
      <ModalFooter>
        <Button variant="outline" onClick={onClose}>Cancel</Button>
        <Button type="submit" form="create-guest-form" loading={loading}>
          {loading ? "Saving..." : "Add Guest"}
        </Button>
      </ModalFooter>
    </Modal>
  );
}

function GuestDetailDrawer({ guest, onClose, onUpdated: _onUpdated }: { guest: Guest; onClose: () => void; onUpdated: () => void }) {
  const api = useApi();
  interface StayRecord { unit_number?: string; status?: string; check_in_date?: string; check_out_date?: string; }
  const [history, setHistory] = useState<StayRecord[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(true);

  useEffect(() => {
    api.get<StayRecord[] | { data: StayRecord[] }>(`/api/v1/guests/${guest.id}/history`).then((data) => {
      setHistory(Array.isArray(data) ? data : (data as { data: StayRecord[] })?.data || []);
    }).catch(() => {}).finally(() => setLoadingHistory(false));
  }, [api, guest.id]);

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex justify-end" onClick={onClose} role="dialog" aria-modal="true" aria-label="Guest details">
      <div
        className="bg-card w-full max-w-md h-full overflow-y-auto shadow-xl animate-slide-in-right"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-foreground">Guest Details</h2>
            <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors rounded-lg p-1 hover:bg-muted" aria-label="Close">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          {/* Guest Info */}
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center text-lg font-semibold text-primary">
                {guest.full_name.charAt(0).toUpperCase()}
              </div>
              <div>
                <p className="font-medium text-foreground">{guest.full_name}</p>
                <p className="text-sm text-muted-foreground">{guest.total_stays || 0} stays</p>
              </div>
            </div>

            <div className="border-t pt-4 space-y-3">
              <InfoRow label="Phone" value={guest.phone} />
              <InfoRow label="Email" value={guest.email} />
              <InfoRow label="ID Type" value={guest.id_type?.replaceAll("_", " ")} />
              <InfoRow label="ID Number" value={guest.id_number} />
              <InfoRow label="Nationality" value={guest.nationality} />
              <InfoRow label="Added" value={guest.created_at ? new Date(guest.created_at).toLocaleDateString() : ""} />
            </div>

            {/* Stay History */}
            <div className="border-t pt-4">
              <h3 className="font-medium text-sm text-foreground mb-3">Stay History</h3>
              {loadingHistory && <div className="flex justify-center py-4"><Spinner size="sm" /></div>}
              {!loadingHistory && history.length === 0 && (
                <p className="text-sm text-muted-foreground">No previous stays</p>
              )}
              {!loadingHistory && history.map((stay, i) => (
                <div key={i} className="border rounded-lg p-3 mb-2 text-sm">
                  <div className="flex justify-between">
                    <span className="font-medium text-foreground">{stay.unit_number || "Unit"}</span>
                    <span className="text-xs text-muted-foreground">{stay.status}</span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
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
      <span className="text-muted-foreground">{label}</span>
      <span className="text-foreground capitalize">{value || "—"}</span>
    </div>
  );
}
