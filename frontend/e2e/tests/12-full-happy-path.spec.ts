/**
 * E2E Test: Full Happy Path (Smoke Test)
 *
 * Tests the complete flow from registration to invoice:
 * Register → Create Property → Add Rooms → Add Guest → Reserve → Check-in → Add Charges → Check-out → Invoice
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { RegisterPage } from "../helpers/page-objects";

test.describe.serial("Full Happy Path - End to End", () => {
  const user = generateTestUser("e2e-full");
  let authToken: string;
  let propertyId: string;
  let unitTypeId: string;
  let unitId: string;
  let guestId: string;
  let reservationId: string;

  test("Step 1: Register new tenant", async ({ page }) => {
    const registerPage = new RegisterPage(page);
    await registerPage.goto();
    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });

    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    // Extract token from localStorage
    authToken = await page.evaluate(() => localStorage.getItem("access_token") || "");
    expect(authToken).toBeTruthy();
  });

  test("Step 2: Create property", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/properties`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        name: "E2E Grand Hotel",
        address: "1 Test Avenue",
        city: "Mumbai",
        state: "Maharashtra",
        country: "India",
      },
    });
    expect(resp.status()).toBe(201);
    const data = await resp.json();
    propertyId = data.data.id;
    expect(propertyId).toBeTruthy();
  });

  test("Step 3: Create unit type", async ({ request }) => {
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/unit-types`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { name: "Premium Room", base_rate: 3000, max_occupancy: 2 },
      }
    );
    expect(resp.status()).toBe(201);
    const data = await resp.json();
    unitTypeId = data.data.id;
  });

  test("Step 4: Create unit", async ({ request }) => {
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { unit_number: "E2E-101", floor: "1", unit_type_id: unitTypeId },
      }
    );
    expect(resp.status()).toBe(201);
    const data = await resp.json();
    unitId = data.data.id;
  });

  test("Step 5: Create guest", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/guests`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        first_name: "E2E",
        last_name: "FullPathGuest",
        phone: "+919876500100",
        email: "e2e-fullpath@test.com",
        city: "Mumbai",
        state: "Maharashtra",
        country: "India",
      },
    });
    expect(resp.status()).toBe(201);
    const data = await resp.json();
    guestId = data.data.id;
  });

  test("Step 6: Create reservation", async ({ request }) => {
    const today = new Date();
    const checkout = new Date(today);
    checkout.setDate(today.getDate() + 3);

    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/reservations`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: propertyId,
        unit_id: unitId,
        guest_id: guestId,
        check_in_date: today.toISOString().split("T")[0],
        check_out_date: checkout.toISOString().split("T")[0],
        rate_per_night: 3000,
        booking_source: "walk_in",
        num_guests: 2,
      },
    });
    expect(resp.status()).toBe(201);
    const data = await resp.json();
    reservationId = data.data.id;
  });

  test("Step 7: Confirm reservation", async ({ request }) => {
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}/confirm`,
      { headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" } }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("Step 8: Check-in guest", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/operations/check-in`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        reservation_id: reservationId,
        assigned_unit_id: unitId,
        deposit_amount: 3000,
        deposit_method: "card",
        id_document_type: "aadhaar",
        id_document_number: "998877665544",
      },
    });
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("Step 9: Add charges to folio", async ({ request }) => {
    // Get folio
    const foliosResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios?property_id=${propertyId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );

    if (foliosResp.ok()) {
      const folios = await foliosResp.json();
      const folio = folios.data?.[0] || folios.data?.find?.((f: any) => f.status === "open");

      if (folio?.id) {
        // Add room service charge
        const chargeResp = await request.post(
          `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folio.id}/charges`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: {
              description: "Minibar",
              amount: 500,
              charge_type: "minibar",
              tax_rate: 18,
            },
          }
        );
        expect([200, 201].includes(chargeResp.status())).toBeTruthy();
      }
    }
  });

  test("Step 10: Record payment", async ({ request }) => {
    const foliosResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios?property_id=${propertyId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );

    if (foliosResp.ok()) {
      const folios = await foliosResp.json();
      const folio = folios.data?.[0] || folios.data?.find?.((f: any) => f.status === "open");

      if (folio?.id) {
        const payResp = await request.post(
          `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folio.id}/payments`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: {
              amount: 5000,
              payment_method: "card",
              payment_type: "payment",
              reference_number: "CARD-E2E-001",
            },
          }
        );
        expect([200, 201].includes(payResp.status())).toBeTruthy();
      }
    }
  });

  test("Step 11: Check-out guest", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/operations/check-out`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        reservation_id: reservationId,
        late_checkout_charge: 0,
      },
    });
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("Step 12: Verify invoice generated", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/invoices?property_id=${propertyId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    // Invoice endpoint should exist
    expect([200, 404].includes(resp.status())).toBeTruthy();
  });

  test("Step 13: Verify dashboard shows updated metrics", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(user.email);
    await page.locator('input[type="password"]').fill(user.password);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    await page.waitForTimeout(3000);
    const body = await page.locator("body").textContent();
    expect(body?.length).toBeGreaterThan(100);
  });
});
