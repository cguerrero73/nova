import { Page, Locator, expect } from '@playwright/test';

/**
 * Page Object for the query-builder modal in the users grid.
 *
 * Encapsulates selectors and user-facing flows so specs stay short and
 * resilient to UI refactors.
 */
export class QueryBuilderPage {
  readonly openButton: Locator;
  readonly modal: Locator;
  readonly querySelect: Locator;
  readonly sortTab: Locator;
  readonly filterTab: Locator;
  readonly sortFieldSelect: Locator;
  readonly sortDirectionSelect: Locator;
  readonly addSortButton: Locator;
  readonly filterFieldSelect: Locator;
  readonly filterOperatorSelect: Locator;
  readonly filterValueInput: Locator;
  readonly addFilterButton: Locator;
  readonly nameInput: Locator;
  readonly saveButton: Locator;
  readonly cancelButton: Locator;

  constructor(private readonly page: Page) {
    this.openButton = page.getByRole('button', { name: /editar query/i });
    this.modal = page.locator('div.fixed.inset-0').filter({ hasText: /editar query|nueva query|copiar query/i });
    this.querySelect = this.modal.locator('select').first();
    this.sortTab = this.modal.getByRole('button', { name: /ordenamiento/i });
    this.filterTab = this.modal.getByRole('button', { name: /filtros/i });
    this.sortFieldSelect = this.modal.locator('select').nth(1); // field dropdown in sort tab
    this.sortDirectionSelect = this.modal.locator('select').nth(2);
    this.addSortButton = this.modal.getByRole('button', { name: /^agregar$/i }).first();
    this.filterFieldSelect = this.modal.locator('select').nth(1);
    this.filterOperatorSelect = this.modal.locator('select').nth(2);
    this.filterValueInput = this.modal.locator('input[placeholder*="valor"]');
    this.addFilterButton = this.modal.getByRole('button', { name: /^agregar$/i });
    this.nameInput = this.modal.locator('input[type="text"]').first();
    this.saveButton = this.modal.getByRole('button', { name: /guardar consulta/i });
    this.cancelButton = this.modal.getByRole('button', { name: /cancelar/i });
  }

  async open(): Promise<void> {
    await this.openButton.click();
    await expect(this.modal).toBeVisible();
  }

  async selectQueryByLabel(label: string): Promise<void> {
    await this.querySelect.selectOption({ label });
  }

  async openSortTab(): Promise<void> {
    await this.sortTab.click();
  }

  async openFilterTab(): Promise<void> {
    await this.filterTab.click();
  }

  async addSort(fieldLabel: string, direction: 'Ascendente' | 'Descendente'): Promise<void> {
    await this.openSortTab();
    await this.sortFieldSelect.selectOption({ label: fieldLabel });
    await this.sortDirectionSelect.selectOption(direction);
    await this.addSortButton.click();
  }

  async addFilter(fieldLabel: string, operatorLabel: string, value: string): Promise<void> {
    await this.openFilterTab();
    await this.filterFieldSelect.selectOption({ label: fieldLabel });
    await this.filterOperatorSelect.selectOption(operatorLabel);
    if (value.length > 0) {
      await this.filterValueInput.fill(value);
    }
    await this.addFilterButton.click();
  }

  async save(): Promise<void> {
    await this.saveButton.click();
    await expect(this.modal).toBeHidden();
  }

  async cancel(): Promise<void> {
    await this.cancelButton.click();
    await expect(this.modal).toBeHidden();
  }
}
