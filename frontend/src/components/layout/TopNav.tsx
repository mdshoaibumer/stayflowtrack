"use client";

import { useAuth } from "@/contexts/AuthContext";
import { useApi } from "@/hooks/useApi";
import { useState, useRef, useEffect } from "react";

interface TopNavProps {
  onMenuToggle: () => void;
}

interface Notification {
  id: string;
  type: string;
  message: string;
  time: string;
  read: boolean;
}

export default function TopNav({ onMenuToggle }: TopNavProps) {
  const { user, logout } = useAuth();
  const api = useApi();
  const [showMenu, setShowMenu] = useState(false);
  const [showNotifs, setShowNotifs] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const menuRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowMenu(false);
      }
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setShowNotifs(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  // Fetch live notifications (overdue checkouts, pending tasks, etc.)
  useEffect(() => {
    if (!user?.property_id) return;
    const propertyId = user.property_id;
    interface ReservationEntry { reservation_id?: string; id?: string; status: string; check_in_date: string; check_out_date: string; guest_name: string; unit_number: string; }
    const fetchNotifs = async () => {
      try {
        const data = await api.get<ReservationEntry[] | { data: ReservationEntry[] }>("/api/v1/reservations", { property_id: propertyId, per_page: "100" });
        const reservations: ReservationEntry[] = Array.isArray(data) ? data : (data as { data: ReservationEntry[] })?.data || [];
        const today = new Date().toISOString().split("T")[0];
        const now = new Date();
        const notifs: Notification[] = [];

        // Overdue checkouts
        const overdue = reservations.filter((r) => r.status === "checked_in" && r.check_out_date <= today);
        overdue.forEach((r) => {
          notifs.push({
            id: `overdue-${r.reservation_id || r.id}`,
            type: "warning",
            message: `Overdue: ${r.guest_name} in ${r.unit_number} (was due ${r.check_out_date})`,
            time: "Now",
            read: false,
          });
        });

        // Expected arrivals not yet checked in
        const pendingArrivals = reservations.filter((r) => r.check_in_date === today && (r.status === "confirmed" || r.status === "pending"));
        if (pendingArrivals.length > 0 && now.getHours() >= 14) {
          notifs.push({
            id: "arrivals-pending",
            type: "info",
            message: `${pendingArrivals.length} guest(s) expected today haven't checked in yet`,
            time: "Today",
            read: false,
          });
        }

        // Tomorrow arrivals reminder
        const tomorrow = new Date(Date.now() + 86400000).toISOString().split("T")[0];
        const tomorrowArrivals = reservations.filter((r) => r.check_in_date === tomorrow && (r.status === "confirmed" || r.status === "pending"));
        if (tomorrowArrivals.length > 0) {
          notifs.push({
            id: "tomorrow-arrivals",
            type: "info",
            message: `${tomorrowArrivals.length} arrival(s) tomorrow — ensure rooms are ready`,
            time: "Upcoming",
            read: false,
          });
        }

        setNotifications(notifs);
      } catch { /* silent */ }
    };
    fetchNotifs();
    const interval = setInterval(fetchNotifs, 60000);
    return () => clearInterval(interval);
  }, [api, user?.property_id]);

  const unreadCount = notifications.filter((n) => !n.read).length;

  return (
    <header className="h-14 border-b border-border bg-card flex items-center justify-between px-4 lg:px-6">
      {/* Mobile menu button */}
      <button
        onClick={onMenuToggle}
        className="lg:hidden text-muted-foreground hover:text-foreground transition-colors"
        aria-label="Toggle menu"
      >
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      {/* Page context - shows property name on desktop */}
      <div className="hidden lg:block">
        <p className="text-sm text-muted-foreground">Property Management</p>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-3">
        {/* Command palette trigger */}
        <button
          onClick={() => {
            document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, ctrlKey: true, bubbles: true }));
          }}
          className="hidden sm:inline-flex items-center gap-2 px-3 py-1.5 text-xs text-muted-foreground bg-muted/50 border rounded-lg hover:bg-muted transition-colors"
          aria-label="Open command palette"
        >
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span>Search...</span>
          <kbd className="ml-1 px-1 py-0.5 text-[10px] font-medium bg-background border rounded">{typeof navigator !== 'undefined' && /Mac/.test(navigator.userAgent) ? '⌘K' : 'Ctrl+K'}</kbd>
        </button>

        {/* Notifications */}
        <div className="relative" ref={notifRef}>
          <button
            onClick={() => setShowNotifs(!showNotifs)}
            className="relative text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Notifications"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 w-4 h-4 bg-destructive text-destructive-foreground text-[10px] font-bold rounded-full flex items-center justify-center">
                {unreadCount > 9 ? "9+" : unreadCount}
              </span>
            )}
          </button>

          {showNotifs && (
            <div className="absolute right-0 mt-2 w-80 bg-card border rounded-xl shadow-lg z-50 overflow-hidden animate-scale-in">
              <div className="px-4 py-3 border-b">
                <h3 className="text-sm font-semibold text-foreground">Notifications</h3>
              </div>
              <div className="max-h-80 overflow-y-auto">
                {notifications.length === 0 ? (
                  <div className="px-4 py-6 text-center text-sm text-muted-foreground">
                    All clear! No notifications.
                  </div>
                ) : (
                  notifications.map((n) => (
                    <div
                      key={n.id}
                      className={`px-4 py-3 border-b last:border-0 ${!n.read ? "bg-primary/5" : ""}`}
                    >
                      <div className="flex items-start gap-2">
                        <div className={`mt-0.5 w-2 h-2 rounded-full shrink-0 ${n.type === "warning" ? "bg-destructive" : "bg-primary"}`} aria-hidden="true" />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-foreground">{n.message}</p>
                          <p className="text-xs text-muted-foreground mt-0.5">{n.time}</p>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
              {notifications.length > 0 && (
                <div className="px-4 py-2 border-t">
                  <button
                    onClick={() => { setNotifications(notifications.map((n) => ({ ...n, read: true }))); }}
                    className="text-xs text-primary hover:underline"
                  >
                    Mark all as read
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* User menu */}
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setShowMenu(!showMenu)}
            className="flex items-center gap-2 text-sm hover:bg-gray-50 rounded-lg px-2 py-1 transition-colors"
            aria-label="User menu"
            aria-expanded={showMenu}
          >
            <div className="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium text-primary">
              {user?.full_name?.charAt(0)?.toUpperCase() || "U"}
            </div>
            <span className="hidden sm:block text-foreground font-medium max-w-[120px] truncate">
              {user?.full_name}
            </span>
            <svg className="w-4 h-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          {showMenu && (
            <div className="absolute right-0 mt-2 w-48 bg-card border rounded-xl shadow-lg py-1 z-50 overflow-hidden animate-scale-in">
              <div className="px-4 py-2 border-b">
                <p className="text-sm font-medium text-foreground">{user?.full_name}</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
              <a
                href="/settings"
                className="block px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                onClick={() => setShowMenu(false)}
              >
                Settings
              </a>
              <button
                onClick={() => {
                  setShowMenu(false);
                  logout();
                }}
                className="w-full text-left px-4 py-2 text-sm text-destructive hover:bg-destructive/5 transition-colors"
              >
                Sign Out
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
