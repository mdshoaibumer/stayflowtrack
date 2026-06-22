"use client";

import { useAuth } from "@/contexts/AuthContext";
import { useCallback, useMemo } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface RequestOptions {
  method?: string;
  body?: unknown;
  params?: Record<string, string>;
}

// Map of known error codes to user-friendly messages
const ERROR_MESSAGES: Record<string, string> = {
  UNAUTHORIZED: "Please log in again.",
  FORBIDDEN: "You don't have permission to perform this action.",
  NOT_FOUND: "The requested resource was not found.",
  CONFLICT: "This operation conflicts with existing data.",
  RATE_LIMITED: "Too many requests. Please wait and try again.",
  VALIDATION_FAILED: "Please check your input and try again.",
};

/**
 * Sanitize error messages from the API to prevent information leakage.
 * Only expose known safe messages; fallback to generic message for unknown errors.
 */
function sanitizeErrorMessage(data: { error?: { code?: string; message?: string } }, status: number): string {
  const code = data?.error?.code;
  if (code && ERROR_MESSAGES[code]) {
    return ERROR_MESSAGES[code];
  }
  // Only pass through validation messages (they're user-facing by design)
  if (status === 400 && data?.error?.message) {
    // Strip anything that looks like internal details (stack traces, SQL, etc.)
    const msg = data.error.message;
    if (msg.length < 200 && !msg.includes("sql") && !msg.includes("panic") && !msg.includes("runtime")) {
      return msg;
    }
  }
  // Generic messages by status code
  if (status === 404) return "Resource not found.";
  if (status === 409) return "This conflicts with existing data.";
  if (status === 422) return "Invalid input. Please check your data.";
  if (status === 429) return "Too many requests. Please slow down.";
  if (status >= 500) return "An unexpected error occurred. Please try again later.";
  return "Request failed. Please try again.";
}

export function useApi() {
  const { accessToken, logout } = useAuth();

  const request = useCallback(
    async <T = unknown>(path: string, options: RequestOptions = {}): Promise<T> => {
      const { method = "GET", body, params } = options;

      let url = `${API_BASE}${path}`;
      if (params) {
        const searchParams = new URLSearchParams(params);
        url += `?${searchParams.toString()}`;
      }

      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }

      const resp = await fetch(url, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });

      if (resp.status === 401) {
        logout();
        throw new Error("Session expired. Please login again.");
      }

      const data = await resp.json();

      if (!resp.ok) {
        throw new Error(sanitizeErrorMessage(data, resp.status));
      }

      return data.data ?? data;
    },
    [accessToken, logout]
  );

  const get = useCallback(<T = unknown>(path: string, params?: Record<string, string>) => {
    return request<T>(path, { params });
  }, [request]);

  const post = useCallback(<T = unknown>(path: string, body?: unknown) => {
    return request<T>(path, { method: "POST", body });
  }, [request]);

  const put = useCallback(<T = unknown>(path: string, body?: unknown) => {
    return request<T>(path, { method: "PUT", body });
  }, [request]);

  const patch = useCallback(<T = unknown>(path: string, body?: unknown) => {
    return request<T>(path, { method: "PATCH", body });
  }, [request]);

  const del = useCallback(<T = unknown>(path: string) => {
    return request<T>(path, { method: "DELETE" });
  }, [request]);

  const upload = useCallback(async <T = unknown>(path: string, formData: FormData): Promise<T> => {
    const url = `${API_BASE}${path}`;
    const headers: Record<string, string> = {};
    if (accessToken) {
      headers["Authorization"] = `Bearer ${accessToken}`;
    }
    const resp = await fetch(url, { method: "POST", headers, body: formData });
    if (resp.status === 401) {
      logout();
      throw new Error("Session expired. Please login again.");
    }
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(sanitizeErrorMessage(data, resp.status));
    }
    return data.data ?? data;
  }, [accessToken, logout]);

  return useMemo(
    () => ({ request, get, post, put, patch, del, upload }),
    [request, get, post, put, patch, del, upload]
  );
}
