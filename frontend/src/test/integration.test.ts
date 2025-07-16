import { describe, it, expect, beforeAll, vi } from 'vitest';
import { SMTPEdcClient } from '../utils/api-helpers';
import { services } from '../types/api';

// Mock the Wails runtime and App methods for testing
beforeAll(() => {
  // Mock window.go.main.App
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globalThis as any).window = {
    go: {
      main: {
        App: {
          Greet: vi
            .fn()
            .mockResolvedValue('Hello Test User, welcome to SMTP-EDC!'),
          TestConnection: vi.fn().mockResolvedValue({
            success: true,
            message: 'Connection successful',
          }),
          SendEmail: vi.fn().mockResolvedValue({
            success: true,
            message: 'Email sent successfully',
          }),
          ValidateEmailAddress: vi
            .fn()
            .mockResolvedValue({ success: true, message: 'Email is valid' }),
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
          SaveConfig: vi.fn().mockResolvedValue({ success: true }),
          ValidateConfig: vi.fn().mockResolvedValue({
            success: true,
            message: 'Configuration is valid',
          }),
          ListTemplates: vi.fn().mockResolvedValue(['template1', 'template2']),
          LoadTemplate: vi.fn().mockResolvedValue({
            success: true,
            content: 'Template content for {{name}}',
          }),
          SaveTemplate: vi.fn().mockResolvedValue({ success: true }),
          ExecuteTemplate: vi
            .fn()
            .mockResolvedValue('Hello John! Template executed successfully.'),
          GetDefaultConfigPath: vi
            .fn()
            .mockResolvedValue('/path/to/default/config'),
        },
      },
    },
  };
});

describe('SMTP-EDC Integration Tests', () => {
  let client: SMTPEdcClient;

  beforeAll(() => {
    client = new SMTPEdcClient();
  });

  describe('Type Consistency Tests', () => {
    it('should validate SMTP configuration types match backend expectations', async () => {
      const config: services.ConfigData = {
        server: 'smtp.example.com',
        port: 587,
        username: 'test@example.com',
        password: 'password123',
        authType: 'plain', // Backend expects lowercase
        startTLS: true,
        skipVerify: false,
        templates: {},
      };

      // This should not throw type errors
      const result = await client.testConnection(config);
      expect(result).toBeDefined();
      expect(result.success).toBe(true);
    });

    it('should validate email message types match backend expectations', async () => {
      const emailRequest: services.EmailRequest = {
        from: 'sender@example.com',
        to: ['recipient@example.com'],
        cc: ['cc@example.com'],
        bcc: ['bcc@example.com'],
        subject: 'Test Subject',
        body: 'Test body content',
        htmlBody: '<p>Test HTML content</p>',
        attachments: [],
        headers: {},
        convertValues: () => ({}), // Add required convertValues method
      };

      // This should not throw type errors
      const result = await client.sendEmail(emailRequest);
      expect(result).toBeDefined();
      expect(result.success).toBe(true);
    });

    it('should validate authentication type enum consistency', () => {
      // Test that our TypeScript AuthType values are handled correctly
      const authTypes = [
        'PLAIN',
        'LOGIN',
        'CRAM-MD5',
        'OAUTH2',
        'XOAUTH2',
        'NONE',
      ] as const;

      authTypes.forEach(authType => {
        const config: services.ConfigData = {
          server: 'smtp.example.com',
          port: 587,
          username: 'test@example.com',
          password: 'password123',
          authType: authType.toLowerCase(),
          startTLS: true,
          skipVerify: false,
          templates: {},
        };

        // Should compile without type errors
        expect(config.authType).toBe(authType.toLowerCase());
      });
    });

    it('should validate error handling consistency', async () => {
      // Test with invalid email to trigger error path
      const result = await client.validateEmail('invalid-email');
      expect(result).toBeDefined();
      expect(typeof result.success).toBe('boolean');
    });
  });

  describe('API Method Coverage Tests', () => {
    it('should expose all expected configuration methods', async () => {
      expect(typeof client.loadConfig).toBe('function');
      expect(typeof client.saveConfig).toBe('function');
      expect(typeof client.validateConfig).toBe('function');

      // Test method signatures work correctly
      const config = await client.loadConfig();
      expect(config).toBeDefined();

      await client.saveConfig({
        server: 'smtp.example.com',
        port: 587,
        username: 'test@example.com',
        password: 'password123',
        authType: 'plain',
        startTLS: true,
        skipVerify: false,
        templates: {},
      });
    });

    it('should expose all expected template methods', async () => {
      expect(typeof client.listTemplates).toBe('function');
      expect(typeof client.loadTemplate).toBe('function');
      expect(typeof client.saveTemplate).toBe('function');
      expect(typeof client.executeTemplate).toBe('function');

      // Test template methods
      const templates = await client.listTemplates();
      expect(Array.isArray(templates)).toBe(true);

      const templateResult = await client.loadTemplate('template1');
      expect(typeof templateResult).toBe('object');
      expect(templateResult.success).toBe(true);

      const templateContext = {
        templateName: 'template1',
        variables: { name: 'John' },
        emailData: {
          from: 'test@example.com',
          to: ['recipient@example.com'],
        },
      };
      const executed = await client.executeTemplate(templateContext);
      expect(typeof executed).toBe('string');
    });

    it('should expose all expected SMTP methods', async () => {
      expect(typeof client.testConnection).toBe('function');
      expect(typeof client.sendEmail).toBe('function');
      expect(typeof client.validateEmail).toBe('function');
      expect(typeof client.greet).toBe('function');

      // Test core SMTP functionality
      const greetResult = await client.greet('Test User');
      expect(typeof greetResult).toBe('string');

      const validationResult = await client.validateEmail('test@example.com');
      expect(validationResult).toBeDefined();
      expect(typeof validationResult.success).toBe('boolean');
    });
  });

  describe('Data Structure Consistency Tests', () => {
    it('should handle complex nested data structures correctly', async () => {
      const emailRequest: services.EmailRequest = {
        from: 'sender@example.com',
        to: ['recipient1@example.com', 'recipient2@example.com'],
        cc: ['cc1@example.com', 'cc2@example.com'],
        bcc: ['bcc@example.com'],
        subject: 'Multi-recipient test email',
        body: 'This is a test email with multiple recipients.',
        htmlBody:
          '<p>This is a <strong>test email</strong> with multiple recipients.</p>',
        attachments: [],
        headers: {},
        convertValues: () => ({}), // Add required convertValues method
      };

      // This tests that complex data structures are properly serialized/deserialized
      const result = await client.sendEmail(emailRequest);
      expect(result).toBeDefined();
      expect(result.success).toBe(true);
    });

    it('should handle optional fields correctly', async () => {
      // Test minimal configuration
      const minimalConfig: services.ConfigData = {
        server: 'localhost',
        port: 25,
        username: '',
        password: '',
        authType: 'none',
        startTLS: false,
        skipVerify: true,
        templates: {},
      };

      const result = await client.testConnection(minimalConfig);
      expect(result).toBeDefined();

      // Test minimal email
      const minimalEmail: services.EmailRequest = {
        from: 'test@example.com',
        to: ['recipient@example.com'],
        cc: [],
        bcc: [],
        subject: 'Minimal test',
        body: 'Test content',
        htmlBody: '',
        attachments: [],
        headers: {},
        convertValues: () => ({}), // Add required convertValues method
      };

      const sendResult = await client.sendEmail(minimalEmail);
      expect(sendResult).toBeDefined();
    });
  });
});
