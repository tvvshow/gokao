/**
 * Property Test: Form Validation Completeness
 * 
 * Feature: gui-audit, Property 5: Form Validation Completeness
 * Validates: Requirements 3.2, 4.6
 * 
 * For any form component with required fields, there must be corresponding
 * validation rules, and validation must be executed before form submission.
 */
import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import * as fs from 'fs'
import * as path from 'path'
import {
  validatePasswordStrength,
  sanitizeInput,
  validateEmail,
  validatePhone,
  validateUsername,
  containsXSSPatterns,
} from '../../utils/validators'

// Helper to generate string from character set
function stringFromChars(chars: string, minLength: number, maxLength: number) {
  return fc.array(
    fc.constantFrom(...chars.split('')),
    { minLength, maxLength }
  ).map(arr => arr.join(''))
}

describe('Property 5: Form Validation Completeness', () => {
  /**
   * Static analysis: Check that form components have validation rules
   */
  describe('Form Components Validation Rules', () => {
    it('LoginPage should have validation rules for all required fields', () => {
      const loginPagePath = path.resolve(__dirname, '../../views/LoginPage.vue')
      const content = fs.readFileSync(loginPagePath, 'utf-8')
      
      // Check for form rules definition
      expect(content).toContain('loginRules')
      expect(content).toContain('registerRules')
      
      // Check for required field validation
      expect(content).toContain('required: true')
      
      // Check for form validation before submit
      expect(content).toContain('validate()')
    })

    it('Forms should use sanitizeFormData before submission', () => {
      const loginPagePath = path.resolve(__dirname, '../../views/LoginPage.vue')
      const content = fs.readFileSync(loginPagePath, 'utf-8')
      
      // Check for sanitization import
      expect(content).toContain('sanitizeFormData')
      
      // Check for sanitization usage
      expect(content).toContain('sanitizeFormData(loginForm)')
      expect(content).toContain('sanitizeFormData(registerForm)')
    })

    it('Password fields should have strength validation', () => {
      const loginPagePath = path.resolve(__dirname, '../../views/LoginPage.vue')
      const content = fs.readFileSync(loginPagePath, 'utf-8')
      
      // Check for password validator
      expect(content).toContain('createPasswordValidator')
      expect(content).toContain('validatePasswordStrength')
    })
  })

  /**
   * Property tests for validation functions
   */
  describe('Password Strength Validation', () => {
    it('property: weak passwords should be rejected', () => {
      // Generate weak passwords (short, no uppercase, no numbers)
      const weakPasswordArb = stringFromChars('abcdefghijklmnopqrstuvwxyz', 1, 7)

      fc.assert(
        fc.property(weakPasswordArb, (password) => {
          const result = validatePasswordStrength(password)
          // Weak passwords should not be valid
          expect(result.isValid).toBe(false)
          expect(result.score).toBeLessThan(4)
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: strong passwords should be accepted', () => {
      // Generate strong passwords (8+ chars, uppercase, lowercase, number)
      const strongPasswordArb = fc.tuple(
        stringFromChars('ABCDEFGHIJKLMNOPQRSTUVWXYZ', 2, 4),
        stringFromChars('abcdefghijklmnopqrstuvwxyz', 2, 4),
        stringFromChars('0123456789', 2, 4),
        stringFromChars('abcdefghijklmnopqrstuvwxyz', 2, 4)
      ).map(([upper, lower, num, extra]) => upper + lower + num + extra)

      fc.assert(
        fc.property(strongPasswordArb, (password) => {
          const result = validatePasswordStrength(password)
          // Strong passwords should be valid (at least 8 chars with upper, lower, number)
          expect(password.length).toBeGreaterThanOrEqual(8)
          expect(result.isValid).toBe(true)
          expect(result.score).toBeGreaterThanOrEqual(4)
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: empty password should be invalid', () => {
      const result = validatePasswordStrength('')
      expect(result.isValid).toBe(false)
      expect(result.score).toBe(0)
    })
  })

  describe('XSS Protection', () => {
    it('property: HTML special characters should be escaped', () => {
      const htmlCharsArb = stringFromChars('<>&"\'`=/', 1, 20)

      fc.assert(
        fc.property(htmlCharsArb, (input) => {
          const sanitized = sanitizeInput(input)
          // Sanitized output should not contain raw HTML chars
          expect(sanitized).not.toContain('<')
          expect(sanitized).not.toContain('>')
          expect(sanitized).not.toContain('"')
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: XSS patterns should be detected', () => {
      const xssPatterns = [
        '<script>alert(1)</script>',
        'javascript:alert(1)',
        '<img onerror=alert(1)>',
        '<iframe src="evil.com">',
        'expression(alert(1))',
      ]

      for (const pattern of xssPatterns) {
        expect(containsXSSPatterns(pattern)).toBe(true)
      }
    })

    it('property: safe text should not be flagged as XSS', () => {
      const safeTextArb = stringFromChars(
        'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ,.!?',
        1, 100
      )

      fc.assert(
        fc.property(safeTextArb, (text) => {
          expect(containsXSSPatterns(text)).toBe(false)
          return true
        }),
        { numRuns: 100 }
      )
    })
  })

  describe('Email Validation', () => {
    it('property: valid emails should be accepted', () => {
      const validEmailArb = fc.tuple(
        stringFromChars('abcdefghijklmnopqrstuvwxyz0123456789', 1, 10),
        stringFromChars('abcdefghijklmnopqrstuvwxyz', 1, 10),
        fc.constantFrom('com', 'org', 'net', 'edu', 'cn')
      ).map(([local, domain, tld]) => `${local}@${domain}.${tld}`)

      fc.assert(
        fc.property(validEmailArb, (email) => {
          expect(validateEmail(email)).toBe(true)
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: invalid emails should be rejected', () => {
      const invalidEmails = [
        'notanemail',
        '@nodomain.com',
        'no@domain',
        'spaces in@email.com',
        '',
      ]

      for (const email of invalidEmails) {
        expect(validateEmail(email)).toBe(false)
      }
    })
  })

  describe('Phone Validation', () => {
    it('property: valid Chinese phone numbers should be accepted', () => {
      // Chinese mobile numbers start with 1[3-9] followed by 9 digits
      const validPhoneArb = fc.tuple(
        fc.constantFrom('13', '14', '15', '16', '17', '18', '19'),
        stringFromChars('0123456789', 9, 9)
      ).map(([prefix, suffix]) => prefix + suffix)

      fc.assert(
        fc.property(validPhoneArb, (phone) => {
          expect(validatePhone(phone)).toBe(true)
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: invalid phone numbers should be rejected', () => {
      const invalidPhones = [
        '12345678901', // starts with 12
        '1234567890',  // too short
        '123456789012', // too long
        'abcdefghijk', // not numbers
        '',
      ]

      for (const phone of invalidPhones) {
        expect(validatePhone(phone)).toBe(false)
      }
    })
  })

  describe('Username Validation', () => {
    it('property: valid usernames should be accepted', () => {
      const validUsernameArb = stringFromChars(
        'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_',
        3, 20
      )

      fc.assert(
        fc.property(validUsernameArb, (username) => {
          expect(validateUsername(username)).toBe(true)
          return true
        }),
        { numRuns: 100 }
      )
    })

    it('property: usernames with special characters should be rejected', () => {
      const invalidUsernames = [
        'user@name',
        'user name',
        'user<script>',
        'ab', // too short
      ]

      for (const username of invalidUsernames) {
        expect(validateUsername(username)).toBe(false)
      }
    })
  })
})
