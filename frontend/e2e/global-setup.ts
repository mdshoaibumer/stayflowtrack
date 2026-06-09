/**
 * Global Setup for Playwright E2E tests
 * Ensures backend is reachable before running tests
 */

import { checkHealth } from "./helpers/api-helpers";

async function globalSetup() {
  console.log("🔍 Checking backend health...");

  const maxRetries = 10;
  for (let i = 0; i < maxRetries; i++) {
    const healthy = await checkHealth();
    if (healthy) {
      console.log("✅ Backend is healthy");
      return;
    }
    console.log(`⏳ Backend not ready, retrying... (${i + 1}/${maxRetries})`);
    await new Promise((r) => setTimeout(r, 3000));
  }

  throw new Error(
    "Backend is not reachable. Please start the backend server before running E2E tests.\n" +
      "Run: cd backend && make run"
  );
}

export default globalSetup;
