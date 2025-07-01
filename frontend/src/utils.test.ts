import { describe, it, expect } from 'vitest';

// Simple utility functions to test
const StringUtils = {
  capitalize: (str: string): string => {
    return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
  },

  isEmail: (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  },

  extractDomain: (email: string): string => {
    return email.split('@')[1] || '';
  },

  formatPort: (port: number | string): number => {
    const parsed = typeof port === 'string' ? parseInt(port, 10) : port;
    return isNaN(parsed) ? 25 : parsed; // default SMTP port
  },
};

describe('String Utilities', () => {
  describe('capitalize', () => {
    it('capitalizes first letter and lowercases rest', () => {
      expect(StringUtils.capitalize('hello')).toBe('Hello');
      expect(StringUtils.capitalize('WORLD')).toBe('World');
      expect(StringUtils.capitalize('tEsT')).toBe('Test');
    });

    it('handles empty strings', () => {
      expect(StringUtils.capitalize('')).toBe('');
    });

    it('handles single characters', () => {
      expect(StringUtils.capitalize('a')).toBe('A');
      expect(StringUtils.capitalize('Z')).toBe('Z');
    });
  });

  describe('email validation', () => {
    it('validates correct email addresses', () => {
      expect(StringUtils.isEmail('test@example.com')).toBe(true);
      expect(StringUtils.isEmail('user.name+tag@domain.co.uk')).toBe(true);
      expect(StringUtils.isEmail('admin@smtp-server.org')).toBe(true);
    });

    it('rejects invalid email addresses', () => {
      expect(StringUtils.isEmail('invalid')).toBe(false);
      expect(StringUtils.isEmail('invalid@')).toBe(false);
      expect(StringUtils.isEmail('@domain.com')).toBe(false);
      expect(StringUtils.isEmail('')).toBe(false);
      expect(StringUtils.isEmail('no spaces@domain.com')).toBe(false);
    });
  });

  describe('domain extraction', () => {
    it('extracts domain from valid emails', () => {
      expect(StringUtils.extractDomain('user@example.com')).toBe('example.com');
      expect(StringUtils.extractDomain('test@mail.domain.org')).toBe(
        'mail.domain.org'
      );
      expect(StringUtils.extractDomain('admin@smtp.server.net')).toBe(
        'smtp.server.net'
      );
    });

    it('handles invalid emails gracefully', () => {
      expect(StringUtils.extractDomain('invalid')).toBe('');
      expect(StringUtils.extractDomain('')).toBe('');
      expect(StringUtils.extractDomain('no-at-sign')).toBe('');
    });
  });

  describe('port formatting', () => {
    it('handles numeric ports', () => {
      expect(StringUtils.formatPort(587)).toBe(587);
      expect(StringUtils.formatPort(25)).toBe(25);
      expect(StringUtils.formatPort(465)).toBe(465);
    });

    it('parses string ports', () => {
      expect(StringUtils.formatPort('587')).toBe(587);
      expect(StringUtils.formatPort('25')).toBe(25);
      expect(StringUtils.formatPort('465')).toBe(465);
    });

    it('handles invalid ports with default', () => {
      expect(StringUtils.formatPort('invalid')).toBe(25);
      expect(StringUtils.formatPort('')).toBe(25);
      expect(StringUtils.formatPort(NaN)).toBe(25);
    });
  });
});
