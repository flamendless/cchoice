import { test, expect } from '@playwright/test';
import { gotoShop } from '../fixtures/shop';

test.describe('static pages', () => {
	test('terms page renders', async ({ page }) => {
		await gotoShop(page, '/terms');
		await expect(page.getByRole('heading', { name: 'Terms and Conditions' })).toBeVisible();
	});

	test('privacy page renders', async ({ page }) => {
		await gotoShop(page, '/privacy');
		await expect(page.getByRole('heading', { name: 'Privacy Policy' })).toBeVisible();
	});

	test('return policy page renders', async ({ page }) => {
		await gotoShop(page, '/return');
		await expect(page.getByRole('heading', { name: 'Return Policy' })).toBeVisible();
	});

	test('unknown path shows 404 page', async ({ page }) => {
		await gotoShop(page, '/this-page-does-not-exist');
		await expect(page.getByRole('heading', { name: 'Page Not Found' })).toBeVisible();
	});
});
