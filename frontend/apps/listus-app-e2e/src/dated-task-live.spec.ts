import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

test.use({ trace: 'off' });

const required = (name: string): string => {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required for Listus acceptance`);
  return value;
};

interface Actor {
  readonly email: string;
  readonly password: string;
  readonly token: string;
}

async function authenticate(request: APIRequestContext): Promise<Actor> {
  const email = required('LISTUS_E2E_EMAIL');
  const password = required('LISTUS_E2E_PASSWORD');
  const response = await request.post(required('LISTUS_E2E_AUTH_SIGN_IN_URL'), {
    data: { email, password, returnSecureToken: true },
  });
  expect(response.ok(), `Authentication HTTP ${response.status()}`).toBeTruthy();
  const body = (await response.json()) as { idToken: string };
  return { email, password, token: body.idToken };
}

async function post<T>(
  request: APIRequestContext,
  actor: Actor,
  endpoint: string,
  data: unknown,
): Promise<T> {
  const response = await request.post(
    new URL(endpoint, required('LISTUS_E2E_API_BASE')).href,
    { headers: { Authorization: `Bearer ${actor.token}` }, data },
  );
  expect(response.ok(), `${endpoint} HTTP ${response.status()}`).toBeTruthy();
  return response.json();
}

async function signIn(page: Page, actor: Actor): Promise<void> {
  const apiBase = required('LISTUS_E2E_API_BASE');
  const appOrigin = new URL(required('BASE_URL')).origin;
  await page.context().addInitScript(
    ({ appOrigin, apiBase }) => {
      if (location.origin === appOrigin)
        sessionStorage.setItem('sneat-app:local-api-base', apiBase);
    },
    { appOrigin, apiBase },
  );
  await page.goto(`/login?apiBase=${encodeURIComponent(apiBase)}`);
  await page.locator('ion-segment-button[value="in"]').click();
  await page.locator('ion-input[name="email"] input').fill(actor.email);
  await page.locator('ion-input[type="password"] input').fill(actor.password);
  await page.getByRole('button', { name: /Sign in.*with password/ }).click();
  await page.waitForURL((url) => !url.pathname.includes('/login'));
}

test('real @authenticated Listus due task stays linked through its lifecycle', async ({
  page,
  request,
}) => {
  test.skip(
    process.env['LISTUS_E2E_DATED_TASK'] !== '1',
    'Requires the integrated local Listus and Calendarius stack',
  );
  test.setTimeout(120_000);
  const actor = await authenticate(request);
  const suffix = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  await post(request, actor, 'users/init_user_record', {
    email: actor.email,
    authProvider: 'password',
    ianaTimezone: 'Europe/Dublin',
  });
  const createdSpace = await post<{ space: { id: string } }>(
    request,
    actor,
    'spaces/create_space',
    {
      type: 'family',
      title: `Listus dated task ${suffix}`,
      requestID: `listus-dated-${suffix}`,
    },
  );
  const spaceID = createdSpace.space.id;
  const createdList = await post<{ id: string }>(request, actor, 'listus/create_list', {
    spaceID,
    type: 'do',
    title: `Payments ${suffix}`,
  });
  const listSubID = createdList.id.startsWith('do!')
    ? createdList.id.slice('do!'.length)
    : createdList.id;
  const listURL = `/space/family/${spaceID}/list/do/${listSubID}`;
  const title = `Renew insurance ${suffix}`;

  await signIn(page, actor);
  await page.goto(listURL);
  await page.locator('ion-input[placeholder="New item"] input').fill(title);
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  const row = page.locator('ion-reorder').filter({ hasText: title });
  await expect(row).toBeVisible();

  const otherTitle = `Second task ${suffix}`;
  await page.locator('ion-input[placeholder="New item"] input').fill(otherTitle);
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  const otherRow = page.locator('ion-reorder').filter({ hasText: otherTitle });
  await expect(otherRow).toBeVisible();

  const initialSave = page.waitForResponse(
    (response) => response.url().includes('/v0/listus/item_date_task_save'),
  );
  await row.getByLabel(`Add due date for ${title}`).fill('2026-09-21');
  expect((await initialSave).ok()).toBeTruthy();

  await page.locator('ion-select').filter({ hasText: /Swipe|Reorder/ }).click();
  await page.getByRole('radio', { name: 'Reorder', exact: true }).click();
  const reorder = page.waitForResponse(
    (response) => response.url().includes('/v0/listus/list_items_reorder'),
  );
  await row.dragTo(otherRow);
  expect((await reorder).ok()).toBeTruthy();
  await page.reload();
  await expect(
    page.locator('ion-reorder').filter({ hasText: title }),
  ).toBeVisible();

  await page.goto(`/space/family/${spaceID}/calendar?tab=day&date=2026-09-21`);
  await expect(page.getByText(title, { exact: true })).toBeVisible();

  await page.goto(listURL);
  const loadedRow = page.locator('ion-reorder').filter({ hasText: title });
  await expect(loadedRow.getByLabel(`Change due date for ${title}`)).toHaveValue(
    '2026-09-21',
  );
  const reschedule = page.waitForResponse(
    (response) => response.url().includes('/v0/listus/item_date_task_save'),
  );
  await loadedRow.getByLabel(`Change due date for ${title}`).fill('2026-09-23');
  expect((await reschedule).ok()).toBeTruthy();

  await page.goto(`/space/family/${spaceID}/calendar?tab=day&date=2026-09-21`);
  await expect(page.getByText(title, { exact: true })).toHaveCount(0);
  await page.goto(`/space/family/${spaceID}/calendar?tab=day&date=2026-09-23`);
  await expect(page.getByText(title, { exact: true })).toBeVisible();

  await page.goto(listURL);
  const currentRow = page.locator('ion-reorder').filter({ hasText: title });
  await currentRow.locator('ion-checkbox').click();
  await expect(currentRow.locator('ion-checkbox')).toBeChecked();
  await currentRow.locator('ion-checkbox').click();
  await expect(currentRow.locator('ion-checkbox')).not.toBeChecked();

  const clearDue = page.waitForResponse(
    (response) => response.url().includes('/v0/listus/item_date_task_save'),
  );
  await currentRow.getByRole('button', { name: `Remove due date for ${title}` }).click();
  expect((await clearDue).ok()).toBeTruthy();
  await expect(currentRow).toBeVisible();
  await expect(currentRow.getByLabel(`Add due date for ${title}`)).toBeVisible();
});
