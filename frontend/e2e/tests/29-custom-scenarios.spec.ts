/**
 * E2E Test: Custom Business Scenarios
 *
 * Scenarios:
 * 1. E2E Guest Lifecycle: Check-in, add room service charge, record payment, check-out.
 * 2. Double-Booking Conflict Validation: Overlapping dates on the same unit must be blocked.
 * 3. Housekeeping Task Lifecycle: Assign task to staff, progress from dirty -> cleaning -> ready.
 */

import { test, expect } from "@playwright/test";
import { generateTestUser } from "../helpers/test-config";
import {
  registerTenant,
  login,
  createProperty,
  createUnitType,
  createUnit,
  createUser,
  createGuest,
  createReservation,
} from "../helpers/api-helpers";

const API_BASE = process.env.E2E_API_URL || "http://localhost:8080";

test.describe("Custom E2E Business Scenarios", () => {
  let ownerToken: string;
  let ownerData: any;
  let propertyId: string;
  let unitTypeId: string;
  let unit101Id: string;
  let unit102Id: string;
  let guestId: string;

  test.beforeAll(async () => {
    // 1. Owner Setup
    const user = generateTestUser("scenario-owner");
    const uniqueSuffix = Date.now().toString().slice(-6);
    const result = await registerTenant({
      full_name: "Scenario Owner",
      email: user.email,
      password: user.password,
      property_name: `Boutique Hotel ${uniqueSuffix}`,
    });
    ownerToken = result.data.access_token;
    ownerData = user;

    // 2. Create Property
    const propResp = await createProperty(ownerToken, {
      name: `Boutique HQ ${uniqueSuffix}`,
      address: "456 Scenario St",
      city: "San Francisco",
      state: "CA",
      country: "USA",
    });
    propertyId = propResp.data.id;

    // 3. Create Unit Type
    const utResp = await createUnitType(ownerToken, propertyId, {
      name: "Luxury Suite",
      base_rate: 3000,
      max_occupancy: 2,
    });
    unitTypeId = utResp.data.id;

    // 4. Create 2 Units
    const u1Resp = await createUnit(ownerToken, propertyId, {
      unit_number: "101",
      floor: "1",
      unit_type_id: unitTypeId,
    });
    unit101Id = u1Resp.data.id;

    const u2Resp = await createUnit(ownerToken, propertyId, {
      unit_number: "102",
      floor: "1",
      unit_type_id: unitTypeId,
    });
    unit102Id = u2Resp.data.id;

    // 5. Create Guest
    const gResp = await createGuest(ownerToken, {
      first_name: "Alice",
      last_name: "Smith",
      phone: "+15551234567",
      email: `alice-${uniqueSuffix}@example.com`,
    });
    guestId = gResp.data.id;
  });

  test("Scenario 1: End-to-End Guest Lifecycle (Check-in -> Add Charge -> Pay -> Check-out)", async () => {
    // A. Book a Reservation for Unit 101
    const checkInDate = new Date();
    const checkOutDate = new Date();
    checkOutDate.setDate(checkOutDate.getDate() + 3); // 3 nights

    const resResp = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: unit101Id,
      guest_id: guestId,
      check_in_date: checkInDate.toISOString().split("T")[0],
      check_out_date: checkOutDate.toISOString().split("T")[0],
      rate_per_night: 3000,
      booking_source: "website",
    });
    const reservationId = resResp.data.id;
    expect(reservationId).toBeTruthy();

    // A.2 Confirm the reservation first
    const confirmResp = await fetch(`${API_BASE}/api/v1/reservations/${reservationId}/confirm`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${ownerToken}`,
      },
    });
    expect([200, 201].includes(confirmResp.status)).toBeTruthy();

    // B. Check-in the Guest
    const checkInResp = await fetch(`${API_BASE}/api/v1/operations/check-in`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        reservation_id: reservationId,
        assigned_unit_id: unit101Id,
        id_document_type: "driving_license",
        id_document_number: "DL-9876543",
        notes: "Guest checking in on time",
        deposit_amount: 1500,
        deposit_method: "cash",
      }),
    });
    expect(checkInResp.status).toBe(200);
    const checkInData = await checkInResp.json();

    // C. Get Open Folio
    const folioResp = await fetch(`${API_BASE}/api/v1/billing/folios/reservation/${reservationId}`, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${ownerToken}`,
      },
    });
    expect(folioResp.status).toBe(200);
    const folioData = await folioResp.json();
    const folioId = folioData.data.id;
    expect(folioId).toBeTruthy();

    // D. Add Room Service Charge (Folio Charge)
    const chargeResp = await fetch(`${API_BASE}/api/v1/billing/charges`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        folio_id: folioId,
        category: "food_beverage",
        description: "Dinner Room Service",
        quantity: 1,
        unit_price: 80,
        tax_rate: 18,
      }),
    });
    expect(chargeResp.status).toBe(201);

    // E. Record Folio Payment (UPI Payment)
    const paymentResp = await fetch(`${API_BASE}/api/v1/billing/payments`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        folio_id: folioId,
        payment_type: "payment",
        payment_method: "upi",
        amount: 9000 + 80 * 1.18, // room total (3 nights * 3000) + food total
        reference_number: "UPI-TXN-12345",
        notes: "Full folio clearance",
      }),
    });
    expect(paymentResp.status).toBe(201);

    // F. Perform Check-out
    const checkOutResp = await fetch(`${API_BASE}/api/v1/operations/check-out`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        reservation_id: reservationId,
        late_checkout_charge: 0,
        notes: "Everything checked out clean",
      }),
    });
    expect(checkOutResp.status).toBe(200);
  });

  test("Scenario 2: Double-Booking Conflict Validation (Overlapping Booking Fails)", async () => {
    const checkInDate = new Date();
    const checkOutDate = new Date();
    checkOutDate.setDate(checkOutDate.getDate() + 3);

    // Booking 1 on Unit 102
    const res1 = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: unit102Id,
      guest_id: guestId,
      check_in_date: checkInDate.toISOString().split("T")[0],
      check_out_date: checkOutDate.toISOString().split("T")[0],
      rate_per_night: 3000,
      booking_source: "website",
    });
    expect(res1.data?.id).toBeTruthy();

    // Booking 2 on Unit 102 (Overlapping dates) -> Expect Conflict
    const checkInDate2 = new Date();
    checkInDate2.setDate(checkInDate2.getDate() + 1); // Overlaps
    const checkOutDate2 = new Date();
    checkOutDate2.setDate(checkOutDate2.getDate() + 4);

    const res2 = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: unit102Id,
      guest_id: guestId,
      check_in_date: checkInDate2.toISOString().split("T")[0],
      check_out_date: checkOutDate2.toISOString().split("T")[0],
      rate_per_night: 3000,
      booking_source: "walk_in",
    });

    expect(res2.error).toBeDefined();
    expect(res2.error?.code).toBe("CONFLICT");
  });

  test("Scenario 3: Housekeeping Task Lifecycle (dirty -> cleaning -> ready)", async () => {
    // 1. Create a Housekeeping Task for Unit 101
    const taskResp = await fetch(`${API_BASE}/api/v1/housekeeping/tasks`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        property_id: propertyId,
        unit_id: unit101Id,
        priority: "high",
        task_type: "checkout_clean",
        notes: "Immediate attention requested",
        estimated_minutes: 45,
      }),
    });
    expect(taskResp.status).toBe(201);
    const taskData = await taskResp.json();
    const taskId = taskData.data.id;
    expect(taskId).toBeTruthy();

    // 2. Update status to 'cleaning'
    const statusCleanResp = await fetch(`${API_BASE}/api/v1/housekeeping/tasks/status`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        task_id: taskId,
        status: "cleaning",
        notes: "Started dusting and washing",
      }),
    });
    expect(statusCleanResp.status).toBe(200);

    // 3. Update status to 'ready' (completed)
    const statusReadyResp = await fetch(`${API_BASE}/api/v1/housekeeping/tasks/status`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${ownerToken}`,
      },
      body: JSON.stringify({
        task_id: taskId,
        status: "ready",
        actual_minutes: 40,
        notes: "Room is clean and inspection passed",
      }),
    });
    expect(statusReadyResp.status).toBe(200);
  });
});
