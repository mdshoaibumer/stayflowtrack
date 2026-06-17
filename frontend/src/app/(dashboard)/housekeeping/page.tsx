"use client";

import React, { useState, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import HousekeepingBoard from "@/components/housekeeping/HousekeepingBoard";

export default function HousekeepingPage() {
  const { user } = useAuth();
  const api = useApi();
  const [showCreate, setShowCreate] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const propertyId = user?.property_id || "";

  const handleCreateTask = async (data: any) => {
    await api.post("/api/v1/housekeeping/tasks", { ...data, property_id: propertyId });
    setShowCreate(false);
    setRefreshKey((k) => k + 1);
  };

  if (!propertyId) {
    return <div className="text-muted-foreground text-sm py-8 text-center">No property configured. Please set up your property in Settings.</div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <h1 className="page-title">Housekeeping</h1>
          <p className="page-description">Manage room cleaning tasks and inspections</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center justify-center gap-2 h-9 px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 shadow-sm transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Task
        </button>
      </div>

      <HousekeepingBoard key={refreshKey} propertyId={propertyId} />

      {/* Create Task Modal */}
      {showCreate && (
        <CreateTaskModal onClose={() => setShowCreate(false)} onSubmit={handleCreateTask} />
      )}
    </div>
  );
}

function CreateTaskModal({ onClose, onSubmit }: { onClose: () => void; onSubmit: (data: any) => Promise<void> }) {
  const { user } = useAuth();
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [units, setUnits] = useState<any[]>([]);
  const [form, setForm] = useState({
    unit_id: "",
    unit_number: "",
    task_type: "checkout_clean",
    priority: "normal",
    assignee_name: "",
    notes: "",
    estimated_minutes: 30,
  });

  // Fetch units for the property to allow selection by ID
  useEffect(() => {
    if (!user?.property_id) return;
    api.get<any>(`/api/v1/properties/${user.property_id}/units`)
      .then((data: any) => {
        const list = Array.isArray(data) ? data : data?.data || [];
        setUnits(list);
      })
      .catch(() => {});
  }, [api, user?.property_id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.unit_id) { setError("Please select a unit"); return; }
    setLoading(true);
    setError(null);
    try {
      await onSubmit({ ...form });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create task");
    } finally {
      setLoading(false);
    }
  };

  const update = (field: string, value: string | number) => setForm((f) => ({ ...f, [field]: value }));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" role="dialog" aria-modal="true">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-[2px]" onClick={onClose} aria-hidden="true" />
      <div className="relative bg-card rounded-xl shadow-xl max-w-md w-full max-h-[85vh] overflow-hidden flex flex-col">
        <div className="px-6 py-4 border-b flex items-center justify-between shrink-0">
          <h2 className="text-lg font-semibold text-gray-900">New Housekeeping Task</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-100 transition-colors" aria-label="Close">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <div className="px-6 py-4 overflow-y-auto flex-1">
          {error && <div className="mb-4 p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700" role="alert">{error}</div>}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Unit *</label>
              <select required value={form.unit_id} onChange={(e) => {
                const selected = units.find((u: any) => u.id === e.target.value);
                setForm(f => ({ ...f, unit_id: e.target.value, unit_number: selected?.unit_number || "" }));
              }} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                <option value="">Select a unit...</option>
                {units.map((u: any) => (
                  <option key={u.id} value={u.id}>{u.unit_number} — {u.status || "available"}</option>
                ))}
              </select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Task Type</label>
                <select value={form.task_type} onChange={(e) => update("task_type", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                  <option value="checkout_clean">Checkout Clean</option>
                  <option value="stay_over_clean">Stay Over Clean</option>
                  <option value="deep_clean">Deep Clean</option>
                  <option value="inspection">Inspection</option>
                  <option value="maintenance">Maintenance</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Priority</label>
                <select value={form.priority} onChange={(e) => update("priority", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                  <option value="urgent">Urgent</option>
                  <option value="high">High</option>
                  <option value="normal">Normal</option>
                  <option value="low">Low</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Assign To</label>
              <input type="text" value={form.assignee_name} onChange={(e) => update("assignee_name", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" placeholder="Staff name" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Estimated Time (minutes)</label>
              <input type="number" min="5" step="5" value={form.estimated_minutes} onChange={(e) => update("estimated_minutes", Number(e.target.value))} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Notes</label>
              <textarea value={form.notes} onChange={(e) => update("notes", e.target.value)} rows={2} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors resize-none" placeholder="Special instructions..." />
            </div>
            <div className="flex gap-3 justify-end pt-2">
              <button type="button" onClick={onClose} className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium border border-gray-300 bg-white text-gray-700 rounded-lg hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors">Cancel</button>
              <button type="submit" disabled={loading} className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 shadow-sm transition-colors">
                {loading ? "Creating..." : "Create Task"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
