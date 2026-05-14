import { Component, input, output, computed, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface ToolbarButton {
  id: string;
  icon: 'create' | 'delete' | 'save' | 'print' | 'refresh' | 'export' | 'import' | 'search' | 'filter' | 'edit' | 'view' | 'copy' | 'close' | 'prev' | 'next' | 'duplicate';
  label: string;
  shortcut?: string;
  disabled?: boolean;
  variant?: 'primary' | 'danger' | 'secondary' | 'ghost';
  position?: 'left' | 'right';
}

@Component({
  selector: 'app-toolbar',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './toolbar.component.html',
  styleUrl: './toolbar.component.css',
})
export class ToolbarComponent {
  // Botones a mostrar (si no se provee, usa los default)
  buttons = input<ToolbarButton[]>([]);
  
  // Traducciones para labels de botones (viene de TranslationService)
  translations = input<Record<string, string>>({});

  // Prefix para las keys de traducciones (ej: 'toolbar.')
  translationPrefix = input<string>('toolbar.');
  
  // Habilitar/deshabilitar botones específicos (override)
  createEnabled = input<boolean>(true);
  deleteEnabled = input<boolean>(true);
  saveEnabled = input<boolean>(true);
  printEnabled = input<boolean>(true);
  refreshEnabled = input<boolean>(true);
  duplicateEnabled = input<boolean>(true);
  exportEnabled = input<boolean>(true);
  importEnabled = input<boolean>(true);
  searchEnabled = input<boolean>(true);
  filterEnabled = input<boolean>(true);

  // Mostrar botones específicos (por defecto todos visibles, usar enabled para deshabilitar)
  // Estos inputs permiten ocultar botones específicos en pantallas que no los necesitan
  showCreate = input<boolean>(true);
  showDelete = input<boolean>(true);
  showSave = input<boolean>(true);
  showPrint = input<boolean>(true);
  showRefresh = input<boolean>(true);
  showDuplicate = input<boolean>(true);
  showExport = input<boolean>(false); // Por defecto oculto
  showImport = input<boolean>(false); // Por defecto oculto
  showSearch = input<boolean>(false); // Por defecto oculto
  showFilter = input<boolean>(false); // Por defecto oculto
  showPrev = input<boolean>(true);
  showNext = input<boolean>(true);

  // Habilitar/deshabilitar navegación
  prevEnabled = input<boolean>(false);
  nextEnabled = input<boolean>(false);

  // Eventos
  create = output<void>();
  delete = output<void>();
  save = output<void>();
  print = output<void>();
  refresh = output<void>();
  duplicate = output<void>();
  export = output<void>();
  import = output<void>();
  search = output<string>();
  filter = output<void>();
  prev = output<void>();
  next = output<void>();

  // Busqueda
  searchTerm = '';

  // Keyboard shortcuts
  @HostListener('window:keydown', ['$event'])
  handleKeydown(event: KeyboardEvent) {
    // Ignore if user is typing in an input field
    const target = event.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;
    
    // Ctrl+S: Save
    if (event.ctrlKey && event.key === 's') {
      event.preventDefault();
      if (this.showSave() && this.saveEnabled()) {
        this.save.emit();
      }
      return;
    }
    
    // Ctrl+N: Create
    if (event.ctrlKey && event.key === 'n') {
      event.preventDefault();
      if (this.showCreate() && this.createEnabled()) {
        this.create.emit();
      }
      return;
    }
    
    // Delete key: Delete
    if (event.key === 'Delete' && !isInput) {
      event.preventDefault();
      if (this.showDelete() && this.deleteEnabled()) {
        this.delete.emit();
      }
      return;
    }
    
    // Ctrl+D: Duplicate
    if (event.ctrlKey && event.key === 'd') {
      event.preventDefault();
      if (this.showDuplicate() && this.duplicateEnabled()) {
        this.duplicate.emit();
      }
      return;
    }
    
    // F5: Refresh
    if (event.key === 'F5') {
      event.preventDefault();
      if (this.showRefresh() && this.refreshEnabled()) {
        this.refresh.emit();
      }
      return;
    }
    
    // Ctrl+P: Print
    if (event.ctrlKey && event.key === 'p') {
      event.preventDefault();
      if (this.showPrint() && this.printEnabled()) {
        this.print.emit();
      }
      return;
    }
    
    // Left Arrow: Prev (only if not in input)
    if (event.key === 'ArrowLeft' && !isInput) {
      if (this.showPrev() && this.prevEnabled()) {
        this.prev.emit();
      }
      return;
    }
    
    // Right Arrow: Next (only if not in input)
    if (event.key === 'ArrowRight' && !isInput) {
      if (this.showNext() && this.nextEnabled()) {
        this.next.emit();
      }
      return;
    }
  }

  // Computed: si se proveen botones custom, usarlos; sino, generar los default
  resolvedButtons = computed(() => {
    const custom = this.buttons();
    if (custom.length > 0) {
      return custom;
    }
    return this.getDefaultButtons();
  });

  private getDefaultButtons(): ToolbarButton[] {
    const buttons: ToolbarButton[] = [];
    const t = this.translations();
    const prefix = this.translationPrefix();

    // Helper para obtener traducción
    const getLabel = (key: string, fallback: string): string => {
      const fullKey = prefix + key;
      return t[fullKey] || fallback;
    };

    // Save (Grabar) - primero
    if (this.showSave()) {
      buttons.push({
        id: 'save',
        icon: 'save',
        label: getLabel('save', 'Salvar'),
        shortcut: 'Ctrl+S',
        disabled: !this.saveEnabled(),
        variant: 'primary',
        position: 'left'
      });
    }

    // Create (Nuevo)
    if (this.showCreate()) {
      buttons.push({
        id: 'create',
        icon: 'create',
        label: getLabel('create', 'Crear'),
        shortcut: 'Ctrl+N',
        disabled: !this.createEnabled(),
        variant: 'primary',
        position: 'left'
      });
    }

    // Delete (Borrar)
    if (this.showDelete()) {
      buttons.push({
        id: 'delete',
        icon: 'delete',
        label: getLabel('delete', 'Borrar'),
        shortcut: 'Del',
        disabled: !this.deleteEnabled(),
        variant: 'danger',
        position: 'left'
      });
    }

    // Prev (Anterior)
    if (this.showPrev()) {
      buttons.push({
        id: 'prev',
        icon: 'prev',
        label: getLabel('prev', 'Prev'),
        shortcut: '←',
        disabled: !this.prevEnabled(),
        variant: 'ghost',
        position: 'left'
      });
    }

    // Next (Siguiente)
    if (this.showNext()) {
      buttons.push({
        id: 'next',
        icon: 'next',
        label: getLabel('next', 'Next'),
        shortcut: '→',
        disabled: !this.nextEnabled(),
        variant: 'ghost',
        position: 'left'
      });
    }

    // Duplicate (Duplicar)
    if (this.showDuplicate()) {
      buttons.push({
        id: 'duplicate',
        icon: 'duplicate',
        label: getLabel('duplicate', 'Duplicar'),
        shortcut: 'Ctrl+D',
        disabled: !this.duplicateEnabled(),
        variant: 'ghost',
        position: 'left'
      });
    }

    // Refresh (Actualizar)
    if (this.showRefresh()) {
      buttons.push({
        id: 'refresh',
        icon: 'refresh',
        label: getLabel('refresh', 'Refrescar'),
        shortcut: 'F5',
        disabled: !this.refreshEnabled(),
        variant: 'ghost',
        position: 'left'
      });
    }

    // Print (Imprimir)
    if (this.showPrint()) {
      buttons.push({
        id: 'print',
        icon: 'print',
        label: getLabel('print', 'Imprimir'),
        shortcut: 'Ctrl+P',
        disabled: !this.printEnabled(),
        variant: 'ghost',
        position: 'left'
      });
    }

    return buttons;
  }

  getLeftButtons() {
    return this.resolvedButtons().filter(b => b.position !== 'right');
  }

  getRightButtons() {
    return this.resolvedButtons().filter(b => b.position === 'right');
  }

  onButtonClick(button: ToolbarButton) {
    switch (button.id) {
      case 'create':
        this.create.emit();
        break;
      case 'delete':
        this.delete.emit();
        break;
      case 'save':
        this.save.emit();
        break;
      case 'print':
        this.print.emit();
        break;
      case 'duplicate':
        this.duplicate.emit();
        break;
      case 'prev':
        this.prev.emit();
        break;
      case 'next':
        this.next.emit();
        break;
      case 'refresh':
        this.refresh.emit();
        break;
      case 'export':
        this.export.emit();
        break;
      case 'import':
        this.import.emit();
        break;
      case 'search':
        this.search.emit(this.searchTerm);
        break;
      case 'filter':
        this.filter.emit();
        break;
    }
  }

  onSearchInput(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    this.searchTerm = value;
  }

  onSearchKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      this.search.emit(this.searchTerm);
    }
  }
}