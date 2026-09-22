import { test, expect } from '@playwright/test';
import {
	addToCart,
	dismissSaleBanner,
	gotoShop,
	openCart,
	openProductViaSearch,
	waitForHtmx,
} from '../fixtures/shop';

test.describe('cart', () => {
	test('adds, updates, and removes items', async ({ page }) => {
		await gotoShop(page, '/');
		await dismissSaleBanner(page);

		const productName = await openProductViaSearch(page);
		await addToCart(page);

		await expect(page.locator('#cart-count-desktop')).toHaveText('1');

		await openCart(page);
		await expect(page.locator('#cart-lines h2').first()).toContainText(productName);

		await Promise.all([
			waitForHtmx(page, '/carts/lines/', 'PATCH'),
			page.getByRole('button', { name: 'Increase quantity' }).first().click(),
		]);
		await expect(page.locator('#cart-lines p[name^="qty-"]').first()).toContainText('Qty: 2');

		await Promise.all([
			waitForHtmx(page, '/carts/lines/', 'DELETE'),
			page.getByRole('button', { name: 'Remove item in cart' }).first().click(),
		]);
		await expect(page.getByRole('heading', { name: 'Your cart is empty' })).toBeVisible();
	});
});
