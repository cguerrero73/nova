import { FormGroup, FormControl, Validators } from '@angular/forms';
import { evaluateCrossFieldRules } from './cross-field-validator';
import { CrossFieldRule } from '../models/layout-definition.model';

describe('evaluateCrossFieldRules', () => {
  describe('requiredIf rule', () => {
    it('should set error on target when source is truthy and target is empty', () => {
      const form = new FormGroup({
        field_a: new FormControl('yes'),
        field_b: new FormControl(''),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'requiredIf', source: 'field_a', target: 'field_b' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['field_b']).toBeDefined();
      expect(result.errors['field_b']['requiredIf']).toBeDefined();
    });

    it('should not set error when source is falsy', () => {
      const form = new FormGroup({
        field_a: new FormControl(''),
        field_b: new FormControl(''),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'requiredIf', source: 'field_a', target: 'field_b' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['field_b']).toBeUndefined();
    });

    it('should not set error when target has a value', () => {
      const form = new FormGroup({
        field_a: new FormControl('yes'),
        field_b: new FormControl('filled'),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'requiredIf', source: 'field_a', target: 'field_b' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['field_b']).toBeUndefined();
    });
  });

  describe('equals rule', () => {
    it('should set error when values differ', () => {
      const form = new FormGroup({
        password: new FormControl('abc123'),
        confirm: new FormControl('xyz789'),
      });

      const rules: CrossFieldRule[] = [
        {
          operator: 'equals',
          source: 'password',
          target: 'confirm',
          message: 'Passwords must match',
        },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['confirm']).toBeDefined();
      expect(result.errors['confirm']['equals']['message']).toBe('Passwords must match');
    });

    it('should not set error when values are equal', () => {
      const form = new FormGroup({
        password: new FormControl('abc123'),
        confirm: new FormControl('abc123'),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'equals', source: 'password', target: 'confirm' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['confirm']).toBeUndefined();
    });
  });

  describe('notEquals rule', () => {
    it('should set error when values are the same', () => {
      const form = new FormGroup({
        old_val: new FormControl('same'),
        new_val: new FormControl('same'),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'notEquals', source: 'old_val', target: 'new_val' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['new_val']).toBeDefined();
      expect(result.errors['new_val']['notEquals']).toBeDefined();
    });

    it('should not set error when values differ', () => {
      const form = new FormGroup({
        old_val: new FormControl('a'),
        new_val: new FormControl('b'),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'notEquals', source: 'old_val', target: 'new_val' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['new_val']).toBeUndefined();
    });
  });

  describe('hiddenIf rule', () => {
    it('should add target to hiddenFields when source is truthy', () => {
      const form = new FormGroup({
        show_extra: new FormControl(true),
        extra_field: new FormControl(''),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'hiddenIf', source: 'show_extra', target: 'extra_field' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.hiddenFields).toContain('extra_field');
    });

    it('should not hide target when source is falsy', () => {
      const form = new FormGroup({
        show_extra: new FormControl(false),
        extra_field: new FormControl(''),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'hiddenIf', source: 'show_extra', target: 'extra_field' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.hiddenFields).not.toContain('extra_field');
    });
  });

  describe('multiple rules', () => {
    it('should evaluate all rules and combine results', () => {
      const form = new FormGroup({
        type: new FormControl('business'),
        company: new FormControl(''),
        email: new FormControl('test@test.com'),
        confirm_email: new FormControl('other@test.com'),
      });

      const rules: CrossFieldRule[] = [
        { operator: 'requiredIf', source: 'type', target: 'company' },
        { operator: 'equals', source: 'email', target: 'confirm_email', message: 'Emails must match' },
      ];

      const result = evaluateCrossFieldRules(rules, form);

      expect(result.errors['company']).toBeDefined();
      expect(result.errors['confirm_email']).toBeDefined();
    });
  });
});
