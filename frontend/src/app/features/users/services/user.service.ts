import { Injectable, signal, computed, inject } from '@angular/core';
import { ApiService } from '@core/services/api.service';
import { UiStore } from '@core/stores/ui.store';
import { User, CreateUserDto, UpdateUserDto, QueryParams } from '../models/user.model';

@Injectable({ providedIn: 'root' })
export class UserService {
  private readonly api = inject(ApiService);
  private readonly uiStore = inject(UiStore);

  readonly users = signal<User[]>([]);
  readonly selectedUser = signal<User | null>(null);
  readonly meta = signal({ page: 1, pageSize: 20, total: 0, totalPages: 0 });

  readonly hasMore = computed(() => this.meta().page < this.meta().totalPages);

  async loadUsers(params: QueryParams = {}) {
    this.uiStore.setLoading(true);
    try {
      const response = await this.api.get<User[]>('/users', params).toPromise();
      if (response?.success && response.data) {
        this.users.set(response.data);
        if (response.meta) {
          this.meta.set({
            page: response.meta.page,
            pageSize: response.meta.pageSize,
            total: response.meta.total,
            totalPages: response.meta.totalPages,
          });
        }
      }
    } finally {
      this.uiStore.setLoading(false);
    }
  }

  async loadUser(id: string) {
    this.uiStore.setLoading(true);
    try {
      const response = await this.api.getById<User>('/users', id).toPromise();
      if (response?.success && response.data) {
        this.selectedUser.set(response.data);
        return response.data;
      }
    } finally {
      this.uiStore.setLoading(false);
    }
    return null;
  }

  async createUser(dto: CreateUserDto) {
    const response = await this.api.post<User>('/users', dto).toPromise();
    if (response?.success && response.data) {
      this.uiStore.success('Usuario creado', 'El usuario se ha creado correctamente');
    }
    return response?.success && response.data ? response.data : null;
  }

  async updateUser(id: string, dto: UpdateUserDto) {
    const response = await this.api.put<User>('/users', id, dto).toPromise();
    if (response?.success && response.data) {
      this.uiStore.success('Usuario actualizado', 'Los cambios se han guardado correctamente');
    }
    return response?.success && response.data ? response.data : null;
  }

  async deleteUser(id: string) {
    const response = await this.api.delete('/users', id).toPromise();
    if (response?.success) {
      this.uiStore.success('Usuario eliminado', 'El usuario ha sido eliminado');
      this.users.update(users => users.filter(u => u.id !== id));
    }
    return response?.success ?? false;
  }
}
