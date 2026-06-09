/**
 * E2E Test Configuration
 * Central configuration for all test constants and URLs
 */

export const TEST_CONFIG = {
  API_URL: process.env.E2E_API_URL || "http://localhost:8080",
  BASE_URL: process.env.E2E_BASE_URL || "http://localhost:3000",

  // Test user credentials (generated fresh per test run)
  TEST_USER: {
    full_name: "E2E Test Admin",
    email: `e2e-${Date.now()}@stayflow-test.com`,
    password: "TestPass@2026!",
    property_name: "E2E Test Hotel",
    phone: "+919876543210",
  },

  // Timeouts
  NAVIGATION_TIMEOUT: 15_000,
  API_TIMEOUT: 10_000,

  // Routes
  ROUTES: {
    LOGIN: "/login",
    REGISTER: "/register",
    DASHBOARD: "/dashboard",
    RESERVATIONS: "/reservations",
    GUESTS: "/guests",
    BILLING: "/billing",
    HOUSEKEEPING: "/housekeeping",
    LAUNDRY: "/laundry",
    OPERATIONS: "/operations",
    SETTINGS: "/settings",
  },
} as const;

/**
 * Generate unique test user for each test run to avoid conflicts
 */
export function generateTestUser(prefix = "e2e") {
  const ts = Date.now();
  return {
    full_name: `${prefix} User ${ts}`,
    email: `${prefix}-${ts}@stayflow-test.com`,
    password: "TestPass@2026!",
    property_name: `${prefix} Hotel ${ts}`,
    phone: "+919876543210",
  };
}
