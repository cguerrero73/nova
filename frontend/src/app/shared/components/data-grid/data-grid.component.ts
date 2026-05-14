import { Component, input, output, signal, computed, ChangeDetectionStrategy, ViewChild, ElementRef, AfterViewInit, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  createAngularTable,
  getCoreRowModel,
  getSortedRowModel,
  ColumnDef,
  Table,
  Row,
  Header,
} from '@tanstack/angular-table';

export interface GridMeta {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

@Component({
  selector: 'app-data-grid',
  standalone: true,
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './data-grid.component.html',
  styleUrl: './data-grid.component.css',
})
export class DataGridComponent<T> implements AfterViewInit {
  @ViewChild('scrollContainer') scrollContainer?: ElementRef<HTMLDivElement>;
  
  columns = input.required<ColumnDef<T>[]>();
  data = input.required<T[]>();
  rowKey = input<string>('id');
  enableSorting = input(true);
  // Registro seleccionado (selección simple)
  selectedRow = input<T | null>(null);
  
  meta = input<GridMeta | null>(null);
  loading = input(false);

  rowSelect = output<T>();
  rowDoubleClick = output<T>();
  pageChange = output<number>();

  sort = signal<{ id: string; desc: boolean }[]>([]);
  private previousDataLength = 0;
  private isRestoringScroll = false;
  private pendingScrollPosition = 0;

  table = computed(() => 
    createAngularTable(() => ({
      data: this.data(),
      columns: this.columns(),
      state: {
        sorting: this.sort(),
      },
      getCoreRowModel: getCoreRowModel(),
      getSortedRowModel: getSortedRowModel(),
      onSortingChange: (updater: unknown) => {
        if (typeof updater === 'function') {
          const result = (updater as (s: { id: string; desc: boolean }[]) => { id: string; desc: boolean }[])(this.sort());
          this.sort.set(result);
        } else {
          this.sort.set(updater as { id: string; desc: boolean }[]);
        }
      },
      getRowId: (originalRow: T) => String((originalRow as Record<string, unknown>)[this.rowKey()]),
    }))
  );

  constructor() {
    // Effect to detect when data changes and restore scroll position
    effect(() => {
      const currentLength = this.data().length;
      
      // If we have more data than before and we're restoring position
      if (currentLength > this.previousDataLength && this.pendingScrollPosition > 0 && this.scrollContainer) {
        // Restore scroll position after new data is rendered
        requestAnimationFrame(() => {
          if (this.scrollContainer && !this.isRestoringScroll) {
            this.isRestoringScroll = true;
            this.scrollContainer.nativeElement.scrollTop = this.pendingScrollPosition;
            this.isRestoringScroll = false;
            this.pendingScrollPosition = 0;
          }
        });
      }
      
      this.previousDataLength = currentLength;
    });
  }

  ngAfterViewInit() {
    this.previousDataLength = this.data().length;
  }

  allRowsSelected = computed(() => {
    // Single selection mode - check if the single selected row is in the data
    const selected = this.selectedRow();
    if (!selected) return false;
    const selectedKey = (selected as Record<string, unknown>)[this.rowKey()];
    return this.data().some(row => (row as Record<string, unknown>)[this.rowKey()] === selectedKey);
  });

  totalColumns = computed(() => {
    return this.columns().length;
  });

  getHeaderValue(column: Header<T, unknown>): string {
    const header = column.column.columnDef.header;
    if (typeof header === 'function') {
      return header(column.getContext()) as string;
    }
    return header as string;
  }

  getCellValue(cell: { column: { columnDef: { cell?: unknown } }; getContext: () => unknown; getValue: () => unknown }): unknown {
    const cellFn = cell.column.columnDef.cell;
    if (typeof cellFn === 'function') {
      return cellFn(cell.getContext());
    }
    return cell.getValue();
  }

  // Verificar si una fila está seleccionada
  isSelected(row: Row<T>): boolean {
    const key = (row.original as Record<string, unknown>)[this.rowKey()];
    const selected = this.selectedRow();
    if (!selected) return false;
    return (selected as Record<string, unknown>)[this.rowKey()] === key;
  }

  // Manejar click en fila
  onRowClickInternal(row: Row<T>) {
    this.rowSelect.emit(row.original);
  }

  // Manejar doble click en fila
  onRowDoubleClickInternal(row: Row<T>, event: MouseEvent) {
    event.stopPropagation();
    this.rowDoubleClick.emit(row.original);
  }

  toggleSort(columnId: string) {
    const current = this.sort();
    if (current.length > 0 && current[0].id === columnId) {
      if (current[0].desc) {
        this.sort.set([]);
      } else {
        this.sort.set([{ id: columnId, desc: true }]);
      }
    } else {
      this.sort.set([{ id: columnId, desc: false }]);
    }
  }

  private scrollThreshold = 150;

  onScroll(event: Event) {
    const el = event.target as HTMLElement;
    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    
    // Guardar posición actual si no estamos restaurando
    if (!this.isRestoringScroll) {
      this.pendingScrollPosition = el.scrollTop;
    }
    
    if (distanceFromBottom < this.scrollThreshold && 
        !this.loading() && 
        this.meta() && 
        this.meta()!.page < this.meta()!.totalPages) {
      this.pageChange.emit(this.meta()!.page + 1);
    }
  }
}
