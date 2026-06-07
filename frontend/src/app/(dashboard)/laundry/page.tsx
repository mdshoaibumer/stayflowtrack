"use client";

import React, { useState, useEffect, Suspense } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { useSearchParams } from "next/navigation";
import LaundryTracker from "@/components/laundry/LaundryTracker";

export default function LaundryPage() {
  return (
    <Suspense fallback={<div className="flex justify-center py-8"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>}>
      <LaundryContent />
    </Suspense>
  );
}

function LaundryContent() {
  const { user } = useAuth();
  const api = useApi();
  const searchParams = useSearchParams();
  const action = searchParams.get("action");

  const [showCreate, setShowCreate] = useState(action === "new");
  const [refreshKey, setRefreshKey] = useState(0);

  const propertyId = user?.property_id || "";

  const handleCreateOrder = async (data: any) => {
    await api.post("/api/v1/laundry/orders", { ...data, property_id: propertyId });
    setShowCreate(false);
    setRefreshKey((k) => k + 1);
  };

  if (!propertyId) {
    return <div className="text-gray-500 text-sm">No property configured</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Laundry</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Order
        </button>
      </div>

      <LaundryTracker key={refreshKey} propertyId={propertyId} />

      {showCreate && (
        <CreateLaundryOrderModal onClose={() => setShowCreate(false)} onSubmit={handleCreateOrder} />
      )}
    </div>
  );
}

interface LaundryItem {
  item_name: string;
  quantity: number;
  rate: number;
}

function CreateLaundryOrderModal({ onClose, onSubmit }: { onClose: () => void; onSubmit: (data: any) => Promise<void> }) {
  const api = useApi();
  const { user } = useAuth();
  const propertyId = user?.property_id || "";
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState({
    order_type: "guest" as "guest" | "house",
    guest_name: "",
    guest_id: "",
    unit_number: "",
    notes: "",
    service_type: "regular" as "regular" | "express",
  });
  const [items, setItems] = useState<LaundryItem[]>([
    { item_name: "", quantity: 1, rate: 0 },
  ]);

  // Rate card (saved prices)
  const [rateCards, setRateCards] = useState<any[]>([]);

  // Load rate cards
  useEffect(() => {
    if (!propertyId) return;
    api.get<any>(`/api/v1/laundry/rate-card/${propertyId}`)
      .then((data) => setRateCards(Array.isArray(data) ? data : []))
      .catch(() => {});
  }, [api, propertyId]);

  // Guest search
  const [guestSearch, setGuestSearch] = useState("");
  const [guestResults, setGuestResults] = useState<any[]>([]);

  const searchGuests = async (query: string) => {
    setGuestSearch(query);
    if (query.length < 2) { setGuestResults([]); return; }
    try {
      const data = await api.get<any>("/api/v1/guests/search", { search: query });
      setGuestResults(Array.isArray(data) ? data : data?.data || []);
    } catch { /* silent */ }
  };

  const addItem = () => setItems([...items, { item_name: "", quantity: 1, rate: 0 }]);
  const removeItem = (idx: number) => setItems(items.filter((_, i) => i !== idx));
  const updateItem = (idx: number, field: keyof LaundryItem, value: string | number) => {
    setItems(items.map((item, i) => i === idx ? { ...item, [field]: value } : item));
  };

  // Quick-add from rate card
  const addFromRateCard = (card: any) => {
    const newItem: LaundryItem = {
      item_name: card.item_name,
      quantity: 1,
      rate: Number(card.default_rate),
    };
    // If there's an empty first row, replace it
    if (items.length === 1 && !items[0].item_name) {
      setItems([newItem]);
    } else {
      setItems([...items, newItem]);
    }
  };

  const total = items.reduce((sum, item) => sum + item.quantity * item.rate, 0) * (form.service_type === "express" ? 2 : 1);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const validItems = items.filter((i) => i.item_name);
    if (validItems.length === 0) { setError("Add at least one item"); return; }
    if (form.order_type === "guest" && !form.unit_number) { setError("Unit number required for guest orders"); return; }

    setLoading(true);
    setError(null);
    try {
      await onSubmit({
        ...form,
        items: validItems.map((i) => ({
          ...i,
          rate: form.service_type === "express" ? i.rate * 2 : i.rate,
        })),
        service_type: form.service_type,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create order");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">New Laundry Order</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        {error && <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Order Type */}
          <div className="flex gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="radio" name="type" checked={form.order_type === "guest"} onChange={() => setForm({ ...form, order_type: "guest" })} className="text-blue-600" />
              <span className="text-sm">Guest Laundry</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="radio" name="type" checked={form.order_type === "house"} onChange={() => setForm({ ...form, order_type: "house" })} className="text-blue-600" />
              <span className="text-sm">House Laundry</span>
            </label>
          </div>

          {/* Service Type (Regular / Express) */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => setForm({ ...form, service_type: "regular" })}
              className={`flex-1 py-2 px-3 text-sm rounded-md border-2 font-medium transition-colors ${
                form.service_type === "regular" ? "border-blue-500 bg-blue-50 text-blue-700" : "border-gray-200 hover:border-gray-300"
              }`}
            >
              Regular
              <span className="block text-xs font-normal text-gray-500 mt-0.5">24-48 hours</span>
            </button>
            <button
              type="button"
              onClick={() => setForm({ ...form, service_type: "express" })}
              className={`flex-1 py-2 px-3 text-sm rounded-md border-2 font-medium transition-colors ${
                form.service_type === "express" ? "border-orange-500 bg-orange-50 text-orange-700" : "border-gray-200 hover:border-gray-300"
              }`}
            >
              Express (2× rate)
              <span className="block text-xs font-normal text-gray-500 mt-0.5">Same day</span>
            </button>
          </div>

          {/* Guest search */}
          {form.order_type === "guest" && (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-700">Guest</label>
                <input
                  type="text"
                  placeholder="Search guest..."
                  value={guestSearch || form.guest_name}
                  onChange={(e) => { searchGuests(e.target.value); setForm({ ...form, guest_name: e.target.value }); }}
                  className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                />
                {guestResults.length > 0 && (
                  <div className="mt-1 border rounded-md max-h-32 overflow-y-auto">
                    {guestResults.map((g) => (
                      <button
                        key={g.id}
                        type="button"
                        onClick={() => {
                          setForm({ ...form, guest_id: g.id, guest_name: g.full_name });
                          setGuestSearch("");
                          setGuestResults([]);
                        }}
                        className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50"
                      >
                        {g.full_name} <span className="text-gray-400">{g.phone}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Unit Number *</label>
                <input type="text" value={form.unit_number} onChange={(e) => setForm({ ...form, unit_number: e.target.value })} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="101" />
              </div>
            </>
          )}

          {/* Rate Card Quick-Pick */}
          {rateCards.length > 0 && (
            <div>
              <label className="text-sm font-medium text-gray-700 mb-2 block">Quick Add (Rate Card)</label>
              <div className="flex flex-wrap gap-2">
                {rateCards.map((card) => (
                  <button
                    key={card.id}
                    type="button"
                    onClick={() => addFromRateCard(card)}
                    className="px-3 py-1.5 text-xs border rounded-full hover:bg-blue-50 hover:border-blue-300 transition-colors"
                  >
                    {card.item_name} <span className="text-gray-400">₹{Number(card.default_rate)}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Items */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-gray-700">Items</label>
              <button type="button" onClick={addItem} className="text-xs text-blue-600 hover:underline">+ Add Item</button>
            </div>
            <div className="space-y-2">
              {items.map((item, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <input
                    type="text"
                    placeholder="Item name"
                    value={item.item_name}
                    onChange={(e) => updateItem(idx, "item_name", e.target.value)}
                    className="flex-1 rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                  />
                  <input
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) => updateItem(idx, "quantity", Number(e.target.value))}
                    className="w-16 rounded-md border border-gray-300 px-2 py-1.5 text-sm text-center focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                    placeholder="Qty"
                  />
                  <input
                    type="number"
                    min="0"
                    value={item.rate}
                    onChange={(e) => updateItem(idx, "rate", Number(e.target.value))}
                    className="w-20 rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                    placeholder="₹ Rate"
                  />
                  {items.length > 1 && (
                    <button type="button" onClick={() => removeItem(idx)} className="text-red-400 hover:text-red-600">
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  )}
                </div>
              ))}
            </div>
            {total > 0 && (
              <p className="text-sm font-medium mt-2">
                Total: ₹{total.toFixed(2)}
                {form.service_type === "express" && <span className="text-orange-600 text-xs ml-2">(Express: 2× rate applied)</span>}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Notes</label>
            <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} rows={2} className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500" placeholder="Special instructions..." />
          </div>

          <div className="flex gap-3 justify-end pt-2">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm border rounded-md hover:bg-gray-50">Cancel</button>
            <button type="submit" disabled={loading} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
              {loading ? "Creating..." : "Create Order"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
