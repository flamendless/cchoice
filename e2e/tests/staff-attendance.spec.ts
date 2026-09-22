import { test, expect } from '@playwright/test';
import {
	expectAdminWelcome,
	expectAttendanceButtonStates,
	expectAttendanceDurationNotDash,
	expectAttendanceRowDash,
	expectAttendanceRowTime,
	gotoStaffAttendance,
	loginAdmin,
	postAttendanceAction,
	punchAttendance,
	resetStaffAttendance,
	selectAttendanceDate,
	todayISO,
	yesterdayISO,
} from '../fixtures/admin';

test.describe('staff attendance', () => {
	test.beforeEach(() => {
		resetStaffAttendance();
	});

	test('records time in, lunch break, and time out', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await expect(page.locator('#staff-attendance-table')).toContainText(
			'No attendance record for this date',
		);
		await expectAttendanceButtonStates(page, {
			'Time In': true,
			'Time Out': false,
			'Start Lunch Break': false,
			'End Lunch Break': false,
		});

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': true,
			'Start Lunch Break': true,
			'End Lunch Break': false,
		});
		await expectAttendanceRowTime(page, 'Time In');

		await punchAttendance(
			page,
			'Start Lunch Break',
			'/admin/staff/lunch-break-start',
			'Lunch break start recorded successfully',
		);
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': true,
			'Start Lunch Break': false,
			'End Lunch Break': true,
		});
		await expectAttendanceRowTime(page, 'Lunch Break Start');

		await punchAttendance(
			page,
			'End Lunch Break',
			'/admin/staff/lunch-break-end',
			'Lunch break end recorded successfully',
		);
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': true,
			'Start Lunch Break': false,
			'End Lunch Break': false,
		});
		await expectAttendanceRowTime(page, 'Lunch Break End');

		await punchAttendance(
			page,
			'Time Out',
			'/admin/staff/time-out',
			'Time out recorded successfully',
		);
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': false,
			'Start Lunch Break': false,
			'End Lunch Break': false,
		});
		await expectAttendanceRowTime(page, 'Time Out');
	});

	test('skips lunch and clocks out directly', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await punchAttendance(
			page,
			'Time Out',
			'/admin/staff/time-out',
			'Time out recorded successfully',
		);

		await expectAttendanceRowDash(page, 'Lunch Break Start');
		await expectAttendanceRowDash(page, 'Lunch Break End');
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': false,
		});
	});

	test('clocks out during lunch without ending break', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await punchAttendance(
			page,
			'Start Lunch Break',
			'/admin/staff/lunch-break-start',
			'Lunch break start recorded successfully',
		);
		await punchAttendance(
			page,
			'Time Out',
			'/admin/staff/time-out',
			'Time out recorded successfully',
		);

		await expectAttendanceRowTime(page, 'Lunch Break Start');
		await expectAttendanceRowDash(page, 'Lunch Break End');
		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': false,
			'Start Lunch Break': false,
			'End Lunch Break': false,
		});
	});

	test('shows partial day with time in only', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);

		await expectAttendanceButtonStates(page, {
			'Time In': false,
			'Time Out': true,
			'Start Lunch Break': true,
			'End Lunch Break': false,
		});
		await expectAttendanceRowTime(page, 'Time In');
		await expectAttendanceRowDash(page, 'Time Out');
		const durationRow = page
			.locator('#staff-attendance-table tr')
			.filter({ hasText: 'Duration' })
			.first();
		await expect(durationRow.locator('td').last()).toHaveText('-');
	});

	test('filters attendance by date', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await expectAttendanceRowTime(page, 'Time In');

		await selectAttendanceDate(page, yesterdayISO());
		await expect(page.locator('#staff-attendance-table')).toContainText(
			'No attendance record for this date',
		);

		await selectAttendanceDate(page, todayISO());
		await expectAttendanceRowTime(page, 'Time In');
	});

	test('shows duration after clocking out', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);
		await punchAttendance(
			page,
			'Time Out',
			'/admin/staff/time-out',
			'Time out recorded successfully',
		);

		await expectAttendanceDurationNotDash(page);
	});

	test('navigates back to staff home', async ({ page }) => {
		const admin = await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await page.getByRole('link', { name: 'Back to Home' }).click();
		await page.waitForURL(/\/admin\/staff$/);
		await expectAdminWelcome(page, admin.fullName);
	});

	test('rejects double time in via API', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		await punchAttendance(
			page,
			'Time In',
			'/admin/staff/time-in',
			'Time in recorded successfully',
		);

		const response = await postAttendanceAction(page, '/admin/staff/time-in');
		expect(response.status()).toBe(400);
	});

	test('rejects time out without time in via API', async ({ page }) => {
		await loginAdmin(page, 'staff');
		await gotoStaffAttendance(page);

		const response = await postAttendanceAction(page, '/admin/staff/time-out');
		expect(response.status()).toBe(400);
	});
});
