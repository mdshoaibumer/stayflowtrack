"use client";

import React, { useState, useEffect, useCallback } from "react";
import { useApi } from "@/hooks/useApi";

interface LaundryOrder {
  id: string;
  order_number: string;
  order_type: string;
  status: string;
  guest_name: string;
  unit_number: string;
  total_items: number;
  grand_total: number;
  posted_to_folio: boolean;
  received_at: string;
}

interface LaundryTrackerProps {
  propertyId: string;
}

const statusFlow = ["received", "washing", "ready", "delivered"];
const statusColors: Record<string, string> = {
  received: "bg-gray-200",
  washing: "bg-blue-200",
  ready: "bg-green-200",
  delivered: "bg-purple-200",
};

export default function LaundryTracker({ propertyId }: LaundryTrackerProps) {
  const api = useApi();
  const [orders, setOrders] = useState<LaundryOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("");

  const fetchOrders = useCallback(async () => {
    const params: Record<string, string> = { property_id: propertyId, per_page: "50" };
    if (filter) params.status = filter;

    try {
      const data = await api.get<LaundryOrder[] | { data: LaundryOrder[] }>("/api/v1/laundry/orders", params);
      setOrders(Array.isArray(data) ? data : (data as { data: LaundryOrder[] })?.data || []);
    } catch {
      // Error handled by useApi (401 logout)
    }
    setLoading(false);
  }, [api, propertyId, filter]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const updateStatus = async (orderId: string, status: string) => {
    try {
      await api.post("/api/v1/laundry/orders/status", { order_id: orderId, status });
      fetchOrders();
    } catch {
      // Error handled by useApi
    }
  };

  const postToFolio = async (orderId: string) => {
    try {
      await api.post(`/api/v1/laundry/orders/${orderId}/post-to-folio`);
      fetchOrders();
    } catch {
      // Error handled by useApi
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-32"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Laundry Tracker</h2>
        <div className="flex gap-2">
          <button onClick={() => setFilter("")} className={`px-3 py-1 text-xs rounded ${!filter ? "bg-blue-600 text-white" : "border"}`}>All</button>
          {statusFlow.map((s) => (
            <button key={s} onClick={() => setFilter(s)} className={`px-3 py-1 text-xs rounded capitalize ${filter === s ? "bg-blue-600 text-white" : "border"}`}>
              {s}
            </button>
          ))}
        </div>
      </div>

      {/* Orders Table (mobile-friendly cards) */}
      <div className="space-y-3">
        {orders.map((order) => (
          <OrderCard key={order.id} order={order} onStatusChange={updateStatus} onPostToFolio={postToFolio} />
        ))}
        {orders.length === 0 && (
          <div className="text-center text-gray-500 py-8">No laundry orders found</div>
        )}
      </div>
    </div>
  );
}

function OrderCard({ order, onStatusChange, onPostToFolio }: {
  order: LaundryOrder;
  onStatusChange: (id: string, status: string) => void;
  onPostToFolio: (id: string) => void;
}) {
  const currentIdx = statusFlow.indexOf(order.status);
  const nextStatus = currentIdx < statusFlow.length - 1 ? statusFlow[currentIdx + 1] : null;

  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <div className="flex items-center justify-between mb-2">
        <div>
          <span className="font-mono text-sm font-bold">{order.order_number}</span>
          <span className={`ml-2 text-xs px-2 py-0.5 rounded-full capitalize ${statusColors[order.status]}`}>{order.status}</span>
        </div>
        <span className="text-xs text-gray-500">{new Date(order.received_at).toLocaleString()}</span>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm mb-3">
        {order.guest_name && <div><span className="text-gray-500">Guest:</span> {order.guest_name}</div>}
        {order.unit_number && <div><span className="text-gray-500">Unit:</span> {order.unit_number}</div>}
        <div><span className="text-gray-500">Items:</span> {order.total_items}</div>
        <div><span className="text-gray-500">Total:</span> ₹{(order.grand_total ?? 0).toFixed(2)}</div>
      </div>

      {/* Progress Steps */}
      <div className="flex items-center gap-1 mb-3">
        {statusFlow.map((s, i) => (
          <React.Fragment key={s}>
            <div className={`h-2 flex-1 rounded-full ${i <= currentIdx ? "bg-blue-500" : "bg-gray-200"}`}></div>
          </React.Fragment>
        ))}
      </div>

      <div className="flex gap-2 justify-end">
        {order.order_type === "guest" && !order.posted_to_folio && (
          <button
            onClick={() => onPostToFolio(order.id)}
            className="text-xs px-3 py-1 border border-orange-500 text-orange-600 rounded hover:bg-orange-50"
          >
            Add to Guest Bill
          </button>
        )}
        {nextStatus && (
          <button
            onClick={() => onStatusChange(order.id, nextStatus)}
            className="text-xs px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Mark as {nextStatus}
          </button>
        )}
      </div>
    </div>
  );
}
