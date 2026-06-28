import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { RegisterPage } from "../helpers/page-objects";
import path from "path";
import fs from "fs";

const ARTIFACTS_DIR = "C:/Users/mdsho/.gemini/antigravity/brain/c5d09d8a-5063-4a22-83a0-7769bb26a614/scratch";

// Ensure the directory exists
if (!fs.existsSync(ARTIFACTS_DIR)) {
  fs.mkdirSync(ARTIFACTS_DIR, { recursive: true });
}

test.describe("StayOwner Visual Assessment & UI/UX Audit", () => {
  const user = generateTestUser("owner-visual");
  let registerPage: RegisterPage;

  test("Capture Desktop and Mobile Layouts", async ({ page }) => {
    // 1. Desktop Viewport Setup
    await page.setViewportSize({ width: 1280, height: 800 });
    registerPage = new RegisterPage(page);

    // Register a new tenant owner
    console.log("➡️ Registering owner for visual assessment...");
    await registerPage.goto();
    
    // Take landing page/registration screenshot
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "desktop_register.png") });

    await registerPage.register({
      full_name: user.full_name,
      email: user.email,
      password: user.password,
      property_name: user.property_name,
    });

    // Wait to land on dashboard
    await expect(page).toHaveURL(/dashboard/, { timeout: 25_000 });
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(2000); // Allow animations to settle

    // Click 'Load Demo Data' button to dismiss dialog and populate stay details
    const dialog = page.locator("[data-testid='demo-data-dialog']");
    const loadDemoBtn = page.locator("[data-testid='load-demo-data']");
    if (await dialog.isVisible()) {
      console.log("➡️ Loading demo data to populate dashboard...");
      await loadDemoBtn.click();
      await expect(dialog).toBeHidden({ timeout: 45_000 });
      console.log("➡️ Demo data loaded successfully!");
    }

    // Refresh context: Log out and log back in to load property_id into AuthState
    console.log("➡️ Logging out to refresh session context...");
    await page.evaluate(() => {
      sessionStorage.clear();
      localStorage.removeItem("sf_logout_signal");
    });
    
    await page.goto(`${TEST_CONFIG.BASE_URL}/login`);
    await page.waitForLoadState("networkidle");

    console.log("➡️ Logging back in as stay owner...");
    await page.locator('input[type="email"]').fill(user.email);
    await page.locator('input[type="password"]').fill(user.password);
    await page.locator('button[type="submit"]').click();

    // Wait to land on dashboard again
    await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(3000); // Allow dashboard charts and UI to load

    // Take Desktop Dashboard screenshot
    console.log("📸 Saving desktop dashboard screenshot...");
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "desktop_dashboard.png"), fullPage: true });

    // Navigate to Reservations Page
    console.log("➡️ Navigating to Reservations page...");
    await page.goto(`${TEST_CONFIG.BASE_URL}/reservations`);
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1000);
    console.log("📸 Saving desktop reservations screenshot...");
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "desktop_reservations.png"), fullPage: true });

    // Navigate to Housekeeping Page
    console.log("➡️ Navigating to Housekeeping page...");
    await page.goto(`${TEST_CONFIG.BASE_URL}/housekeeping`);
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1000);
    console.log("📸 Saving desktop housekeeping screenshot...");
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "desktop_housekeeping.png"), fullPage: true });

    // 2. Mobile Viewport Setup (simulate iPhone 13/14)
    console.log("➡️ Switching to mobile viewport...");
    await page.setViewportSize({ width: 390, height: 844 });
    
    // Refresh to trigger mobile layout adjustments
    await page.goto(`${TEST_CONFIG.BASE_URL}/dashboard`);
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(2000);

    // Take Mobile Dashboard screenshot
    console.log("📸 Saving mobile dashboard screenshot...");
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "mobile_dashboard.png") });

    // Click mobile hamburger menu if present
    const menuBtn = page.locator('button[aria-label="Toggle Menu"], button[aria-label="menu"], button:has-text("Menu"), svg.lucide-menu').first();
    if (await menuBtn.isVisible()) {
      console.log("➡️ Clicking mobile menu button...");
      await menuBtn.click();
      await page.waitForTimeout(500);
      console.log("📸 Saving mobile navigation menu screenshot...");
      await page.screenshot({ path: path.join(ARTIFACTS_DIR, "mobile_menu.png") });
      // Close menu by clicking overlay or toggle again
      await menuBtn.click();
      await page.waitForTimeout(300);
    }

    // Navigate to Mobile Reservations Page
    console.log("➡️ Navigating to mobile reservations page...");
    await page.goto(`${TEST_CONFIG.BASE_URL}/reservations`);
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(1000);
    console.log("📸 Saving mobile reservations screenshot...");
    await page.screenshot({ path: path.join(ARTIFACTS_DIR, "mobile_reservations.png") });

    console.log("✅ Visual assessment screenshots captured successfully!");
  });
});
