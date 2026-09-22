import { test, expect } from '@playwright/test';
import {
	expectPortalWelcome,
	registerCustomer,
	submitLogin,
} from '../fixtures/customer';

test.describe('customer registration', () => {
	test('registers an individual customer', async ({ page }) => {
		await registerCustomer(page, 'customer');
		await expect(page).toHaveURL(/\/customer$/);
	});

	test('registers a business customer', async ({ page }) => {
		await registerCustomer(page, 'company');
		await expect(page).toHaveURL(/\/customer$/);
	});
});

test.describe('customer login', () => {
	test('logs in an individual customer', async ({ page }) => {
		const data = await registerCustomer(page, 'customer');
		await submitLogin(page, data.email, data.password);
		await expectPortalWelcome(page, data.fullName);
	});

	test('logs in a business customer', async ({ page }) => {
		const data = await registerCustomer(page, 'company');
		await submitLogin(page, data.email, data.password);
		await expectPortalWelcome(page, data.fullName);
	});
});
