import { test, expect } from '@playwright/test';
import {
	gotoAdminLogin,
	loginAdmin,
	logoutAdmin,
	submitAdminLogin,
	testAdminData,
} from '../fixtures/admin';
import { gotoShop } from '../fixtures/shop';

test.describe('admin login', () => {
	test('logs in a superuser', async ({ page }) => {
		await loginAdmin(page, 'superuser');
	});

	test('logs in a staff member', async ({ page }) => {
		await loginAdmin(page, 'staff');
	});

	test('rejects invalid password', async ({ page }) => {
		const admin = testAdminData('staff');
		await gotoAdminLogin(page);
		await submitAdminLogin(page, admin.email, 'WrongPassword123');
		await page.waitForURL(/\/admin\?error=/);
		await expect(page.locator('#error_banner_text')).toContainText('Invalid email or password');
	});

	test('redirects unauthenticated users from attendance page', async ({ page }) => {
		await gotoShop(page, '/admin/staff/attendance');
		await expect(page).toHaveURL(/\/cchoice\/admin$/);
		await expect(page.getByRole('heading', { name: 'C-Choice Admin Portal' })).toBeVisible();
		await expect(page.locator('#error_banner_text')).toContainText('Login to access page');
	});

	test('prevents staff from accessing superuser home', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoShop(page, '/admin/superuser');
		await expect(page).toHaveURL(/\/cchoice\/admin\/staff$/);
		await expect(page.locator('#error_banner_text')).toContainText('Login to access page');
	});

	test('logs out and clears session', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await logoutAdmin(page);
		await expect(page.getByRole('heading', { name: 'C-Choice Admin Portal' })).toBeVisible();

		await gotoShop(page, '/admin/staff/attendance');
		await expect(page).toHaveURL(/\/cchoice\/admin$/);
		await expect(page.locator('#error_banner_text')).toContainText('Login to access page');
	});
});
