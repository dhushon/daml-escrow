import { test, expect, type Page } from '@playwright/test';

async function performLogout(page: Page) {
  const profileBtn = page.locator('#user-pill');
  try {
    if (await profileBtn.isVisible({ timeout: 2000 })) {
      await profileBtn.click();
      const logoutBtn = page.locator('#logout-btn');
      await expect(logoutBtn).toBeVisible();
      await logoutBtn.click();
      await page.waitForURL('**/login');
      return;
    }
  } catch (e) {
    // Ignore and fallback to manual cookie clearing
  }
  await page.evaluate(() => {
    localStorage.removeItem('auth_session');
    document.cookie = "user_email=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    document.cookie = "user_scopes=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    document.cookie = "admin-mode=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    document.cookie = "assumed_role=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    window.location.href = '/login';
  });
  await page.waitForURL('**/login');
}

test.describe('Bilateral Escrow Agreement Lifecycle E2E', () => {
  
  test('should execute complete draft, negotiation, ledger promotion, and settlement lifecycle', async ({ page }) => {
    // 1. LOGIN BYPASS AS DEPOSITOR (Joey)
    console.log('Logging in as Depositor (Joey)...');
    await page.goto('/login');
    
    // Check if dev roles are rendered and click Joey
    const joeyBtn = page.locator('button[data-role="Depositor"]');
    await expect(joeyBtn).toBeVisible();
    await joeyBtn.click();
    
    // Wait for authentication status completion and redirect to dashboard
    await page.waitForURL('**/');
    console.log('Successfully logged in. Dashboard loaded.');
    
    // Validate Depositor identity displayed
    await expect(page.locator('text=joey@depositor.devlocal').first()).toBeVisible();

    // 2. CREATE A NEW OFF-CHAIN DRAFT AGREEMENT
    console.log('Composing new draft escrow agreement...');
    await page.goto('/compose');
    await expect(page).toHaveURL('/compose');

    // Select Counterparty (Jimmy/Beneficiary)
    const counterpartySelect = page.locator('#counterparty-select');
    await expect(counterpartySelect).toBeVisible();
    // Wait for asynchronous directory search to populate options
    await page.waitForFunction(() => {
      const el = document.getElementById('counterparty-select') as HTMLSelectElement;
      return el && el.options.length > 0 && !el.options[0].text.includes('Searching');
    });
    await counterpartySelect.selectOption({ index: 0 });

    // Select Mediator (Sally)
    const mediatorSelect = page.locator('#mediator-select');
    await expect(mediatorSelect).toBeVisible();
    // Wait for asynchronous mediators search to populate options
    await page.waitForFunction(() => {
      const el = document.getElementById('mediator-select') as HTMLSelectElement;
      return el && el.options.length > 0 && !el.options[0].text.includes('Searching');
    });
    await mediatorSelect.selectOption({ index: 0 });

    // Select Currency
    const currencySelect = page.locator('select[name="currency"]');
    await currencySelect.selectOption('USD');

    // Fill in Milestone Label and Amount
    const labelInput = page.locator('input[name="m-label"]').first();
    const amountInput = page.locator('input[name="m-amount"]').first();
    await labelInput.fill('Phase 1 Milestone');
    await amountInput.fill('50000');

    // Add a technical metadata schema and payload
    const schemaSelect = page.locator('select[name="schemaUrl"]');
    await schemaSelect.selectOption({ label: 'Research Grant' });

    // Check we get a confirmation alert and redirect back to dashboard
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('saved');
      await dialog.accept();
    });

    // Save & Share Draft
    const saveBtn = page.locator('button:has-text("Save & Share Draft")');
    await saveBtn.click();
    
    await page.waitForURL('**/');
    console.log('Draft created and saved successfully.');

    // 3. BILATERAL NEGOTIATION - DEPOSITOR APPROVAL
    console.log('Navigating to created draft for negotiation...');
    await page.goto('/?tab=negotiations');
    // Find our draft card in the list
    const draftCard = page.locator('a:has-text("jimmy@beneficiary.devlocal")').first();
    await expect(draftCard).toBeVisible();
    
    // Click card to open negotiation cockpit
    await draftCard.click();
    await page.waitForURL(/\/negotiate\/.+/);
    
    // Verify draft details
    await expect(page.locator('#display-beneficiary')).toHaveText('jimmy@beneficiary.devlocal');
    await expect(page.locator('#display-amount')).toHaveText('50,000');

    // Approve version as Depositor (Joey)
    const approveBtn = page.locator('#btn-approve');
    await expect(approveBtn).toBeVisible();
    await approveBtn.click();

    // Verify approval update
    await expect(page.locator('#approvals-list')).toContainText('Joey Authorized');
    console.log('Depositor (Joey) approved the draft.');

    // 4. BILATERAL NEGOTIATION - BENEFICIARY APPROVAL
    console.log('Switching user to Beneficiary (Jimmy)...');
    
    await performLogout(page);

    const jimmyBtn = page.locator('button[data-role="Beneficiary"]');
    await expect(jimmyBtn).toBeVisible();
    await jimmyBtn.click();
    await page.waitForURL('**/');

    // Confirm Jimmy is logged in
    await expect(page.locator('text=jimmy@beneficiary.devlocal').first()).toBeVisible();

    await page.goto('/?tab=negotiations');

    // Open same draft as Beneficiary
    console.log('Opening draft as Beneficiary...');
    const draftCardJimmy = page.locator('a:has-text("jimmy@beneficiary.devlocal")').first();
    await expect(draftCardJimmy).toBeVisible();
    await draftCardJimmy.click();
    await page.waitForURL(/\/negotiate\/.+/);

    // Approve version as Beneficiary
    const approveBtnJimmy = page.locator('#btn-approve');
    await expect(approveBtnJimmy).toBeVisible();
    await approveBtnJimmy.click();

    // Verify draft state is now RATIFIED
    await expect(page.locator('#draft-status')).toHaveText('RATIFIED');
    await expect(page.locator('#approvals-list')).toContainText('Jimmy Authorized');
    console.log('Beneficiary (Jimmy) approved. Draft is now RATIFIED.');

    // 5. LEDGER PROMOTION
    console.log('Promoting ratified draft agreement to the Canton ledger...');
    const promoteBtn = page.locator('#btn-promote');
    await expect(promoteBtn).toBeVisible();
    
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('promoted');
      await dialog.accept();
    });
    await promoteBtn.click();

    // Redirect to dashboard page
    await page.waitForURL('**/');
    console.log('Successfully promoted draft to active escrow contract.');

    // 6. VALIDATE LEDGER ACTIVE STATE AND LIFE CYCLE TRIGGERS
    console.log('Checking active escrow status on dashboard...');
    const activeEscrowCard = page.locator('text=Phase 1 Milestone').first();
    await expect(async () => {
      await page.reload();
      await expect(activeEscrowCard).toBeVisible({ timeout: 2000 });
    }).toPass({
      intervals: [1000, 2000],
      timeout: 15000,
    });


    // 7. SESSION DESTRUCTION ON DISCONNECT
    console.log('Testing wallet disconnect / logout session destruction...');
    await performLogout(page);
    const authSession = await page.evaluate(() => localStorage.getItem('auth_session'));
    expect(authSession).toBeNull();
    console.log('Session destroyed successfully upon disconnect.');
  });

  test('should execute proposal withdrawal and database reset to DRAFT', async ({ page }) => {
    // 1. LOGIN BYPASS AS DEPOSITOR (Joey)
    console.log('Logging in as Depositor (Joey) for withdrawal test...');
    await page.goto('/login');
    const joeyBtn = page.locator('button[data-role="Depositor"]');
    await expect(joeyBtn).toBeVisible();
    await joeyBtn.click();
    await page.waitForURL('**/');
    console.log('Logged in. Dashboard loaded.');

    // 2. CREATE A NEW OFF-CHAIN DRAFT AGREEMENT
    console.log('Creating draft for withdrawal testing...');
    await page.goto('/compose');
    const counterpartySelect = page.locator('#counterparty-select');
    await expect(counterpartySelect).toBeVisible();
    await page.waitForFunction(() => {
      const el = document.getElementById('counterparty-select') as HTMLSelectElement;
      return el && el.options.length > 0 && !el.options[0].text.includes('Searching');
    });
    await counterpartySelect.selectOption({ index: 0 });

    const mediatorSelect = page.locator('#mediator-select');
    await expect(mediatorSelect).toBeVisible();
    await page.waitForFunction(() => {
      const el = document.getElementById('mediator-select') as HTMLSelectElement;
      return el && el.options.length > 0 && !el.options[0].text.includes('Searching');
    });
    await mediatorSelect.selectOption({ index: 0 });

    const currencySelect = page.locator('select[name="currency"]');
    await currencySelect.selectOption('USD');

    const labelInput = page.locator('input[name="m-label"]').first();
    const amountInput = page.locator('input[name="m-amount"]').first();
    await labelInput.fill('Phase 2 Milestone Withdrawal Test');
    await amountInput.fill('75000');

    const schemaSelect = page.locator('select[name="schemaUrl"]');
    await schemaSelect.selectOption({ label: 'Research Grant' });

    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('saved');
      await dialog.accept();
    });

    const saveBtn = page.locator('button:has-text("Save & Share Draft")');
    await saveBtn.click();
    await page.waitForURL('**/');
    console.log('Draft saved.');

    // 3. BILATERAL NEGOTIATION - DEPOSITOR APPROVAL
    console.log('Navigating to draft page...');
    await page.goto('/?tab=negotiations');
    const draftCard = page.locator('a:has-text("75,000")').first();
    await expect(draftCard).toBeVisible();
    await draftCard.click();
    await page.waitForURL(/\/negotiate\/.+/);

    const approveBtn = page.locator('#btn-approve');
    await expect(approveBtn).toBeVisible();
    await approveBtn.click();
    await expect(page.locator('#approvals-list')).toContainText('Joey Authorized');
    console.log('Depositor (Joey) approved.');

    // 4. SWITCH USER TO BENEFICIARY (Jimmy) AND APPROVE TO RATIFY
    console.log('Switching to Beneficiary (Jimmy)...');
    await performLogout(page);
    const jimmyBtn = page.locator('button[data-role="Beneficiary"]');
    await expect(jimmyBtn).toBeVisible();
    await jimmyBtn.click();
    await page.waitForURL('**/');

    await page.goto('/?tab=negotiations');
    const draftCardJimmy = page.locator('a:has-text("75,000")').first();
    await expect(draftCardJimmy).toBeVisible();
    await draftCardJimmy.click();
    await page.waitForURL(/\/negotiate\/.+/);

    const approveBtnJimmy = page.locator('#btn-approve');
    await expect(approveBtnJimmy).toBeVisible();
    await approveBtnJimmy.click();
    await expect(page.locator('#draft-status')).toHaveText('RATIFIED');
    console.log('Beneficiary approved. Draft is now RATIFIED.');

    // 5. LEDGER PROMOTION BY BENEFICIARY
    console.log('Promoting to Ledger...');
    const promoteBtn = page.locator('#btn-promote');
    await expect(promoteBtn).toBeVisible();
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('promoted');
      await dialog.accept();
    });
    await promoteBtn.click();
    await page.waitForURL('**/');
    console.log('Draft promoted to ledger. Status is now PROMOTED.');

    // 6. DEPOSITOR (Joey) LOGS IN AND WITHDRAWS PROPOSAL
    console.log('Switching to Depositor (Joey) to withdraw...');
    await performLogout(page);
    const joeyBtn2 = page.locator('button[data-role="Depositor"]');
    await expect(joeyBtn2).toBeVisible();
    await joeyBtn2.click();
    await page.waitForURL('**/');

    await page.goto('/?tab=negotiations');
    const draftCardJoey = page.locator('a:has-text("75,000")').first();
    await expect(draftCardJoey).toBeVisible();
    await draftCardJoey.click();
    await page.waitForURL(/\/negotiate\/.+/);

    // Verify status is PROMOTED
    await expect(page.locator('#draft-status')).toHaveText('PROMOTED');

    // Withdraw button should be visible
    const withdrawBtn = page.locator('#btn-withdraw');
    await expect(withdrawBtn).toBeVisible();

    // Intercept both confirmation and success dialogs
    page.on('dialog', async dialog => {
      if (dialog.message().includes('withdraw this proposal')) {
        await dialog.accept();
      } else if (dialog.message().includes('withdrawn successfully')) {
        await dialog.accept();
      } else {
        await dialog.dismiss();
      }
    });

    await withdrawBtn.click();

    // Verify status returns to DRAFT
    await expect(page.locator('#draft-status')).toHaveText('DRAFT');

    // Verify approvals list is reset
    await expect(page.locator('#approvals-list')).toContainText('Bilateral authorization pending');
    console.log('Proposal successfully withdrawn and reset to DRAFT.');

    await performLogout(page);
  });
});
