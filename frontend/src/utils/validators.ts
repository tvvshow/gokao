/**
 * Form validation utilities
 * Provides password strength validation and XSS protection
 */

// Password strength validation rules
export interface PasswordStrengthResult {
  isValid: boolean;
  score: number; // 0-4
  message: string;
  suggestions: string[];
}

/**
 * Validate password strength
 * Requirements: at least 8 characters, uppercase, lowercase, and number
 */
export function validatePasswordStrength(
  password: string
): PasswordStrengthResult {
  const suggestions: string[] = [];
  let score = 0;

  if (!password) {
    return {
      isValid: false,
      score: 0,
      message: '请输入密码',
      suggestions: ['密码不能为空'],
    };
  }

  // Check length
  if (password.length >= 8) {
    score += 1;
  } else {
    suggestions.push('密码长度至少8位');
  }

  // Check for uppercase
  if (/[A-Z]/.test(password)) {
    score += 1;
  } else {
    suggestions.push('需要包含大写字母');
  }

  // Check for lowercase
  if (/[a-z]/.test(password)) {
    score += 1;
  } else {
    suggestions.push('需要包含小写字母');
  }

  // Check for numbers
  if (/\d/.test(password)) {
    score += 1;
  } else {
    suggestions.push('需要包含数字');
  }

  // Bonus for special characters
  if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
    score = Math.min(score + 1, 4);
  }

  const isValid = score >= 4 && password.length >= 8;

  let message = '';
  if (score <= 1) {
    message = '密码强度：弱';
  } else if (score === 2) {
    message = '密码强度：较弱';
  } else if (score === 3) {
    message = '密码强度：中等';
  } else {
    message = '密码强度：强';
  }

  return {
    isValid,
    score,
    message,
    suggestions,
  };
}

import type { FormItemRule } from 'element-plus';

/**
 * Create Element Plus form validator for password strength
 */
export function createPasswordValidator(minLength: number = 8): FormItemRule['validator'] {
  return (_rule, value: string, callback) => {
    if (!value) {
      callback(new Error('请输入密码'));
      return;
    }

    if (value.length < minLength) {
      callback(new Error(`密码长度至少${minLength}位`));
      return;
    }

    if (!/[A-Z]/.test(value)) {
      callback(new Error('密码需要包含大写字母'));
      return;
    }

    if (!/[a-z]/.test(value)) {
      callback(new Error('密码需要包含小写字母'));
      return;
    }

    if (!/\d/.test(value)) {
      callback(new Error('密码需要包含数字'));
      return;
    }

    callback();
  };
}

/**
 * Sanitize user input to prevent XSS attacks
 * Escapes HTML special characters
 */
export function sanitizeInput(input: string): string {
  if (!input) return '';

  const htmlEntities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;',
    '`': '&#x60;',
    '=': '&#x3D;',
  };

  return input.replace(/[&<>"'`=/]/g, (char) => htmlEntities[char] || char);
}

/**
 * Validate and sanitize form data object
 */
export function sanitizeFormData<T extends Record<string, unknown>>(data: T): T {
  const sanitized = { ...data };

  for (const key in sanitized) {
    if (typeof sanitized[key] === 'string') {
      sanitized[key] = sanitizeInput(sanitized[key] as string) as T[Extract<
        keyof T,
        string
      >];
    }
  }

  return sanitized;
}

/**
 * Validate email format
 */
export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validate Chinese phone number
 */
export function validatePhone(phone: string): boolean {
  const phoneRegex = /^1[3-9]\d{9}$/;
  return phoneRegex.test(phone);
}

/**
 * Validate username format
 * Allows letters, numbers, and underscores, 3-20 characters
 */
export function validateUsername(username: string): boolean {
  const usernameRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]{3,20}$/;
  return usernameRegex.test(username);
}

/**
 * Check if input contains potential XSS patterns
 */
export function containsXSSPatterns(input: string): boolean {
  const xssPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /javascript:/gi,
    /on\w+\s*=/gi,
    /<iframe/gi,
    /<object/gi,
    /<embed/gi,
    /expression\s*\(/gi,
    /vbscript:/gi,
  ];

  return xssPatterns.some((pattern) => pattern.test(input));
}

/**
 * Create a safe text validator that checks for XSS
 */
export function createSafeTextValidator(fieldName: string = '输入'): FormItemRule['validator'] {
  return (_rule, value: string, callback) => {
    if (value && containsXSSPatterns(value)) {
      callback(new Error(`${fieldName}包含不安全的内容`));
      return;
    }
    callback();
  };
}
