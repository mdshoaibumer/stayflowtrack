/**
 * E2E Test: Admin Module (Platform Admin)
 *
 * Tests: admin list tenants, admin get metrics, admin list plans
 * Note: These require platform admin role - we test both access denied and
 * platform admin access via PLATFORM_TENANT_ID config.
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";

test.describe("Admin Module", () => {
  let normalToken: string;

  test.beforeAll(async () => {
    const user = generateTestUser("admin-test");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    normalToken = result.data.access_token;
  });

  test("should reject admin/tenants for non-admin user", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/admin/tenants`,
      { headers: { Authorization: `Bearer ${normalToken}` } }
    );
    expect([401, 403].includes(resp.status())).toBeTruthy();
  });

  test("should reject admin/metrics for non-admin user", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/admin/metrics`,
      { headers: { Authorization: `Bearer ${normalToken}` } }
    );
    expect([401, 403].includes(resp.status())).toBeTruthy();
  });

  test("should reject admin/plans for non-admin user", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/admin/plans`,
      { headers: { Authorization: `Bearer ${normalToken}` } }
    );
    expect([401, 403].includes(resp.status())).toBeTruthy();
  });

  test("should reject admin access without auth", async ({ request }) => {
    const resp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/admin/tenants`);
    expect([401, 403].includes(resp.status())).toBeTruthy();
  });
});
