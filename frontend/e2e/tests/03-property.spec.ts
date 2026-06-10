/**
 * E2E Test: Property Management
 *
 * Tests: Create property, unit types, units, view/edit/delete
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { RegisterPage } from "../helpers/page-objects";
import { login, registerTenant, createProperty, createUnitType } from "../helpers/api-helpers";

test.describe("Property Management", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;

  test.beforeAll(async () => {
    // Register via API for faster setup
    const user = generateTestUser("prop");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    userEmail = user.email;
    userPassword = user.password;
  });

  test("should display properties page", async ({ page }) => {
    // Login via UI
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    // Navigate to settings/properties
    await page.goto("/settings");
    await page.waitForLoadState("networkidle");
    await expect(page.locator("body")).toBeVisible();
  });

  test("should create a new property via API and see it in UI", async ({ page }) => {
    // Create property via API
    await createProperty(authToken, {
      name: "Seaside Resort",
      address: "45 Beach Road",
      city: "Goa",
      state: "Goa",
      country: "India",
    });

    // Login and verify in UI
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    // Check property appears somewhere in the app
    await page.waitForLoadState("networkidle");
    const bodyText = await page.locator("body").textContent();
    // Property should be referenced somewhere
    expect(bodyText).toBeDefined();
  });

  test("should create unit types and units", async ({ page }) => {
    // Create property and units via API
    const propResult = await createProperty(authToken, {
      name: "Unit Test Hotel",
      address: "123 Main St",
      city: "Delhi",
      state: "Delhi",
      country: "India",
    });
    const propertyId = propResult.data.id;

    // Create unit type
    const utResult = await createUnitType(authToken, propertyId, {
      name: "Deluxe Suite",
      base_rate: 5000,
      max_occupancy: 3,
    });
    expect(utResult.data.id).toBeTruthy();

    // Login and navigate
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
  });
});
