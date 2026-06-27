/**
 * E2E Test: Reservation Extended - Update, Cancel, Availability Check
 *
 * Tests: update reservation, cancel reservation, check availability
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createReservation } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Reservation Extended", () => {
  let authToken: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;
  let reservationId: string;

  test.beforeAll(async () => {
    const user = generateTestUser("res-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    demoData = await loadMinimalDemoData(authToken);

    // Create a reservation for testing
    const nextWeek = new Date();
    nextWeek.setDate(nextWeek.getDate() + 20);
    const checkout = new Date(nextWeek);
    checkout.setDate(checkout.getDate() + 3);

    const res = await createReservation(authToken, {
      property_id: demoData.property.id,
      unit_id: demoData.unit1.id,
      guest_id: demoData.guest.id,
      check_in_date: nextWeek.toISOString().split("T")[0],
      check_out_date: checkout.toISOString().split("T")[0],
      rate_per_night: 2500,
      booking_source: "other",
    });
    reservationId = res.data?.id;
  });

  test("should get reservation by ID", async ({ request }) => {
    test.skip(!reservationId, "reservation creation failed - skipping dependent test");
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.id).toBe(reservationId);
    expect(body.data.status).toBe("pending");
  });

  test("should update reservation", async ({ request }) => {
    test.skip(!reservationId, "reservation creation failed - skipping dependent test");
    const nextWeek = new Date();
    nextWeek.setDate(nextWeek.getDate() + 20);
    const newCheckout = new Date(nextWeek);
    newCheckout.setDate(newCheckout.getDate() + 5); // extend by 2 nights

    const resp = await request.put(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          check_in_date: nextWeek.toISOString().split("T")[0],
          check_out_date: newCheckout.toISOString().split("T")[0],
          rate_per_night: 2800,
          num_guests: 2,
          notes: "Updated via E2E test",
        },
      }
    );
    expect([200, 204].includes(resp.status())).toBeTruthy();
  });

  test("should confirm reservation", async ({ request }) => {
    test.skip(!reservationId, "reservation creation failed - skipping dependent test");
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}/confirm`,
      { headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" } }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("should cancel reservation", async ({ request }) => {
    test.skip(!reservationId, "reservation creation failed - skipping dependent test");
    const resp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}/cancel`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { reason: "E2E test - guest cancelled" },
      }
    );
    expect([200, 201].includes(resp.status())).toBeTruthy();

    // Verify status changed
    const getResp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/${reservationId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    const body = await getResp.json();
    expect(body.data.status).toBe("cancelled");
  });

  test("should check availability", async ({ request }) => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 30);
    const dayAfter = new Date(tomorrow);
    dayAfter.setDate(dayAfter.getDate() + 2);

    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/reservations/availability?property_id=${demoData.property.id}&check_in=${tomorrow.toISOString().split("T")[0]}&check_out=${dayAfter.toISOString().split("T")[0]}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data).toBeTruthy();
  });

  test("should list reservations with pagination", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/reservations?page=1&per_page=10`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(Array.isArray(body.data)).toBeTruthy();
  });
});
