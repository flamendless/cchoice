import { test, expect } from '@playwright/test';
import {
	expectAttendanceButtonStates,
	expectAttendanceRowTime,
	gotoStaffAttendance,
	loginAdmin,
	punchAttendance,
	resetStaffAttendance,
} from '../fixtures/admin';

test.describe('staff attendance', () => {
	test.beforeEach(() => {
		resetStaffAttendance();
	});

	test('logs in and records time in, lunch break, and time out', async ({ page }) => {
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
});
