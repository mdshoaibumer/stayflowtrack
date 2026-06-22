"use client";

import React, { useState, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";

export default function SettingsPage() {
  useAuth(); // ensure authenticated
  useApi(); // init api context
  const [tab, setTab] = useState("property");
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const tabs = [
    { id: "property", label: "Property" },
    { id: "profile", label: "Profile" },
    { id: "password", label: "Password" },
    { id: "team", label: "Team" },
  ];

  return (
    <div className="page-container">
      <h1 className="page-title">Settings</h1>

      {successMsg && (
        <div className="p-3 rounded-lg bg-emerald-50 border border-emerald-200 text-sm text-emerald-700" role="alert">{successMsg}</div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
              tab === t.id ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300"
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "property" && <PropertySettings onSuccess={(msg) => { setSuccessMsg(msg); setTimeout(() => setSuccessMsg(null), 3000); }} />}
      {tab === "profile" && <ProfileSettings onSuccess={(msg) => { setSuccessMsg(msg); setTimeout(() => setSuccessMsg(null), 3000); }} />}
      {tab === "password" && <PasswordSettings onSuccess={(msg) => { setSuccessMsg(msg); setTimeout(() => setSuccessMsg(null), 3000); }} />}
      {tab === "team" && <TeamSettings />}
    </div>
  );
}

function PropertySettings({ onSuccess }: { onSuccess: (msg: string) => void }) {
  const { user } = useAuth();
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState({
    name: "",
    address: "",
    city: "",
    state: "",
    pincode: "",
    gstin: "",
    phone: "",
    email: "",
  });

  useEffect(() => {
    if (!user?.property_id) return;
    api.get<Record<string, string>>(`/api/v1/properties/${user.property_id}`).then((data) => {
      if (data) setForm((prev) => ({ ...prev, ...data }));
    }).catch(() => {});
  }, [api, user?.property_id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user?.property_id) return;
    setLoading(true);
    setError(null);
    try {
      await api.put(`/api/v1/properties/${user.property_id}`, form);
      onSuccess("Property settings saved");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setLoading(false);
    }
  };

  const update = (field: string, value: string) => setForm((f) => ({ ...f, [field]: value }));

  return (
    <form onSubmit={handleSubmit} className="max-w-lg space-y-4 rounded-xl border bg-card p-6">
      <h3 className="font-medium text-gray-900">Property Information</h3>
      {error && <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">{error}</div>}
      <div>
        <label className="block text-sm font-medium text-gray-700">Property Name</label>
        <input type="text" value={form.name} onChange={(e) => update("name", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Address</label>
        <textarea value={form.address} onChange={(e) => update("address", e.target.value)} rows={2} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">City</label>
          <input type="text" value={form.city} onChange={(e) => update("city", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">State</label>
          <input type="text" value={form.state} onChange={(e) => update("state", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Pincode</label>
          <input type="text" value={form.pincode} onChange={(e) => update("pincode", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">GSTIN</label>
          <input type="text" value={form.gstin} onChange={(e) => update("gstin", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" placeholder="22AAAAA0000A1Z5" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Phone</label>
          <input type="tel" value={form.phone} onChange={(e) => update("phone", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">Email</label>
          <input type="email" value={form.email} onChange={(e) => update("email", e.target.value)} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
        </div>
      </div>
      <button type="submit" disabled={loading} className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 shadow-sm transition-colors">
        {loading ? "Saving..." : "Save Changes"}
      </button>
    </form>
  );
}

function ProfileSettings({ onSuccess: _onSuccess }: { onSuccess: (msg: string) => void }) {
  const { user } = useAuth();

  return (
    <div className="max-w-lg rounded-xl border bg-card p-6 space-y-4">
      <h3 className="font-medium text-gray-900">Profile</h3>
      <div className="space-y-3">
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Name</span>
          <span className="font-medium">{user?.full_name}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Email</span>
          <span className="font-medium">{user?.email}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Role</span>
          <span className="font-medium capitalize">{user?.role}</span>
        </div>
      </div>
    </div>
  );
}

function PasswordSettings({ onSuccess }: { onSuccess: (msg: string) => void }) {
  const api = useApi();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState({ current_password: "", new_password: "", confirm_password: "" });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (form.new_password !== form.confirm_password) { setError("Passwords don't match"); return; }
    if (form.new_password.length < 8) { setError("Min 8 characters"); return; }
    setLoading(true);
    setError(null);
    try {
      await api.post("/api/v1/auth/change-password", {
        current_password: form.current_password,
        new_password: form.new_password,
      });
      setForm({ current_password: "", new_password: "", confirm_password: "" });
      onSuccess("Password changed successfully");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to change password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="max-w-lg rounded-xl border bg-card p-6 space-y-4">
      <h3 className="font-medium text-gray-900">Change Password</h3>
      {error && <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">{error}</div>}
      <div>
        <label className="block text-sm font-medium text-gray-700">Current Password</label>
        <input type="password" required value={form.current_password} onChange={(e) => setForm({ ...form, current_password: e.target.value })} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">New Password</label>
        <input type="password" required value={form.new_password} onChange={(e) => setForm({ ...form, new_password: e.target.value })} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" placeholder="Min 8 characters" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Confirm New Password</label>
        <input type="password" required value={form.confirm_password} onChange={(e) => setForm({ ...form, confirm_password: e.target.value })} className="mt-1 block w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/20 focus:border-ring transition-colors" />
      </div>
      <button type="submit" disabled={loading} className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 shadow-sm transition-colors">
        {loading ? "Changing..." : "Change Password"}
      </button>
    </form>
  );
}

function TeamSettings() {
  const api = useApi();
  interface TeamMember { id?: string; email: string; full_name?: string; role?: string; }
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [showInvite, setShowInvite] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("receptionist");
  const [inviteLoading, setInviteLoading] = useState(false);

  useEffect(() => {
    api.get<TeamMember[] | { data: TeamMember[] }>("/api/v1/users", {}).then((data) => {
      setMembers(Array.isArray(data) ? data : (data as { data: TeamMember[] })?.data || []);
    }).catch(() => {}).finally(() => setLoading(false));
  }, [api]);

  const handleInvite = async () => {
    if (!inviteEmail) return;
    setInviteLoading(true);
    try {
      await api.post("/api/v1/users", { email: inviteEmail, role_name: inviteRole });
      setShowInvite(false);
      setInviteEmail("");
      // Refresh
      const data = await api.get<TeamMember[] | { data: TeamMember[] }>("/api/v1/users", {});
      setMembers(Array.isArray(data) ? data : (data as { data: TeamMember[] })?.data || []);
    } catch { /* silent */ }
    finally { setInviteLoading(false); }
  };

  return (
    <div className="max-w-lg rounded-xl border bg-card p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-gray-900">Team Members</h3>
        <button onClick={() => setShowInvite(true)} className="text-sm text-blue-600 hover:underline">+ Invite</button>
      </div>

      {loading && <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600 mx-auto"></div>}

      {!loading && members.length === 0 && <p className="text-sm text-gray-400">No team members</p>}
      {!loading && members.map((m) => (
        <div key={m.id || m.email} className="flex items-center justify-between py-2 border-b last:border-0">
          <div>
            <p className="text-sm font-medium">{m.full_name || m.email}</p>
            <p className="text-xs text-gray-500">{m.email}</p>
          </div>
          <span className="text-xs px-2 py-0.5 bg-gray-100 rounded capitalize">{m.role}</span>
        </div>
      ))}

      {showInvite && (
        <div className="border-t pt-4 space-y-3">
          <input type="email" placeholder="Email address" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          <select value={inviteRole} onChange={(e) => setInviteRole(e.target.value)} className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="property_admin">Property Admin</option>
            <option value="receptionist">Receptionist</option>
            <option value="housekeeping">Housekeeping</option>
          </select>
          <div className="flex gap-2">
            <button onClick={() => setShowInvite(false)} className="px-3 py-1.5 text-sm border rounded-md">Cancel</button>
            <button onClick={handleInvite} disabled={inviteLoading} className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md disabled:opacity-50">
              {inviteLoading ? "Inviting..." : "Send Invite"}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
