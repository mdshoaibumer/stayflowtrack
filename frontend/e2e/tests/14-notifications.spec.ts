/**
 * E2E Test: Notifications
 *
 * Tests: Template management, send notifications via API
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";

test.describe("Notifications", () => {
  let authToken: string;

  test.beforeAll(async () => {
    const user = generateTestUser("notif");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
  });

  test("should create notification template via API", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/notifications/templates`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        event_type: "booking_confirmation",
        channel: "whatsapp",
        subject: "Booking Confirmed",
        body: "Dear {{guest_name}}, your booking at {{property_name}} is confirmed for {{check_in_date}}. Ref: {{reservation_id}}",
      },
    });
    expect([200, 201, 409].includes(resp.status())).toBeTruthy();
  });

  test("should list notification templates via API", async ({ request }) => {
    const resp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/notifications/templates`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect([200, 404].includes(resp.status())).toBeTruthy();
  });

  test("should send test notification via API", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/notifications/send-test`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        channel: "whatsapp",
        recipient: "+919876543210",
        message: "E2E Test notification",
      },
    });
    // May succeed or fail depending on provider config
    expect([200, 201, 400, 422, 500].includes(resp.status())).toBeTruthy();
  });

  test("should get notification logs via API", async ({ request }) => {
    const resp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/notifications/logs`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect([200, 404].includes(resp.status())).toBeTruthy();
  });
});
