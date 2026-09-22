import { test, expect } from '@playwright/test';
import { defaultSearchQuery, dismissSaleBanner, gotoShop, searchProducts } from '../fixtures/shop';

test.describe('search', () => {
	test('finds products and navigates to product page', async ({ page }) => {
		await gotoShop(page, '/');
		await dismissSaleBanner(page);

		await searchProducts(page, defaultSearchQuery);

		const firstResult = page.locator('#search-results ul li a').first();
		await expect(firstResult).toBeVisible();
		await firstResult.click();

		await page.waitForURL(/\/product\//);
		await expect(page.locator('#btn-add-to-cart')).toBeVisible();
	});
});
