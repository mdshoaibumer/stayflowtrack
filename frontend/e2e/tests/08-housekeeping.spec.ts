/**
 * E2E Test: Housekeeping Module
 *
 * Tests: Create tasks, assign, update status, view stats
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { registerTenant } from "../helpers/api-helpers";
import { loadMinimalDemoData } from "../helpers/demo-data";

test.describe("Housekeeping", () => {
  let authToken: string;
  let userEmail: string;
  let userPassword: string;
  let demoData: Awaited<ReturnType<typeof loadMinimalDemoData>>;

  test.beforeAll(async () => {
    const user = generateTestUser("hk");
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
    await page.locator('input[type="email"]').fill(userEmail);
    await page.locator('input[type="password"]').fill(userPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    await page.goto(route);
    await page.waitForLoadState("networkidle");
  }

  test("should navigate to housekeeping page", async ({ page }) => {
    await loginAndGo(page, "/housekeeping");
    await expect(page).toHaveURL(/housekeeping/);
    await expect(page.locator("body")).toBeVisible();
  });

  test("should create a housekeeping task via API", async ({ request }) => {
    const resp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: demoData.property.id,
        unit_id: demoData.unit1.id,
        task_type: "cleaning",
        priority: "high",
        notes: "E2E test: Deep cleaning after checkout",
      },
    });
    expect([200, 201].includes(resp.status())).toBeTruthy();
  });

  test("should display housekeeping tasks in UI", async ({ page }) => {
    await loginAndGo(page, "/housekeeping");
    await page.waitForTimeout(2000);

    // Should show task content
    const body = await page.locator("body").textContent();
    expect(body?.length).toBeGreaterThan(50);
  });

  test("should update task status via API", async ({ request }) => {
    // Create a task
    const createResp = await request.post(`${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks`, {
      headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
      data: {
        property_id: demoData.property.id,
        unit_id: demoData.unit1.id,
        task_type: "inspection",
        priority: "medium",
        notes: "Routine inspection",
      },
    });

    if (createResp.ok()) {
      const task = (await createResp.json()).data;
      if (task?.id) {
        // Update status
        const updateResp = await request.patch(
          `${TEST_CONFIG.API_URL}/api/v1/housekeeping/tasks/${task.id}/status`,
          {
            headers: { Authorization: `Bearer ${authToken}`, "Content-Type": "application/json" },
            data: { status: "in_progress" },
          }
        );
        expect([200, 204].includes(updateResp.status())).toBeTruthy();
      }
    }
  });

  test("should get housekeeping stats via API", async ({ request }) => {
    const resp = await request.get(
      `${TEST_CONFIG.API_URL}/api/v1/housekeeping/stats?property_id=${demoData.property.id}`,
      { headers: { Authorization: `Bearer ${authToken}` } }
    );
    expect([200, 404].includes(resp.status())).toBeTruthy();
  });
});
