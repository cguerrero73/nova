import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UiStore } from '@core/stores/ui.store';

@Component({
  selector: 'app-toast',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="toast-container">
      @for (notification of uiStore.notifications(); track notification.id) {
        <div 
          class="toast" 
          [class]="'toast--' + notification.type"
          (click)="uiStore.dismissNotification(notification.id)"
        >
          <div class="toast__icon">
            @switch (notification.type) {
              @case ('success') { ✓ }
              @case ('error') { ✕ }
              @case ('warning') { ⚠ }
              @case ('info') { ℹ }
            }
          </div>
          <div class="toast__content">
            <div class="toast__title">{{ notification.title }}</div>
            <div class="toast__message">{{ notification.message }}</div>
          </div>
          <button class="toast__close" (click)="uiStore.dismissNotification(notification.id); $event.stopPropagation()">×</button>
        </div>
      }
    </div>
  `,
  styles: [`
    .toast-container {
      position: fixed;
      top: 1rem;
      right: 1rem;
      z-index: 9999;
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      max-width: 400px;
    }

    .toast {
      display: flex;
      align-items: flex-start;
      gap: 0.75rem;
      padding: 1rem;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      cursor: pointer;
      animation: slideIn 0.3s ease-out;
    }

    @keyframes slideIn {
      from {
        transform: translateX(100%);
        opacity: 0;
      }
      to {
        transform: translateX(0);
        opacity: 1;
      }
    }

    .toast--success {
      background: #ecfdf5;
      border: 1px solid #10b981;
      color: #065f46;
    }

    .toast--error {
      background: #fef2f2;
      border: 1px solid #ef4444;
      color: #991b1b;
    }

    .toast--warning {
      background: #fffbeb;
      border: 1px solid #f59e0b;
      color: #92400e;
    }

    .toast--info {
      background: #eff6ff;
      border: 1px solid #3b82f6;
      color: #1e40af;
    }

    .toast__icon {
      font-size: 1.25rem;
      font-weight: bold;
    }

    .toast__content {
      flex: 1;
    }

    .toast__title {
      font-weight: 600;
      margin-bottom: 0.25rem;
    }

    .toast__message {
      font-size: 0.875rem;
      opacity: 0.9;
    }

    .toast__close {
      background: none;
      border: none;
      font-size: 1.25rem;
      cursor: pointer;
      opacity: 0.5;
      padding: 0;
      line-height: 1;
    }

    .toast__close:hover {
      opacity: 1;
    }
  `]
})
export class ToastComponent {
  readonly uiStore = inject(UiStore);
}