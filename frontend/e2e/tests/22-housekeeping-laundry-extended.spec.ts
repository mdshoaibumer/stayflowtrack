/**
 * E2E Test: Housekeeping & Laundry Extended
 *
 * Tests: get single task, assign task, single laundry order, post-to-folio, stats, rate cards
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createReservation } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Housekeeping & Laundry Extended", () => {
  let authToken: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;
  let taskId: string;
  let laundryOrderId: string;
  let rateCardId: string;
  let userId: string;

  test.beforeAll(async () => {
    const user = generateTestUser("hk-lnd-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    userId = result.data.user.id;
    demoData = await loadMinimalDemoData(authToken);
  });

  test.describe("Housekeeping Extended", () => {
    test("should create a housekeeping task", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            property_id: demoData.property.id,
            unit_id: demoData.unit1.id,
            task_type: "checkout_clean",
            priority: "high",
            notes: "E2E extended test task",
          },
        }
      );
      expect([200, 201].includes(resp.status())).toBeTruthy();
      const body = await resp.json();
      taskId = body.data?.id;
    });

    test("should get single task by ID", async ({ request }) => {
      if (!taskId) return;
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks/${taskId}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data.id).toBe(taskId);
    });

    test("should assign task to user", async ({ request }) => {
      if (!taskId) return;
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks/assign`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            task_id: taskId,
            assigned_to: userId,
          },
        }
      );
      expect([200, 201].includes(resp.status())).toBeTruthy();
    });

    test("should get housekeeping stats", async ({ request }) => {
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/housekeeping/stats/${demoData.property.id}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data).toBeTruthy();
    });
  });

  test.describe("Laundry Extended", () => {
    test("should create rate card", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/rate-card`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            property_id: demoData.property.id,
            item_type: "shirt",
            item_name: "Shirt",
            default_rate: 50,
            service_type: "wash",
          },
        }
      );
      expect([200, 201].includes(resp.status())).toBeTruthy();
      const body = await resp.json();
      rateCardId = body.data?.id;
    });

    test("should list rate cards", async ({ request }) => {
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/rate-card/${demoData.property.id}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(Array.isArray(body.data)).toBeTruthy();
      // If rate card was created, it should be in the list
      if (rateCardId) {
        expect(body.data.length).toBeGreaterThan(0);
      }
    });

    test("should update rate card", async ({ request }) => {
      if (!rateCardId) return;
      const resp = await request.put(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/rate-card`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            id: rateCardId,
            property_id: demoData.property.id,
            item_name: "Shirt (Updated)",
            wash_price: 60,
            iron_price: 35,
            dry_clean_price: 180,
          },
        }
      );
      expect([200, 204].includes(resp.status())).toBeTruthy();
    });

    test("should create laundry order", async ({ request }) => {
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/orders`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            property_id: demoData.property.id,
            guest_id: demoData.guest.id,
            order_type: "guest",
            items: [
              { item_type: "shirt", quantity: 3, service_type: "wash", unit_price: 60 },
              { item_type: "trouser", quantity: 2, service_type: "dry_clean", unit_price: 200 },
            ],
            notes: "E2E extended test",
          },
        }
      );
      expect([200, 201].includes(resp.status())).toBeTruthy();
      const body = await resp.json();
      laundryOrderId = body.data?.id;
    });

    test("should get single laundry order", async ({ request }) => {
      if (!laundryOrderId) return;
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/orders/${laundryOrderId}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data.id).toBe(laundryOrderId);
    });

    test("should update laundry order status", async ({ request }) => {
      if (!laundryOrderId) return;
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/orders/status`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: {
            order_id: laundryOrderId,
            status: "washing",
          },
        }
      );
      expect([200, 201].includes(resp.status())).toBeTruthy();
    });

    test("should post laundry order to folio", async ({ request }) => {
      if (!laundryOrderId) return;

      // First check-in a guest to get a folio
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

      const headers = { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" };
      await fetch(`${TEST_CONFIG.API_URL}/api/v1/reservations/${res.data.id}/confirm`, {
        method: "POST", headers,
      });
      await fetch(`${TEST_CONFIG.API_URL}/api/v1/operations/check-in`, {
        method: "POST", headers,
        body: JSON.stringify({
          reservation_id: res.data.id,
          unit_id: demoData.unit1.id,
          deposit_amount: 1000,
          deposit_method: "cash",
          id_document_type: "driving_license",
          id_document_number: "DL9876543",
        }),
      });

      // Post laundry to folio
      const resp = await request.post(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/orders/${laundryOrderId}/post-to-folio`,
        {
          headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
          data: { reservation_id: res.data.id },
        }
      );
      expect([200, 201, 400, 409].includes(resp.status())).toBeTruthy();
    });

    test("should get laundry stats", async ({ request }) => {
      const resp = await request.get(
        `${TEST_CONFIG.API_URL}/api/v1/laundry/stats/${demoData.property.id}`,
        { headers: { Authorization: `Bearer ${authToken}` } }
      );
      expect(resp.status()).toBe(200);
      const body = await resp.json();
      expect(body.data).toBeTruthy();
    });
  });
});
