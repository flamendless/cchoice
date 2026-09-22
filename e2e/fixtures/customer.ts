import { expect, type Page } from '@playwright/test';
import { gotoShop, waitForHtmx } from './shop';

export type CustomerType = 'customer' | 'company';

export type TestCustomer = {
	type: CustomerType;
	email: string;
	password: string;
	firstName: string;
	middleName: string;
	lastName: string;
	fullName: string;
	companyName?: string;
};

const defaultPassword = 'Password123';
const defaultBirthdate = '1990-01-15';
const defaultMobile = '9171234567';

export function testCustomerData(type: CustomerType): TestCustomer {
	const stamp = Date.now();
	const firstName = 'Test';
	const middleName = 'E2E';
	const lastName = type === 'company' ? 'Company' : 'Customer';
	const companyName = type === 'company' ? 'E2E Test Corp' : undefined;

	return {
		type,
		email: `e2e-${type}-${stamp}@cchoice.test`,
		password: defaultPassword,
		firstName,
		middleName,
		lastName,
		fullName: `${firstName} ${middleName} ${lastName}`,
		companyName,
	};
}

export async function gotoRegister(page: Page) {
	await gotoShop(page, '/customer/register');
}

export async function gotoLogin(page: Page) {
	await gotoShop(page, '/customer');
}

export async function fillRegistrationForm(page: Page, data: TestCustomer) {
	await page.selectOption('#customer_type', data.type);
	await expect(page.locator('#register-form-fields')).toBeVisible();

	if (data.type === 'company' && data.companyName) {
		await expect(page.locator('#company-name-group')).toBeVisible();
		await page.locator('#company_name').fill(data.companyName);
	}

	await page.locator('#first_name').fill(data.firstName);
	await page.locator('#middle_name').fill(data.middleName);
	await page.locator('#last_name').fill(data.lastName);
	await page.locator('#birthdate').fill(defaultBirthdate);
	await page.selectOption('#sex', 'male');
	await page.locator('#email').fill(data.email);
	await page.locator('#mobile_no').fill(defaultMobile);
	await page.locator('#password').fill(data.password);
	await page.locator('#confirm_password').fill(data.password);
}

const registrationSuccessMessage = 'Registration successful! Please log in.';

export async function submitRegistration(page: Page) {
	await Promise.all([
		waitForHtmx(page, '/customer/register', 'POST'),
		page.getByRole('button', { name: 'Register' }).click(),
	]);
	await page.waitForURL(/\/customer/);
	await expect(page.getByRole('heading', { name: 'C-Choice Customer Portal' })).toBeVisible();
	await expect(page.getByText(registrationSuccessMessage)).toBeVisible();
}

export async function registerCustomer(page: Page, type: CustomerType): Promise<TestCustomer> {
	const data = testCustomerData(type);
	await gotoRegister(page);
	await fillRegistrationForm(page, data);
	await submitRegistration(page);
	return data;
}

export async function submitLogin(page: Page, email: string, password: string) {
	await page.locator('#email').fill(email);
	await page.locator('#password').fill(password);
	await Promise.all([
		waitForHtmx(page, '/customer/login', 'POST'),
		page.getByRole('button', { name: 'Log In' }).click(),
	]);
	await page.waitForURL(/\/customer\/portal/);
}

export async function expectPortalWelcome(page: Page, fullName: string) {
	await expect(page.getByText(`Welcome, ${fullName}`)).toBeVisible();
}

export async function loginCustomer(page: Page, data: TestCustomer) {
	await gotoLogin(page);
	await submitLogin(page, data.email, data.password);
	await expectPortalWelcome(page, data.fullName);
}
