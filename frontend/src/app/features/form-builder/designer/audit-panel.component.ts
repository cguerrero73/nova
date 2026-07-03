import { Component, inject, signal, computed, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FormDesignerStore } from '../state/designer.store';
import { FormDesignerService } from '../services/designer.service';
import { AuditEntry } from '../models/designer.model';

/**
 * Audit panel — displays form audit log with filters.
 * Wired to GET /forms/:formKey/audit.
 */
@Component({
  selector: 'app-audit-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="audit-panel p-3">
      <h3 class="text-xs font-semibold text-gray-700 mb-2">Audit Log</h3>

      <!-- Filters -->
      <div class="flex gap-1 mb-2">
        <select
          [ngModel]="actionFilter()"
          (ngModelChange)="onActionFilterChange($event)"
          class="flex-1 px-1.5 py-1 text-xs border rounded"
        >
          <option value="">All actions</option>
          @for (action of actionOptions; track action) {
            <option [value]="action">{{ action }}</option>
          }
        </select>
        <select
          [ngModel]="entityFilter()"
          (ngModelChange)="onEntityFilterChange($event)"
          class="flex-1 px-1.5 py-1 text-xs border rounded"
        >
          <option value="">All entities</option>
          @for (entity of entityOptions; track entity) {
            <option [value]="entity">{{ entity }}</option>
          }
        </select>
      </div>

      <!-- Loading -->
      @if (loading()) {
        <p class="text-xs text-gray-400 text-center py-4">Loading...</p>
      } @else if (entries().length === 0) {
        <p class="text-xs text-gray-400 text-center py-4">No audit entries.</p>
      } @else {
        <!-- Entries -->
        <div class="space-y-1.5 max-h-64 overflow-y-auto">
          @for (entry of entries(); track entry.id) {
            <div class="text-xs border-b border-gray-100 pb-1.5">
              <div class="flex items-center justify-between">
                <span class="font-medium text-gray-700">{{ entry.action }}</span>
                <span class="text-gray-400">{{ formatTime(entry.createdAt) }}</span>
              </div>
              <div class="text-gray-500">
                {{ entry.entityType }} #{{ entry.entityId }}
                @if (entry.actorUserId) {
                  <span class="text-gray-400">by {{ entry.actorUserId }}</span>
                }
              </div>
              @if (entry.note) {
                <div class="text-gray-400 italic">{{ entry.note }}</div>
              }
            </div>
          }
        </div>

        <!-- Pagination -->
        @if (totalPages() > 1) {
          <div class="flex items-center justify-between mt-2 text-xs text-gray-500">
            <button
              class="px-2 py-0.5 border rounded disabled:opacity-30"
              [disabled]="page() <= 1"
              (click)="goToPage(page() - 1)"
            >
              Prev
            </button>
            <span>{{ page() }} / {{ totalPages() }}</span>
            <button
              class="px-2 py-0.5 border rounded disabled:opacity-30"
              [disabled]="page() >= totalPages()"
              (click)="goToPage(page() + 1)"
            >
              Next
            </button>
          </div>
        }
      }
    </div>
  `,
  styles: [`
    .audit-panel {
      border-top: 1px solid #e5e7eb;
    }
  `],
})
export class AuditPanelComponent implements OnInit {
  private readonly store = inject(FormDesignerStore);
  private readonly designerService = inject(FormDesignerService);

  entries = signal<AuditEntry[]>([]);
  loading = signal(false);
  page = signal(1);
  total = signal(0);
  pageSize = 20;

  actionFilter = signal('');
  entityFilter = signal('');

  readonly actionOptions = [
    'form.create',
    'form.archive',
    'layout.create',
    'layout.archive',
    'layout.assign',
    'layout.unassign',
    'version.publish',
    'version.revert',
    'version.draft_save',
  ];

  readonly entityOptions = ['form', 'layout', 'version', 'assignment'];

  totalPages = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));

  ngOnInit(): void {
    this.loadAudit();
  }

  onActionFilterChange(value: string): void {
    this.actionFilter.set(value);
    this.page.set(1);
    this.loadAudit();
  }

  onEntityFilterChange(value: string): void {
    this.entityFilter.set(value);
    this.page.set(1);
    this.loadAudit();
  }

  goToPage(p: number): void {
    this.page.set(p);
    this.loadAudit();
  }

  loadAudit(): void {
    const formKey = this.store.formKey();
    if (!formKey) return;

    this.loading.set(true);
    this.designerService
      .listAudit(formKey, {
        page: this.page(),
        pageSize: this.pageSize,
        action: this.actionFilter() || undefined,
        entityType: this.entityFilter() || undefined,
      })
      .subscribe({
        next: (res) => {
          this.entries.set(res.items);
          this.total.set(res.total);
          this.loading.set(false);
        },
        error: () => {
          this.entries.set([]);
          this.loading.set(false);
        },
      });
  }

  formatTime(iso: string): string {
    if (!iso) return '';
    const d = new Date(iso);
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) +
      ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
}
