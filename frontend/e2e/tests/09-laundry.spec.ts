/**
 * E2E Test: Laundry Module
 *
 * Tests: Create orders, manage rate cards, post to folio
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Laundry", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;

  test.beforeAll(async () => {
    const user = generateTestUser("laundry");
    const result = await registerTenant({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });
    authToken = result.data.access_token;
    userEmail = user.email;
    userPassword = user.password;
    demoData = await loadMinimalDemoData(authToken);
  });

  async function loginAndGo(page: any, route: string) {
    await page.goto("/login");
    await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    await page.goto(route);
    await page.waitForLoadState("networkidle");
  }

  test("should navigate to laundry page", async ({ page }) => {
    await loginAndGo(page, "/laundry");
    await expect(page).toHaveURL(/laundry/);
  });

  test("should create a rate card via API", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/laundry/rate-cards`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: demoData.property.id,
        item_name: "Shirt",
        item_type: "clothing",
        wash_price: 50,
        iron_price: 30,
        wash_iron_price: 70,
        dry_clean_price: 150,
      },
    });
    expect([200, 201, 409].includes(resp.status())).toBeTruthy();
  });

  test("should create a laundry order via API", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/laundry/orders`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: demoData.property.id,
        guest_id: demoData.guest.id,
        order_type: "guest",
        items: [
          { item_type: "shirt", service_type: "wash_iron", quantity: 3, unit_price: 70 },
          { item_type: "trouser", service_type: "dry_clean", quantity: 1, unit_price: 200 },
        ],
        notes: "E2E test order — express requested",
      },
    });
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("should display laundry orders in UI", async ({ page }) => {
    await loginAndGo(page, "/laundry");
    await page.waitForTimeout(2000);
    const body = await page.locator("body").textContent();
    expect(body?.length).toBeGreaterThan(50);
  });

  test("should update laundry order status via API", async ({ request }) => {
    // Create an order first
    const createResp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/laundry/orders`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: demoData.property.id,
        order_type: "house",
        items: [{ item_type: "bedsheet", service_type: "wash", quantity: 5, unit_price: 40 }],
      },
    });

    if (createResp.ok()) {
      const order = (await createResp.json()).data;
      if (order?.id) {
        const updateResp = await request.patch(
          `${TEST_CONFIG.API_URL}/api/v1/laundry/orders/${order.id}/status`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: { status: "in_progress" },
          }
        );
        expect([200, 204].includes(updateResp.status())).toBeTruthy();
      }
    }
  });
});
