# SMTP-EDC API Documentation

This documentation covers the complete API surface exposed by the SMTP-EDC Wails application, providing detailed information about all available methods, their parameters, return types, and usage examples.

## Table of Contents

- [Overview](#overview)
- [Configuration Service](#configuration-service)
- [SMTP Service](#smtp-service)
- [Template Service](#template-service)
- [Type Definitions](#type-definitions)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Overview

The SMTP-EDC API exposes **19 methods** across three main service areas:

- **Configuration Management (7 methods)**: Config loading, saving, validation
- **SMTP Operations (7 methods)**: Connection testing, email sending, validation
- **Template Management (5 methods)**: Template creation, execution, management

All methods are available through Wails bindings in the frontend as well as through the CLI interface.

## Configuration Service

### Methods

#### `GetCurrentConfig(): Promise<ConfigData>`
Returns the currently loaded configuration.

**Returns:** Current SMTP configuration or default values if none loaded.

```typescript
const config = await GetCurrentConfig();
console.log(`Server: ${config.server}:${config.port}`);
```

---

#### `LoadConfig(filename?: string): Promise<ConfigData>`
Loads configuration from a YAML file.

**Parameters:**
- `filename` (optional): Path to config file. Uses default if not provided.

**Returns:** Loaded configuration data.

```typescript
// Load default config
const config = await LoadConfig();

// Load specific config file
const prodConfig = await LoadConfig('/path/to/prod-config.yaml');
```

---

#### `SaveConfig(config: ConfigData, filename?: string): Promise<void>`
Saves configuration to a YAML file.

**Parameters:**
- `config`: Configuration object to save
- `filename` (optional): Target file path. Uses default if not provided.

```typescript
const config: ConfigData = {
  server: 'smtp.gmail.com',
  port: 587,
  username: 'user@gmail.com',
  password: 'app-password',
  authType: 'PLAIN',
  startTLS: true,
  skipVerify: false,
  templates: {}
};

await SaveConfig(config, 'gmail-config.yaml');
```

---

#### `ValidateConfig(config: ConfigData): Promise<void>`
Validates configuration structure and required fields.

**Parameters:**
- `config`: Configuration to validate

**Throws:** Error if configuration is invalid.

```typescript
try {
  await ValidateConfig(config);
  console.log('Configuration is valid');
} catch (error) {
  console.error('Invalid configuration:', error.message);
}
```

---

#### `GetDefaultConfigPath(): Promise<string>`
Returns the default configuration file path.

**Returns:** Path to default config file location.

```typescript
const defaultPath = await GetDefaultConfigPath();
// Returns: ~/.smtp-edc/config.yaml
```

---

#### `ListConfigFiles(): Promise<string[]>`
Lists available configuration files in the config directory.

**Returns:** Array of configuration file names.

```typescript
const configFiles = await ListConfigFiles();
configFiles.forEach(file => console.log(`Available config: ${file}`));
```

---

#### `CreateDefaultTemplates(): Promise<TemplateResult>`
Creates default email templates in the templates directory.

**Returns:** Result indicating success/failure of template creation.

```typescript
const result = await CreateDefaultTemplates();
if (result.success) {
  console.log('Default templates created successfully');
}
```

## SMTP Service

### Methods

#### `TestConnection(config: ConfigData): Promise<ConnectionResult>`
Tests SMTP server connection with provided configuration.

**Parameters:**
- `config`: SMTP configuration to test

**Returns:** Connection test results including capabilities and timing.

```typescript
const result = await TestConnection(config);
if (result.success) {
  console.log(`Connected successfully: ${result.message}`);
  console.log(`Server capabilities: ${result.capabilities}`);
} else {
  console.error(`Connection failed: ${result.error}`);
}
```

---

#### `SendEmail(request: EmailRequest): Promise<SendResult>`
Sends an email using the specified configuration and message details.

**Parameters:**
- `request`: Complete email request with config and message details

**Returns:** Send result with success status and message ID.

```typescript
const emailRequest: EmailRequest = {
  config: await GetCurrentConfig(),
  from: 'sender@example.com',
  to: ['recipient@example.com'],
  cc: [],
  bcc: [],
  subject: 'Test Email',
  body: 'This is a test email from SMTP-EDC',
  htmlBody: '<h1>Test Email</h1><p>This is a test email from SMTP-EDC</p>',
  attachments: [],
  headers: { 'X-Priority': '1' }
};

const result = await SendEmail(emailRequest);
if (result.success) {
  console.log(`Email sent successfully. Message ID: ${result.messageId}`);
}
```

---

#### `ValidateEmailAddress(email: string): Promise<SendResult>`
Validates a single email address syntax and optionally MX records.

**Parameters:**
- `email`: Email address to validate

**Returns:** Validation result with detailed information.

```typescript
const result = await ValidateEmailAddress('test@example.com');
if (result.success) {
  console.log('Email address is valid');
} else {
  console.error(`Invalid email: ${result.error}`);
}
```

---

#### `ValidateEmailList(emails: string[], validateMX: boolean): Promise<SendResult>`
Validates multiple email addresses with optional MX record checking.

**Parameters:**
- `emails`: Array of email addresses to validate
- `validateMX`: Whether to perform MX record validation

**Returns:** Validation results for all provided addresses.

```typescript
const emails = ['valid@example.com', 'invalid-email', 'test@domain.com'];
const result = await ValidateEmailList(emails, true);
console.log(`Validation results: ${result.message}`);
```

---

#### `GetAuthStats(): Promise<Record<string, unknown>>`
Returns authentication statistics and connection metrics.

**Returns:** Object containing authentication statistics.

```typescript
const stats = await GetAuthStats();
console.log('Auth stats:', stats);
```

---

#### `SetDebugMode(enabled: boolean): Promise<void>`
Enables or disables debug mode for SMTP operations.

**Parameters:**
- `enabled`: Whether to enable debug mode

```typescript
await SetDebugMode(true);
console.log('Debug mode enabled');
```

---

#### `GetDebugMode(): Promise<boolean>`
Returns current debug mode status.

**Returns:** Current debug mode state.

```typescript
const isDebugEnabled = await GetDebugMode();
console.log(`Debug mode: ${isDebugEnabled ? 'enabled' : 'disabled'}`);
```

## Template Service

### Methods

#### `ListTemplates(): Promise<TemplateInfo[]>`
Lists all available email templates.

**Returns:** Array of template information objects.

```typescript
const templates = await ListTemplates();
templates.forEach(template => {
  console.log(`Template: ${template.name} (${template.path})`);
});
```

---

#### `LoadTemplate(name: string): Promise<TemplateResult>`
Loads a specific template by name.

**Parameters:**
- `name`: Name of the template to load

**Returns:** Template content and metadata.

```typescript
const template = await LoadTemplate('welcome-email');
if (template.success) {
  console.log(`Template content: ${template.content}`);
}
```

---

#### `SaveTemplate(name: string, content: string): Promise<TemplateResult>`
Saves a template with the specified name and content.

**Parameters:**
- `name`: Template name
- `content`: Template content with placeholders

**Returns:** Save operation result.

```typescript
const templateContent = `
Subject: Welcome {{.Name}}!

Hello {{.Name}},

Welcome to our service! Your account has been created successfully.

Best regards,
The {{.Company}} Team
`;

const result = await SaveTemplate('welcome-email', templateContent);
if (result.success) {
  console.log('Template saved successfully');
}
```

---

#### `DeleteTemplate(name: string): Promise<TemplateResult>`
Deletes a template by name.

**Parameters:**
- `name`: Name of the template to delete

**Returns:** Deletion operation result.

```typescript
const result = await DeleteTemplate('old-template');
if (result.success) {
  console.log('Template deleted successfully');
}
```

---

#### `ExecuteTemplate(templateName: string, subjectTemplate: string, data: TemplateData): Promise<EmailRequest>`
Executes a template with provided data, returning a ready-to-send email request.

**Parameters:**
- `templateName`: Name of the template to execute
- `subjectTemplate`: Subject line template
- `data`: Template data including variables and recipient info

**Returns:** Populated email request ready for sending.

```typescript
const templateData: TemplateData = {
  from: 'noreply@company.com',
  to: ['user@example.com'],
  cc: [],
  bcc: [],
  subject: '', // Will be populated by template
  data: {
    Name: 'John Doe',
    Company: 'ACME Corp',
    ActivationLink: 'https://example.com/activate/abc123'
  }
};

const emailRequest = await ExecuteTemplate(
  'welcome-email',
  'Welcome {{.Name}} to {{.Company}}!',
  templateData
);

// Now send the email
const result = await SendEmail(emailRequest);
```

## Type Definitions

### Core Types

```typescript
// Authentication types
type AuthType = 'PLAIN' | 'LOGIN' | 'CRAM-MD5' | 'OAUTH2' | 'XOAUTH2' | 'NONE';

// Configuration structure
interface ConfigData {
  server: string;
  port: number;
  username: string;
  password: string;
  authType: AuthType;
  startTLS: boolean;
  skipVerify: boolean;
  templates: Record<string, string>;
}

// Email request structure
interface EmailRequest {
  config?: ConfigData;
  from: string;
  to: string[];
  cc: string[];
  bcc: string[];
  subject: string;
  body: string;
  htmlBody: string;
  attachments: string[];
  headers: Record<string, string>;
}

// Template variables
interface TemplateVariables {
  [key: string]: string | number | boolean | Date;
}

// Template data for execution
interface TemplateData {
  from: string;
  to: string[];
  cc: string[];
  bcc: string[];
  subject: string;
  data: TemplateVariables;
}
```

### Result Types

```typescript
// Connection test result
interface ConnectionResult {
  success: boolean;
  message: string;
  error?: string;
  timestamp: string;
  duration?: number;
  capabilities?: string[];
}

// Email send result
interface SendResult {
  success: boolean;
  message: string;
  error?: string;
  timestamp: string;
  messageId?: string;
}

// Template operation result
interface TemplateResult {
  success: boolean;
  message: string;
  error?: string;
  content?: string;
  path?: string;
}

// Template information
interface TemplateInfo {
  name: string;
  path: string;
  size: number;
  modified: string;
}
```

## Error Handling

### Custom Error Classes

The API provides custom error classes for different types of failures:

```typescript
class SMTPError extends Error {
  constructor(
    message: string,
    public code?: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'SMTPError';
  }
}

class ConfigError extends Error {
  constructor(
    message: string,
    public field?: string,
    public value?: unknown
  ) {
    super(message);
    this.name = 'ConfigError';
  }
}

class TemplateError extends Error {
  constructor(
    message: string,
    public templateName?: string
  ) {
    super(message);
    this.name = 'TemplateError';
  }
}
```

### Error Handling Examples

```typescript
try {
  const result = await SendEmail(emailRequest);
  if (!result.success) {
    throw new SMTPError(result.error || 'Unknown error');
  }
} catch (error) {
  if (error instanceof SMTPError) {
    console.error('SMTP Error:', error.message, error.code);
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## Examples

### Complete Email Sending Workflow

```typescript
import {
  GetCurrentConfig,
  LoadConfig,
  TestConnection,
  SendEmail,
  ValidateEmailAddress,
  ExecuteTemplate
} from './wailsjs/go/main/App';

async function sendWelcomeEmail(userEmail: string, userName: string) {
  try {
    // 1. Load configuration
    const config = await LoadConfig('production.yaml');

    // 2. Validate recipient email
    const validation = await ValidateEmailAddress(userEmail);
    if (!validation.success) {
      throw new Error(`Invalid email address: ${userEmail}`);
    }

    // 3. Test connection
    const connectionTest = await TestConnection(config);
    if (!connectionTest.success) {
      throw new Error(`Cannot connect to SMTP server: ${connectionTest.error}`);
    }

    // 4. Execute template
    const templateData = {
      from: config.username,
      to: [userEmail],
      cc: [],
      bcc: [],
      subject: '',
      data: {
        Name: userName,
        Company: 'SMTP-EDC',
        Date: new Date().toLocaleDateString()
      }
    };

    const emailRequest = await ExecuteTemplate(
      'welcome-email',
      'Welcome {{.Name}} to {{.Company}}!',
      templateData
    );

    // 5. Send email
    const sendResult = await SendEmail(emailRequest);
    if (sendResult.success) {
      console.log(`Welcome email sent to ${userEmail}. Message ID: ${sendResult.messageId}`);
      return sendResult.messageId;
    } else {
      throw new Error(`Failed to send email: ${sendResult.error}`);
    }

  } catch (error) {
    console.error('Failed to send welcome email:', error);
    throw error;
  }
}

// Usage
sendWelcomeEmail('newuser@example.com', 'John Doe')
  .then(messageId => console.log('Email sent successfully:', messageId))
  .catch(error => console.error('Error:', error.message));
```

### Bulk Email Sending

```typescript
async function sendBulkEmails(recipientList: Array<{email: string, name: string}>) {
  const config = await GetCurrentConfig();
  const results = [];

  for (const recipient of recipientList) {
    try {
      const emailRequest = {
        config,
        from: config.username,
        to: [recipient.email],
        cc: [],
        bcc: [],
        subject: `Hello ${recipient.name}!`,
        body: `Dear ${recipient.name},\n\nThis is a personalized message for you.`,
        htmlBody: `<h1>Hello ${recipient.name}!</h1><p>This is a personalized message for you.</p>`,
        attachments: [],
        headers: {}
      };

      const result = await SendEmail(emailRequest);
      results.push({ email: recipient.email, success: result.success, messageId: result.messageId });

      // Add delay to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 1000));

    } catch (error) {
      results.push({ email: recipient.email, success: false, error: error.message });
    }
  }

  return results;
}
```

### Configuration Management

```typescript
async function setupConfiguration() {
  // Create a new configuration
  const newConfig = {
    server: 'smtp.gmail.com',
    port: 587,
    username: 'your-email@gmail.com',
    password: 'your-app-password',
    authType: 'PLAIN' as AuthType,
    startTLS: true,
    skipVerify: false,
    templates: {}
  };

  // Validate the configuration
  try {
    await ValidateConfig(newConfig);
    console.log('Configuration is valid');
  } catch (error) {
    console.error('Configuration validation failed:', error);
    return;
  }

  // Save the configuration
  await SaveConfig(newConfig, 'gmail-config.yaml');
  console.log('Configuration saved successfully');

  // Test the connection
  const testResult = await TestConnection(newConfig);
  if (testResult.success) {
    console.log('SMTP connection test successful');
    console.log('Server capabilities:', testResult.capabilities);
  } else {
    console.error('SMTP connection test failed:', testResult.error);
  }
}
```

This API provides a comprehensive interface for all SMTP-EDC operations, with full TypeScript support, proper error handling, and extensive customization options.
