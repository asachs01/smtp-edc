import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Global test configuration
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock IntersectionObserver
Object.defineProperty(globalThis, 'IntersectionObserver', {
  writable: true,
  value: class IntersectionObserver {
    constructor() {}
    disconnect() {}
    observe() {}
    unobserve() {}
    root = null;
    rootMargin = '';
    thresholds = [];
    takeRecords() {
      return [];
    }
  },
});

// Mock the Wails runtime for testing
Object.defineProperty(window, 'go', {
  writable: true,
  value: {
    main: {
      App: {
        Greet: vi.fn().mockResolvedValue('Hello Test!'),
        TestConnection: vi.fn().mockResolvedValue({
          success: true,
          message: 'Connection successful',
          timestamp: new Date().toISOString(),
        }),
        SendEmail: vi.fn().mockResolvedValue({
          success: true,
          message: 'Email sent successfully',
          timestamp: new Date().toISOString(),
          messageId: 'test-message-id',
        }),
        ValidateEmailAddress: vi.fn().mockResolvedValue({
          success: true,
          message: 'Email is valid',
          timestamp: new Date().toISOString(),
        }),
        ValidateConfig: vi.fn().mockResolvedValue(undefined),
        LoadConfig: vi.fn().mockResolvedValue({
          server: 'smtp.example.com',
          port: 587,
          username: 'test@example.com',
          password: 'password123',
          authType: 'plain',
          startTLS: true,
          skipVerify: false,
          templates: {},
        }),
        GetCurrentConfig: vi.fn().mockResolvedValue({
          server: 'smtp.example.com',
          port: 587,
          username: 'test@example.com',
          password: 'password123',
          authType: 'plain',
          startTLS: true,
          skipVerify: false,
          templates: {},
        }),
        ListTemplates: vi.fn().mockResolvedValue([]),
        LoadTemplate: vi.fn().mockResolvedValue({
          success: true,
          message: 'Template loaded successfully',
          content: 'Test template content',
        }),
        ExecuteTemplate: vi.fn().mockResolvedValue('Hello John!'),
        SaveConfig: vi.fn().mockResolvedValue(undefined),
        GetDefaultConfigPath: vi.fn().mockResolvedValue('/path/to/config.yaml'),
        ValidateEmailList: vi.fn().mockResolvedValue({
          success: true,
          message: 'All emails are valid',
          timestamp: new Date().toISOString(),
        }),
      },
    },
  },
});
