import { test, expect } from '@playwright/test';
import {
	fillTimeOffForm,
	gotoStaffTimeOff,
	loginAdmin,
	postTimeOffRequest,
	resetStaffTimeOff,
	submitTimeOffRequest,
	expectTimeOffTableRow,
	todayISO,
	tomorrowISO,
} from '../fixtures/admin';

test.describe('staff time off', () => {
	test.beforeEach(() => {
		resetStaffTimeOff();
	});

	test('submits a vacation leave request', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffTimeOff(page);

		await expect(page.locator('#time-off-table')).toContainText('No time off requests yet');

		const today = todayISO();
		await fillTimeOffForm(page, {
			type: 'VL',
			startDate: today,
			endDate: today,
			description: 'E2E vacation leave',
		});
		await submitTimeOffRequest(page);

		await expect(page.locator('#success_banner_text')).toHaveText(
			'Time off request submitted',
		);
		await expectTimeOffTableRow(page, 'E2E vacation leave');
		await expect(page.locator('#time-off-table')).toContainText('Vacation Leave');
	});

	test('rejects end date before start date', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffTimeOff(page);

		const response = await postTimeOffRequest(page, {
			type: 'VL',
			startDate: tomorrowISO(),
			endDate: todayISO(),
			description: 'Invalid date range',
		});
		expect(response.status()).toBe(400);
		await expect(response.text()).resolves.toContain('end date must not be before start date');
	});

	test('rejects missing time off type', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffTimeOff(page);

		const response = await postTimeOffRequest(page, {
			type: '',
			startDate: todayISO(),
			endDate: todayISO(),
			description: 'Missing type',
		});
		expect(response.status()).toBe(400);
		await expect(response.text()).resolves.toMatch(/type/i);
	});

	test('lists prior requests after page reload', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffTimeOff(page);

		const today = todayISO();
		await fillTimeOffForm(page, {
			type: 'SL',
			startDate: today,
			endDate: today,
			description: 'E2E sick leave',
		});
		await submitTimeOffRequest(page);
		await expectTimeOffTableRow(page, 'E2E sick leave');

		await page.reload();
		await expect(page.getByRole('heading', { name: 'Request Time Off' })).toBeVisible();
		await expectTimeOffTableRow(page, 'E2E sick leave');
		await expect(page.locator('#time-off-table')).toContainText('Sick Leave');
	});
});
