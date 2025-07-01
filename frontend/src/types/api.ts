/**
 * SMTP-EDC API Contract Definitions
 *
 * This file provides TypeScript interfaces and types for the SMTP-EDC API,
 * enhancing the auto-generated Wails bindings with better documentation,
 * validation helpers, and developer experience improvements.
 */

// Re-export Wails-generated types for convenience
import { services } from '../../wailsjs/go/models';
export { services } from '../../wailsjs/go/models';
export * from '../../wailsjs/go/main/App';

// Enhanced type definitions with better documentation and validation

/**
 * SMTP Authentication Types
 */
export type AuthType =
  | 'PLAIN'
  | 'LOGIN'
  | 'CRAM-MD5'
  | 'OAUTH2'
  | 'XOAUTH2'
  | 'NONE';

/**
 * Email validation modes
 */
export type ValidationMode = 'syntax' | 'mx' | 'both';

/**
 * Template variable types for better type safety
 */
export interface TemplateVariables {
  [key: string]: string | number | boolean | Date;
}

/**
 * Enhanced configuration with validation and defaults
 */
export interface SMTPConfigInput {
  server: string;
  port?: number; // defaults to 587
  username?: string;
  password?: string;
  authType?: AuthType; // defaults to 'PLAIN'
  startTLS?: boolean; // defaults to true
  skipVerify?: boolean; // defaults to false
  templates?: Record<string, string>;
}

/**
 * Email composition helper interface
 */
export interface EmailComposition {
  from: string;
  to: string | string[];
  cc?: string | string[];
  bcc?: string | string[];
  subject: string;
  body?: string;
  htmlBody?: string;
  attachments?: string[];
  headers?: Record<string, string>;
  replyTo?: string;
  priority?: 'low' | 'normal' | 'high';
}

/**
 * Template execution context
 */
export interface TemplateContext {
  templateName: string;
  subjectTemplate?: string;
  variables: TemplateVariables;
  emailData: {
    from: string;
    to: string[];
    cc?: string[];
    bcc?: string[];
  };
}

/**
 * Connection test options
 */
export interface ConnectionTestOptions {
  timeout?: number; // seconds, defaults to 30
  checkCapabilities?: boolean; // defaults to true
  testAuth?: boolean; // defaults to true
}

/**
 * Validation options for email addresses
 */
export interface ValidationOptions {
  mode: ValidationMode;
  timeout?: number; // seconds for MX lookup
  allowLocalDomains?: boolean;
}

/**
 * API Response wrapper for better error handling
 */
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  timestamp: string;
}

/**
 * Configuration validation result
 */
export interface ConfigValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

/**
 * Template validation result with detailed feedback
 */
export interface TemplateValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  variables: string[];
  hasHtml: boolean;
  hasText: boolean;
}

/**
 * Server capabilities with human-readable descriptions
 */
export interface ServerCapabilities {
  pipelining: boolean;
  startTLS: boolean;
  authMethods: AuthType[];
  maxMessageSize: number;
  supportsEightBit: boolean;
  extensions: string[];
}

/**
 * Authentication statistics for monitoring
 */
export interface AuthStats {
  totalAttempts: number;
  successfulAttempts: number;
  failedAttempts: number;
  lastSuccess?: string; // ISO timestamp
  lastFailure?: string; // ISO timestamp
  rateLimitHits: number;
  blockedIPs: string[];
}

/**
 * Email sending statistics
 */
export interface SendingStats {
  totalSent: number;
  totalFailed: number;
  lastSent?: string; // ISO timestamp
  averageResponseTime: number; // milliseconds
  dailyQuota?: number;
  dailySent: number;
}

/**
 * Template management interface
 */
export interface TemplateManager {
  listTemplates(): Promise<services.TemplateInfo[]>;
  loadTemplate(name: string): Promise<services.TemplateResult>;
  saveTemplate(name: string, content: string): Promise<services.TemplateResult>;
  deleteTemplate(name: string): Promise<services.TemplateResult>;
  validateTemplate(content: string): Promise<services.TemplateResult>;
  getDefaultTemplates(): Promise<Record<string, string>>;
  createDefaultTemplates(): Promise<services.TemplateResult>;
}

/**
 * Configuration manager interface
 */
export interface ConfigManager {
  loadConfig(filename?: string): Promise<services.ConfigData>;
  saveConfig(config: services.ConfigData, filename?: string): Promise<void>;
  validateConfig(config: services.ConfigData): Promise<void>;
  getCurrentConfig(): Promise<services.ConfigData>;
  getDefaultConfigPath(): Promise<string>;
  listConfigFiles(): Promise<string[]>;
}

/**
 * SMTP client interface
 */
export interface SMTPClient {
  testConnection(
    config: services.ConfigData
  ): Promise<services.ConnectionResult>;
  sendEmail(request: services.EmailRequest): Promise<services.SendResult>;
  validateEmailAddress(email: string): Promise<services.SendResult>;
  validateEmailList(
    emails: string[],
    validateMX: boolean
  ): Promise<services.SendResult>;
  getAuthStats(): Promise<Record<string, unknown>>;
  setDebugMode(enabled: boolean): Promise<void>;
  getDebugMode(): Promise<boolean>;
}

/**
 * Complete API interface combining all services
 */
export interface SMTPEdcApi extends TemplateManager, ConfigManager, SMTPClient {
  // Convenience methods for better UX
  greet(name: string): Promise<string>;

  // Helper methods (could be implemented client-side)
  createEmailFromTemplate(
    context: TemplateContext
  ): Promise<services.EmailRequest>;
  validateEmailComposition(email: EmailComposition): ConfigValidationResult;
  getServerCapabilities(
    config: services.ConfigData
  ): Promise<ServerCapabilities>;
}

/**
 * Error types for better error handling
 */
export class SMTPError extends Error {
  constructor(
    message: string,
    public code?: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'SMTPError';
  }
}

export class ConfigError extends Error {
  constructor(
    message: string,
    public field?: string,
    public value?: unknown
  ) {
    super(message);
    this.name = 'ConfigError';
  }
}

export class TemplateError extends Error {
  constructor(
    message: string,
    public templateName?: string,
    public line?: number
  ) {
    super(message);
    this.name = 'TemplateError';
  }
}

/**
 * Type guards for runtime type checking
 */
export const isConfigData = (obj: unknown): obj is services.ConfigData => {
  if (!obj || typeof obj !== 'object' || obj === null) return false;

  const record = obj as Record<string, unknown>;
  return (
    'server' in obj &&
    'port' in obj &&
    'username' in obj &&
    'authType' in obj &&
    'startTLS' in obj &&
    'skipVerify' in obj &&
    typeof record.server === 'string' &&
    typeof record.port === 'number' &&
    typeof record.username === 'string' &&
    typeof record.authType === 'string' &&
    typeof record.startTLS === 'boolean' &&
    typeof record.skipVerify === 'boolean'
  );
};

export const isEmailRequest = (obj: unknown): obj is services.EmailRequest => {
  if (!obj || typeof obj !== 'object' || obj === null) return false;

  const record = obj as Record<string, unknown>;
  return (
    'from' in obj &&
    'to' in obj &&
    'subject' in obj &&
    typeof record.from === 'string' &&
    Array.isArray(record.to) &&
    typeof record.subject === 'string'
  );
};

export const isConnectionResult = (
  obj: unknown
): obj is services.ConnectionResult => {
  if (!obj || typeof obj !== 'object' || obj === null) return false;

  const record = obj as Record<string, unknown>;
  return (
    'success' in obj &&
    'message' in obj &&
    'timestamp' in obj &&
    typeof record.success === 'boolean' &&
    typeof record.message === 'string' &&
    typeof record.timestamp === 'string'
  );
};

/**
 * Validation helpers
 */
export const validateEmailSyntax = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validatePort = (port: number): boolean => {
  return port > 0 && port <= 65535;
};

export const validateAuthType = (authType: string): authType is AuthType => {
  return (
    ['PLAIN', 'LOGIN', 'CRAM-MD5', 'OAUTH2', 'XOAUTH2', 'NONE'] as const
  ).includes(authType as AuthType);
};

/**
 * Default values for configuration
 */
export const DEFAULT_CONFIG: Partial<services.ConfigData> = {
  port: 587,
  authType: 'PLAIN' as AuthType,
  startTLS: true,
  skipVerify: false,
  templates: {},
};

/**
 * Common SMTP ports for reference
 */
export const SMTP_PORTS = {
  PLAIN: 25,
  TLS: 587,
  SSL: 465,
  SUBMISSION: 587,
} as const;

/**
 * Template placeholders for common use cases
 */
export const TEMPLATE_PLACEHOLDERS = {
  USER_NAME: '{{.UserName}}',
  EMAIL: '{{.Email}}',
  DATE: '{{.Date}}',
  TIME: '{{.Time}}',
  SUBJECT: '{{.Subject}}',
  COMPANY: '{{.Company}}',
  CUSTOM: (name: string) => `{{.${name}}}`,
} as const;
