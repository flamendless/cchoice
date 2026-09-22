import { test, expect } from '@playwright/test';
import { dismissSaleBanner, gotoShop, openProductViaSearch } from '../fixtures/shop';

test.describe('product page', () => {
	test('shows product details and quantity controls', async ({ page }) => {
		await gotoShop(page, '/');
		await dismissSaleBanner(page);

		const productName = await openProductViaSearch(page);

		await expect(page.locator('h1')).toContainText(productName);
		await expect(page.locator('.text-3xl.font-bold.text-primary').first()).toBeVisible();
		await expect(page.locator('#btn-add-to-cart')).toBeEnabled();
		await expect(page.locator('#product-qty')).toHaveValue('1');

		await page.locator('#btn-qty-increase').click();
		await expect(page.locator('#product-qty')).toHaveValue('2');
	});
});
