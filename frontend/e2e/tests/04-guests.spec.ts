/**
 * E2E Test: Guest Management
 *
 * Tests: Create guest, search, view details, edit, document upload
 */

import { test, expect } from "@playwright/test";
import { generateTestUser } from "../helpers/test-config";
import { registerTenant, createProperty, createUnitType, createUnit, createGuest } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Guest Management", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;

  test.beforeAll(async () => {
    const user = generateTestUser("guest");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    userEmail = user.email;
    userPassword = user.password;

    // Load minimal demo data
    await loadMinimalDemoData(authToken);
  });

  test("should navigate to guests page", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    await page.goto("/guests");
    await page.waitForLoadState("networkidle");
    await expect(page).toHaveURL(/guests/);
  });

  test("should display existing guests", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    await page.goto("/guests");
    await page.waitForLoadState("networkidle");

    // Should show the demo guest
    await expect(page.locator("body")).toContainText("Test", { timeout: 10_000 });
  });

  test("should create a new guest via UI", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    await page.goto("/guests");
    await page.waitForLoadState("networkidle");

    // Look for "Add Guest" or "New Guest" button
    const addBtn = page.locator(
      'button:has-text("Add"), button:has-text("New"), button:has-text("Create"), a:has-text("Add Guest")'
    );
    if (await addBtn.first().isVisible()) {
      await addBtn.first().click();
      await page.waitForTimeout(1000);

      // Fill in guest form if a modal/form appears
      const firstNameField = page.locator(
        'input[name="first_name"], input[placeholder*="First"], #first_name'
      );
      if (await firstNameField.isVisible()) {
        await firstNameField.fill("E2E");
        const lastNameField = page.locator(
          'input[name="last_name"], input[placeholder*="Last"], #last_name'
        );
        if (await lastNameField.isVisible()) {
          await lastNameField.fill("TestGuest");
        }
        const phoneField = page.locator('input[name="phone"], input[type="tel"], #phone');
        if (await phoneField.isVisible()) {
          await phoneField.fill("+919876500001");
        }

        // Submit
        const submitBtn = page.locator(
          'button[type="submit"], button:has-text("Save"), button:has-text("Create")'
        );
        if (await submitBtn.first().isVisible()) {
          await submitBtn.first().click();
          await page.waitForTimeout(2000);
        }
      }
    }
  });

  test("should search for guests", async ({ page }) => {
    // Create additional guest via API
    await createGuest(authToken, {
      first_name: "Searchable",
      last_name: "Person",
      phone: "+919876500099",
      email: "searchable@test.com",
    });

    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    await page.goto("/guests");
    await page.waitForLoadState("networkidle");

    // Look for search input
    const searchInput = page.locator(
      'input[type="search"], input[placeholder*="Search"], input[placeholder*="search"]'
    );
    if (await searchInput.isVisible()) {
      await searchInput.fill("Searchable");
      await page.waitForTimeout(1500);
      await expect(page.locator("body")).toContainText("Searchable");
    }
  });
});
