import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { FormRuntimeComponent } from './form-runtime.component';
import { LayoutDefinition } from '../models/layout-definition.model';

describe('FormRuntimeComponent', () => {
  let component: FormRuntimeComponent;
  let fixture: ComponentFixture<FormRuntimeComponent>;

  const basicLayout: LayoutDefinition = {
    formKey: 'test-form',
    layoutName: 'default',
    sections: [
      {
        name: 'personal',
        title: 'Personal Info',
        order: 0,
        fields: [
          {
            type: 'text',
            name: 'first_name',
            ui: { label: 'First Name', placeholder: 'Enter first name', width: 'half' },
            validators: [{ kind: 'required' }],
          },
          {
            type: 'text',
            name: 'last_name',
            ui: { label: 'Last Name', width: 'half' },
            validators: [],
          },
        ],
      },
      {
        name: 'contact',
        title: 'Contact',
        order: 1,
        fields: [
          {
            type: 'text',
            name: 'email',
            ui: { label: 'Email', width: 'full' },
            validators: [{ kind: 'required' }, { kind: 'email' }],
          },
        ],
      },
    ],
    rules: [],
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReactiveFormsModule, FormRuntimeComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(FormRuntimeComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('layout', basicLayout);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should build FormGroup from layout sections', () => {
    fixture.componentRef.setInput('layout', basicLayout);
    fixture.detectChanges();

    expect(component.form).toBeDefined();
    expect(component.form.get('first_name')).toBeTruthy();
    expect(component.form.get('last_name')).toBeTruthy();
    expect(component.form.get('email')).toBeTruthy();
  });

  it('should sort sections by order', () => {
    fixture.componentRef.setInput('layout', basicLayout);
    fixture.detectChanges();

    const sections = component.sortedSections();
    expect(sections[0].name).toBe('personal');
    expect(sections[1].name).toBe('contact');
  });

  it('should apply required validator', () => {
    fixture.componentRef.setInput('layout', basicLayout);
    fixture.detectChanges();

    const firstNameCtrl = component.form.get('first_name')!;
    expect(firstNameCtrl.hasError('required')).toBeTrue();
  });

  it('should apply email validator', () => {
    fixture.componentRef.setInput('layout', basicLayout);
    fixture.detectChanges();

    const emailCtrl = component.form.get('email')!;
    emailCtrl.setValue('invalid-email');
    expect(emailCtrl.hasError('email')).toBeTrue();
  });

  it('should disable control when ui.readOnly is true', () => {
    const layoutWithReadOnly: LayoutDefinition = {
      ...basicLayout,
      sections: [
        {
          name: 's1',
          order: 0,
          fields: [
            {
              type: 'text',
              name: 'readonly_field',
              ui: { label: 'Read Only', readOnly: true },
              validators: [],
            },
          ],
        },
      ],
    };

    fixture.componentRef.setInput('layout', layoutWithReadOnly);
    fixture.detectChanges();

    expect(component.form.get('readonly_field')!.disabled).toBeTrue();
  });

  describe('grid layout', () => {
    it('should track field widths from layout definition', () => {
      fixture.componentRef.setInput('layout', basicLayout);
      fixture.detectChanges();

      const personalSection = basicLayout.sections[0];
      expect(personalSection.fields[0].ui.width).toBe('half');
      expect(personalSection.fields[1].ui.width).toBe('half');
    });
  });

  describe('cross-field rules', () => {
    it('should evaluate requiredIf rule on value changes', () => {
      const layoutWithRules: LayoutDefinition = {
        formKey: 'test',
        layoutName: 'default',
        sections: [
          {
            name: 's1',
            order: 0,
            fields: [
              {
                type: 'checkbox',
                name: 'has_company',
                ui: { label: 'Has Company' },
                validators: [],
              },
              {
                type: 'text',
                name: 'company_name',
                ui: { label: 'Company Name' },
                validators: [],
              },
            ],
          },
        ],
        rules: [
          { operator: 'requiredIf', source: 'has_company', target: 'company_name' },
        ],
      };

      fixture.componentRef.setInput('layout', layoutWithRules);
      fixture.detectChanges();

      // Initially, has_company is false, so company_name should not have error
      expect(component.form.get('company_name')!.errors).toBeNull();

      // Set has_company to true
      component.form.get('has_company')!.setValue(true);
      fixture.detectChanges();

      // Now company_name should have requiredIf error
      expect(component.form.get('company_name')!.errors).toBeDefined();
      expect(component.form.get('company_name')!.errors!['requiredIf']).toBeDefined();
    });

    it('should hide fields based on hiddenIf rule', () => {
      const layoutWithHidden: LayoutDefinition = {
        formKey: 'test',
        layoutName: 'default',
        sections: [
          {
            name: 's1',
            order: 0,
            fields: [
              {
                type: 'checkbox',
                name: 'show_details',
                ui: { label: 'Show Details' },
                validators: [],
              },
              {
                type: 'text',
                name: 'details',
                ui: { label: 'Details' },
                validators: [],
              },
            ],
          },
        ],
        rules: [
          { operator: 'hiddenIf', source: 'show_details', target: 'details' },
        ],
      };

      fixture.componentRef.setInput('layout', layoutWithHidden);
      fixture.detectChanges();

      // Initially show_details is false, details should be visible
      expect(component.hiddenFields()).not.toContain('details');

      // Set show_details to true
      component.form.get('show_details')!.setValue(true);
      fixture.detectChanges();

      // Now details should be hidden
      expect(component.hiddenFields()).toContain('details');
    });
  });

  describe('pattern validation', () => {
    it('should validate pattern on text field', () => {
      const layoutWithPattern: LayoutDefinition = {
        formKey: 'test',
        layoutName: 'default',
        sections: [
          {
            name: 's1',
            order: 0,
            fields: [
              {
                type: 'text',
                name: 'code',
                ui: { label: 'Code' },
                validators: [{ kind: 'pattern', value: '^[A-Z]{3}$' }],
              },
            ],
          },
        ],
      };

      fixture.componentRef.setInput('layout', layoutWithPattern);
      fixture.detectChanges();

      const codeCtrl = component.form.get('code')!;

      codeCtrl.setValue('abc');
      expect(codeCtrl.hasError('pattern')).toBeTrue();

      codeCtrl.setValue('ABC');
      expect(codeCtrl.hasError('pattern')).toBeFalse();
    });
  });
});
