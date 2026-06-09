/**
 * Page Object Models for E2E tests
 * Encapsulates page interactions for reuse across tests
 */

import { type Page, type Locator, expect } from "@playwright/test";
import { TEST_CONFIG } from "./test-config";

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;
  readonly registerLink: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.locator('input[type="email"]');
    this.passwordInput = page.locator('input[type="password"]');
    this.submitButton = page.locator('button[type="submit"]');
    this.errorMessage = page.locator('[class*="red"]');
    this.registerLink = page.locator('a[href="/register"]');
  }

  async goto() {
    await this.page.goto(TEST_CONFIG.ROUTES.LOGIN);
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }

  async expectError(message?: string) {
    await expect(this.errorMessage).toBeVisible();
    if (message) {
      await expect(this.errorMessage).toContainText(message);
    }
  }
}

export class RegisterPage {
  readonly page: Page;
  readonly fullNameInput: Locator;
  readonly emailInput: Locator;
  readonly phoneInput: Locator;
  readonly propertyNameInput: Locator;
  readonly passwordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;
  readonly loginLink: Locator;

  constructor(page: Page) {
    this.page = page;
    this.fullNameInput = page.locator("#full_name");
    this.emailInput = page.locator("#email");
    this.phoneInput = page.locator("#phone");
    this.propertyNameInput = page.locator("#property_name");
    this.passwordInput = page.locator("#password");
    this.confirmPasswordInput = page.locator("#confirm_password");
    this.submitButton = page.locator('button[type="submit"]');
    this.errorMessage = page.locator('[class*="red"]');
    this.loginLink = page.locator('a[href="/login"]');
  }

  async goto() {
    await this.page.goto(TEST_CONFIG.ROUTES.REGISTER);
  }

  async register(data: {
    full_name: string;
    email: string;
    password: string;
    property_name: string;
    phone?: string;
  }) {
    await this.fullNameInput.fill(data.full_name);
    await this.emailInput.fill(data.email);
    if (data.phone) {
      await this.phoneInput.fill(data.phone);
    }
    await this.propertyNameInput.fill(data.property_name);
    await this.passwordInput.fill(data.password);
    await this.confirmPasswordInput.fill(data.password);
    await this.submitButton.click();
  }

  async expectError(message?: string) {
    await expect(this.errorMessage).toBeVisible();
    if (message) {
      await expect(this.errorMessage).toContainText(message);
    }
  }
}

export class DashboardPage {
  readonly page: Page;
  readonly sidebar: Locator;
  readonly mainContent: Locator;

  constructor(page: Page) {
    this.page = page;
    this.sidebar = page.locator("nav, [class*='sidebar']");
    this.mainContent = page.locator("main");
  }

  async goto() {
    await this.page.goto(TEST_CONFIG.ROUTES.DASHBOARD);
  }

  async expectLoaded() {
    await expect(this.page).toHaveURL(/dashboard/);
    await expect(this.mainContent).toBeVisible();
  }

  async navigateTo(route: string) {
    await this.page.locator(`a[href="${route}"]`).click();
    await this.page.waitForURL(`**${route}**`);
  }
}

export class DemoDataDialog {
  readonly page: Page;
  readonly dialog: Locator;
  readonly loadButton: Locator;
  readonly skipButton: Locator;
  readonly loadingIndicator: Locator;
  readonly successMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.dialog = page.locator("[data-testid='demo-data-dialog']");
    this.loadButton = page.locator("[data-testid='load-demo-data']");
    this.skipButton = page.locator("[data-testid='skip-demo-data']");
    this.loadingIndicator = page.locator("[data-testid='demo-loading']");
    this.successMessage = page.locator("[data-testid='demo-success']");
  }

  async expectVisible() {
    await expect(this.dialog).toBeVisible();
  }

  async loadDemoData() {
    await this.loadButton.click();
  }

  async skipDemoData() {
    await this.skipButton.click();
  }

  async waitForCompletion() {
    // Wait for loading to finish (up to 30s for demo data)
    await expect(this.loadingIndicator).toBeHidden({ timeout: 30_000 });
  }
}
