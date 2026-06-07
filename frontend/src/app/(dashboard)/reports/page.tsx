"use client";

import React, { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";

type ReportType = "occupancy" | "revenue" | "gst" | "laundry" | "daily_collection" | "outstanding" | "end_of_day";

export default function ReportsPage() {
  const { user } = useAuth();
  const api = useApi();
  const [activeReport, setActiveReport] = useState<ReportType>("end_of_day");
  const [dateRange, setDateRange] = useState({
    start: new Date().toISOString().split("T")[0],
    end: new Date().toISOString().split("T")[0],
  });
  const [reportData, setReportData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCloseDayConfirm, setShowCloseDayConfirm] = useState(false);

  const propertyId = user?.property_id || "";

  const reports = [
    { id: "end_of_day" as ReportType, label: "End of Day", desc: "Complete day summary" },
    { id: "daily_collection" as ReportType, label: "Daily Collection", desc: "Cash/UPI/Card breakdown" },
    { id: "outstanding" as ReportType, label: "Outstanding Dues", desc: "Guests with pending balance" },
    { id: "occupancy" as ReportType, label: "Occupancy", desc: "Unit occupancy rates" },
    { id: "revenue" as ReportType, label: "Revenue", desc: "Revenue by date range" },
    { id: "gst" as ReportType, label: "GST", desc: "Tax summary for filing" },
    { id: "laundry" as ReportType, label: "Laundry", desc: "Laundry order summary" },
  ];

  const generateReport = useCallback(async () => {
    if (!propertyId) return;
    setLoading(true);
    setError(null);
    try {
      let data;
      switch (activeReport) {
        case "end_of_day":
          data = await api.get(`/api/v1/dashboard/${propertyId}/end-of-day`, { date: dateRange.start });
          break;
        case "daily_collection":
          data = await api.get(`/api/v1/dashboard/${propertyId}/daily-collection`, { date: dateRange.start });
          break;
        case "outstanding":
          data = await api.get(`/api/v1/dashboard/${propertyId}/outstanding-dues`);
          break;
        case "occupancy":
          data = await api.get(`/api/v1/calendar/${propertyId}/occupancy`, { start_date: dateRange.start, end_date: dateRange.end });
          break;
        case "revenue":
          data = await api.get(`/api/v1/dashboard/${propertyId}/revenue-trend`, { start_date: dateRange.start, end_date: dateRange.end });
          break;
        case "gst":
          data = await api.get("/api/v1/billing/invoices", { property_id: propertyId, start_date: dateRange.start, end_date: dateRange.end });
          break;
        case "laundry":
          data = await api.get(`/api/v1/laundry/stats/${propertyId}`);
          break;
      }
      setReportData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate report");
    } finally {
      setLoading(false);
    }
  }, [api, propertyId, activeReport, dateRange]);

  // Auto-load report when report type or date changes
  useEffect(() => {
    if (propertyId) {
      generateReport();
    }
  }, [generateReport, propertyId]);

  const exportCSV = () => {
    if (!reportData) return;
    let data: any[];
    if (activeReport === "outstanding") {
      data = reportData.dues || [];
    } else if (activeReport === "daily_collection") {
      data = [reportData];
    } else {
      data = Array.isArray(reportData) ? reportData : reportData?.data || [reportData];
    }
    if (data.length === 0) return;

    const headers = Object.keys(data[0]).filter((k) => !k.includes("id") || k === "reservation_id");
    const csv = [
      headers.join(","),
      ...data.map((row: any) => headers.map((h) => JSON.stringify(row[h] ?? "")).join(",")),
    ].join("\n");

    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${activeReport}_report_${dateRange.start}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const fmt = (n: number) => new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", minimumFractionDigits: 0 }).format(n);

  return (
    <div className="space-y-4">
      <h1 className="text-xl lg:text-2xl font-bold text-gray-900">Reports</h1>

      {/* Report selector */}
      <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-3">
        {reports.map((r) => (
          <button
            key={r.id}
            onClick={() => { setActiveReport(r.id); setReportData(null); }}
            className={`p-3 rounded-lg border text-left transition-colors ${
              activeReport === r.id ? "border-blue-500 bg-blue-50" : "hover:bg-gray-50"
            }`}
          >
            <p className="text-sm font-medium">{r.label}</p>
            <p className="text-xs text-gray-500 mt-0.5">{r.desc}</p>
          </button>
        ))}
      </div>

      {/* Date Range & Generate */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3 p-4 bg-white border rounded-lg">
        {activeReport !== "outstanding" && activeReport !== "laundry" && (
          <>
            <div className="flex items-center gap-2">
              <label className="text-sm text-gray-600">{activeReport === "daily_collection" ? "Date:" : "From:"}</label>
              <input type="date" value={dateRange.start} onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })} className="rounded-md border border-gray-300 px-3 py-1.5 text-sm" />
            </div>
            {activeReport !== "daily_collection" && (
              <div className="flex items-center gap-2">
                <label className="text-sm text-gray-600">To:</label>
                <input type="date" value={dateRange.end} onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })} className="rounded-md border border-gray-300 px-3 py-1.5 text-sm" />
              </div>
            )}
            <div className="flex items-center gap-1">
              <button
                onClick={() => {
                  const yesterday = new Date();
                  yesterday.setDate(yesterday.getDate() - 1);
                  setDateRange({ start: yesterday.toISOString().split("T")[0], end: yesterday.toISOString().split("T")[0] });
                }}
                className="px-2 py-1 text-xs border rounded hover:bg-gray-50"
              >
                Yesterday
              </button>
              <button
                onClick={() => {
                  const today = new Date().toISOString().split("T")[0];
                  setDateRange({ start: today, end: today });
                }}
                className="px-2 py-1 text-xs border rounded hover:bg-gray-50"
              >
                Today
              </button>
            </div>
          </>
        )}
        <div className="flex gap-2">
          <button onClick={generateReport} disabled={loading} className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
            {loading ? "Loading..." : "Refresh"}
          </button>
          {reportData && (
            <button onClick={exportCSV} className="px-4 py-1.5 text-sm border rounded-md hover:bg-gray-50">Export CSV</button>
          )}
        </div>
      </div>

      {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>}

      {/* End of Day Report */}
      {reportData && !loading && activeReport === "end_of_day" && (
        <div className="border rounded-lg bg-white p-6 space-y-6">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-lg">End of Day Summary - {reportData.date}</h3>
            {!reportData.is_closed ? (
              <div className="relative group">
                <button
                  onClick={() => setShowCloseDayConfirm(true)}
                  className="px-4 py-2 bg-red-600 text-white text-sm rounded-md hover:bg-red-700"
                >
                  Close Day
                </button>
                <div className="absolute right-0 top-full mt-1 w-64 p-2 bg-gray-900 text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                  Closing the day locks all billing for this date. No more charges or payments can be added. Use this at end of business day after verifying collections.
                </div>
              </div>
            ) : (
              <span className="px-3 py-1 bg-green-100 text-green-700 text-xs font-medium rounded-full">Day Closed ✓</span>
            )}
          </div>

          {/* Occupancy */}
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-blue-50 rounded-lg p-4 text-center">
              <p className="text-3xl font-bold text-blue-700">{reportData.occupancy_rate?.toFixed(0) || 0}%</p>
              <p className="text-xs text-blue-600 mt-1">Occupancy</p>
            </div>
            <div className="bg-green-50 rounded-lg p-4 text-center">
              <p className="text-3xl font-bold text-green-700">{reportData.occupied_units || 0}/{reportData.total_units || 0}</p>
              <p className="text-xs text-green-600 mt-1">Units Occupied</p>
            </div>
            <div className="bg-purple-50 rounded-lg p-4 text-center">
              <p className="text-3xl font-bold text-purple-700">{fmt(reportData.collection?.net_collection || 0)}</p>
              <p className="text-xs text-purple-600 mt-1">Net Collection</p>
            </div>
          </div>

          {/* Operations */}
          <div>
            <h4 className="font-medium text-sm text-gray-700 mb-2">Operations</h4>
            <div className="grid grid-cols-3 sm:grid-cols-6 gap-2 text-center">
              <div className="border rounded p-2"><p className="text-lg font-bold text-green-600">{reportData.check_ins || 0}</p><p className="text-[10px] text-gray-500">Check-ins</p></div>
              <div className="border rounded p-2"><p className="text-lg font-bold text-orange-600">{reportData.check_outs || 0}</p><p className="text-[10px] text-gray-500">Check-outs</p></div>
              <div className="border rounded p-2"><p className="text-lg font-bold text-blue-600">{reportData.walk_ins || 0}</p><p className="text-[10px] text-gray-500">Walk-ins</p></div>
              <div className="border rounded p-2"><p className="text-lg font-bold text-red-600">{reportData.no_shows || 0}</p><p className="text-[10px] text-gray-500">No-shows</p></div>
              <div className="border rounded p-2"><p className="text-lg font-bold text-gray-600">{reportData.cancellations || 0}</p><p className="text-[10px] text-gray-500">Cancellations</p></div>
              <div className="border rounded p-2"><p className="text-lg font-bold text-indigo-600">{reportData.extensions || 0}</p><p className="text-[10px] text-gray-500">Extensions</p></div>
            </div>
          </div>

          {/* Collection Breakdown */}
          <div>
            <h4 className="font-medium text-sm text-gray-700 mb-2">Collection Breakdown</h4>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
              {[
                { label: "Cash", value: reportData.collection?.cash },
                { label: "UPI", value: reportData.collection?.upi },
                { label: "Card", value: reportData.collection?.card },
                { label: "Transfer", value: reportData.collection?.bank_transfer },
                { label: "Cheque", value: reportData.collection?.cheque },
              ].map((m) => (
                <div key={m.label} className="border rounded p-2 text-center">
                  <p className="font-medium text-sm">{fmt(m.value || 0)}</p>
                  <p className="text-[10px] text-gray-500">{m.label}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Outstanding */}
          <div>
            <h4 className="font-medium text-sm text-gray-700 mb-2">
              Outstanding Dues
              <span className="ml-2 text-red-600 font-bold">{fmt(reportData.outstanding?.total_outstanding || 0)}</span>
              <span className="text-xs text-gray-400 ml-1">({reportData.outstanding?.count || 0} folios)</span>
            </h4>
            {reportData.outstanding?.dues?.length > 0 && (
              <div className="space-y-1 max-h-40 overflow-y-auto">
                {reportData.outstanding.dues.slice(0, 10).map((d: any, i: number) => (
                  <div key={i} className="flex justify-between text-sm py-1 border-b border-gray-50">
                    <span>{d.guest_name} <span className="text-gray-400">({d.unit_number})</span></span>
                    <span className="font-medium text-red-600">{fmt(d.balance)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Daily Collection Report */}
      {reportData && !loading && activeReport === "daily_collection" && (
        <div className="border rounded-lg bg-white p-6">
          <h3 className="font-semibold text-lg mb-4">Daily Collection - {reportData.date}</h3>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <div className="bg-green-50 rounded-lg p-4 text-center">
              <p className="text-2xl font-bold text-green-700">{fmt(reportData.total_collected || 0)}</p>
              <p className="text-xs text-green-600 mt-1">Total Collected</p>
            </div>
            <div className="bg-blue-50 rounded-lg p-4 text-center">
              <p className="text-2xl font-bold text-blue-700">{fmt(reportData.net_collection || 0)}</p>
              <p className="text-xs text-blue-600 mt-1">Net (after refunds)</p>
            </div>
            <div className="bg-red-50 rounded-lg p-4 text-center">
              <p className="text-2xl font-bold text-red-700">{fmt(reportData.total_refunds || 0)}</p>
              <p className="text-xs text-red-600 mt-1">Refunds</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4 text-center">
              <p className="text-2xl font-bold text-gray-700">{reportData.transactions || 0}</p>
              <p className="text-xs text-gray-600 mt-1">Transactions</p>
            </div>
          </div>
          <h4 className="font-medium text-sm text-gray-700 mb-3">Breakdown by Method</h4>
          <div className="space-y-2">
            {[
              { label: "Cash", value: reportData.cash, color: "bg-green-500" },
              { label: "UPI", value: reportData.upi, color: "bg-blue-500" },
              { label: "Card", value: reportData.card, color: "bg-purple-500" },
              { label: "Bank Transfer", value: reportData.bank_transfer, color: "bg-indigo-500" },
              { label: "Cheque", value: reportData.cheque, color: "bg-yellow-500" },
            ].filter((m) => m.value > 0).map((method) => (
              <div key={method.label} className="flex items-center gap-3">
                <div className={`w-3 h-3 rounded-full ${method.color}`}></div>
                <span className="text-sm flex-1">{method.label}</span>
                <span className="font-medium text-sm">{fmt(method.value)}</span>
              </div>
            ))}
            {reportData.total_collected === 0 && <p className="text-sm text-gray-400">No collections today</p>}
          </div>
        </div>
      )}

      {/* Outstanding Dues Report */}
      {reportData && !loading && activeReport === "outstanding" && (
        <div className="border rounded-lg bg-white p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="font-semibold text-lg">Outstanding Dues</h3>
            <div className="bg-red-50 rounded-lg px-4 py-2">
              <span className="text-sm text-red-600">Total: </span>
              <span className="font-bold text-red-700">{fmt(reportData.total_outstanding || 0)}</span>
              <span className="text-xs text-red-500 ml-2">({reportData.count || 0} guests)</span>
            </div>
          </div>
          {reportData.dues && reportData.dues.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="text-left px-3 py-2 font-medium">Guest</th>
                    <th className="text-left px-3 py-2 font-medium">Unit</th>
                    <th className="text-left px-3 py-2 font-medium">Status</th>
                    <th className="text-right px-3 py-2 font-medium">Total</th>
                    <th className="text-right px-3 py-2 font-medium">Paid</th>
                    <th className="text-right px-3 py-2 font-medium text-red-600">Balance</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {reportData.dues.map((d: any, i: number) => (
                    <tr key={i} className="hover:bg-gray-50">
                      <td className="px-3 py-2 font-medium">{d.guest_name}</td>
                      <td className="px-3 py-2">{d.unit_number}</td>
                      <td className="px-3 py-2"><span className={`px-2 py-0.5 text-xs rounded-full ${d.status === "checked_in" ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-600"}`}>{d.status}</span></td>
                      <td className="px-3 py-2 text-right">{fmt(d.total_amount)}</td>
                      <td className="px-3 py-2 text-right text-green-600">{fmt(d.paid_amount)}</td>
                      <td className="px-3 py-2 text-right font-bold text-red-600">{fmt(d.balance)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-sm text-gray-400 text-center py-8">No outstanding dues. All clear!</p>
          )}
        </div>
      )}

      {/* Generic Report Table for other types */}
      {reportData && !loading && !["daily_collection", "outstanding", "end_of_day"].includes(activeReport) && (
        <div className="border rounded-lg bg-white p-4">
          <h3 className="font-medium text-sm text-gray-700 mb-4 capitalize">{activeReport.replace(/_/g, " ")} Report</h3>
          <ReportTable data={reportData} type={activeReport} />
        </div>
      )}

      {/* Close Day Confirmation Dialog */}
      {showCloseDayConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
                <svg className="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Close Day?</h3>
                <p className="text-sm text-gray-500">This action cannot be undone</p>
              </div>
            </div>
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
              <p className="text-sm text-yellow-800">Closing the day will:</p>
              <ul className="text-sm text-yellow-700 mt-1 space-y-1 list-disc pl-4">
                <li>Lock all billing and payments for {dateRange.start}</li>
                <li>Mark the day as audited (no further changes allowed)</li>
                <li>Finalize the daily collection totals</li>
              </ul>
              <p className="text-xs text-yellow-600 mt-2 font-medium">Only do this after verifying your cash and online collections match.</p>
            </div>
            <div className="flex gap-3 justify-end">
              <button onClick={() => setShowCloseDayConfirm(false)} className="px-4 py-2 text-sm border rounded-md hover:bg-gray-50">Cancel</button>
              <button
                onClick={async () => {
                  setShowCloseDayConfirm(false);
                  await api.post(`/api/v1/dashboard/${propertyId}/close-day`, { date: dateRange.start });
                  generateReport();
                }}
                className="px-4 py-2 text-sm bg-red-600 text-white rounded-md hover:bg-red-700"
              >
                Yes, Close Day
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ReportTable({ data, type }: { data: any; type: ReportType }) {
  const items = Array.isArray(data) ? data : data?.data || (typeof data === "object" ? [data] : []);
  if (items.length === 0) return <p className="text-sm text-gray-400">No data for this period</p>;

  // GST Report: show summary + invoice table
  if (type === "gst" && items.length > 0) {
    const totalTaxable = items.reduce((sum: number, inv: any) => sum + (Number(inv.total_amount) || 0) - (Number(inv.tax_amount) || 0), 0);
    const totalTax = items.reduce((sum: number, inv: any) => sum + (Number(inv.tax_amount) || 0), 0);
    const totalInvoiceAmount = items.reduce((sum: number, inv: any) => sum + (Number(inv.total_amount) || 0), 0);
    const cgst = totalTax / 2;
    const sgst = totalTax / 2;

    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div className="bg-blue-50 rounded-lg p-3 text-center">
            <p className="text-lg font-bold text-blue-700">₹{totalTaxable.toLocaleString()}</p>
            <p className="text-xs text-blue-600">Taxable Amount</p>
          </div>
          <div className="bg-green-50 rounded-lg p-3 text-center">
            <p className="text-lg font-bold text-green-700">₹{cgst.toLocaleString()}</p>
            <p className="text-xs text-green-600">CGST (9%)</p>
          </div>
          <div className="bg-green-50 rounded-lg p-3 text-center">
            <p className="text-lg font-bold text-green-700">₹{sgst.toLocaleString()}</p>
            <p className="text-xs text-green-600">SGST (9%)</p>
          </div>
          <div className="bg-purple-50 rounded-lg p-3 text-center">
            <p className="text-lg font-bold text-purple-700">₹{totalInvoiceAmount.toLocaleString()}</p>
            <p className="text-xs text-purple-600">Total Invoice</p>
          </div>
        </div>
        <p className="text-xs text-gray-500">SAC Code: 996311 (Hotel accommodation) • {items.length} invoices</p>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-left px-3 py-2 font-medium">Invoice #</th>
                <th className="text-left px-3 py-2 font-medium">Guest</th>
                <th className="text-right px-3 py-2 font-medium">Taxable</th>
                <th className="text-right px-3 py-2 font-medium">Tax</th>
                <th className="text-right px-3 py-2 font-medium">Total</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {items.slice(0, 50).map((inv: any, i: number) => (
                <tr key={i}>
                  <td className="px-3 py-2 font-mono text-xs">{inv.invoice_number || "—"}</td>
                  <td className="px-3 py-2">{inv.guest_name || "—"}</td>
                  <td className="px-3 py-2 text-right">₹{((Number(inv.total_amount) || 0) - (Number(inv.tax_amount) || 0)).toLocaleString()}</td>
                  <td className="px-3 py-2 text-right text-green-600">₹{(Number(inv.tax_amount) || 0).toLocaleString()}</td>
                  <td className="px-3 py-2 text-right font-medium">₹{(Number(inv.total_amount) || 0).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  if (items.length === 1 && !Array.isArray(data)) {
    const obj = items[0];
    return (
      <div className="space-y-2">
        {Object.entries(obj).map(([key, value]) => (
          <div key={key} className="flex justify-between text-sm py-1 border-b border-gray-50">
            <span className="text-gray-500 capitalize">{key.replace(/_/g, " ")}</span>
            <span className="font-medium">{typeof value === "number" ? (key.includes("rate") ? `${(value as number).toFixed(1)}%` : `₹${(value as number).toLocaleString()}`) : String(value ?? "—")}</span>
          </div>
        ))}
      </div>
    );
  }

  const headers = Object.keys(items[0]).filter((k) => !k.includes("id"));
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead className="bg-gray-50">
          <tr>{headers.map((h) => (<th key={h} className="text-left px-3 py-2 font-medium text-gray-600 capitalize">{h.replace(/_/g, " ")}</th>))}</tr>
        </thead>
        <tbody className="divide-y">
          {items.slice(0, 50).map((row: any, i: number) => (
            <tr key={i}>{headers.map((h) => (<td key={h} className="px-3 py-2">{typeof row[h] === "number" ? (h.includes("amount") || h.includes("rate") || h.includes("total") || h.includes("revenue") ? `₹${row[h].toLocaleString()}` : row[h].toLocaleString()) : String(row[h] ?? "—")}</td>))}</tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
