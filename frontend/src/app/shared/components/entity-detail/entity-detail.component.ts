import { Component, input, output, signal, computed, effect, TemplateRef, ViewChild, ContentChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

export interface EntityTab {
  id: string;
  label: string;
  icon?: string;
}

export interface EntityAction {
  id: string;
  label: string;
  icon?: string;
  variant?: 'primary' | 'danger' | 'secondary';
  disabled?: boolean;
}

@Component({
  selector: 'app-entity-detail',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './entity-detail.component.html',
  styleUrl: './entity-detail.component.css',
})
export class EntityDetailComponent {
  // Entity a editar (any para flexibilidad)
  entity = input<unknown>(null);
  entityId = input<number | string | null>(null);
  entityLabel = input<string>('');
  
  // Configuración
  title = input<string>('Detalle');
  subtitle = input<string>('');
  isOpen = input<boolean>(false);
  isLoading = input<boolean>(false);
  
  // isSaving necesita ser un signal para poder mutar desde el componente
  private _isSaving = signal<boolean>(false);
  get isSaving() { return this._isSaving.asReadonly(); }
  
  // Helper para usar en template como función
  isSavingCheck(): boolean {
    return this._isSaving();
  }
  
  // Tabs disponibles
  tabs = input<EntityTab[]>([
    { id: 'view', label: 'Visualización del Registro' },
    { id: 'comments', label: 'Comentarios' },
    { id: 'documents', label: 'Documentos' },
    { id: 'audit', label: 'Historial' }
  ]);
  activeTab = signal<string>('view');
  
  // Eventos
  close = output<void>();
  save = output<unknown>();
  action = output<{ id: string; entity: unknown }>();

  // Form editable (copia del entity)
  editableEntity = signal<Record<string, unknown>>({});
  
  // Templates para cada tab (inyectados desde el padre)
  @ContentChild('viewTemplate') viewTemplate?: TemplateRef<unknown>;
  @ContentChild('commentsTemplate') commentsTemplate?: TemplateRef<unknown>;
  @ContentChild('documentsTemplate') documentsTemplate?: TemplateRef<unknown>;
  @ContentChild('auditTemplate') auditTemplate?: TemplateRef<unknown>;
  
  constructor() {
    // Sincronizar editableEntity cuando cambia el entity
    effect(() => {
      this.updateEditable(this.entity());
    });
  }
  
  // Actualizar editableEntity cuando cambia el entity
  updateEditable(newEntity: unknown) {
    if (newEntity && typeof newEntity === 'object') {
      this.editableEntity.set({ ...newEntity as Record<string, unknown> });
    } else {
      this.editableEntity.set({});
    }
  }
  
  // Track de cambios
  hasChanges = computed(() => {
    const original = this.entity();
    const current = this.editableEntity();
    if (!original || typeof original !== 'object') return false;
    
    const originalObj = original as Record<string, unknown>;
    return Object.keys(current).some(key => 
      JSON.stringify(originalObj[key]) !== JSON.stringify(current[key])
    );
  });

  setActiveTab(tabId: string) {
    this.activeTab.set(tabId);
  }

  onClose() {
    this.close.emit();
  }

  onSave() {
    this._isSaving.set(true);
    this.save.emit(this.editableEntity());
  }

  onAction(actionId: string) {
    this.action.emit({ id: actionId, entity: this.editableEntity() });
  }

  // Método para que las tabs children actualicen sus datos
  getEditableEntity() {
    return this.editableEntity;
  }

  updateField(field: string, value: unknown) {
    this.editableEntity.update(current => ({
      ...current,
      [field]: value
    }));
  }
  
  // Obtener el template activo
  getActiveTemplate(): TemplateRef<unknown> | null {
    const tab = this.activeTab();
    switch (tab) {
      case 'view': return this.viewTemplate || null;
      case 'comments': return this.commentsTemplate || null;
      case 'documents': return this.documentsTemplate || null;
      case 'audit': return this.auditTemplate || null;
      default: return this.viewTemplate || null;
    }
  }
}