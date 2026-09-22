import { test } from '@playwright/test';
import { loginAdmin } from '../fixtures/admin';

test.describe('admin login', () => {
	test('logs in a superuser', async ({ page }) => {
		await loginAdmin(page, 'superuser');
	});

	test('logs in a staff member', async ({ page }) => {
		await loginAdmin(page, 'staff');
	});
});
