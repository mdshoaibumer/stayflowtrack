"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";

interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
  tenant_id: string;
  property_id?: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  requestPasswordReset: (email: string) => Promise<void>;
}

interface RegisterData {
  email: string;
  password: string;
  full_name: string;
  property_name: string;
  phone?: string;
}

const AuthContext = createContext<AuthContextType | null>(null);

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Secure token storage key prefix
const STORAGE_KEY_PREFIX = "sf_";

/**
 * Safe storage helpers that handle SSR and storage errors gracefully.
 * Tokens are stored in sessionStorage (cleared on tab close) instead of localStorage
 * to reduce XSS persistence window.
 */
function getStorageItem(key: string): string | null {
  try {
    if (typeof window === "undefined") return null;
    return sessionStorage.getItem(`${STORAGE_KEY_PREFIX}${key}`);
  } catch {
    return null;
  }
}

function setStorageItem(key: string, value: string): void {
  try {
    if (typeof window === "undefined") return;
    sessionStorage.setItem(`${STORAGE_KEY_PREFIX}${key}`, value);
  } catch {
    // Storage quota exceeded or blocked — fail silently
  }
}

function removeStorageItem(key: string): void {
  try {
    if (typeof window === "undefined") return;
    sessionStorage.removeItem(`${STORAGE_KEY_PREFIX}${key}`);
  } catch {
    // Fail silently
  }
}

function clearStorage(): void {
  removeStorageItem("at");
  removeStorageItem("rt");
  removeStorageItem("user");
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [state, setState] = useState<AuthState>({
    user: null,
    accessToken: null,
    isAuthenticated: false,
    isLoading: true,
  });
  const logoutRef = useRef<() => void>(() => {});

  // Keep logoutRef current to avoid stale closures in interval
  const logout = useCallback(() => {
    const token = getStorageItem("at");
    if (token) {
      fetch(`${API_BASE}/api/v1/auth/logout`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {});
    }
    clearStorage();
    setState({ user: null, accessToken: null, isAuthenticated: false, isLoading: false });
    router.push("/login");
  }, [router]);

  logoutRef.current = logout;

  // Restore session on mount
  useEffect(() => {
    const token = getStorageItem("at");
    const userStr = getStorageItem("user");
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        setState({ user, accessToken: token, isAuthenticated: true, isLoading: false });
      } catch {
        clearStorage();
        setState((s) => ({ ...s, isLoading: false }));
      }
    } else {
      setState((s) => ({ ...s, isLoading: false }));
    }
  }, []);

  // Tab synchronization — logout if another tab clears the session
  useEffect(() => {
    function handleStorageChange(e: StorageEvent) {
      // sessionStorage doesn't fire storage events across tabs,
      // but we broadcast logout via a localStorage signal
      if (e.key === `${STORAGE_KEY_PREFIX}logout_signal` && e.newValue) {
        clearStorage();
        setState({ user: null, accessToken: null, isAuthenticated: false, isLoading: false });
        router.push("/login");
      }
    }
    window.addEventListener("storage", handleStorageChange);
    return () => window.removeEventListener("storage", handleStorageChange);
  }, [router]);

  // Auto-refresh token with jitter to prevent thundering herd
  useEffect(() => {
    if (!state.isAuthenticated) return;

    let timeoutId: ReturnType<typeof setTimeout>;
    let failCount = 0;
    const MAX_FAILURES = 3;

    const scheduleRefresh = () => {
      // Base: 4 minutes, jitter: 0-30 seconds
      const baseMs = 4 * 60 * 1000;
      const jitterMs = Math.random() * 30_000;
      // Exponential backoff on failures
      const backoffMs = failCount > 0 ? Math.min(failCount * 10_000, 60_000) : 0;
      const delay = baseMs + jitterMs + backoffMs;

      timeoutId = setTimeout(async () => {
        try {
          const refreshToken = getStorageItem("rt");
          if (!refreshToken) {
            logoutRef.current();
            return;
          }
          const resp = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: refreshToken }),
          });
          if (resp.ok) {
            const data = await resp.json();
            setStorageItem("at", data.data.access_token);
            if (data.data.refresh_token) {
              setStorageItem("rt", data.data.refresh_token);
            }
            setState((s) => ({ ...s, accessToken: data.data.access_token }));
            failCount = 0;
          } else {
            failCount++;
            if (failCount >= MAX_FAILURES) {
              logoutRef.current();
              return;
            }
          }
        } catch {
          failCount++;
          if (failCount >= MAX_FAILURES) {
            logoutRef.current();
            return;
          }
        }
        scheduleRefresh();
      }, delay);
    };

    scheduleRefresh();
    return () => clearTimeout(timeoutId);
  }, [state.isAuthenticated]);

  // Automatically fetch and set default property if none is assigned (e.g. for super_admin owners)
  useEffect(() => {
    if (!state.isAuthenticated || !state.accessToken || state.user?.property_id) return;

    let active = true;
    const fetchDefaultProperty = async () => {
      try {
        const resp = await fetch(`${API_BASE}/api/v1/properties`, {
          headers: { Authorization: `Bearer ${state.accessToken}` },
        });
        if (resp.ok && active) {
          const data = await resp.json();
          const properties = data.data || [];
          if (properties.length > 0) {
            const updatedUser = { ...state.user!, property_id: properties[0].id };
            setStorageItem("user", JSON.stringify(updatedUser));
            setState((s) => ({ ...s, user: updatedUser }));
          }
        }
      } catch (err) {
        console.error("Failed to fetch default property:", err);
      }
    };

    fetchDefaultProperty();
    return () => {
      active = false;
    };
  }, [state.isAuthenticated, state.accessToken, state.user?.property_id]);

  const login = useCallback(async (email: string, password: string) => {
    const resp = await fetch(`${API_BASE}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data.error?.message || "Login failed");
    }
    const { access_token, refresh_token, user } = data.data;
    setStorageItem("at", access_token);
    setStorageItem("rt", refresh_token);
    setStorageItem("user", JSON.stringify(user));
    setState({ user, accessToken: access_token, isAuthenticated: true, isLoading: false });
    router.push("/dashboard");
  }, [router]);

  const register = useCallback(async (regData: RegisterData) => {
    const resp = await fetch(`${API_BASE}/api/v1/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(regData),
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data.error?.message || "Registration failed");
    }

    // Clear demo seeder flags on fresh registration
    localStorage.removeItem("demo_data_shown");
    localStorage.removeItem("demo_data_loaded");

    // Auto-login after register
    if (data.data?.access_token) {
      const { access_token, refresh_token, user } = data.data;
      setStorageItem("at", access_token);
      setStorageItem("rt", refresh_token);
      setStorageItem("user", JSON.stringify(user));
      setState({ user, accessToken: access_token, isAuthenticated: true, isLoading: false });
      router.push("/dashboard");
    } else {
      router.push("/login?registered=true");
    }
  }, [router]);

  const requestPasswordReset = useCallback(async (email: string) => {
    const resp = await fetch(`${API_BASE}/api/v1/auth/password-reset/request`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    });
    if (!resp.ok) {
      const data = await resp.json();
      throw new Error(data.error?.message || "Request failed");
    }
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout, requestPasswordReset }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
