/**
 * API Helpers for E2E tests
 * Direct API calls for setup/teardown and verification
 */

import { TEST_CONFIG } from "./test-config";

interface ApiResponse<T = unknown> {
  data: T;
  error?: { message: string; code: string };
}

async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const resp = await fetch(`${TEST_CONFIG.API_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  return resp.json();
}

/**
 * Register a new tenant and get tokens
 */
export async function registerTenant(data: {
  full_name: string;
  email: string;
  password: string;
  property_name: string;
  phone?: string;
}) {
  return apiCall<{
    tenant: { id: string; name: string; slug: string };
    user: { id: string; email: string; role_name: string; tenant_id: string };
    access_token: string;
    refresh_token: string;
  }>("POST", "/api/v1/auth/register", data);
}

/**
 * Login and get tokens
 */
export async function login(email: string, password: string) {
  return apiCall<{
    access_token: string;
    refresh_token: string;
    user: { id: string; email: string; role_name: string; tenant_id: string; property_id?: string };
  }>("POST", "/api/v1/auth/login", { email, password });
}

/**
 * Create a property
 */
export async function createProperty(
  token: string,
  data: { name: string; address?: string; city?: string; state?: string; country?: string }
) {
  return apiCall<{ id: string; name: string }>("POST", "/api/v1/properties", data, token);
}

/**
 * Create a unit type
 */
export async function createUnitType(
  token: string,
  propertyId: string,
  data: { name: string; base_rate: number; max_occupancy: number }
) {
  return apiCall<{ id: string; name: string }>(
    "POST",
    `/api/v1/properties/${propertyId}/unit-types`,
    data,
    token
  );
}

/**
 * Create a unit
 */
export async function createUnit(
  token: string,
  propertyId: string,
  data: { unit_number: string; floor: number; unit_type_id: string }
) {
  return apiCall<{ id: string; unit_number: string }>(
    "POST",
    `/api/v1/properties/${propertyId}/units`,
    data,
    token
  );
}

/**
 * Create a guest
 */
export async function createGuest(
  token: string,
  data: {
    first_name: string;
    last_name: string;
    email?: string;
    phone: string;
    address?: string;
    city?: string;
    state?: string;
    country?: string;
  }
) {
  return apiCall<{ id: string; first_name: string; last_name: string }>(
    "POST",
    "/api/v1/guests",
    data,
    token
  );
}

/**
 * Create a reservation
 */
export async function createReservation(
  token: string,
  data: {
    property_id: string;
    unit_id: string;
    guest_id: string;
    check_in_date: string;
    check_out_date: string;
    rate_per_night: number;
    booking_source?: string;
  }
) {
  return apiCall<{ id: string; status: string }>("POST", "/api/v1/reservations", data, token);
}

/**
 * Check health
 */
export async function checkHealth() {
  try {
    const resp = await fetch(`${TEST_CONFIG.API_URL}/health`);
    return resp.ok;
  } catch {
    return false;
  }
}
