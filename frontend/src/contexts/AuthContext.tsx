"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
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

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [state, setState] = useState<AuthState>({
    user: null,
    accessToken: null,
    isAuthenticated: false,
    isLoading: true,
  });

  // Restore session on mount
  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const userStr = localStorage.getItem("user");
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        setState({ user, accessToken: token, isAuthenticated: true, isLoading: false });
      } catch {
        localStorage.clear();
        setState((s) => ({ ...s, isLoading: false }));
      }
    } else {
      setState((s) => ({ ...s, isLoading: false }));
    }
  }, []);

  // Auto-refresh token
  useEffect(() => {
    if (!state.isAuthenticated) return;
    const interval = setInterval(async () => {
      try {
        const refreshToken = localStorage.getItem("refresh_token");
        if (!refreshToken) return;
        const resp = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (resp.ok) {
          const data = await resp.json();
          localStorage.setItem("access_token", data.data.access_token);
          if (data.data.refresh_token) {
            localStorage.setItem("refresh_token", data.data.refresh_token);
          }
          setState((s) => ({ ...s, accessToken: data.data.access_token }));
        } else {
          logout();
        }
      } catch {
        // Silent fail — will retry next interval
      }
    }, 4 * 60 * 1000); // Refresh every 4 minutes
    return () => clearInterval(interval);
  }, [state.isAuthenticated]);

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
    localStorage.setItem("access_token", access_token);
    localStorage.setItem("refresh_token", refresh_token);
    localStorage.setItem("user", JSON.stringify(user));
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
    // Auto-login after register
    if (data.data?.access_token) {
      const { access_token, refresh_token, user } = data.data;
      localStorage.setItem("access_token", access_token);
      localStorage.setItem("refresh_token", refresh_token);
      localStorage.setItem("user", JSON.stringify(user));
      setState({ user, accessToken: access_token, isAuthenticated: true, isLoading: false });
      router.push("/dashboard");
    } else {
      router.push("/login?registered=true");
    }
  }, [router]);

  const logout = useCallback(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      fetch(`${API_BASE}/api/v1/auth/logout`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {});
    }
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
    setState({ user: null, accessToken: null, isAuthenticated: false, isLoading: false });
    router.push("/login");
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
