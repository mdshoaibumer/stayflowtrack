"use client";

import { useAuth } from "@/contexts/AuthContext";
import { useCallback, useMemo } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface RequestOptions {
  method?: string;
  body?: unknown;
  params?: Record<string, string>;
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
        throw new Error(data.error?.message || `Request failed (${resp.status})`);
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
      throw new Error(data.error?.message || `Upload failed (${resp.status})`);
    }
    return data.data ?? data;
  }, [accessToken, logout]);

  return useMemo(
    () => ({ request, get, post, put, patch, del, upload }),
    [request, get, post, put, patch, del, upload]
  );
}
