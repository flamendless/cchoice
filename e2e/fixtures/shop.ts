import { expect, type Page } from '@playwright/test';

export function shopPath(path = '/') {
	if (path === '/') {
		return '/cchoice/';
	}
	return path.startsWith('/cchoice') ? path : `/cchoice${path}`;
}

export async function gotoShop(page: Page, path = '/') {
	await page.goto(shopPath(path));
}

export async function dismissSaleBanner(page: Page) {
	const banner = page.locator('#sale-banner');
	if (await banner.isVisible().catch(() => false)) {
		await page.getByRole('button', { name: 'Close banner' }).click();
		await expect(banner).toBeHidden();
	}
}

export async function waitForHtmx(page: Page, urlPart: string, method?: string) {
	await page.waitForResponse(
		(response) => {
			if (!response.url().includes(urlPart) || !response.ok()) {
				return false;
			}
			if (method && response.request().method() !== method) {
				return false;
			}
			return true;
		},
		{ timeout: 15_000 },
	);
}

export const defaultSearchQuery = 'GEX';

export async function searchProducts(page: Page, query = defaultSearchQuery) {
	const search = page.locator('#search');
	await search.click();
	await search.fill(query);
	await waitForHtmx(page, '/search', 'POST');
	await expect(page.locator('#search-results:not(.hidden) ul li a').first()).toBeVisible({
		timeout: 10_000,
	});
}

export async function openFirstSearchResult(page: Page) {
	const firstResult = page.locator('#search-results ul li a').first();
	await expect(firstResult).toBeVisible();
	const productName = (await firstResult.locator('p').textContent())?.trim() ?? '';
	await firstResult.click();
	await page.waitForURL(/\/product\/[^/]+$/);
	await expect(page.locator('#btn-add-to-cart')).toBeVisible();
	return productName;
}

export async function openProductViaSearch(page: Page, query = defaultSearchQuery) {
	await searchProducts(page, query);
	return openFirstSearchResult(page);
}

export async function addToCart(page: Page) {
	const addButton = page.locator('#btn-add-to-cart');
	await expect(addButton).toBeEnabled();
	await Promise.all([
		waitForHtmx(page, '/carts/lines', 'POST'),
		addButton.click(),
	]);
}

export async function openCart(page: Page) {
	await Promise.all([
		waitForHtmx(page, '/carts/lines', 'GET'),
		gotoShop(page, '/carts'),
	]);
	await expect(page.locator('#cart-lines h2').first()).toBeVisible({ timeout: 15_000 });
}
