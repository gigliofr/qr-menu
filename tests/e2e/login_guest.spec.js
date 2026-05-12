const { test, expect } = require('@playwright/test');

test('guest flow works', async ({ page }) => {
  await page.goto('/login');
  const guestBtn = page.locator('button:has-text("Continua come ospite")');
  await expect(guestBtn).toBeVisible();
  await guestBtn.click();
  await expect(page).toHaveURL(/.*guest.*|.*menu.*/);
});

test('login page accessible', async ({ page }) => {
  await page.goto('/login');
  const heading = page.locator('h1');
  await expect(heading).toContainText(/Login|Accedi/i);
});
