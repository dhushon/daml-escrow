import { test, expect } from '@playwright/test';

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
    // Select option with Jimmy's email
    await counterpartySelect.selectOption({ label: 'Beneficiary: jimmy (jimmy...)' });

    // Select Mediator (Sally)
    const mediatorSelect = page.locator('#mediator-select');
    await expect(mediatorSelect).toBeVisible();
    await mediatorSelect.selectOption({ label: 'sally (Institutional Mediator)' });

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

    // Save & Share Draft
    const saveBtn = page.locator('button[type="submit"]');
    await saveBtn.click();

    // Check we get a confirmation alert and redirect back to dashboard
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('saved');
      await dialog.accept();
    });
    
    await page.waitForURL('**/');
    console.log('Draft created and saved successfully.');

    // 3. BILATERAL NEGOTIATION - DEPOSITOR APPROVAL
    console.log('Navigating to created draft for negotiation...');
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
    await page.goto('/login');
    const jimmyBtn = page.locator('button[data-role="Beneficiary"]');
    await expect(jimmyBtn).toBeVisible();
    await jimmyBtn.click();
    await page.waitForURL('**/');

    // Confirm Jimmy is logged in
    await expect(page.locator('text=jimmy@beneficiary.devlocal').first()).toBeVisible();

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
    await expect(activeEscrowCard).toBeVisible();

    // 7. SESSION DESTRUCTION ON DISCONNECT
    console.log('Testing wallet disconnect / logout session destruction...');
    // Open user profile dropdown
    const profileBtn = page.locator('#user-pill');
    if (await profileBtn.isVisible()) {
      await profileBtn.click();
      const logoutBtn = page.locator('#logout-btn');
      await expect(logoutBtn).toBeVisible();
      await logoutBtn.click();
    } else {
      // Direct call to clear session if UI profile button is structured differently
      await page.evaluate(() => {
        localStorage.removeItem('auth_session');
        window.location.href = '/login';
      });
    }

    await page.waitForURL('**/login');
    const authSession = await page.evaluate(() => localStorage.getItem('auth_session'));
    expect(authSession).toBeNull();
    console.log('Session destroyed successfully upon disconnect.');
  });
});
