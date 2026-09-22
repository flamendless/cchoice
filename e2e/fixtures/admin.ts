import { execSync } from 'node:child_process';
import path from 'node:path';
import { expect, type Page } from '@playwright/test';
import { gotoShop, shopPath, waitForHtmx } from './shop';

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

export type StaffDataScope = 'attendance' | 'time-off' | 'all';

function clearStaffData(scope: StaffDataScope = 'all') {
	execSync(`bash scripts/e2e-clear-staff-data.sh ${scope}`, { cwd: repoRoot });
}

export function resetStaffData(scope: StaffDataScope = 'all') {
	clearStaffData(scope);
}

export function resetStaffAttendance() {
	clearStaffData('attendance');
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

export async function expectAttendanceRowDash(page: Page, rowLabel: string) {
	const row = page.locator('#staff-attendance-table tr').filter({ hasText: rowLabel });
	await expect(row.locator('td').last()).toHaveText('-');
}

export async function expectAttendanceDurationNotDash(page: Page) {
	const row = page.locator('#staff-attendance-table tr').filter({ hasText: 'Duration' }).first();
	await expect(row.locator('td').last()).not.toHaveText('-');
}

export function toISODate(date: Date): string {
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');
	return `${year}-${month}-${day}`;
}

export function todayISO(): string {
	return toISODate(new Date());
}

export function yesterdayISO(): string {
	const date = new Date();
	date.setDate(date.getDate() - 1);
	return toISODate(date);
}

export function tomorrowISO(): string {
	const date = new Date();
	date.setDate(date.getDate() + 1);
	return toISODate(date);
}

export async function selectAttendanceDate(page: Page, isoDate: string) {
	await Promise.all([
		waitForHtmx(page, '/admin/staff/attendance/rows', 'GET'),
		page.locator('#date-selector').fill(isoDate),
	]);
}

export async function postAttendanceAction(page: Page, endpoint: string) {
	return page.request.post(shopPath(endpoint));
}

export async function logoutAdmin(page: Page) {
	const response = await page.request.post(shopPath('/admin/logout'));
	expect(response.ok()).toBeTruthy();
	await gotoAdminLogin(page);
}

export function resetStaffTimeOff() {
	clearStaffData('time-off');
}

export type TimeOffRequest = {
	type: string;
	startDate: string;
	endDate: string;
	description: string;
};

export async function gotoStaffTimeOff(page: Page) {
	await gotoShop(page, '/admin/staff/time-off');
	await expect(page.getByRole('heading', { name: 'Request Time Off' })).toBeVisible();
}

export async function fillTimeOffForm(page: Page, request: TimeOffRequest) {
	if (request.type) {
		await page.locator('#type').selectOption(request.type);
	}
	await page.locator('#start-date').fill(request.startDate);
	await page.locator('#end-date').fill(request.endDate);
	await page.locator('#description').fill(request.description);
}

export async function submitTimeOffRequest(page: Page) {
	await Promise.all([
		waitForHtmx(page, '/admin/staff/time-off', 'POST'),
		page.getByRole('button', { name: 'Submit' }).click(),
	]);
	await page.waitForURL(/\/admin\/staff\/time-off/);
}

export async function postTimeOffRequest(page: Page, request: TimeOffRequest) {
	return page.request.post(shopPath('/admin/staff/time-off'), {
		form: {
			type: request.type,
			'start-date': request.startDate,
			'end-date': request.endDate,
			description: request.description,
		},
	});
}

export async function expectTimeOffTableRow(page: Page, description: string) {
	await expect(page.locator('#time-off-table')).toContainText(description);
}

export async function gotoSuperuserAttendance(page: Page) {
	await Promise.all([
		waitForHtmx(page, '/admin/superuser/attendance/table', 'GET'),
		gotoShop(page, '/admin/superuser/attendance'),
	]);
	await expect(page.getByRole('heading', { name: 'Employee Attendance' })).toBeVisible();
}
