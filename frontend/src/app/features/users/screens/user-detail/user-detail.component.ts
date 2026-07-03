import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user.model';
import { FormRuntimeComponent } from '../../../form-builder/runtime/form-runtime.component';
import { FormRuntimeService } from '../../../form-builder/services/runtime.service';
import { LayoutDefinition } from '../../../form-builder/models/layout-definition.model';

@Component({
  selector: 'app-user-detail',
  standalone: true,
  imports: [RouterLink, FormRuntimeComponent],
  templateUrl: './user-detail.component.html',
})
export class UserDetailComponent implements OnInit {
  readonly userService = inject(UserService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly formRuntimeService = inject(FormRuntimeService);
  
  readonly user = signal<User | null>(null);
  readonly isNewMode = signal<boolean>(false);
  readonly isLoading = signal<boolean>(false);
  readonly layout = signal<LayoutDefinition | null>(null);
  readonly layoutLoading = signal<boolean>(true);

  formInitialValues = computed<Record<string, any>>(() => {
    const u = this.user();
    if (!u) return {};
    return {
      name: u.name,
      email: u.email,
      status: u.status,
    };
  });

  async ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    
    if (id === 'new') {
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
      this.isNewMode.set(false);
      const userData = await this.userService.loadUser(id);
      this.user.set(userData);
    }

    this.loadLayout();
  }

  private loadLayout(): void {
    this.layoutLoading.set(true);
    this.formRuntimeService.resolveForm('user-form').subscribe({
      next: (layout) => {
        this.layout.set(layout);
        this.layoutLoading.set(false);
      },
      error: () => {
        this.layout.set(null);
        this.layoutLoading.set(false);
      },
    });
  }

  async onFormSubmit(data: Record<string, any>) {
    this.isLoading.set(true);

    try {
      if (this.isNewMode()) {
        const newUser = await this.userService.createUser({
          name: data['name'],
          email: data['email'],
          status: data['status'],
        });
        if (newUser) {
          this.router.navigate(['/users']);
        }
      } else {
        const user = this.user();
        if (!user) return;
        const updatedUser = await this.userService.updateUser(user.id, {
          name: data['name'],
          email: data['email'],
          status: data['status'],
        });
        if (updatedUser) {
          this.user.set(updatedUser);
          this.router.navigate(['/users']);
        }
      }
    } finally {
      this.isLoading.set(false);
    }
  }

  onFormCancel(): void {
    this.router.navigate(['/users']);
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
}
