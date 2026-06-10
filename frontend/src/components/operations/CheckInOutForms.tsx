"use client";

import React, { useState, useEffect } from "react";
import { useApi } from "@/hooks/useApi";
import { useAuth } from "@/contexts/AuthContext";

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
  id_document_number: string;
  notes: string;
}

export function CheckInForm({ reservationId, unitId, onSubmit, onCancel }: CheckInFormProps) {
  const api = useApi();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [docFile, setDocFile] = useState<File | null>(null);
  const [docUploading, setDocUploading] = useState(false);
  const [waiveDeposit, setWaiveDeposit] = useState(false);
  const [roomStatus, setRoomStatus] = useState<string | null>(null);
  const [form, setForm] = useState<CheckInData>({
    reservation_id: reservationId,
    assigned_unit_id: unitId,
    deposit_amount: 2000,
    deposit_method: "cash",
    deposit_reference: "",
    id_document_type: "aadhaar",
    id_document_number: "",
    notes: "",
  });

  // Check room readiness
  useEffect(() => {
    if (!unitId || !user?.property_id) return;
    api.get<any>(`/api/v1/properties/${user.property_id}/units`)
      .then((data) => {
        const units = Array.isArray(data) ? data : data?.data || [];
        const unit = units.find((u: any) => u.id === unitId);
        if (unit) setRoomStatus(unit.status);
      })
      .catch(() => {});
  }, [api, unitId, user?.property_id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!waiveDeposit && form.deposit_amount <= 0) { setError("Deposit amount is required (or check 'Corporate / No Deposit')."); return; }
    if (!form.deposit_method && !waiveDeposit) { setError("Select deposit payment method."); return; }
    if (!form.id_document_type || !form.id_document_number || form.id_document_number.length < 4) {
      setError("ID document type and number are mandatory (legal requirement)."); return;
    }
    setLoading(true);
    setError(null);
    try { await onSubmit({ ...form, deposit_amount: waiveDeposit ? 0 : form.deposit_amount }); } catch (err) { setError(err instanceof Error ? err.message : "Check-in failed"); } finally { setLoading(false); }
  };

  // Upload document photo after check-in form submission
  const handleDocUpload = async (guestId: string) => {
    if (!docFile) return;
    setDocUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", docFile);
      formData.append("document_type", form.id_document_type);
      await fetch(`/api/v1/guests/${guestId}/documents`, {
        method: "POST",
        body: formData,
      });
    } catch { /* non-blocking */ }
    finally { setDocUploading(false); }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h3 className="text-lg font-semibold">Check-In</h3>

      {/* Room Readiness Indicator */}
      {roomStatus && (
        <div className={`p-3 rounded-lg border flex items-center gap-2 ${
          roomStatus === "available" ? "bg-green-50 border-green-200" :
          roomStatus === "cleaning" ? "bg-orange-50 border-orange-200" :
          roomStatus === "maintenance" ? "bg-red-50 border-red-200" :
          "bg-yellow-50 border-yellow-200"
        }`}>
          <div className={`w-3 h-3 rounded-full ${
            roomStatus === "available" ? "bg-green-500" :
            roomStatus === "cleaning" ? "bg-orange-500" :
            roomStatus === "maintenance" ? "bg-red-500" :
            "bg-yellow-500"
          }`} />
          <span className={`text-sm font-medium ${
            roomStatus === "available" ? "text-green-800" :
            roomStatus === "cleaning" ? "text-orange-800" :
            roomStatus === "maintenance" ? "text-red-800" :
            "text-yellow-800"
          }`}>
            Room Status: {roomStatus === "available" ? "Ready ✓" : roomStatus === "cleaning" ? "Cleaning in Progress — Not Ready" : roomStatus === "maintenance" ? "Under Maintenance — Cannot Check In" : roomStatus.replace("_", " ").toUpperCase()}
          </span>
        </div>
      )}

      {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>}

      <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
        <p className="text-xs font-medium text-yellow-800 mb-2">ID Verification (Legal Requirement)</p>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm font-medium text-gray-700">ID Type *</label>
            <select required value={form.id_document_type} onChange={(e) => setForm({ ...form, id_document_type: e.target.value })} className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
              <option value="aadhaar">Aadhaar Card</option>
              <option value="passport">Passport</option>
              <option value="driving_license">Driving License</option>
              <option value="voter_id">Voter ID</option>
              <option value="pan_card">PAN Card</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">ID Number *</label>
            <input type="text" required minLength={4} value={form.id_document_number} onChange={(e) => setForm({ ...form, id_document_number: e.target.value })} placeholder="Enter ID number" className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" />
          </div>
        </div>
        <div className="mt-3">
          <label className="block text-sm font-medium text-gray-700">Upload ID Photo (optional)</label>
          <input
            type="file"
            accept="image/*,.pdf"
            onChange={(e) => setDocFile(e.target.files?.[0] || null)}
            className="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-1.5 file:px-3 file:rounded file:border file:border-gray-300 file:text-sm file:bg-white hover:file:bg-gray-50"
          />
          {docFile && <p className="text-xs text-green-600 mt-1">Photo ready: {docFile.name}</p>}
        </div>
      </div>

      <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
        <div className="flex items-center justify-between mb-2">
          <p className="text-xs font-medium text-blue-800">Security Deposit</p>
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" checked={waiveDeposit} onChange={(e) => setWaiveDeposit(e.target.checked)} className="rounded border-gray-300 text-blue-600" />
            <span className="text-xs text-gray-600">Corporate / No Deposit</span>
          </label>
        </div>
        {!waiveDeposit && (
          <>
            <p className="text-xs text-blue-600 mb-2">Suggested: ₹2,000 (1 night rate). Adjust as needed.</p>
            <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700">Amount *</label>
              <input type="number" required min={1} value={form.deposit_amount} onChange={(e) => setForm({ ...form, deposit_amount: parseFloat(e.target.value) || 0 })} className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Method *</label>
              <select required value={form.deposit_method} onChange={(e) => setForm({ ...form, deposit_method: e.target.value })} className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
                <option value="cash">Cash</option>
                <option value="upi">UPI</option>
                <option value="card">Card</option>
                <option value="bank_transfer">Bank Transfer</option>
                <option value="cheque">Cheque</option>
              </select>
            </div>
          </div>
          </>
        )}
        {!waiveDeposit && form.deposit_method !== "cash" && (
          <div className="mt-2">
            <label className="block text-sm font-medium text-gray-700">Reference</label>
            <input type="text" value={form.deposit_reference} onChange={(e) => setForm({ ...form, deposit_reference: e.target.value })} placeholder="Transaction ID" className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" />
          </div>
        )}
        {waiveDeposit && (
          <p className="text-xs text-gray-500 mt-1">Deposit waived — suitable for corporate bookings or trusted repeat guests.</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700">Notes</label>
        <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} rows={2} className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" />
      </div>

      <div className="flex gap-2 justify-end">
        <button type="button" onClick={onCancel} className="px-4 py-2 text-sm border rounded hover:bg-gray-100">Cancel</button>
        <button type="submit" disabled={loading} className="px-4 py-2 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50">{loading ? "Processing..." : "Complete Check-In"}</button>
      </div>
    </form>
  );
}

interface CheckOutFormProps {
  reservationId: string;
  guestName: string;
  onSubmit: (data: { reservation_id: string; notes: string; late_checkout_charge: number }) => Promise<void>;
  onCancel: () => void;
}

interface BillSummary {
  folio_id: string;
  guest_name: string;
  unit_number: string;
  check_in_date: string;
  check_out_date: string;
  nights: number;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  paid_amount: number;
  balance: number;
  deposit_held: number;
  is_late_check_out: boolean;
  late_by_hours: number;
  line_items: { category: string; description: string; quantity: number; unit_price: number; amount: number; tax_amount: number; total: number }[];
  payments: { payment_type: string; payment_method: string; amount: number; reference: string; paid_at: string }[];
}

export function CheckOutForm({ reservationId, guestName, onSubmit, onCancel }: CheckOutFormProps) {
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [loadingSummary, setLoadingSummary] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notes, setNotes] = useState("");
  const [summary, setSummary] = useState<BillSummary | null>(null);
  const [showConfirm, setShowConfirm] = useState(false);
  const [lateCharge, setLateCharge] = useState(0);
  const [pendingLaundry, setPendingLaundry] = useState<any[]>([]);
  const [showPayment, setShowPayment] = useState(false);
  const [paymentAmount, setPaymentAmount] = useState(0);
  const [paymentMethod, setPaymentMethod] = useState("cash");
  const [paymentReference, setPaymentReference] = useState("");
  const [paymentRecorded, setPaymentRecorded] = useState(false);
  const { user } = useAuth();

  useEffect(() => {
    setLoadingSummary(true);
    Promise.all([
      api.get<BillSummary>("/api/v1/operations/pre-checkout", { reservation_id: reservationId }),
      api.get<any>("/api/v1/laundry/orders", { property_id: user?.property_id || "", status: "received,washing,ready", per_page: "50" }).catch(() => ({ data: [] })),
    ])
      .then(([billData, laundryData]) => {
        setSummary(billData);
        // Filter laundry orders for this guest/unit
        const allOrders = Array.isArray(laundryData) ? laundryData : laundryData?.data || [];
        const pending = allOrders.filter((o: any) => o.unit_number === billData.unit_number && o.status !== "delivered");
        setPendingLaundry(pending);
        setPaymentAmount(Math.max(0, Number(billData.balance)));
        setError(null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load bill"))
      .finally(() => setLoadingSummary(false));
  }, [api, reservationId, user?.property_id]);

  const handleRecordPayment = async () => {
    if (paymentAmount <= 0) return;
    setLoading(true);
    try {
      await api.post("/api/v1/billing/payments", {
        folio_id: summary?.folio_id,
        amount: paymentAmount,
        method: paymentMethod,
        reference: paymentReference,
      });
      setPaymentRecorded(true);
      setShowPayment(false);
      // Refresh summary
      const billData = await api.get<BillSummary>("/api/v1/operations/pre-checkout", { reservation_id: reservationId });
      setSummary(billData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Payment failed");
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    setLoading(true); setError(null);
    try { await onSubmit({ reservation_id: reservationId, notes, late_checkout_charge: lateCharge }); }
    catch (err) { setError(err instanceof Error ? err.message : "Check-out failed"); setShowConfirm(false); }
    finally { setLoading(false); }
  };

  const fmt = (n: number) => new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", minimumFractionDigits: 0 }).format(n);

  if (loadingSummary) {
    return (<div className="space-y-4"><h3 className="text-lg font-semibold">Check-Out</h3><div className="flex justify-center py-8"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-orange-600"></div></div><p className="text-sm text-gray-500 text-center">Loading bill...</p></div>);
  }

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">Check-Out - {guestName}</h3>
      {error && <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">{error}</div>}

      {summary && (
        <div className="border rounded-lg overflow-hidden">
          <div className="bg-gray-50 px-4 py-3 border-b">
            <div className="flex justify-between text-sm"><span>Unit <strong>{summary.unit_number}</strong></span><span>{summary.nights} night(s)</span></div>
            <div className="text-xs text-gray-500 mt-1">{summary.check_in_date} to {summary.check_out_date}</div>
          </div>

          <div className="px-4 py-3 divide-y divide-gray-100">
            {summary.line_items?.map((item, i) => (
              <div key={i} className="flex justify-between py-1.5 text-sm">
                <span className="text-gray-700">{item.description}</span>
                <span className="font-medium">{fmt(item.total)}</span>
              </div>
            ))}
          </div>

          {summary.is_late_check_out && (
            <div className="px-4 py-3 bg-orange-50 border-t border-orange-200">
              <p className="text-sm text-orange-800 font-medium">Late checkout by {summary.late_by_hours.toFixed(1)} hours (after 11:00 AM)</p>
              <div className="flex items-center gap-2 mt-2">
                <label className="text-sm text-orange-700">Add charge:</label>
                <input type="number" min={0} value={lateCharge} onChange={(e) => setLateCharge(Number(e.target.value))} className="w-28 rounded border-orange-300 px-2 py-1 text-sm" placeholder="0" />
              </div>
            </div>
          )}

          <div className="px-4 py-3 bg-gray-50 border-t space-y-1 text-sm">
            <div className="flex justify-between"><span>Subtotal</span><span>{fmt(summary.subtotal)}</span></div>
            <div className="flex justify-between text-gray-500"><span>Tax (GST)</span><span>{fmt(summary.tax_total)}</span></div>
            {lateCharge > 0 && <div className="flex justify-between text-orange-600"><span>Late checkout</span><span>+ {fmt(lateCharge * 1.18)}</span></div>}
            <div className="flex justify-between font-bold text-base pt-2 border-t"><span>Total</span><span>{fmt(Number(summary.total_amount) + lateCharge * 1.18)}</span></div>
            <div className="flex justify-between text-green-700"><span>Paid</span><span>- {fmt(summary.paid_amount)}</span></div>
            {summary.deposit_held > 0 && (
              <div className="flex justify-between text-blue-700">
                <span>Deposit Held</span>
                <span>{fmt(summary.deposit_held)}</span>
              </div>
            )}
            <div className={`flex justify-between font-bold text-lg pt-1 ${Number(summary.balance) + lateCharge * 1.18 > 0 ? "text-red-600" : "text-green-600"}`}>
              <span>Balance Due</span><span>{fmt(Number(summary.balance) + lateCharge * 1.18)}</span>
            </div>
            {summary.deposit_held > 0 && Number(summary.balance) + lateCharge * 1.18 > 0 && (
              <div className="pt-2 border-t">
                <p className="text-xs text-blue-700 font-medium">
                  After applying deposit: {fmt(Math.max(0, Number(summary.balance) + lateCharge * 1.18 - summary.deposit_held))} due
                  {Number(summary.balance) + lateCharge * 1.18 < summary.deposit_held && (
                    <span className="text-green-600"> • Refund ₹{(summary.deposit_held - Number(summary.balance) - lateCharge * 1.18).toFixed(0)} to guest</span>
                  )}
                </p>
                {Number(summary.balance) + lateCharge * 1.18 < summary.deposit_held && (
                  <div className="mt-2">
                    <label className="block text-xs text-gray-600 mb-1">Refund Method:</label>
                    <select className="w-full rounded-lg border border-input bg-background px-3 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                      <option value="cash">Cash</option>
                      <option value="upi">UPI (to original account)</option>
                      <option value="bank_transfer">Bank Transfer</option>
                      <option value="adjust">Adjust against future booking</option>
                    </select>
                  </div>
                )}
              </div>
            )}
          </div>

          {summary.payments && summary.payments.length > 0 && (
            <details className="px-4 py-2 border-t">
              <summary className="text-xs text-gray-500 cursor-pointer">Payments received ({summary.payments.length})</summary>
              <div className="mt-2 space-y-1">{summary.payments.map((p, i) => (<div key={i} className="flex justify-between text-xs text-gray-600"><span>{p.payment_type} ({p.payment_method})</span><span>{fmt(p.amount)}</span></div>))}</div>
            </details>
          )}
        </div>
      )}

      {/* Pending Laundry Warning */}
      {pendingLaundry.length > 0 && (
        <div className="p-3 bg-yellow-50 border border-yellow-300 rounded-lg">
          <div className="flex items-center gap-2 mb-1">
            <svg className="w-4 h-4 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
            <p className="text-sm font-medium text-yellow-800">Pending Laundry ({pendingLaundry.length} order{pendingLaundry.length > 1 ? "s" : ""})</p>
          </div>
          <div className="space-y-1 ml-6">
            {pendingLaundry.map((order: any, i: number) => (
              <div key={i} className="flex items-center justify-between text-xs text-yellow-700 py-0.5">
                <span>
                  #{order.order_number} — Status: <span className="font-medium capitalize">{order.status}</span>
                  {order.grand_total > 0 && !order.posted_to_folio && <span className="text-red-600 ml-1">(₹{order.grand_total} not billed)</span>}
                </span>
                {!order.posted_to_folio && order.grand_total > 0 && (
                  <button
                    onClick={async () => {
                      try {
                        await api.post(`/api/v1/laundry/orders/${order.id}/post-to-folio`);
                        // Refresh pre-checkout summary
                        const billData = await api.get<BillSummary>("/api/v1/operations/pre-checkout", { reservation_id: reservationId });
                        setSummary(billData);
                        setPendingLaundry((prev: any[]) => prev.map((o: any) => o.id === order.id ? { ...o, posted_to_folio: true } : o));
                      } catch { /* silent */ }
                    }}
                    className="px-2 py-0.5 text-[10px] bg-yellow-600 text-white rounded hover:bg-yellow-700"
                  >
                    Post to Bill
                  </button>
                )}
              </div>
            ))}
          </div>
          <p className="text-xs text-yellow-600 mt-2 ml-6">Unbilled laundry charges will be lost if you proceed without posting.</p>
        </div>
      )}

      {/* Collect Payment Section */}
      {summary && Number(summary.balance) + lateCharge * 1.18 > 0 && !paymentRecorded && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
          {!showPayment ? (
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-green-800">Collect Payment</p>
                <p className="text-xs text-green-600">Balance due: {fmt(Number(summary.balance) + lateCharge * 1.18)}</p>
              </div>
              <button onClick={() => { setShowPayment(true); setPaymentAmount(Math.max(0, Number(summary.balance) + lateCharge * 1.18)); }} className="px-3 py-1.5 text-sm bg-green-600 text-white rounded-md hover:bg-green-700">
                Record Payment
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-green-800">Record Payment</h4>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-gray-600">Amount (₹)</label>
                  <input type="number" min={0} value={paymentAmount || ""} onChange={(e) => setPaymentAmount(Number(e.target.value))} className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
                </div>
                <div>
                  <label className="block text-xs text-gray-600">Method</label>
                  <select value={paymentMethod} onChange={(e) => setPaymentMethod(e.target.value)} className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors">
                    <option value="cash">Cash</option>
                    <option value="upi">UPI</option>
                    <option value="card">Card</option>
                    <option value="bank_transfer">Bank Transfer</option>
                  </select>
                </div>
              </div>
              {paymentMethod !== "cash" && (
                <div>
                  <label className="block text-xs text-gray-600">Reference / Transaction ID</label>
                  <input type="text" value={paymentReference} onChange={(e) => setPaymentReference(e.target.value)} className="mt-1 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" placeholder="UPI ref / Card last 4" />
                </div>
              )}
              <div className="flex gap-2 justify-end">
                <button onClick={() => setShowPayment(false)} className="px-3 py-1.5 text-xs border rounded-md">Skip</button>
                <button onClick={handleRecordPayment} disabled={loading || paymentAmount <= 0} className="px-3 py-1.5 text-xs bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50">
                  {loading ? "Recording..." : "Confirm Payment"}
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {paymentRecorded && (
        <div className="p-2 bg-green-50 border border-green-200 rounded text-sm text-green-700 flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
          Payment recorded successfully
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-gray-700">Notes</label>
        <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} placeholder="Remarks..." className="mt-1 block w-full rounded border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm" />
      </div>

      {showConfirm ? (
        <div className="p-4 bg-orange-50 border border-orange-200 rounded-lg">
          <p className="text-sm font-medium text-orange-800">Confirm checkout for {guestName}?</p>
          {summary && Number(summary.balance) + lateCharge * 1.18 > 0 && !paymentRecorded && (<p className="text-sm text-red-600 mt-1">Outstanding: {fmt(Number(summary.balance) + lateCharge * 1.18)}</p>)}
          {pendingLaundry.length > 0 && <p className="text-sm text-yellow-600 mt-1">⚠️ {pendingLaundry.length} laundry order(s) still pending</p>}
          <div className="flex gap-2 mt-3">
            <button onClick={() => setShowConfirm(false)} className="px-3 py-1.5 text-sm border rounded hover:bg-white">Go Back</button>
            <button onClick={handleSubmit} disabled={loading} className="px-3 py-1.5 text-sm bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50">{loading ? "Processing..." : "Confirm Check-Out"}</button>
          </div>
        </div>
      ) : (
        <div className="flex gap-2 justify-end">
          <button type="button" onClick={onCancel} className="px-4 py-2 text-sm border rounded hover:bg-gray-100">Cancel</button>
          <button type="button" onClick={() => setShowConfirm(true)} className="px-4 py-2 text-sm bg-orange-600 text-white rounded hover:bg-orange-700">Proceed to Check-Out</button>
        </div>
      )}
    </div>
  );
}
