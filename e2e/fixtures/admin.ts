import { execSync } from 'node:child_process';
import path from 'node:path';
import { expect, type Page } from '@playwright/test';
import { gotoShop, waitForHtmx } from './shop';

const repoRoot = path.resolve(__dirname, '../..');

export type AdminUserType = 'superuser' | 'staff';

export type TestAdmin = {
	type: AdminUserType;
	email: string;
	password: string;
	fullName: string;
	homePath: string;
};

const defaultPassword = 'Password123';

const adminAccounts: Record<AdminUserType, Omit<TestAdmin, 'password'>> = {
	superuser: {
		type: 'superuser',
		email: 'e2e-superuser@cchoice.test',
		fullName: 'Test E2E Superuser',
		homePath: '/admin/superuser',
	},
	staff: {
		type: 'staff',
		email: 'e2e-staff@cchoice.test',
		fullName: 'Test E2E Staff',
		homePath: '/admin/staff',
	},
};

export function testAdminData(type: AdminUserType): TestAdmin {
	return {
		...adminAccounts[type],
		password: defaultPassword,
	};
}

export async function gotoAdminLogin(page: Page) {
	await gotoShop(page, '/admin');
}

export async function submitAdminLogin(page: Page, email: string, password: string) {
	await page.locator('#email').fill(email);
	await page.locator('#password').fill(password);
	await Promise.all([
		waitForHtmx(page, '/admin/login', 'POST'),
		page.getByRole('button', { name: 'Log In' }).click(),
	]);
}

export async function expectAdminWelcome(page: Page, fullName: string) {
	await expect(page.getByText(`Welcome, ${fullName}`)).toBeVisible();
}

export async function loginAdmin(page: Page, type: AdminUserType): Promise<TestAdmin> {
	const data = testAdminData(type);
	await gotoAdminLogin(page);
	await expect(page.getByRole('heading', { name: 'C-Choice Admin Portal' })).toBeVisible();
	await submitAdminLogin(page, data.email, data.password);
	await page.waitForURL(new RegExp(`${data.homePath.replace('/', '\\/')}$`));
	await expectAdminWelcome(page, data.fullName);
	return data;
}

export function resetStaffAttendance() {
	execSync('bash scripts/e2e-clear-staff-attendance.sh', { cwd: repoRoot });
}

export async function gotoStaffAttendance(page: Page) {
	await gotoShop(page, '/admin/staff/attendance');
	await expect(page.getByRole('button', { name: 'Time In' })).toBeVisible();
}

export async function punchAttendance(
	page: Page,
	buttonName: string,
	endpoint: string,
	successMessage: string,
) {
	await Promise.all([
		waitForHtmx(page, endpoint, 'POST'),
		page.getByRole('button', { name: buttonName }).click(),
	]);
	await page.waitForURL(/\/admin\/staff\/attendance/);
	await expect(page.locator('#success_banner_text')).toHaveText(successMessage);
}

export async function expectAttendanceButtonStates(
	page: Page,
	states: Record<string, boolean>,
) {
	for (const [name, enabled] of Object.entries(states)) {
		const button = page.getByRole('button', { name });
		if (enabled) {
			await expect(button).toBeEnabled();
		} else {
			await expect(button).toBeDisabled();
		}
	}
}

export async function expectAttendanceRowTime(page: Page, rowLabel: string) {
	const row = page.locator('#staff-attendance-table tr').filter({ hasText: rowLabel });
	await expect(row.locator('td').last()).toHaveText(/\d{2}:\d{2}:\d{2}/);
}
