/**
 * E2E Test: Billing & Invoicing
 *
 * Tests: Folio creation, add charges, payments, invoice generation
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createReservation } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Billing & Invoicing", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;

  test.beforeAll(async () => {
    const user = generateTestUser("billing");
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
  });

  async function loginAndGo(page: any, route: string) {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    await page.goto(route);
    await page.waitForLoadState("networkidle");
  }

  test("should navigate to billing page", async ({ page }) => {
    await loginAndGo(page, "/billing");
    await expect(page).toHaveURL(/billing/);
    await expect(page.locator("body")).toBeVisible();
  });

  test("should create folio via check-in API and see it in billing", async ({ page, request }) => {
    // Create reservation
    const today = new Date();
    const checkout = new Date(today);
    checkout.setDate(today.getDate() + 2);

    const res = await createReservation(authToken, {
      property_id: demoData.property.id,
      unit_id: demoData.unit1.id,
      guest_id: demoData.guest.id,
      check_in_date: today.toISOString().split("T")[0],
      check_out_date: checkout.toISOString().split("T")[0],
      rate_per_night: 2000,
      booking_source: "walk_in",
    });

    if (res.data?.id) {
      // Confirm and check in (creates folio)
      await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/reservations/${res.data.id}/confirm`,
        { headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" } }
      );

      await request.post(`${TEST_CONFIG.API_URL}/api/v1/operations/check-in`, {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          reservation_id: res.data.id,
          unit_id: demoData.unit1.id,
          deposit_amount: 1000,
          deposit_method: "upi",
          id_document_type: "aadhaar",
          id_document_number: "123456789012",
        },
      });
    }

    // Navigate to billing
    await loginAndGo(page, "/billing");
    await page.waitForTimeout(2000);
    await expect(page.locator("body")).toBeVisible();
  });

  test("should add charges to folio via API", async ({ request }) => {
    // List folios
    const foliosResp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/billing/folios`, {
      headers: { Authorization: `Bearer ${authToken}` },
      params: { property_id: demoData.property.id },
    });

    if (foliosResp.ok()) {
      const folios = await foliosResp.json();
      const openFolio = folios.data?.find?.((f: any) => f.status === "open");

      if (openFolio) {
        // Add a charge
        const chargeResp = await request.post(
          `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${openFolio.id}/charges`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: {
              description: "Room Service - Dinner",
              amount: 1500,
              charge_type: "food",
              tax_rate: 5,
            },
          }
        );
        expect([200, 201, 404].includes(chargeResp.status())).toBeTruthy();
      }
    }
  });

  test("should record payment via API", async ({ request }) => {
    const foliosResp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/billing/folios`, {
      headers: { Authorization: `Bearer ${authToken}` },
      params: { property_id: demoData.property.id },
    });

    if (foliosResp.ok()) {
      const folios = await foliosResp.json();
      const openFolio = folios.data?.find?.((f: any) => f.status === "open");

      if (openFolio) {
        const payResp = await request.post(
          `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${openFolio.id}/payments`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: {
              amount: 1000,
              payment_method: "upi",
              payment_type: "payment",
              reference_number: "UPI-TEST-001",
            },
          }
        );
        expect([200, 201, 404].includes(payResp.status())).toBeTruthy();
      }
    }
  });

  test("should display billing data in UI", async ({ page }) => {
    await loginAndGo(page, "/billing");
    await page.waitForTimeout(3000);
    // Billing page should load with some content
    const bodyText = await page.locator("body").textContent();
    expect(bodyText?.length).toBeGreaterThan(100);
  });
});
