import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-user-detail',
  standalone: true,
  imports: [RouterLink, FormsModule],
  templateUrl: './user-detail.component.html',
})
export class UserDetailComponent implements OnInit {
  readonly userService = inject(UserService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  
  readonly user = signal<User | null>(null);
  readonly isNewMode = signal<boolean>(false);
  readonly isLoading = signal<boolean>(false);

  async ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    
    if (id === 'new') {
      // Modo creación: crear usuario vacío
      this.isNewMode.set(true);
      this.user.set({
        id: '',
        name: '',
        email: '',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
    } else if (id) {
      // Modo edición: cargar usuario existente
      this.isNewMode.set(false);
      const userData = await this.userService.loadUser(id);
      this.user.set(userData);
    }
  }

  async saveUser() {
    const user = this.user();
    if (!user) return;
    
    this.isLoading.set(true);
    
    try {
      if (this.isNewMode()) {
        // Crear nuevo usuario
        const newUser = await this.userService.createUser(user);
        if (newUser) {
          this.router.navigate(['/users']);
        }
      } else {
        // Actualizar usuario existente
        const updatedUser = await this.userService.updateUser(user.id, user);
        if (updatedUser) {
          this.user.set(updatedUser);
        }
      }
    } finally {
      this.isLoading.set(false);
    }
  }

  async deleteUser() {
    const user = this.user();
    if (user && confirm('Are you sure you want to delete this user?')) {
      const success = await this.userService.deleteUser(user.id);
      if (success) {
        this.router.navigate(['/users']);
      }
    }
  }

  onFieldChange(field: string, value: unknown) {
    this.user.update(u => {
      if (!u) return u;
      return { ...u, [field]: value };
    });
  }
}
