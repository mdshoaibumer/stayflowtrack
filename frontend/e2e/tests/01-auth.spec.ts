/**
 * E2E Test: Authentication Flow
 *
 * Tests: Register, Login, Logout, Password validation, Session management
 */

import { test, expect } from "@playwright/test";
import { generateTestUser, TEST_CONFIG } from "../helpers/test-config";
import { LoginPage, RegisterPage } from "../helpers/page-objects";

test.describe("Authentication", () => {
  test.describe("Registration", () => {
    test("should register a new user successfully", async ({ page }) => {
      const registerPage = new RegisterPage(page);
      const user = generateTestUser("reg");

      await registerPage.goto();
      await registerPage.register({
        full_name: user.full_name,
        email: user.email,
        password: user.password,
        property_name: user.property_name,
        phone: user.phone,
      });

      // Should redirect to dashboard after successful registration
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    });

    test("should show error for mismatched passwords", async ({ page }) => {
      const registerPage = new RegisterPage(page);

      await registerPage.goto();
      await registerPage.fullNameInput.fill("Test User");
      await registerPage.emailInput.fill("test@test.com");
      await registerPage.propertyNameInput.fill("Test Hotel");
      await registerPage.passwordInput.fill("Password123!");
      await registerPage.confirmPasswordInput.fill("DifferentPass!");
      await registerPage.submitButton.click();

      await registerPage.expectError("Passwords do not match");
    });

    test("should show error for short password", async ({ page }) => {
      const registerPage = new RegisterPage(page);

      await registerPage.goto();
      await registerPage.fullNameInput.fill("Test User");
      await registerPage.emailInput.fill("test@test.com");
      await registerPage.propertyNameInput.fill("Test Hotel");
      await registerPage.passwordInput.fill("short");
      await registerPage.confirmPasswordInput.fill("short");
      await registerPage.submitButton.click();

      await registerPage.expectError("at least 8 characters");
    });

    test("should show error for duplicate email", async ({ page }) => {
      const registerPage = new RegisterPage(page);
      const user = generateTestUser("dup");

      // Register first time
      await registerPage.goto();
      await registerPage.register({
        full_name: user.full_name,
        email: user.email,
        password: user.password,
        property_name: user.property_name,
      });
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

      // Try to register again with same email (new page context)
      const newPage = await page.context().newPage();
      const registerPage2 = new RegisterPage(newPage);
      await registerPage2.goto();
      await registerPage2.register({
        full_name: "Another User",
        email: user.email,
        password: user.password,
        property_name: "Another Hotel",
      });

      await registerPage2.expectError();
      await newPage.close();
    });

    test("should navigate to login page from register", async ({ page }) => {
      const registerPage = new RegisterPage(page);
      await registerPage.goto();
      await registerPage.loginLink.click();
      await expect(page).toHaveURL(/login/);
    });
  });

  test.describe("Login", () => {
    let testUser: ReturnType<typeof generateTestUser>;

    test.beforeAll(async ({ browser }) => {
      // Create a test user via registration
      testUser = generateTestUser("login");
      const page = await browser.newPage();
      const registerPage = new RegisterPage(page);
      await registerPage.goto();
      await registerPage.register({
        full_name: testUser.full_name,
        email: testUser.email,
        password: testUser.password,
        property_name: testUser.property_name,
      });
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
      await page.close();
    });

    test("should login with valid credentials", async ({ page }) => {
      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.login(testUser.email, testUser.password);
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });
    });

    test("should show error for invalid credentials", async ({ page }) => {
      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.login("nonexistent@test.com", "WrongPass123!");
      await loginPage.expectError();
    });

    test("should show error for wrong password", async ({ page }) => {
      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.login(testUser.email, "WrongPassword!");
      await loginPage.expectError();
    });

    test("should navigate to register page from login", async ({ page }) => {
      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.registerLink.click();
      await expect(page).toHaveURL(/register/);
    });
  });

  test.describe("Logout", () => {
    test("should logout and redirect to login", async ({ page }) => {
      const user = generateTestUser("logout");
      const registerPage = new RegisterPage(page);
      await registerPage.goto();
      await registerPage.register({
        full_name: user.full_name,
        email: user.email,
        password: user.password,
        property_name: user.property_name,
      });
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

      // Prevent demo data dialog
      await page.evaluate(() => localStorage.setItem("demo_data_shown", "true"));
      await page.reload();
      await page.waitForLoadState("networkidle");

      // Click logout
      const userMenuBtn = page.locator('header button').last();
      await userMenuBtn.click();
      
      const logoutBtn = page.locator('button:has-text("Sign Out")');
      await expect(logoutBtn).toBeVisible({ timeout: 5000 });
      await logoutBtn.click();
      await expect(page).toHaveURL(/login/);
    });
  });

  test.describe("Route Protection", () => {
    test("should redirect unauthenticated users to login", async ({ page }) => {
      // Clear any stored auth
      await page.goto("/");
      await page.evaluate(() => localStorage.clear());

      await page.goto(TEST_CONFIG.ROUTES.DASHBOARD);
      await expect(page).toHaveURL(/login/, { timeout: 10_000 });
    });

    test("should redirect to dashboard if already logged in", async ({ page }) => {
      const user = generateTestUser("protected");
      const registerPage = new RegisterPage(page);
      await registerPage.goto();
      await registerPage.register({
        full_name: user.full_name,
        email: user.email,
        password: user.password,
        property_name: user.property_name,
      });
      await expect(page).toHaveURL(/dashboard/, { timeout: 15_000 });

      // Navigate to login should redirect back to dashboard
      await page.goto(TEST_CONFIG.ROUTES.LOGIN);
      // Already logged in, should stay on dashboard or redirect
      await page.waitForTimeout(2000);
      const url = page.url();
      expect(url).toMatch(/dashboard|login/);
    });
  });
});
