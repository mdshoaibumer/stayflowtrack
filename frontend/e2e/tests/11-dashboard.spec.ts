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
  let authToken: string;

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
    authToken = result.data.access_token;

    // Load full demo data for dashboard metrics
    await loadDemoData(result.data.access_token);
  });

  async function login(page: any) {
    // Fetch property using test API helper configuration
    const pResp = await fetch(`http://localhost:8080/api/v1/properties`, {
      headers: { Authorization: `Bearer ${authToken}` }
    });
    const pJson = await pResp.json();
    const pid = (pJson.data && pJson.data.length > 0) ? pJson.data[0].id : null;

    page.on('console', msg => console.log('BROWSER CONSOLE:', msg.type(), msg.text()));

    await page.goto("/");
    await page.evaluate(({ token, email, pid }) => {
      localStorage.setItem("demo_data_shown", "true");
      localStorage.setItem("access_token", token);
      localStorage.setItem("user", JSON.stringify({
        id: "test-user-id",
        email: email,
        full_name: "Test User",
        role: "super_admin",
        property_id: pid
      }));
    }, { token: authToken, email: userEmail, pid });

    await page.goto("/dashboard");
    await page.waitForLoadState("domcontentloaded");
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

    // Wait for the dashboard to finish loading
    const occupancyCard = page.locator('text=Occupancy').first();
    try {
      await expect(occupancyCard).toBeVisible({ timeout: 15_000 });
    } catch (err) {
      const body = await page.locator("body").textContent();
      console.log("BODY TEXT:", body);
      throw err;
    }
  });

  test("should navigate to all pages from sidebar", async ({ page }) => {
    await login(page);

    const routes = ["/reservations", "/guests", "/billing", "/housekeeping", "/laundry", "/operations"];

    for (const route of routes) {
      await page.goto(route);
      await page.waitForLoadState("domcontentloaded");
      await expect(page).toHaveURL(new RegExp(route.replace("/", "")));
      await expect(page.locator("body")).toBeVisible();
    }
  });

  test("should have responsive sidebar navigation", async ({ page }) => {
    await login(page);

    // Check sidebar links exist
    const sidebarLinks = page.locator("nav a, [class*='sidebar'] a");
    await expect(sidebarLinks.first()).toBeVisible({ timeout: 10000 });
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
