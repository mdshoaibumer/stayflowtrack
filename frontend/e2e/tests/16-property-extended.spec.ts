/**
 * E2E Test: Property Extended - Update, Delete, Unit Status, Search
 *
 * Tests: update property, update/delete/status-change units, search units
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant, createProperty, createUnitType, createUnit } from "../helpers/api-helpers";

test.describe("Property Extended", () => {
  let authToken: string;
  let propertyId: string;
  let unitTypeId: string;
  let unitId: string;

  test.beforeAll(async () => {
    const user = generateTestUser("prop-ext");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;

    const prop = await createProperty(authToken, {
      name: "Prop Extended Hotel",
      address: "99 Test Rd",
      city: "Mumbai",
      state: "Maharashtra",
      country: "India",
    });
    propertyId = prop.data.id;

    const ut = await createUnitType(authToken, propertyId, {
      name: "Deluxe",
      base_rate: 3500,
      max_occupancy: 2,
    });
    unitTypeId = ut.data.id;

    const u = await createUnit(authToken, propertyId, {
      unit_number: "501",
      floor: "5",
      unit_type_id: unitTypeId,
    });
    unitId = u.data.id;
  });

  test("should update property details", async ({ request }) => {
    const resp = await request.put(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          name: "Updated Hotel Name",
          address: "99 Test Rd Updated",
          city: "Delhi",
          state: "Delhi",
          country: "India",
        },
      }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.name).toBe("Updated Hotel Name");
  });

  test("should get property by ID", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.id).toBe(propertyId);
  });

  test("should list properties", async ({ request }) => {
    const resp = await request.get(`${TEST_CONFIG.API_URL}/api/v1/properties`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.length).toBeGreaterThan(0);
  });

  test("should list unit types", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/unit-types`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.data.length).toBeGreaterThan(0);
  });

  test("should update a unit", async ({ request }) => {
    const resp = await request.put(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units/${unitId}`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: {
          unit_number: "501A",
          floor: "5",
          unit_type_id: unitTypeId,
        },
      }
    );
    expect([200, 204].includes(resp.status())).toBeTruthy();
  });

  test("should change unit status", async ({ request }) => {
    const resp = await request.patch(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units/${unitId}/status`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { status: "maintenance" },
      }
    );
    expect([200, 204].includes(resp.status())).toBeTruthy();
  });

  test("should search units", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units/search?q=501`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect([200].includes(resp.status())).toBeTruthy();
  });

  test("should create second unit and delete it", async ({ request }) => {
    // Create a unit to delete
    const createResp = await request.post(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units`,
      {
        headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
        data: { unit_number: "999", floor: "9", unit_type_id: unitTypeId },
      }
    );
    expect(createResp.status()).toBe(201);
    const created = await createResp.json();
    const deleteId = created.data.id;

    // Delete it
    const delResp = await request.delete(
      `${TEST_CONFIG.API_URL}/api/v1/properties/${propertyId}/units/${deleteId}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect([200, 204].includes(delResp.status())).toBeTruthy();
  });
});
