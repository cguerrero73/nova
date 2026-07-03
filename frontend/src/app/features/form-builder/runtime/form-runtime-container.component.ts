import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { FormRuntimeStore } from '../state/form-runtime.store';
import { FormRuntimeComponent } from './form-runtime.component';

/**
 * Container component that loads the layout by formKey from the route params
 * and passes it to the FormRuntimeComponent.
 */
@Component({
  selector: 'app-form-runtime-container',
  standalone: true,
  imports: [CommonModule, FormRuntimeComponent],
  template: `
    @if (store.loading()) {
      <div class="loading-state">Loading form...</div>
    } @else if (store.error(); as err) {
      <div class="error-state">
        <p>{{ err }}</p>
      </div>
    } @else if (store.layout(); as layoutDef) {
      <app-form-runtime [layout]="layoutDef" />
    }
  `,
  styles: [`
    .loading-state, .error-state {
      text-align: center;
      padding: 2rem;
      color: #6b7280;
    }
    .error-state { color: #dc2626; }
  `],
})
export class FormRuntimeContainerComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  readonly store = inject(FormRuntimeStore);

  ngOnInit(): void {
    const formKey = this.route.snapshot.paramMap.get('formKey');
    if (formKey) {
      this.store.loadForm(formKey);
    }
  }
}
