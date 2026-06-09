/**
 * E2E Test: Demo Data Loading
 *
 * Tests: After registration, user sees demo data prompt, can load or skip
 */

import { test, expect } from "@playwright/test";
import { generateTestUser } from "../helpers/test-config";
import { RegisterPage, DemoDataDialog } from "../helpers/page-objects";

test.describe("Demo Data Loading", () => {
  test("should show demo data dialog after registration", async ({ page }) => {
    const user = generateTestUser("demo");
    const registerPage = new RegisterPage(page);

    await registerPage.goto();
    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });

    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    // Demo data dialog should appear for new users
    const dialog = new DemoDataDialog(page);
    await dialog.expectVisible();
  });

  test("should load demo data when user clicks Load", async ({ page }) => {
    const user = generateTestUser("demo-load");
    const registerPage = new RegisterPage(page);

    await registerPage.goto();
    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });

    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    const dialog = new DemoDataDialog(page);
    await dialog.expectVisible();
    await dialog.loadDemoData();
    await dialog.waitForCompletion();

    // After demo data loads, dashboard should show data
    await expect(page.locator("body")).not.toContainText("No properties");
  });

  test("should skip demo data when user clicks Skip", async ({ page }) => {
    const user = generateTestUser("demo-skip");
    const registerPage = new RegisterPage(page);

    await registerPage.goto();
    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });

    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    const dialog = new DemoDataDialog(page);
    await dialog.expectVisible();
    await dialog.skipDemoData();

    // Dialog should close
    await expect(dialog.dialog).toBeHidden();
  });

  test("should not show demo data dialog on subsequent logins", async ({ page }) => {
    const user = generateTestUser("demo-once");
    const registerPage = new RegisterPage(page);

    // Register and skip demo data
    await registerPage.goto();
    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    const dialog = new DemoDataDialog(page);
    await dialog.expectVisible();
    await dialog.skipDemoData();
    await expect(dialog.dialog).toBeHidden();

    // Logout and login again
    await page.evaluate(() => localStorage.clear());
    await page.goto("/login");

    const loginPage = page;
    await loginPage.locator('input[type="email"]').fill(user.email);
    await loginPage.locator('input[type="password"]').fill(user.password);
    await loginPage.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

    // Dialog should NOT appear again
    await page.waitForTimeout(3000);
    await expect(page.locator("[data-testid='demo-data-dialog']")).toBeHidden();
  });
});
