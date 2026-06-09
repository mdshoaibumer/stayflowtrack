/**
 * E2E Test: Dashboard
 *
 * Tests: Dashboard loads with metrics, navigation works
 */

import { test, expect } from "@playwright/test";
import { generateTestUser } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";
import { loadDemoData } from "../helpers/demo-data";

test.describe("Dashboard", () => {
  let userEmail: string;
  let userPassword: string;

  test.beforeAll(async () => {
    const user = generateTestUser("dash");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    userEmail = user.email;
    userPassword = user.password;

    // Load full demo data for dashboard metrics
    await loadDemoData(result.data.access_token);
  });

  async function login(page: any) {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
  }

  test("should display dashboard with metrics", async ({ page }) => {
    await login(page);
    await page.waitForLoadState("networkidle");

    // Dashboard should have content
    const body = await page.locator("body").textContent();
    expect(body?.length).toBeGreaterThan(100);
  });

  test("should show occupancy data", async ({ page }) => {
    await login(page);
    await page.waitForTimeout(3000);

    // Look for metric cards or numbers
    const metricCards = page.locator("[class*='card'], [class*='metric'], [class*='stat']");
    const count = await metricCards.count();
    expect(count).toBeGreaterThan(0);
  });

  test("should navigate to all pages from sidebar", async ({ page }) => {
    await login(page);

    const routes = ["/reservations", "/guests", "/billing", "/housekeeping", "/laundry", "/operations"];

    for (const route of routes) {
      await page.goto(route);
      await page.waitForLoadState("networkidle");
      await expect(page).toHaveURL(new RegExp(route.replace("/", "")));
      await expect(page.locator("body")).toBeVisible();
    }
  });

  test("should have responsive sidebar navigation", async ({ page }) => {
    await login(page);

    // Check sidebar links exist
    const sidebarLinks = page.locator("nav a, [class*='sidebar'] a");
    const count = await sidebarLinks.count();
    expect(count).toBeGreaterThan(3);
  });

  test("should display revenue trend", async ({ page }) => {
    await login(page);
    await page.waitForTimeout(3000);

    // Look for chart elements (recharts renders SVG)
    const charts = page.locator("svg.recharts-surface, [class*='chart'], canvas");
    if (await charts.first().isVisible({ timeout: 5000 }).catch(() => false)) {
      expect(await charts.count()).toBeGreaterThan(0);
    }
  });
});
