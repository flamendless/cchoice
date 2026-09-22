import { test, expect } from '@playwright/test';
import {
	gotoSuperuserAttendance,
	loginAdmin,
	logoutAdmin,
	punchAttendance,
	resetStaffAttendance,
	testAdminData,
} from '../fixtures/admin';
import { gotoShop } from '../fixtures/shop';

test.describe('superuser attendance', () => {
	test.beforeEach(() => {
		resetStaffAttendance();
	});

	test('loads the attendance report page', async ({ page }) => {
		await loginAdmin(page, 'superuser');
		await gotoSuperuserAttendance(page);
		await expect(page.locator('#attendance-table table')).toBeVisible();
	});

	test('shows staff punch in the report', async ({ page }) => {
		const staff = testAdminData('staff');

		await loginAdmin(page, 'staff');
		await gotoShop(page, '/admin/staff/attendance');
		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await logoutAdmin(page);

		await loginAdmin(page, 'superuser');
		await gotoSuperuserAttendance(page);

		const table = page.locator('#attendance-table table');
		await expect(table).toBeVisible();
		await expect(table).toContainText(staff.fullName);
		await expect(table.locator('tbody tr').first()).toContainText(/\d{2}:\d{2}:\d{2}/);
	});
});
