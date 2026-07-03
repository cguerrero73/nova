import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideAnimations } from '@angular/platform-browser/animations';
import { SectionCanvasComponent } from './section-canvas.component';
import { FormDesignerStore } from '../state/designer.store';
import { LayoutDefinition, Section } from '../models/layout-definition.model';

describe('SectionCanvasComponent', () => {
  let component: SectionCanvasComponent;
  let fixture: ComponentFixture<SectionCanvasComponent>;
  let store: FormDesignerStore;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SectionCanvasComponent],
      providers: [
        FormDesignerStore,
        provideHttpClient(),
        provideHttpClientTesting(),
        provideAnimations(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SectionCanvasComponent);
    component = fixture.componentInstance;
    store = TestBed.inject(FormDesignerStore);

    // Set up a minimal definition in the store
    (store as any)._formKey.set('test-form');
    (store as any)._definition.set({
      formKey: 'test-form',
      layoutName: 'default',
      sections: [
        {
          name: 'section_1',
          title: 'Section 1',
          order: 0,
          fields: [
            {
              name: 'field_a',
              type: 'text' as const,
              ui: { label: 'Field A', width: 'full' as const },
              validators: [],
            },
            {
              name: 'field_b',
              type: 'number' as const,
              ui: { label: 'Field B', width: 'half' as const },
              validators: [],
            },
          ],
        },
      ],
      rules: [],
    });

    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should render sections from the store', () => {
    const sections = store.sortedSections();
    expect(sections.length).toBe(1);
    expect(sections[0].name).toBe('section_1');
    expect(sections[0].fields.length).toBe(2);
  });

  it('should add a new section', () => {
    component.addSection();
    const sections = store.sortedSections();
    expect(sections.length).toBe(2);
    expect(sections[1].name).toBe('section_2');
    expect(sections[1].title).toBe('Section 2');
  });

  it('should remove a section', () => {
    component.onRemoveSection('section_1');
    const sections = store.sortedSections();
    expect(sections.length).toBe(0);
  });

  it('should update section title', () => {
    component.onTitleChange('section_1', 'Updated Title');
    const sections = store.sortedSections();
    expect(sections[0].title).toBe('Updated Title');
  });

  it('should remove a field from a section', () => {
    component.onRemoveField('section_1', 'field_a');
    const sections = store.sortedSections();
    expect(sections[0].fields.length).toBe(1);
    expect(sections[0].fields[0].name).toBe('field_b');
  });

  it('should reorder fields within a section', () => {
    component.onFieldMoved({
      sectionName: 'section_1',
      previousIndex: 1,
      currentIndex: 0,
    });
    const sections = store.sortedSections();
    expect(sections[0].fields[0].name).toBe('field_b');
    expect(sections[0].fields[1].name).toBe('field_a');
  });

  it('should mark definition as dirty after changes', () => {
    expect(store.dirty()).toBe(false);
    component.addSection();
    expect(store.dirty()).toBe(true);
  });

  it('should drop a new field into a section', () => {
    const newField = {
      name: 'field_c',
      type: 'text' as const,
      ui: { label: 'Field C', width: 'full' as const },
      validators: [],
    };
    component.onFieldDropped({
      sectionName: 'section_1',
      field: newField,
      index: 2,
    });
    const sections = store.sortedSections();
    expect(sections[0].fields.length).toBe(3);
    expect(sections[0].fields[2].name).toBe('field_c');
  });
});
