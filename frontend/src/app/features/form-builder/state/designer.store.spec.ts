import { TestBed } from '@angular/core/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { FormDesignerStore } from './designer.store';
import { LayoutDefinition } from '../models/layout-definition.model';

describe('FormDesignerStore', () => {
  let store: FormDesignerStore;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        FormDesignerStore,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    store = TestBed.inject(FormDesignerStore);
  });

  it('should be created', () => {
    expect(store).toBeTruthy();
  });

  it('should start with empty state', () => {
    expect(store.formKey()).toBe('');
    expect(store.layouts()).toEqual([]);
    expect(store.currentLayout()).toBeNull();
    expect(store.definition()).toBeNull();
    expect(store.dirty()).toBe(false);
    expect(store.selectedFieldId()).toBeNull();
  });

  it('should update definition and mark dirty', () => {
    (store as any)._formKey.set('test-form');
    const def: LayoutDefinition = {
      formKey: 'test-form',
      layoutName: 'default',
      sections: [],
      rules: [],
    };
    store.updateDefinition(def);
    expect(store.definition()).toEqual(def);
    expect(store.dirty()).toBe(true);
  });

  it('should add a section', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [],
      rules: [],
    });

    store.addSection({
      name: 'section_1',
      title: 'Section 1',
      order: 0,
      fields: [],
    });

    expect(store.sortedSections().length).toBe(1);
    expect(store.sortedSections()[0].name).toBe('section_1');
    expect(store.dirty()).toBe(true);
  });

  it('should remove a section', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [
        { name: 'section_1', title: 'S1', order: 0, fields: [] },
        { name: 'section_2', title: 'S2', order: 1, fields: [] },
      ],
      rules: [],
    });

    store.removeSection('section_1');
    expect(store.sortedSections().length).toBe(1);
    expect(store.sortedSections()[0].name).toBe('section_2');
  });

  it('should update a field in a section', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [
        {
          name: 'section_1',
          title: 'S1',
          order: 0,
          fields: [
            {
              name: 'field_a',
              type: 'text' as const,
              ui: { label: 'Old Label', width: 'full' as const },
              validators: [],
            },
          ],
        },
      ],
      rules: [],
    });

    store.updateField('section_1', 'field_a', {
      name: 'field_a',
      type: 'text' as const,
      ui: { label: 'New Label', width: 'full' as const },
      validators: [],
    });

    const field = store.sortedSections()[0].fields[0];
    expect(field.ui.label).toBe('New Label');
  });

  it('should remove a field and clear selection if selected', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [
        {
          name: 'section_1',
          title: 'S1',
          order: 0,
          fields: [
            {
              name: 'field_a',
              type: 'text' as const,
              ui: { label: 'A', width: 'full' as const },
              validators: [],
            },
          ],
        },
      ],
      rules: [],
    });
    (store as any)._selectedFieldId.set('field_a');

    store.removeField('section_1', 'field_a');

    expect(store.sortedSections()[0].fields.length).toBe(0);
    expect(store.selectedFieldId()).toBeNull();
  });

  it('should compute selectedField correctly', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [
        {
          name: 'section_1',
          title: 'S1',
          order: 0,
          fields: [
            {
              name: 'field_a',
              type: 'text' as const,
              ui: { label: 'A', width: 'full' as const },
              validators: [],
            },
          ],
        },
      ],
      rules: [],
    });
    (store as any)._selectedFieldId.set('field_a');

    expect(store.selectedField()).toBeTruthy();
    expect(store.selectedField()!.name).toBe('field_a');
    expect(store.selectedFieldSection()).toBe('section_1');
  });

  it('should validate with Zod before save and block invalid drafts', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._currentLayout.set({ fl_name: 'default' } as any);
    // Invalid definition: missing formKey
    (store as any)._definition.set({
      formKey: '',
      layoutName: 'default',
      sections: [],
      rules: [],
    } as any);

    store.saveDraft();

    // Should have validation errors and not attempt save
    expect(store.validationErrors().length).toBeGreaterThan(0);
  });

  it('should reset state', () => {
    (store as any)._formKey.set('test-form');
    (store as any)._dirty.set(true);
    store.reset();

    expect(store.formKey()).toBe('');
    expect(store.dirty()).toBe(false);
    expect(store.definition()).toBeNull();
  });
});
