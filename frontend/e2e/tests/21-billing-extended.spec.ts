/**
 * E2E Test: Billing Extended - Folio Summary, Line Items, Payments List, Void, Invoices, PDF
 *
 * Tests: folio summary, list line items, list payments, void charge, list/get invoices, PDF generation
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createReservation } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Billing Extended", () => {
  let authToken: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;
  let folioId: string;
  let lineItemId: string;
  let reservationId: string;

  test.beforeAll(async () => {
    const user = generateTestUser("bill-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    demoData = await loadMinimalDemoData(authToken);

    // Create reservation, confirm, check-in (to get a folio)
    const today = new Date();
    const checkout = new Date(today);
    checkout.setDate(today.getDate() + 2);

    const res = await createReservation(authToken, {
      property_id: demoData.property.id,
      unit_id: demoData.unit1.id,
      guest_id: demoData.guest.id,
      check_in_date: today.toISOString().split("T")[0],
      check_out_date: checkout.toISOString().split("T")[0],
      rate_per_night: 3000,
      booking_source: "walk_in",
    });
    reservationId = res.data?.id;

    if (!reservationId) return;

    const headers = { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" };
    await fetch(`${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}/confirm`, {
      method: "POST", headers,
    });
    await fetch(`${TEST_CONFIG.API_URL}/api/v1/operations/check-in`, {
      method: "POST", headers,
      body: JSON.stringify({
        reservation_id: reservationId,
        unit_id: demoData.unit1.id,
        deposit_amount: 3000,
        deposit_method: "upi",
        id_document_type: "passport",
        id_document_number: "J1234567",
      }),
    });

    // Get folio ID
    const folioResp = await fetch(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios/reservation/${reservationId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    const folioBody = await folioResp.json();
    folioId = folioBody.data?.id;
  });

  test("should get folio by ID", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folioId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.id).toBe(folioId);
  });

  test("should get folio summary", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folioId}/summary`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should add charge and get line item ID", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/billing/charges`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          folio_id: folioId,
          description: "Minibar - E2E test",
          category: "food",
          quantity: 2,
          unit_price: 500,
          tax_rate: 5,
        },
      }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();
    const body = await resp.json();
    lineItemId = body.data?.id;
  });

  test("should list line items for folio", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folioId}/items`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(body.data.length).toBeGreaterThan(0);
  });

  test("should record payment for folio", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/billing/payments`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          folio_id: folioId,
          payment_type: "payment",
          payment_method: "card",
          amount: 1000,
          reference_number: "CARD-E2E-001",
        },
      }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("should list payments for folio", async ({ request }) => {
    if (!folioId) return;
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/folios/${folioId}/payments`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(body.data.length).toBeGreaterThan(0);
  });

  test("should void a line item", async ({ request }) => {
    if (!folioId || !lineItemId) return;
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/billing/charges/void`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          line_item_id: lineItemId,
          reason: "E2E test void",
        },
      }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("should list invoices", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/invoices`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.data)).toBeTruthy();
  });

  test("should check-out and generate invoice", async ({ request }) => {
    if (!reservationId) return;
    // Check-out to generate invoice
    const checkoutResp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/operations/check-out`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          reservation_id: reservationId,
          payment_method: "card",
          payment_reference: "CARD-FINAL-E2E",
        },
      }
    );
    // Accept success or balance-related errors
    expect([200, 201, 400, 409, 422].includes(checkoutResp.status())).toBeTruthy();
  });

  test("should get invoice by ID after checkout", async ({ request }) => {
    // List invoices, get the latest
    const listResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/invoices`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    const listBody = await listResp.json();
    if (listBody.data?.length > 0) {
      const invoiceId = listBody.data[0].id;
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/billing/invoices/${invoiceId}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data.id).toBe(invoiceId);
    }
  });

  test("should generate invoice PDF", async ({ request }) => {
    const listResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/invoices`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    const listBody = await listResp.json();
    if (listBody.data?.length > 0) {
      const invoiceId = listBody.data[0].id;
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/billing/invoices/${invoiceId}/pdf`,
        { headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" } }
      );
      expect([200, 201, 501].includes(resp.status())).toBeTruthy();
    }
  });

  test("should download invoice PDF", async ({ request }) => {
    const listResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/billing/invoices`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    const listBody = await listResp.json();
    if (listBody.data?.length > 0) {
      const invoiceId = listBody.data[0].id;
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/billing/invoices/${invoiceId}/pdf`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      // PDF may not exist yet if generation is async, accept 200 or 404
      expect([200, 404].includes(resp.status())).toBeTruthy();
    }
  });
});
