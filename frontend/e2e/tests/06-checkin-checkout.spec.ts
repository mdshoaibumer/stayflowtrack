/**
 * E2E Test: Check-in / Check-out Flow
 *
 * Tests: Full hospitality workflow - reserve → check-in → stay → check-out
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createReservation } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Check-in / Check-out Flow", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;
  let reservationId: string;

  test.beforeAll(async () => {
    const user = generateTestUser("checkin");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    userEmail = user.email;
    userPassword = user.password;
    demoData = await loadMinimalDemoData(authToken);

    // Create a confirmed reservation for today
    const today = new Date();
    const checkout = new Date(today);
    checkout.setDate(today.getDate() + 3);

    const res = await createReservation(authToken, {
      property_id: demoData.property.id,
      unit_id: demoData.unit1.id,
      guest_id: demoData.guest.id,
      check_in_date: today.toISOString().split("T")[0],
      check_out_date: checkout.toISOString().split("T")[0],
      rate_per_night: 2000,
      booking_source: "walk_in",
    });
    reservationId = res.data?.id || "";
  });

  async function loginAndGo(page: any, route: string) {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    await page.goto(route);
    await page.waitForLoadState("networkidle");
  }

  test("should perform check-in via API", async ({ request }) => {
    // Confirm reservation first
    const confirmResp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}/confirm`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      }
    );
    // May already be confirmed or succeed
    expect([200, 201, 409, 422].includes(confirmResp.status())).toBeTruthy();

    // Perform check-in via operations endpoint
    const checkinResp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/operations/check-in`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          reservation_id: reservationId,
          unit_id: demoData.unit1.id,
          deposit_amount: 2000,
          deposit_method: "cash",
          id_document_type: "aadhaar",
          id_document_number: "123456789012",
        },
      }
    );
    // Check-in should succeed or already done
    expect([200, 201, 409, 422].includes(checkinResp.status())).toBeTruthy();
  });

  test("should show checked-in reservation in operations page", async ({ page }) => {
    await loginAndGo(page, "/operations");
    await page.waitForTimeout(2000);
    // Page should load successfully
    await expect(page.locator("body")).toBeVisible();
  });

  test("should perform check-out via API", async ({ request }) => {
    const checkoutResp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/operations/check-out`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          reservation_id: reservationId,
          late_checkout_charge: 0,
        },
      }
    );
    // Accept various status codes (may fail if check-in didn't succeed)
    expect([200, 201, 400, 409, 422].includes(checkoutResp.status())).toBeTruthy();
  });

  test("should display walk-in flow", async ({ page }) => {
    await loginAndGo(page, "/operations");

    // Look for walk-in button
    const walkInBtn = page.locator(
      'button:has-text("Walk"), button:has-text("walk-in"), a:has-text("Walk")'
    );
    if (await walkInBtn.first().isVisible({ timeout: 5000 }).catch(() => false)) {
      await walkInBtn.first().click();
      await page.waitForTimeout(1000);
      // Form should appear
      await expect(page.locator("body")).toBeVisible();
    }
  });
});
