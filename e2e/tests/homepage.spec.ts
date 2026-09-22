import { test, expect } from '@playwright/test';
import { dismissSaleBanner, gotoShop } from '../fixtures/shop';

test.describe('homepage', () => {
	test('loads with header and category sections', async ({ page }) => {
		await gotoShop(page, '/');
		await dismissSaleBanner(page);

		await expect(page).toHaveTitle(/.+/);
		await expect(page.locator('#search-wrapper')).toBeVisible();
		await expect(page.locator('#category-sections')).toBeVisible();
		await expect(page.locator('[data-category-product-section]').first()).toBeVisible();
	});
});
