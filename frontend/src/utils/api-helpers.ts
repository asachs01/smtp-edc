/**
 * SMTP-EDC API Helper Functions
 *
 * Utility functions for working with the SMTP-EDC API,
 * providing convenient wrappers and validation helpers.
 */

import { services } from '../../wailsjs/go/models';
import * as App from '../../wailsjs/go/main/App';
import {
  SMTPConfigInput,
  EmailComposition,
  TemplateContext,
  TemplateVariables,
  ConfigValidationResult,
  DEFAULT_CONFIG,
  validateEmailSyntax,
  validatePort,
  SMTPError,
  ConfigError,
  TemplateError,
} from '../types/api';

/**
 * API Client wrapper providing enhanced functionality
 */
export class SMTPEdcClient {
  /**
   * Configuration Management
   */

  async loadConfig(filename?: string): Promise<services.ConfigData> {
    try {
      if (filename) {
        return await App.LoadConfig(filename);
      }
      return await App.GetCurrentConfig();
    } catch (error) {
      throw new ConfigError(`Failed to load configuration: ${error}`, 'load');
    }
  }

  async saveConfig(
    config: services.ConfigData,
    filename?: string
  ): Promise<void> {
    try {
      const validationResult = this.validateConfigData(config);
      if (!validationResult.valid) {
        throw new ConfigError(
          `Invalid configuration: ${validationResult.errors.join(', ')}`
        );
      }

      const path = filename || (await App.GetDefaultConfigPath());
      await App.SaveConfig(config, path);
    } catch (error) {
      throw new ConfigError(`Failed to save configuration: ${error}`, 'save');
    }
  }

  async validateConfig(
    config: services.ConfigData
  ): Promise<ConfigValidationResult> {
    try {
      await App.ValidateConfig(config);
      const clientValidation = this.validateConfigData(config);
      return clientValidation;
    } catch (error) {
      return {
        valid: false,
        errors: [`Backend validation failed: ${error}`],
        warnings: [],
      };
    }
  }

  async createConfig(input: SMTPConfigInput): Promise<services.ConfigData> {
    const config: services.ConfigData = {
      server: input.server,
      port: input.port ?? DEFAULT_CONFIG.port!,
      username: input.username ?? '',
      password: input.password ?? '',
      authType: input.authType ?? DEFAULT_CONFIG.authType!,
      startTLS: input.startTLS ?? DEFAULT_CONFIG.startTLS!,
      skipVerify: input.skipVerify ?? DEFAULT_CONFIG.skipVerify!,
      templates: input.templates ?? DEFAULT_CONFIG.templates!,
    };

    return config;
  }

  validateConfigData(config: services.ConfigData): ConfigValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Required fields
    if (!config.server || config.server.trim() === '') {
      errors.push('Server is required');
    }

    // Port validation
    if (!validatePort(config.port)) {
      errors.push('Port must be between 1 and 65535');
    }

    // Auth type validation
    if (
      config.authType &&
      !['plain', 'login', 'cram-md5', 'oauth2', 'xoauth2', 'none'].includes(
        config.authType.toLowerCase()
      )
    ) {
      errors.push('Invalid authentication type');
    }

    // Auth validation
    if (config.authType !== 'NONE' && (!config.username || !config.password)) {
      warnings.push(
        'Username and password are required for authenticated connections'
      );
    }

    // Security warnings
    if (!config.startTLS && config.port !== 465) {
      warnings.push('Consider enabling StartTLS for secure connections');
    }

    if (config.skipVerify) {
      warnings.push(
        'Certificate verification is disabled - this may be insecure'
      );
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * SMTP Operations
   */

  async testConnection(
    config: services.ConfigData
  ): Promise<services.ConnectionResult> {
    try {
      const validationResult = this.validateConfigData(config);
      if (!validationResult.valid) {
        throw new ConfigError(
          `Invalid configuration: ${validationResult.errors.join(', ')}`
        );
      }

      return await App.TestConnection(config);
    } catch (error) {
      throw new SMTPError(
        `Connection test failed: ${error}`,
        'CONNECTION_FAILED'
      );
    }
  }

  async sendEmail(
    request: services.EmailRequest
  ): Promise<services.SendResult> {
    try {
      const validationResult = this.validateEmailRequest(request);
      if (!validationResult.valid) {
        throw new SMTPError(
          `Invalid email request: ${validationResult.errors.join(', ')}`
        );
      }

      return await App.SendEmail(request);
    } catch (error) {
      throw new SMTPError(`Failed to send email: ${error}`, 'SEND_FAILED');
    }
  }

  async sendEmailFromComposition(
    composition: EmailComposition,
    config?: services.ConfigData
  ): Promise<services.SendResult> {
    const emailRequest = this.compositionToEmailRequest(composition, config);
    return await this.sendEmail(emailRequest);
  }

  compositionToEmailRequest(
    composition: EmailComposition,
    config?: services.ConfigData
  ): services.EmailRequest {
    return services.EmailRequest.createFrom({
      config,
      from: composition.from,
      to: Array.isArray(composition.to) ? composition.to : [composition.to],
      cc: composition.cc
        ? Array.isArray(composition.cc)
          ? composition.cc
          : [composition.cc]
        : [],
      bcc: composition.bcc
        ? Array.isArray(composition.bcc)
          ? composition.bcc
          : [composition.bcc]
        : [],
      subject: composition.subject,
      body: composition.body || '',
      htmlBody: composition.htmlBody || '',
      attachments: composition.attachments || [],
      headers: {
        ...composition.headers,
        ...(composition.replyTo && { 'Reply-To': composition.replyTo }),
        ...(composition.priority && {
          'X-Priority': this.priorityToHeader(composition.priority),
        }),
      },
    });
  }

  private priorityToHeader(priority: 'low' | 'normal' | 'high'): string {
    const priorities = { low: '5', normal: '3', high: '1' };
    return priorities[priority];
  }

  validateEmailRequest(request: services.EmailRequest): ConfigValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Required fields
    if (!request.from || !validateEmailSyntax(request.from)) {
      errors.push('Valid sender email address is required');
    }

    if (!request.to || request.to.length === 0) {
      errors.push('At least one recipient is required');
    }

    // Validate all email addresses
    const allEmails = [
      ...(request.to || []),
      ...(request.cc || []),
      ...(request.bcc || []),
    ];

    allEmails.forEach(email => {
      if (email && !validateEmailSyntax(email)) {
        errors.push(`Invalid email address: ${email}`);
      }
    });

    if (!request.subject || request.subject.trim() === '') {
      warnings.push('Email subject is empty');
    }

    if (!request.body && !request.htmlBody) {
      warnings.push('Email body is empty');
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
    };
  }

  async validateEmailAddress(email: string): Promise<services.SendResult> {
    return await App.ValidateEmailAddress(email);
  }

  // Alias for backward compatibility
  async validateEmail(email: string): Promise<services.SendResult> {
    return await this.validateEmailAddress(email);
  }

  async validateEmailList(
    emails: string[],
    validateMX: boolean = false
  ): Promise<services.SendResult> {
    return await App.ValidateEmailList(emails, validateMX);
  }

  /**
   * Template Management
   */

  async listTemplates(): Promise<services.TemplateInfo[]> {
    return await App.ListTemplates();
  }

  async loadTemplate(name: string): Promise<services.TemplateResult> {
    try {
      return await App.LoadTemplate(name);
    } catch (error) {
      throw new TemplateError(`Failed to load template: ${error}`, name);
    }
  }

  async saveTemplate(
    name: string,
    content: string
  ): Promise<services.TemplateResult> {
    try {
      // Validate template before saving
      const validation = await App.ValidateTemplate(content);
      if (!validation.success) {
        throw new TemplateError(
          `Template validation failed: ${validation.error}`,
          name
        );
      }

      return await App.SaveTemplate(name, content);
    } catch (error) {
      throw new TemplateError(`Failed to save template: ${error}`, name);
    }
  }

  async deleteTemplate(name: string): Promise<services.TemplateResult> {
    try {
      return await App.DeleteTemplate(name);
    } catch (error) {
      throw new TemplateError(`Failed to delete template: ${error}`, name);
    }
  }

  async executeTemplate(
    context: TemplateContext
  ): Promise<services.EmailRequest> {
    try {
      const templateData = services.TemplateData.createFrom({
        from: context.emailData.from,
        to: context.emailData.to,
        cc: context.emailData.cc || [],
        bcc: context.emailData.bcc || [],
        subject: '', // Will be set by template
        data: context.variables,
      });

      return await App.ExecuteTemplate(
        context.templateName,
        context.subjectTemplate || '',
        templateData
      );
    } catch (error) {
      throw new TemplateError(
        `Failed to execute template: ${error}`,
        context.templateName
      );
    }
  }

  async getDefaultTemplates(): Promise<Record<string, string>> {
    return await App.GetDefaultTemplates();
  }

  async createDefaultTemplates(): Promise<services.TemplateResult> {
    return await App.CreateDefaultTemplates();
  }

  /**
   * Utility Methods
   */

  async getAuthStats(): Promise<Record<string, unknown>> {
    return await App.GetAuthStats();
  }

  async setDebugMode(enabled: boolean): Promise<void> {
    await App.SetDebugMode(enabled);
  }

  async getDebugMode(): Promise<boolean> {
    return await App.GetDebugMode();
  }

  async greet(name: string): Promise<string> {
    return await App.Greet(name);
  }

  /**
   * Batch Operations
   */

  async sendBulkEmails(
    template: string,
    recipients: Array<{ email: string; variables: Record<string, unknown> }>,
    config: services.ConfigData
  ): Promise<services.SendResult[]> {
    const results: services.SendResult[] = [];

    for (const recipient of recipients) {
      try {
        const context: TemplateContext = {
          templateName: template,
          variables: recipient.variables as TemplateVariables,
          emailData: {
            from: config.username,
            to: [recipient.email],
          },
        };

        const emailRequest = await this.executeTemplate(context);
        emailRequest.config = config;

        const result = await this.sendEmail(emailRequest);
        results.push(result);
      } catch (error) {
        results.push({
          success: false,
          message: '',
          error: `Failed to send to ${recipient.email}: ${error}`,
          timestamp: new Date().toISOString(),
        });
      }
    }

    return results;
  }

  /**
   * Configuration Helpers
   */

  getCommonConfigs(): Record<string, Partial<SMTPConfigInput>> {
    return {
      gmail: {
        server: 'smtp.gmail.com',
        port: 587,
        authType: 'OAUTH2',
        startTLS: true,
      },
      outlook: {
        server: 'smtp-mail.outlook.com',
        port: 587,
        authType: 'LOGIN',
        startTLS: true,
      },
      yahoo: {
        server: 'smtp.mail.yahoo.com',
        port: 587,
        authType: 'LOGIN',
        startTLS: true,
      },
      sendgrid: {
        server: 'smtp.sendgrid.net',
        port: 587,
        authType: 'LOGIN',
        startTLS: true,
      },
      mailgun: {
        server: 'smtp.mailgun.org',
        port: 587,
        authType: 'LOGIN',
        startTLS: true,
      },
    };
  }
}

// Create and export the API client instance
const apiClient = new SMTPEdcClient();

// Export the API client
export { apiClient as default };

// Convenience exports for common operations
export const testConnection = apiClient.testConnection.bind(apiClient);
export const sendEmail = apiClient.sendEmail.bind(apiClient);
export const loadConfig = apiClient.loadConfig.bind(apiClient);
export const saveConfig = apiClient.saveConfig.bind(apiClient);
export const validateEmailAddress =
  apiClient.validateEmailAddress.bind(apiClient);
