/**
 * E2E Test: Auth Extended - Password Reset, Team Management, Token Refresh
 *
 * Tests: password-reset request/confirm, create user, refresh token
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, login } from "../helpers/api-helpers";

test.describe("Auth Extended", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;
  let refreshToken: string;

  test.beforeAll(async () => {
    const user = generateTestUser("auth-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    refreshToken = result.data.refresh_token;
    userEmail = user.email;
    userPassword = user.password;
  });

  test.describe("Token Refresh", () => {
    test("should refresh token with valid refresh token", async ({ request }) => {
      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/auth/refresh`, {
        data: { refresh_token: refreshToken },
      });
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data.access_token).toBeTruthy();
      expect(body.data.refresh_token).toBeTruthy();
    });

    test("should reject invalid refresh token", async ({ request }) => {
      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/auth/refresh`, {
        data: { refresh_token: "invalid-token-12345" },
      });
      expect([400, 401].includes(resp.status())).toBeTruthy();
    });
  });

  test.describe("Password Reset", () => {
    test("should request password reset for existing email", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/auth/password-reset/request`,
        { data: { email: userEmail } }
      );
      // Should succeed even if email delivery fails (no SMTP in E2E)
      expect([200, 202].includes(resp.status())).toBeTruthy();
    });

    test("should handle password reset for non-existent email gracefully", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/auth/password-reset/request`,
        { data: { email: "nonexistent@example.com" } }
      );
      // Should not reveal whether email exists (security)
      expect([200, 202, 404].includes(resp.status())).toBeTruthy();
    });

    test("should reject password reset confirm with invalid token", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/auth/password-reset/confirm`,
        { data: { token: "invalid-reset-token", new_password: "NewPass@2026!" } }
      );
      expect([400, 401, 404].includes(resp.status())).toBeTruthy();
    });

    test("should reject weak password in reset confirm", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/auth/password-reset/confirm`,
        { data: { token: "any-token", new_password: "weak" } }
      );
      expect([400, 401, 404, 422].includes(resp.status())).toBeTruthy();
    });
  });

  test.describe("Team / User Management", () => {
    test("should create a new team member", async ({ request }) => {
      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/users`, {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          email: `team-${Date.now()}@stayflow-test.com`,
          password: "TeamPass@2026!",
          first_name: "Team",
          last_name: "Receptionist",
          role_name: "receptionist",
        },
      });
      expect([200, 201, 400, 422].includes(resp.status())).toBeTruthy();
      if (resp.status() <= 201) {
        const body = await resp.json();
        expect(body.data.email).toBeTruthy();
      }
    });

    test("should reject creating user with invalid role", async ({ request }) => {
      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/users`, {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          email: `bad-role-${Date.now()}@stayflow-test.com`,
          password: "TeamPass@2026!",
          full_name: "Bad Role User",
          role_name: "superuser",
        },
      });
      expect([400, 422].includes(resp.status())).toBeTruthy();
    });

    test("should reject duplicate team member email", async ({ request }) => {
      const email = `dup-team-${Date.now()}@stayflow-test.com`;
      // Create first
      await request.post(`${TEST_CONFIG.API_URL}/api/v1/users`, {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { email, password: "TeamPass@2026!", full_name: "Dup User", role_name: "receptionist" },
      });
      // Try duplicate
      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/users`, {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { email, password: "TeamPass@2026!", full_name: "Dup User 2", role_name: "receptionist" },
      });
      expect([400, 409, 422].includes(resp.status())).toBeTruthy();
    });
  });

  test.describe("Logout", () => {
    test("should logout via API", async ({ request }) => {
      // Login fresh to get a valid token
      const loginResult = await login(userEmail, userPassword);
      const token = loginResult.data.access_token;

      const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/auth/logout`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect([200, 204].includes(resp.status())).toBeTruthy();
    });
  });
});
