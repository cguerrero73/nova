import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-6" [style.color]="'var(--color-text)'">Dashboard</h1>
      
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- Card 1 -->
        <div class="p-6 rounded-lg" [style.background-color]="'var(--color-surface)'" [style.border]="'1px solid var(--color-border)'">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-full bg-blue-100 dark:bg-blue-900">
              <svg class="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m9 5.197v1" />
              </svg>
            </div>
            <div>
              <p class="text-sm" [style.color]="'var(--color-text-secondary)'">Usuarios</p>
              <p class="text-2xl font-bold" [style.color]="'var(--color-text)'">1,234</p>
            </div>
          </div>
        </div>

        <!-- Card 2 -->
        <div class="p-6 rounded-lg" [style.background-color]="'var(--color-surface)'" [style.border]="'1px solid var(--color-border)'">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-full bg-green-100 dark:bg-green-900">
              <svg class="w-6 h-6 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm" [style.color]="'var(--color-text-secondary)'">Activos</p>
              <p class="text-2xl font-bold" [style.color]="'var(--color-text)'">987</p>
            </div>
          </div>
        </div>

        <!-- Card 3 -->
        <div class="p-6 rounded-lg" [style.background-color]="'var(--color-surface)'" [style.border]="'1px solid var(--color-border)'">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-full bg-yellow-100 dark:bg-yellow-900">
              <svg class="w-6 h-6 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm" [style.color]="'var(--color-text-secondary)'">Pendientes</p>
              <p class="text-2xl font-bold" [style.color]="'var(--color-text)'">42</p>
            </div>
          </div>
        </div>

        <!-- Card 4 -->
        <div class="p-6 rounded-lg" [style.background-color]="'var(--color-surface)'" [style.border]="'1px solid var(--color-border)'">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-full bg-red-100 dark:bg-red-900">
              <svg class="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div>
              <p class="text-sm" [style.color]="'var(--color-text-secondary)'">Inactivos</p>
              <p class="text-2xl font-bold" [style.color]="'var(--color-text)'">205</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class DashboardComponent {}