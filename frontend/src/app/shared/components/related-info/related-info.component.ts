import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface RelatedItem {
  id: string | number;
  title?: string;
  description?: string;
  date?: string;
  icon?: string;
  metadata?: Record<string, unknown>;
}

export interface RelatedSection {
  id: string;
  title: string;
  icon?: string;
  items: RelatedItem[];
  emptyMessage?: string;
  loading?: boolean;
}

@Component({
  selector: 'app-related-info',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6">
      @for (section of sections(); track section.id) {
        <div class="border rounded-lg" [style.border-color]="'var(--color-border)'">
          <!-- Section Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b"
               [style.border-color]="'var(--color-border)'"
               [style.background-color]="'var(--color-surface)'">
            <div class="flex items-center gap-2">
              @if (section.icon) {
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" [attr.d]="section.icon" />
                </svg>
              }
              <h3 class="font-medium">{{ section.title }}</h3>
              <span class="text-xs text-gray-500">({{ section.items.length }})</span>
            </div>
            
            @if (section.loading) {
              <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
            }
          </div>
          
          <!-- Section Content -->
          <div class="p-4">
            @if (section.loading) {
              <div class="flex items-center justify-center py-8">
                <span class="text-gray-500">Cargando...</span>
              </div>
            } @else if (section.items.length === 0) {
              <div class="text-center py-8 text-gray-500">
                {{ section.emptyMessage || 'Sin elementos' }}
              </div>
            } @else {
              <div class="space-y-3">
                @for (item of section.items; track item.id) {
                  <div class="flex items-start gap-3 p-3 rounded hover:bg-muted cursor-pointer transition-colors"
                       [style.background-color]="'var(--color-muted)'">
                    @if (item.icon) {
                      <div class="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center">
                        <svg class="w-4 h-4 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" [attr.d]="item.icon" />
                        </svg>
                      </div>
                    }
                    <div class="flex-1 min-w-0">
                      @if (item.title) {
                        <p class="font-medium truncate">{{ item.title }}</p>
                      }
                      @if (item.description) {
                        <p class="text-sm text-gray-500 dark:text-gray-400 truncate">{{ item.description }}</p>
                      }
                      @if (item.date) {
                        <p class="text-xs text-gray-400 mt-1">{{ item.date }}</p>
                      }
                    </div>
                    <!-- Actions -->
                    <button 
                      (click)="onItemAction(section.id, item, $event)"
                      class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700"
                    >
                      <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                      </svg>
                    </button>
                  </div>
                }
              </div>
            }
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    :host { display: block; }
    .animate-spin { animation: spin 1s linear infinite; }
    @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
  `]
})
export class RelatedInfoComponent {
  // Secciones de información relacionada
  sections = input.required<RelatedSection[]>();
  
  // Evento cuando se hace click en un item
  itemClick = output<{ sectionId: string; item: RelatedItem }>();
  
  // Evento para acciones del item (menu)
  itemAction = output<{ sectionId: string; item: RelatedItem; action: string }>();
  
  onItemClick(sectionId: string, item: RelatedItem) {
    this.itemClick.emit({ sectionId, item });
  }
  
  onItemAction(sectionId: string, item: RelatedItem, event: Event) {
    event.stopPropagation();
    // Por ahora solo emitimos, el componente padre puede mostrar un menú
    this.itemAction.emit({ sectionId, item, action: 'menu' });
  }
}