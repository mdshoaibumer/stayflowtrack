"use client";

import React, { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";

interface Invoice {
  id: string;
  invoice_number: string;
  guest_name: string;
  unit_number: string;
  total_amount: number;
  paid_amount: number;
  balance: number;
  status: string;
  created_at: string;
  reservation_id: string;
}

interface FolioItem {
  id: string;
  description: string;
  amount: number;
  item_type: string;
  date: string;
  voided: boolean;
}

interface FolioPayment {
  id: string;
  amount: number;
  method: string;
  reference: string;
  date: string;
}

interface FolioSummary {
  folio_id: string;
  total_charges: number;
  total_payments: number;
  balance: number;
  guest_name: string;
  unit_number: string;
  reservation_id: string;
  status: string;
}

interface ActiveFolio {
  id: string;
  guest_name: string;
  unit_number: string;
  balance: number;
  status: string;
}

const statusColors: Record<string, string> = {
  open: "bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20",
  settled: "bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/20",
  partial: "bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-600/20",
  overdue: "bg-red-50 text-red-700 ring-1 ring-inset ring-red-600/20",
};

export default function BillingPage() {
  const { user } = useAuth();
  const api = useApi();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [activeFolios, setActiveFolios] = useState<ActiveFolio[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedFolio, setSelectedFolio] = useState<string | null>(null);
  const [tab, setTab] = useState<"active" | "invoices">("active");
  const [showQuickCharge, setShowQuickCharge] = useState(false);

  const propertyId = user?.property_id || "";

  // Check for ?action=charge query param
  useEffect(() => {
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);
      if (params.get("action") === "charge") {
        setShowQuickCharge(true);
      }
    }
  }, []);

  const fetchData = useCallback(async () => {
    if (!propertyId) return;
    setLoading(true);
    try {
      const data = await api.get<any>("/api/v1/billing/invoices", { property_id: propertyId, per_page: "50" });
      setInvoices(Array.isArray(data) ? data : data?.data || []);
      // Fetch active (open) folios
      const folioData = await api.get<any>("/api/v1/billing/invoices", { property_id: propertyId, status: "open", per_page: "50" });
      setActiveFolios(Array.isArray(folioData) ? folioData : folioData?.data || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load billing data");
    } finally {
      setLoading(false);
    }
  }, [api, propertyId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <div className="space-y-5 max-w-[1600px] mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <h1 className="text-xl lg:text-2xl font-bold text-gray-900 tracking-tight">Billing</h1>
        <button
          onClick={() => setShowQuickCharge(true)}
          className="inline-flex items-center justify-center gap-2 px-4 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 shadow-sm transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>
          Quick Charge
        </button>
      </div>

      {/* Tabs: Active Folios vs Invoices */}
      <div className="flex gap-1 border-b">
        <button onClick={() => setTab("active")} className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === "active" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300"}`}>
          Current Bills
          {activeFolios.length > 0 && <span className="ml-1.5 px-1.5 py-0.5 text-xs rounded-full bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20">{activeFolios.length}</span>}
        </button>
        <button onClick={() => setTab("invoices")} className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === "invoices" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300"}`}>
          Past Bills
        </button>
      </div>

      {error && (
        <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700" role="alert">
          {error}
          <button onClick={fetchData} className="ml-2 underline">Retry</button>
        </div>
      )}

      {loading && (
        <div className="flex justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      )}

      {/* Active Folios Tab */}
      {!loading && tab === "active" && (
        <div className="space-y-3">
          {activeFolios.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <p className="font-medium">No active folios</p>
              <p className="text-sm mt-1">Active folios appear when guests are checked in</p>
            </div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {activeFolios.map((folio) => (
                <div key={folio.id} onClick={() => setSelectedFolio(folio.id)} className="border rounded-lg p-4 bg-white shadow-sm cursor-pointer hover:shadow-md transition-shadow">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium">{folio.guest_name}</span>
                    <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[folio.status] || "bg-gray-100"}`}>{folio.status}</span>
                  </div>
                  <p className="text-sm text-gray-500">{folio.unit_number}</p>
                  <div className="mt-2 text-right">
                    <span className={`text-lg font-bold ${folio.balance > 0 ? "text-red-600" : "text-green-600"}`}>₹{folio.balance?.toLocaleString()}</span>
                    <p className="text-xs text-gray-400">balance due</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Invoices Tab */}
      {!loading && tab === "invoices" && invoices.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <svg className="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2z" />
          </svg>
          <p className="font-medium">No invoices yet</p>
          <p className="text-sm mt-1">Invoices will appear here once guests are billed</p>
        </div>
      )}

      {!loading && tab === "invoices" && invoices.length > 0 && (
        <>
          {/* Mobile cards */}
          <div className="space-y-3 lg:hidden">
            {invoices.map((inv) => (
              <div key={inv.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex items-center justify-between mb-2" onClick={() => setSelectedFolio(inv.id)}>
                  <span className="font-mono text-xs text-gray-500">{inv.invoice_number}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[inv.status] || "bg-gray-100"}`}>
                    {inv.status}
                  </span>
                </div>
                <p className="font-medium" onClick={() => setSelectedFolio(inv.id)}>{inv.guest_name}</p>
                <p className="text-sm text-gray-500">{inv.unit_number}</p>
                <div className="flex justify-between mt-2 text-sm">
                  <span className="text-gray-500">Total: ₹{inv.total_amount?.toLocaleString()}</span>
                  <span className={inv.balance > 0 ? "text-red-600 font-medium" : "text-green-600"}>
                    Balance: ₹{inv.balance?.toLocaleString()}
                  </span>
                </div>
                <div className="flex gap-2 mt-2 pt-2 border-t">
                  <button
                    onClick={(e) => { e.stopPropagation(); setSelectedFolio(inv.id); }}
                    className="flex-1 px-2 py-1.5 text-xs border rounded-md hover:bg-gray-50 text-center"
                  >View Details</button>
                  <InvoicePDFButton invoiceId={inv.id} />
                </div>
              </div>
            ))}
          </div>

          {/* Desktop table */}
          <div className="hidden lg:block border rounded-lg overflow-hidden bg-white">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Invoice #</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Guest</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Unit</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-600">Total</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-600">Paid</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-600">Balance</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Status</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-600">Date</th>
                  <th className="text-center px-4 py-3 font-medium text-gray-600">Invoice</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {invoices.map((inv) => (
                  <tr key={inv.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs cursor-pointer" onClick={() => setSelectedFolio(inv.id)}>{inv.invoice_number}</td>
                    <td className="px-4 py-3 font-medium cursor-pointer" onClick={() => setSelectedFolio(inv.id)}>{inv.guest_name}</td>
                    <td className="px-4 py-3 cursor-pointer" onClick={() => setSelectedFolio(inv.id)}>{inv.unit_number}</td>
                    <td className="px-4 py-3 text-right">₹{inv.total_amount?.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-green-600">₹{inv.paid_amount?.toLocaleString()}</td>
                    <td className={`px-4 py-3 text-right font-medium ${inv.balance > 0 ? "text-red-600" : "text-green-600"}`}>
                      ₹{inv.balance?.toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[inv.status] || "bg-gray-100"}`}>{inv.status}</span>
                    </td>
                    <td className="px-4 py-3 text-gray-500">{inv.created_at ? new Date(inv.created_at).toLocaleDateString() : "—"}</td>
                    <td className="px-4 py-3 text-center">
                      <InvoicePDFButton invoiceId={inv.id} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Folio Detail Drawer */}
      {selectedFolio && (
        <FolioDetailDrawer folioId={selectedFolio} onClose={() => setSelectedFolio(null)} onUpdated={fetchData} />
      )}

      {/* Quick Charge Modal */}
      {showQuickCharge && (
        <QuickChargeModal
          propertyId={propertyId}
          onClose={() => setShowQuickCharge(false)}
          onSuccess={() => { setShowQuickCharge(false); fetchData(); }}
        />
      )}
    </div>
  );
}

function FolioDetailDrawer({ folioId, onClose, onUpdated }: { folioId: string; onClose: () => void; onUpdated: () => void }) {
  const api = useApi();
  const [summary, setSummary] = useState<FolioSummary | null>(null);
  const [items, setItems] = useState<FolioItem[]>([]);
  const [payments, setPayments] = useState<FolioPayment[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddCharge, setShowAddCharge] = useState(false);
  const [showAddPayment, setShowAddPayment] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const [sumData, itemsData, payData] = await Promise.all([
          api.get<any>(`/api/v1/billing/folios/${folioId}/summary`),
          api.get<any>(`/api/v1/billing/folios/${folioId}/items`),
          api.get<any>(`/api/v1/billing/folios/${folioId}/payments`),
        ]);
        setSummary(sumData);
        setItems(Array.isArray(itemsData) ? itemsData : itemsData?.data || []);
        setPayments(Array.isArray(payData) ? payData : payData?.data || []);
      } catch { /* silent */ }
      finally { setLoading(false); }
    };
    load();
  }, [api, folioId]);

  const addCharge = async (description: string, amount: number) => {
    setActionLoading(true);
    try {
      await api.post("/api/v1/billing/charges", { folio_id: folioId, description, amount });
      setShowAddCharge(false);
      setSuccessMsg("Charge added");
      setTimeout(() => setSuccessMsg(null), 2000);
      // Refresh
      const [sumData, itemsData] = await Promise.all([
        api.get<any>(`/api/v1/billing/folios/${folioId}/summary`),
        api.get<any>(`/api/v1/billing/folios/${folioId}/items`),
      ]);
      setSummary(sumData);
      setItems(Array.isArray(itemsData) ? itemsData : itemsData?.data || []);
      onUpdated();
    } catch { /* silent */ }
    finally { setActionLoading(false); }
  };

  const addPayment = async (amount: number, method: string, reference: string) => {
    setActionLoading(true);
    try {
      await api.post("/api/v1/billing/payments", { folio_id: folioId, amount, method, reference });
      setShowAddPayment(false);
      setSuccessMsg("Payment recorded");
      setTimeout(() => setSuccessMsg(null), 2000);
      const [sumData, payData] = await Promise.all([
        api.get<any>(`/api/v1/billing/folios/${folioId}/summary`),
        api.get<any>(`/api/v1/billing/folios/${folioId}/payments`),
      ]);
      setSummary(sumData);
      setPayments(Array.isArray(payData) ? payData : payData?.data || []);
      onUpdated();
    } catch { /* silent */ }
    finally { setActionLoading(false); }
  };

  const generatePdf = async () => {
    setActionLoading(true);
    try {
      const result = await api.post<any>(`/api/v1/billing/invoices/${folioId}/pdf`);
      // Try to get the PDF URL and open it
      const pdfUrl = result?.url || result?.pdf_url || result?.download_url;
      if (pdfUrl) {
        window.open(pdfUrl, "_blank");
        setSuccessMsg("PDF generated — opening in new tab");
      } else {
        // Fallback: fetch the PDF URL
        const pdfData = await api.get<any>(`/api/v1/billing/invoices/${folioId}/pdf`);
        const url = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
        if (url) {
          window.open(url, "_blank");
          setSuccessMsg("PDF generated — opening in new tab");
        } else {
          setSuccessMsg("PDF generated");
        }
      }
      setTimeout(() => setSuccessMsg(null), 3000);
    } catch { setSuccessMsg("PDF generation failed"); setTimeout(() => setSuccessMsg(null), 3000); }
    finally { setActionLoading(false); }
  };

  const downloadPdf = async () => {
    setActionLoading(true);
    try {
      const pdfData = await api.get<any>(`/api/v1/billing/invoices/${folioId}/pdf`);
      const url = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
      if (url) {
        const a = document.createElement("a");
        a.href = url;
        a.target = "_blank";
        a.download = `invoice_${summary?.guest_name || "guest"}.pdf`;
        a.click();
      }
    } catch { /* silent */ }
    finally { setActionLoading(false); }
  };

  const printInvoice = async () => {
    setActionLoading(true);
    try {
      // Generate if needed, then get URL to print
      await api.post<any>(`/api/v1/billing/invoices/${folioId}/pdf`).catch(() => {});
      const pdfData = await api.get<any>(`/api/v1/billing/invoices/${folioId}/pdf`);
      const url = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
      if (url) {
        const printWindow = window.open(url, "_blank");
        if (printWindow) {
          printWindow.onload = () => printWindow.print();
        }
      } else {
        // Fallback: print current page content
        window.print();
      }
    } catch { window.print(); }
    finally { setActionLoading(false); }
  };

  const shareWhatsApp = async () => {
    setActionLoading(true);
    try {
      // Ensure PDF is generated first
      await api.post<any>(`/api/v1/billing/invoices/${folioId}/pdf`).catch(() => {});
      const pdfData = await api.get<any>(`/api/v1/billing/invoices/${folioId}/pdf`);
      const url = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
      const msg = encodeURIComponent(
        `Invoice for ${summary?.guest_name || "Guest"}\nTotal: ₹${summary?.total_charges?.toLocaleString() || 0}\nPaid: ₹${summary?.total_payments?.toLocaleString() || 0}\nBalance: ₹${summary?.balance?.toLocaleString() || 0}${url ? `\n\nDownload Invoice: ${url}` : ""}\n\nThank you for staying with us!`
      );
      window.open(`https://wa.me/?text=${msg}`, "_blank");
    } catch {
      const msg = encodeURIComponent(`Invoice ${summary?.guest_name || ""}\nTotal: ₹${summary?.total_charges?.toLocaleString() || 0}\nBalance: ₹${summary?.balance?.toLocaleString() || 0}`);
      window.open(`https://wa.me/?text=${msg}`, "_blank");
    } finally { setActionLoading(false); }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex justify-end" onClick={onClose}>
      <div className="bg-white w-full max-w-lg h-full overflow-y-auto shadow-xl" onClick={(e) => e.stopPropagation()}>
        <div className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Folio</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          {loading && <div className="flex justify-center py-8"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>}

          {successMsg && (
            <div className="p-2 bg-green-50 border border-green-200 rounded text-sm text-green-700">{successMsg}</div>
          )}

          {!loading && summary && (
            <>
              {/* Summary */}
              <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                <div className="flex justify-between text-sm"><span className="text-gray-500">Guest</span><span className="font-medium">{summary.guest_name}</span></div>
                <div className="flex justify-between text-sm"><span className="text-gray-500">Unit</span><span>{summary.unit_number}</span></div>
                <div className="border-t pt-2 mt-2 space-y-1">
                  <div className="flex justify-between text-sm"><span className="text-gray-500">Total Charges</span><span>₹{summary.total_charges?.toLocaleString()}</span></div>
                  <div className="flex justify-between text-sm"><span className="text-gray-500">Total Payments</span><span className="text-green-600">₹{summary.total_payments?.toLocaleString()}</span></div>
                  <div className="flex justify-between font-medium text-sm border-t pt-1">
                    <span>Balance Due</span>
                    <span className={summary.balance > 0 ? "text-red-600" : "text-green-600"}>₹{summary.balance?.toLocaleString()}</span>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-2 flex-wrap">
                <button onClick={() => setShowAddCharge(true)} className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded-md hover:bg-blue-700">Add Charge</button>
                <button onClick={() => setShowAddPayment(true)} className="px-3 py-1.5 text-xs bg-green-600 text-white rounded-md hover:bg-green-700">Record Payment</button>
                <button onClick={generatePdf} disabled={actionLoading} className="px-3 py-1.5 text-xs border rounded-md hover:bg-gray-50">Generate & View PDF</button>
                <button onClick={downloadPdf} disabled={actionLoading} className="px-3 py-1.5 text-xs border rounded-md hover:bg-gray-50">Download PDF</button>
                <button onClick={printInvoice} disabled={actionLoading} className="px-3 py-1.5 text-xs border rounded-md hover:bg-gray-50">
                  🖨️ Print
                </button>
                <button
                  onClick={shareWhatsApp}
                  disabled={actionLoading}
                  className="px-3 py-1.5 text-xs border border-green-500 text-green-600 rounded-md hover:bg-green-50"
                >
                  Share via WhatsApp
                </button>
              </div>

              {/* Charges */}
              <div>
                <h3 className="text-sm font-medium text-gray-700 mb-2">Charges</h3>
                {items.length === 0 && <p className="text-sm text-gray-400">No charges</p>}
                <div className="space-y-1">
                  {items.map((item) => (
                    <div key={item.id} className={`flex justify-between text-sm py-1 ${item.voided ? "line-through text-gray-400" : ""}`}>
                      <div>
                        <span>{item.description}</span>
                        <span className="text-xs text-gray-400 ml-2">{item.date}</span>
                      </div>
                      <span className="font-medium">₹{item.amount?.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Payments */}
              <div>
                <h3 className="text-sm font-medium text-gray-700 mb-2">Payments</h3>
                {payments.length === 0 && <p className="text-sm text-gray-400">No payments</p>}
                <div className="space-y-1">
                  {payments.map((p) => (
                    <div key={p.id} className="flex justify-between text-sm py-1">
                      <div>
                        <span className="capitalize">{p.method}</span>
                        {p.reference && <span className="text-xs text-gray-400 ml-2">({p.reference})</span>}
                      </div>
                      <span className="font-medium text-green-600">₹{p.amount?.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}

          {/* Add Charge Form */}
          {showAddCharge && (
            <AddChargeForm onSubmit={addCharge} onCancel={() => setShowAddCharge(false)} loading={actionLoading} />
          )}

          {/* Add Payment Form */}
          {showAddPayment && (
            <AddPaymentForm onSubmit={addPayment} onCancel={() => setShowAddPayment(false)} loading={actionLoading} balance={summary?.balance || 0} />
          )}
        </div>
      </div>
    </div>
  );
}

function AddChargeForm({ onSubmit, onCancel, loading }: { onSubmit: (desc: string, amount: number) => void; onCancel: () => void; loading: boolean }) {
  const [desc, setDesc] = useState("");
  const [amount, setAmount] = useState(0);
  const [category, setCategory] = useState("");

  // Only show categories relevant to service apartments (no spa/restaurant/minibar)
  const categories = [
    { value: "room_charge", label: "Room Charge" },
    { value: "laundry", label: "Laundry" },
    { value: "parking", label: "Parking" },
    { value: "extra_bed", label: "Extra Bed" },
    { value: "late_checkout", label: "Late Checkout" },
    { value: "damage", label: "Damage" },
    { value: "other", label: "Other" },
  ];

  return (
    <div className="border rounded-lg p-4 bg-blue-50 space-y-3">
      <h4 className="text-sm font-medium">Add Charge</h4>
      <select value={category} onChange={(e) => setCategory(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm">
        <option value="">Select category...</option>
        {categories.map((c) => <option key={c.value} value={c.value}>{c.label}</option>)}
      </select>
      <input type="text" placeholder="Description" value={desc} onChange={(e) => setDesc(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm" />
      <input type="number" min="0" step="0.01" placeholder="Amount (₹)" value={amount || ""} onChange={(e) => setAmount(Number(e.target.value))} className="w-full rounded-md border px-3 py-2 text-sm" />
      <div className="flex gap-2 justify-end">
        <button onClick={onCancel} className="px-3 py-1.5 text-xs border rounded-md">Cancel</button>
        <button onClick={() => { if (desc && amount > 0) onSubmit(desc, amount); }} disabled={loading || !category} className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded-md disabled:opacity-50">
          {loading ? "Adding..." : "Add"}
        </button>
      </div>
    </div>
  );
}

function AddPaymentForm({ onSubmit, onCancel, loading, balance }: { onSubmit: (amount: number, method: string, ref: string) => void; onCancel: () => void; loading: boolean; balance: number }) {
  const [amount, setAmount] = useState(balance);
  const [method, setMethod] = useState("cash");
  const [reference, setReference] = useState("");
  return (
    <div className="border rounded-lg p-4 bg-green-50 space-y-3">
      <h4 className="text-sm font-medium">Record Payment</h4>
      <input type="number" min="0" step="0.01" value={amount || ""} onChange={(e) => setAmount(Number(e.target.value))} className="w-full rounded-md border px-3 py-2 text-sm" placeholder="Amount (₹)" />
      <select value={method} onChange={(e) => setMethod(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm">
        <option value="cash">Cash</option>
        <option value="upi">UPI</option>
        <option value="card">Card</option>
        <option value="bank_transfer">Bank Transfer</option>
        <option value="cheque">Cheque</option>
      </select>
      <input type="text" placeholder="Reference / Transaction ID" value={reference} onChange={(e) => setReference(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm" />
      <div className="flex gap-2 justify-end">
        <button onClick={onCancel} className="px-3 py-1.5 text-xs border rounded-md">Cancel</button>
        <button onClick={() => { if (amount > 0) onSubmit(amount, method, reference); }} disabled={loading} className="px-3 py-1.5 text-xs bg-green-600 text-white rounded-md disabled:opacity-50">
          {loading ? "Recording..." : "Record"}
        </button>
      </div>
    </div>
  );
}

// Invoice PDF Button - Generate & Download on the spot
function InvoicePDFButton({ invoiceId }: { invoiceId: string }) {
  const api = useApi();
  const [loading, setLoading] = useState(false);

  const handleGenerate = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setLoading(true);
    try {
      const result = await api.post<any>(`/api/v1/billing/invoices/${invoiceId}/pdf`);
      const url = result?.url || result?.pdf_url || result?.download_url;
      if (url) {
        window.open(url, "_blank");
      } else {
        const pdfData = await api.get<any>(`/api/v1/billing/invoices/${invoiceId}/pdf`);
        const pdfUrl = pdfData?.url || pdfData?.pdf_url || pdfData?.download_url;
        if (pdfUrl) window.open(pdfUrl, "_blank");
      }
    } catch { /* silent */ }
    finally { setLoading(false); }
  };

  return (
    <button
      onClick={handleGenerate}
      disabled={loading}
      className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
    >
      <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
      {loading ? "..." : "PDF"}
    </button>
  );
}

// Quick Charge Modal - allows adding a charge to any currently checked-in guest from dashboard
function QuickChargeModal({ propertyId, onClose, onSuccess }: { propertyId: string; onClose: () => void; onSuccess: () => void }) {
  const api = useApi();
  const [units, setUnits] = useState<any[]>([]);
  const [selectedUnit, setSelectedUnit] = useState("");
  const [category, setCategory] = useState("other");
  const [description, setDescription] = useState("");
  const [amount, setAmount] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchUnits = async () => {
      try {
        const data = await api.get<any>(`/api/v1/properties/${propertyId}/units`);
        const allUnits = Array.isArray(data) ? data : data?.data || [];
        // Only show occupied units (they have active folios)
        setUnits(allUnits.filter((u: any) => u.status === "occupied"));
      } catch { /* silent */ }
    };
    fetchUnits();
  }, [api, propertyId]);

  const categories = [
    { value: "food_beverage", label: "Food / Delivery" },
    { value: "laundry", label: "Laundry" },
    { value: "parking", label: "Parking" },
    { value: "extra_bed", label: "Extra Bed" },
    { value: "damage", label: "Damage" },
    { value: "other", label: "Other" },
  ];

  const handleSubmit = async () => {
    if (!selectedUnit || !description || amount <= 0) {
      setError("Please fill all fields");
      return;
    }
    setLoading(true);
    setError("");
    try {
      // Get folio for the unit's reservation
      const reservations = await api.get<any>("/api/v1/reservations", { property_id: propertyId, status: "checked_in", per_page: "50" });
      const resList = Array.isArray(reservations) ? reservations : reservations?.data || [];
      const res = resList.find((r: any) => r.unit_id === selectedUnit);
      if (!res) {
        setError("No active reservation found for this unit");
        setLoading(false);
        return;
      }
      const folio = await api.get<any>(`/api/v1/billing/folios/reservation/${res.id}`);
      if (!folio?.id) {
        setError("No open folio found");
        setLoading(false);
        return;
      }
      await api.post("/api/v1/billing/charges", {
        folio_id: folio.id,
        category,
        description,
        quantity: 1,
        unit_price: amount,
        tax_rate: 18,
      });
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add charge");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" onClick={onClose}>
      <div className="bg-white rounded-lg w-full max-w-md p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Quick Charge</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">✕</button>
        </div>
        <p className="text-sm text-gray-500">Add a charge to any occupied room</p>

        {error && <div className="p-2 bg-red-50 text-red-700 text-sm rounded">{error}</div>}

        {/* Preset Quick Charges */}
        <div>
          <p className="text-xs font-medium text-gray-600 mb-2">Quick Presets:</p>
          <div className="grid grid-cols-2 gap-2">
            {[
              { label: "Extra Bed", amount: 500, cat: "extra_bed", desc: "Extra bed charge" },
              { label: "Parking (Day)", amount: 200, cat: "parking", desc: "Parking charge - 1 day" },
              { label: "Late Checkout", amount: 1000, cat: "late_checkout", desc: "Late checkout charge" },
              { label: "Extra Pillow/Blanket", amount: 200, cat: "other", desc: "Extra pillow/blanket" },
              { label: "Food Delivery", amount: 100, cat: "food_beverage", desc: "Food delivery handling fee" },
              { label: "Iron Box Usage", amount: 100, cat: "other", desc: "Iron box usage" },
            ].map((preset) => (
              <button
                key={preset.label}
                type="button"
                onClick={() => { setCategory(preset.cat); setDescription(preset.desc); setAmount(preset.amount); }}
                className="text-left px-2 py-1.5 text-xs border rounded hover:bg-blue-50 hover:border-blue-300"
              >
                <span className="font-medium">{preset.label}</span>
                <span className="text-gray-500 ml-1">₹{preset.amount}</span>
              </button>
            ))}
          </div>
        </div>

        <select value={selectedUnit} onChange={(e) => setSelectedUnit(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm">
          <option value="">Select occupied room...</option>
          {units.map((u) => (
            <option key={u.id} value={u.id}>Unit {u.unit_number}{u.guest_name ? ` - ${u.guest_name}` : ""}</option>
          ))}
        </select>

        <select value={category} onChange={(e) => setCategory(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm">
          {categories.map((c) => <option key={c.value} value={c.value}>{c.label}</option>)}
        </select>

        <input type="text" placeholder="Description (e.g., Food delivery charge)" value={description} onChange={(e) => setDescription(e.target.value)} className="w-full rounded-md border px-3 py-2 text-sm" />
        <input type="number" min="0" step="1" placeholder="Amount (₹)" value={amount || ""} onChange={(e) => setAmount(Number(e.target.value))} className="w-full rounded-md border px-3 py-2 text-sm" />

        <div className="flex gap-2 justify-end pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm border rounded-md">Cancel</button>
          <button onClick={handleSubmit} disabled={loading} className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
            {loading ? "Adding..." : "Add Charge"}
          </button>
        </div>
      </div>
    </div>
  );
}
