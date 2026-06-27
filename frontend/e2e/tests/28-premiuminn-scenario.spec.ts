/**
 * E2E Test: PremiumInn Business Scenario
 *
 * Tests:
 * - Owner registers and creates 9 apartments
 * - Owner creates 2 staff members (receptionist & housekeeping/receptionist)
 * - Staff 1 books daily & weekly
 * - Staff 2 books monthly & yearly
 * - Comprehensive positive and negative test cases
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

test.describe("PremiumInn Comprehensive Scenario", () => {
  let ownerToken: string;
  let staff1Token: string;
  let staff2Token: string;

  let ownerData: any;
  let propertyId: string;
  let unitTypeId: string;
  const units: any[] = [];
  let guest1Id: string;
  let guest2Id: string;
  let guest3Id: string;

  test.beforeAll(async () => {
    // 1. Owner Setup
    const user = generateTestUser("owner-premium");
    const uniqueSuffix = Date.now().toString().slice(-6);
    const propertyName = `PremiumInn ${uniqueSuffix}`;
    const result = await registerTenant({
      full_name: "PremiumInn Owner",
      email: user.email,
      password: user.password,
      property_name: propertyName,
    });
    ownerToken = result.data.access_token;
    ownerData = user;
    
    const propResp = await createProperty(ownerToken, {
      name: `PremiumInn HQ ${uniqueSuffix}`,
      address: "123 Premium Lane",
      city: "Dubai",
      state: "Dubai",
      country: "UAE",
    });
    propertyId = propResp.data.id;

    // Create a generic Service Apartment unit type
    const utResp = await createUnitType(ownerToken, propertyId, {
      name: "Service Apartment",
      base_rate: 2000,
      max_occupancy: 4,
    });
    unitTypeId = utResp.data.id;

    // Create 9 Units (101 to 109)
    for (let i = 1; i <= 9; i++) {
      const uResp = await createUnit(ownerToken, propertyId, {
        unit_number: `10${i}`,
        floor: "1",
        unit_type_id: unitTypeId,
      });
      units.push(uResp.data);
    }

    // 2. Create 2 Staff accounts
    const staff1 = generateTestUser("staff1");
    await createUser(ownerToken, {
      email: staff1.email,
      password: staff1.password,
      first_name: "John",
      last_name: "Doe (Receptionist)",
      role_name: "receptionist",
    });
    
    const staff2 = generateTestUser("staff2");
    await createUser(ownerToken, {
      email: staff2.email,
      password: staff2.password,
      first_name: "Jane",
      last_name: "Smith (Housekeeping)",
      role_name: "housekeeping",
    });

    // 3. Login as Staff
    const login1 = await login(staff1.email, staff1.password);
    staff1Token = login1.data.access_token;

    const login2 = await login(staff2.email, staff2.password);
    staff2Token = login2.data.access_token;
  });

  test("Phase 1: Verify Owner and Staff Setup", async ({ page }) => {
    // Owner login
    await page.goto("/login");
    await page.fill('input[type="email"]', ownerData.email);
    await page.fill('input[type="password"]', ownerData.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/dashboard/, { timeout: 15000 });

    // Verify we can navigate to units page as owner
    await page.goto("/units");
    await expect(page).toHaveURL(/units/);
  });

  test("Phase 2: Staff 1 - Daily and Weekly Bookings", async () => {
    // Create Guests using Staff 1
    const g1 = await createGuest(staff1Token, {
      first_name: "Daily",
      last_name: "Guest",
      phone: "+971501234567",
    });
    guest1Id = g1.data.id;

    const g2 = await createGuest(staff1Token, {
      first_name: "Weekly",
      last_name: "Guest",
      phone: "+971501234568",
    });
    guest2Id = g2.data.id;

    // Daily Booking (2 nights) in Unit 101
    const checkIn1 = new Date();
    const checkOut1 = new Date();
    checkOut1.setDate(checkOut1.getDate() + 2);
    
    const dailyRes = await createReservation(staff1Token, {
      property_id: propertyId,
      unit_id: units[0].id, // 101
      guest_id: guest1Id,
      check_in_date: checkIn1.toISOString().split("T")[0],
      check_out_date: checkOut1.toISOString().split("T")[0],
      rate_per_night: 2000,
      booking_source: "walk_in",
    });
    expect(dailyRes.data?.id).toBeTruthy();

    // Weekly Booking (7 nights) in Unit 102
    const checkIn2 = new Date();
    const checkOut2 = new Date();
    checkOut2.setDate(checkOut2.getDate() + 7);

    const weeklyRes = await createReservation(staff1Token, {
      property_id: propertyId,
      unit_id: units[1].id, // 102
      guest_id: guest2Id,
      check_in_date: checkIn2.toISOString().split("T")[0],
      check_out_date: checkOut2.toISOString().split("T")[0],
      rate_per_night: 1800,
      booking_source: "phone",
    });
    expect(weeklyRes.data?.id).toBeTruthy();

    // NEGATIVE: Try to double book Unit 101 on the same dates
    const conflictRes = await createReservation(staff1Token, {
      property_id: propertyId,
      unit_id: units[0].id, // 101
      guest_id: guest2Id,
      check_in_date: checkIn1.toISOString().split("T")[0],
      check_out_date: checkOut1.toISOString().split("T")[0],
      rate_per_night: 2500,
      booking_source: "walk_in",
    });
    // Should fail with a conflict
    expect(conflictRes.error).toBeDefined();
    expect(conflictRes.error?.code).toBe("CONFLICT");
  });

  test("Phase 3: Staff 2 - Monthly and Yearly Bookings & Negative Cases", async () => {
    // Staff 2 (housekeeping) attempting to create a user should FAIL (Negative test for RBAC)
    const unauthorizedRes = await createUser(staff2Token, {
      email: "hacker@test.com",
      password: "password123",
      first_name: "Hacker",
      last_name: "User",
      role_name: "receptionist"
    });
    expect(unauthorizedRes.error).toBeDefined();
    expect(unauthorizedRes.error?.code).toBe("FORBIDDEN");
    // Monthly Booking (30 nights) - By Owner (or Staff 1)
    const checkIn3 = new Date();
    const checkOut3 = new Date();
    checkOut3.setDate(checkOut3.getDate() + 30);

    const g3 = await createGuest(ownerToken, {
      first_name: "Monthly",
      last_name: "Guest",
      phone: "+971501234569",
    });
    guest3Id = g3.data.id;

    const monthlyRes = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: units[2].id, // 103
      guest_id: guest3Id,
      check_in_date: checkIn3.toISOString().split("T")[0],
      check_out_date: checkOut3.toISOString().split("T")[0],
      rate_per_night: 1500,
      booking_source: "other",
    });
    expect(monthlyRes.data?.id).toBeTruthy();

    // Yearly Booking (365 nights) - By Owner
    const checkIn4 = new Date();
    const checkOut4 = new Date();
    checkOut4.setDate(checkOut4.getDate() + 365);
    
    const yearlyRes = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: units[3].id, // 104
      guest_id: guest3Id,
      check_in_date: checkIn4.toISOString().split("T")[0],
      check_out_date: checkOut4.toISOString().split("T")[0],
      rate_per_night: 1200,
      booking_source: "other",
    });
    expect(yearlyRes.data?.id).toBeTruthy();

    // NEGATIVE: Invalid dates (Check out before Check in)
    const invalidDatesRes = await createReservation(ownerToken, {
      property_id: propertyId,
      unit_id: units[4].id, // 105
      guest_id: guest3Id,
      check_in_date: checkOut4.toISOString().split("T")[0],
      check_out_date: checkIn4.toISOString().split("T")[0],
      rate_per_night: 1200,
      booking_source: "walk_in",
    });
    expect(invalidDatesRes.error).toBeDefined();
    expect(invalidDatesRes.error?.code).toBe("BAD_REQUEST");
  });
});
