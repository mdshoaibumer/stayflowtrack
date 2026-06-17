/**
 * E2E Test: Dashboard Extended - Daily Collection, Outstanding Dues, End of Day, Close Day
 *
 * Tests: daily-collection, outstanding-dues, end-of-day summary, close-day
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Dashboard Extended", () => {
  let authToken: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;

  test.beforeAll(async () => {
    const user = generateTestUser("dash-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    demoData = await loadMinimalDemoData(authToken);
  });

  test("should get dashboard summary", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should get revenue trend", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}/revenue-trend`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should get daily collection", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}/daily-collection`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should get outstanding dues", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}/outstanding-dues`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should get end-of-day summary", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}/end-of-day`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should close day (night audit)", async ({ request }) => {
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/dashboard/${demoData.property.id}/close-day`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          date: new Date().toISOString().split("T")[0],
          notes: "E2E night audit close",
        },
      }
    );
    // Accept success or "already closed" error
    expect([200, 201, 400, 409].includes(resp.status())).toBeTruthy();
  });
});
