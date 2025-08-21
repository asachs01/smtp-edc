#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListResourcesRequestSchema,
  ListToolsRequestSchema,
  ReadResourceRequestSchema,
  ErrorCode,
  McpError
} from '@modelcontextprotocol/sdk/types.js';
import nodemailer from 'nodemailer';
import dns from 'dns/promises';
import validator from 'validator';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import os from 'os';
import { SecurityManager, AuditLogger } from './security.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Configuration management
class Config {
  constructor() {
    this.userConfig = {};
    this.defaultConfig = {
      default_server: '',
      default_port: 587,
      debug_mode: false,
      timeout: 30000,
      rate_limit: 10
    };
    this.loadConfig();
  }

  async loadConfig() {
    try {
      const configPath = path.join(os.homedir(), '.smtp-edc', 'config.json');
      const data = await fs.readFile(configPath, 'utf8');
      this.userConfig = JSON.parse(data);
    } catch (error) {
      // Config file doesn't exist or is invalid, use defaults
      this.userConfig = {};
    }
  }

  get(key) {
    return this.userConfig[key] ?? this.defaultConfig[key];
  }
}

// Rate limiter implementation
class RateLimiter {
  constructor(maxPerMinute) {
    this.maxPerMinute = maxPerMinute;
    this.timestamps = [];
  }

  async checkLimit() {
    if (this.maxPerMinute === 0) return true;
    
    const now = Date.now();
    const oneMinuteAgo = now - 60000;
    
    // Remove timestamps older than 1 minute
    this.timestamps = this.timestamps.filter(ts => ts > oneMinuteAgo);
    
    if (this.timestamps.length >= this.maxPerMinute) {
      throw new McpError(
        ErrorCode.InvalidRequest,
        `Rate limit exceeded. Maximum ${this.maxPerMinute} operations per minute.`
      );
    }
    
    this.timestamps.push(now);
    return true;
  }
}

// Template manager
class TemplateManager {
  constructor() {
    this.templatesDir = path.join(os.homedir(), '.smtp-edc', 'templates');
    this.ensureTemplatesDir();
  }

  async ensureTemplatesDir() {
    try {
      await fs.mkdir(this.templatesDir, { recursive: true });
    } catch (error) {
      console.error('Failed to create templates directory:', error);
    }
  }

  async listTemplates() {
    try {
      const files = await fs.readdir(this.templatesDir);
      return files.filter(f => f.endsWith('.html') || f.endsWith('.txt'));
    } catch (error) {
      return [];
    }
  }

  async loadTemplate(name) {
    const templatePath = path.join(this.templatesDir, name);
    try {
      const content = await fs.readFile(templatePath, 'utf8');
      return content;
    } catch (error) {
      throw new McpError(
        ErrorCode.InvalidRequest,
        `Template '${name}' not found`
      );
    }
  }

  processTemplate(template, variables = {}) {
    let processed = template;
    for (const [key, value] of Object.entries(variables)) {
      const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
      processed = processed.replace(regex, value);
    }
    return processed;
  }
}

// SMTP service implementation
class SMTPService {
  constructor(config, rateLimiter, securityManager, auditLogger) {
    this.config = config;
    this.rateLimiter = rateLimiter;
    this.securityManager = securityManager;
    this.auditLogger = auditLogger;
    this.connectionCache = new Map();
  }

  async testConnection(params) {
    const {
      server,
      port = this.config.get('default_port'),
      username,
      password,
      authType = 'plain',
      starttls = true,
      skipVerify = false
    } = params;

    if (!server) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Server parameter is required'
      );
    }

    // Security validation
    this.securityManager.validateServerParams(server, port);
    this.securityManager.validateCredentials(username, password);
    
    // Audit log
    this.auditLogger.log('CONNECTION_TEST', {
      server: server,
      port: port,
      auth: !!username
    });

    const transportConfig = {
      host: server,
      port: port,
      secure: port === 465,
      auth: username && password ? {
        user: username,
        pass: password,
        type: authType
      } : undefined,
      requireTLS: starttls,
      tls: {
        rejectUnauthorized: !skipVerify
      },
      connectionTimeout: this.config.get('timeout'),
      greetingTimeout: this.config.get('timeout'),
      socketTimeout: this.config.get('timeout')
    };

    try {
      const transporter = nodemailer.createTransporter(transportConfig);
      
      // Add timeout to the verify operation
      const verifyPromise = transporter.verify();
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error('Connection timeout')), this.config.get('timeout'));
      });
      
      await Promise.race([verifyPromise, timeoutPromise]);
      
      return {
        success: true,
        server: server,
        port: port,
        authType: authType,
        starttls: starttls,
        message: 'Connection successful',
        capabilities: {
          auth: !!username,
          tls: starttls || port === 465
        }
      };
    } catch (error) {
      return {
        success: false,
        server: server,
        port: port,
        error: error.message,
        code: error.code || 'CONNECTION_FAILED'
      };
    }
  }

  async sendEmail(params) {
    await this.rateLimiter.checkLimit();

    // Sanitize inputs
    params = {
      ...params,
      subject: this.securityManager.sanitizeInput(params.subject),
      body: this.securityManager.sanitizeInput(params.body)
    };

    const {
      server,
      port = this.config.get('default_port'),
      username,
      password,
      authType = 'plain',
      starttls = true,
      skipVerify = false,
      from,
      to,
      cc,
      bcc,
      subject,
      body,
      isHTML = false,
      attachments
    } = params;

    // Validate required parameters
    if (!server || !from || !to || !subject || !body) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Missing required parameters: server, from, to, subject, and body are required'
      );
    }

    // Security validations
    this.securityManager.validateServerParams(server, port);
    this.securityManager.validateRecipients(to, cc, bcc);
    this.securityManager.validateEmailContent({ subject, body, attachments });
    
    // Audit log
    this.auditLogger.log('SEND_EMAIL', {
      server: server,
      from: this.securityManager.maskSensitiveData(from),
      recipients: (Array.isArray(to) ? to.length : 1),
      hasAttachments: !!attachments
    });

    // Validate email addresses
    const validateEmails = (emails) => {
      const emailList = Array.isArray(emails) ? emails : [emails];
      return emailList.every(email => validator.isEmail(email));
    };

    if (!validator.isEmail(from)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Invalid sender email address: ${from}`
      );
    }

    if (!validateEmails(to)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Invalid recipient email address(es)'
      );
    }

    if (cc && !validateEmails(cc)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Invalid CC email address(es)'
      );
    }

    if (bcc && !validateEmails(bcc)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Invalid BCC email address(es)'
      );
    }

    const transportConfig = {
      host: server,
      port: port,
      secure: port === 465,
      auth: username && password ? {
        user: username,
        pass: password,
        type: authType
      } : undefined,
      requireTLS: starttls,
      tls: {
        rejectUnauthorized: !skipVerify
      },
      connectionTimeout: this.config.get('timeout'),
      greetingTimeout: this.config.get('timeout'),
      socketTimeout: this.config.get('timeout')
    };

    try {
      const transporter = nodemailer.createTransporter(transportConfig);
      
      const mailOptions = {
        from: from,
        to: Array.isArray(to) ? to.join(', ') : to,
        cc: cc ? (Array.isArray(cc) ? cc.join(', ') : cc) : undefined,
        bcc: bcc ? (Array.isArray(bcc) ? bcc.join(', ') : bcc) : undefined,
        subject: subject,
        [isHTML ? 'html' : 'text']: body,
        attachments: attachments
      };

      // Add timeout to the send operation
      const sendPromise = transporter.sendMail(mailOptions);
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error('Send timeout')), this.config.get('timeout'));
      });
      
      const info = await Promise.race([sendPromise, timeoutPromise]);
      
      return {
        success: true,
        messageId: info.messageId,
        accepted: info.accepted,
        rejected: info.rejected,
        response: info.response
      };
    } catch (error) {
      return {
        success: false,
        error: error.message,
        code: error.code || 'SEND_FAILED',
        command: error.command
      };
    }
  }

  async validateAddresses(params) {
    const { addresses, checkMX = false } = params;

    if (!addresses || !Array.isArray(addresses) || addresses.length === 0) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Addresses parameter must be a non-empty array'
      );
    }

    const results = [];

    for (const address of addresses) {
      const result = {
        address: address,
        valid: false,
        reason: '',
        mx: null
      };

      // Basic validation
      if (!validator.isEmail(address)) {
        result.reason = 'Invalid email format';
        results.push(result);
        continue;
      }

      result.valid = true;
      result.reason = 'Valid email format';

      // MX record checking
      if (checkMX) {
        const domain = address.split('@')[1];
        try {
          // Add timeout to DNS lookup
          const dnsPromise = dns.resolveMx(domain);
          const timeoutPromise = new Promise((_, reject) => {
            setTimeout(() => reject(new Error('DNS lookup timeout')), this.config.get('timeout'));
          });
          
          const mxRecords = await Promise.race([dnsPromise, timeoutPromise]);
          if (mxRecords && mxRecords.length > 0) {
            result.mx = mxRecords.sort((a, b) => a.priority - b.priority);
            result.reason = 'Valid email format with MX records';
          } else {
            result.valid = false;
            result.reason = 'Domain has no MX records';
          }
        } catch (error) {
          result.valid = false;
          result.reason = `MX lookup failed: ${error.message}`;
        }
      }

      results.push(result);
    }

    return {
      total: addresses.length,
      valid: results.filter(r => r.valid).length,
      invalid: results.filter(r => !r.valid).length,
      results: results
    };
  }
}

// Main MCP Server
class SMTPEDCServer {
  constructor() {
    this.config = new Config();
    this.rateLimiter = new RateLimiter(this.config.get('rate_limit'));
    this.templateManager = new TemplateManager();
    this.securityManager = new SecurityManager();
    this.auditLogger = new AuditLogger(this.config.get('debug_mode'));
    this.smtpService = new SMTPService(
      this.config, 
      this.rateLimiter, 
      this.securityManager, 
      this.auditLogger
    );
    this.server = new Server(
      {
        name: 'smtp-edc',
        version: '1.0.0'
      },
      {
        capabilities: {
          resources: {},
          tools: {}
        }
      }
    );

    this.setupHandlers();
  }

  setupHandlers() {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        {
          name: 'smtp_test_connection',
          description: 'Test SMTP server connection and capabilities',
          inputSchema: {
            type: 'object',
            properties: {
              server: { type: 'string', description: 'SMTP server hostname or IP' },
              port: { type: 'integer', description: 'SMTP server port', default: 587 },
              username: { type: 'string', description: 'Authentication username' },
              password: { type: 'string', description: 'Authentication password' },
              authType: { 
                type: 'string', 
                description: 'Authentication type',
                enum: ['plain', 'login', 'cram-md5'],
                default: 'plain'
              },
              starttls: { type: 'boolean', description: 'Use STARTTLS', default: true },
              skipVerify: { type: 'boolean', description: 'Skip TLS certificate verification', default: false }
            },
            required: ['server']
          }
        },
        {
          name: 'smtp_send_email',
          description: 'Send an email via SMTP',
          inputSchema: {
            type: 'object',
            properties: {
              server: { type: 'string', description: 'SMTP server hostname or IP' },
              port: { type: 'integer', description: 'SMTP server port', default: 587 },
              username: { type: 'string', description: 'Authentication username' },
              password: { type: 'string', description: 'Authentication password' },
              authType: { type: 'string', enum: ['plain', 'login', 'cram-md5'], default: 'plain' },
              starttls: { type: 'boolean', default: true },
              skipVerify: { type: 'boolean', default: false },
              from: { type: 'string', description: 'Sender email address' },
              to: { 
                oneOf: [
                  { type: 'string' },
                  { type: 'array', items: { type: 'string' } }
                ],
                description: 'Recipient email address(es)'
              },
              cc: { 
                oneOf: [
                  { type: 'string' },
                  { type: 'array', items: { type: 'string' } }
                ],
                description: 'CC recipient email address(es)'
              },
              bcc: { 
                oneOf: [
                  { type: 'string' },
                  { type: 'array', items: { type: 'string' } }
                ],
                description: 'BCC recipient email address(es)'
              },
              subject: { type: 'string', description: 'Email subject' },
              body: { type: 'string', description: 'Email body content' },
              isHTML: { type: 'boolean', description: 'Whether body is HTML', default: false }
            },
            required: ['server', 'from', 'to', 'subject', 'body']
          }
        },
        {
          name: 'smtp_validate_addresses',
          description: 'Validate email addresses with optional MX record checking',
          inputSchema: {
            type: 'object',
            properties: {
              addresses: {
                type: 'array',
                items: { type: 'string' },
                description: 'Email addresses to validate'
              },
              checkMX: {
                type: 'boolean',
                description: 'Check MX records for domains',
                default: false
              }
            },
            required: ['addresses']
          }
        },
        {
          name: 'smtp_load_template',
          description: 'Load and process an email template',
          inputSchema: {
            type: 'object',
            properties: {
              templateName: { type: 'string', description: 'Name of the template file' },
              variables: { 
                type: 'object',
                description: 'Variables to substitute in the template',
                additionalProperties: { type: 'string' }
              }
            },
            required: ['templateName']
          }
        }
      ]
    }));

    // List available resources
    this.server.setRequestHandler(ListResourcesRequestSchema, async () => ({
      resources: [
        {
          uri: 'smtp-edc://templates',
          name: 'Email Templates',
          description: 'List of available email templates',
          mimeType: 'application/json'
        },
        {
          uri: 'smtp-edc://config',
          name: 'Configuration',
          description: 'Current SMTP-EDC configuration',
          mimeType: 'application/json'
        }
      ]
    }));

    // Read resources
    this.server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
      const { uri } = request.params;

      switch (uri) {
        case 'smtp-edc://templates':
          const templates = await this.templateManager.listTemplates();
          return {
            contents: [{
              uri: uri,
              mimeType: 'application/json',
              text: JSON.stringify({ templates }, null, 2)
            }]
          };

        case 'smtp-edc://config':
          const config = {
            default_server: this.config.get('default_server'),
            default_port: this.config.get('default_port'),
            debug_mode: this.config.get('debug_mode'),
            timeout: this.config.get('timeout'),
            rate_limit: this.config.get('rate_limit')
          };
          return {
            contents: [{
              uri: uri,
              mimeType: 'application/json',
              text: JSON.stringify(config, null, 2)
            }]
          };

        default:
          throw new McpError(
            ErrorCode.InvalidRequest,
            `Unknown resource: ${uri}`
          );
      }
    });

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      try {
        let result;
        switch (name) {
          case 'smtp_test_connection':
            result = await this.smtpService.testConnection(args);
            break;

          case 'smtp_send_email':
            result = await this.smtpService.sendEmail(args);
            break;

          case 'smtp_validate_addresses':
            result = await this.smtpService.validateAddresses(args);
            break;

          case 'smtp_load_template':
            const { templateName, variables } = args;
            const template = await this.templateManager.loadTemplate(templateName);
            const processed = this.templateManager.processTemplate(template, variables);
            result = {
              templateName: templateName,
              original: template,
              processed: processed,
              variables: variables
            };
            break;

          default:
            throw new McpError(
              ErrorCode.MethodNotFound,
              `Unknown tool: ${name}`
            );
        }

        // Return result in proper MCP content format
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(result, null, 2)
            }
          ]
        };
      } catch (error) {
        if (error instanceof McpError) {
          throw error;
        }
        
        // Log unexpected errors in debug mode
        if (this.config.get('debug_mode')) {
          console.error(`Error in tool ${name}:`, error);
        }
        
        throw new McpError(
          ErrorCode.InternalError,
          `Tool execution failed: ${error.message}`
        );
      }
    });
  }

  async start() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    
    this.transport = transport;
    
    if (this.config.get('debug_mode')) {
      console.error('SMTP-EDC MCP Server started in debug mode');
    }
  }

  async shutdown() {
    try {
      if (this.config.get('debug_mode')) {
        console.error('SMTP-EDC MCP Server shutting down...');
      }
      
      // Close any open SMTP connections
      if (this.smtpService && this.smtpService.connectionCache) {
        for (const [key, connection] of this.smtpService.connectionCache) {
          try {
            if (connection && connection.close) {
              await connection.close();
            }
          } catch (error) {
            if (this.config.get('debug_mode')) {
              console.error(`Error closing connection ${key}:`, error);
            }
          }
        }
        this.smtpService.connectionCache.clear();
      }
      
      // Disconnect MCP server
      if (this.server) {
        await this.server.close();
      }
      
      if (this.config.get('debug_mode')) {
        console.error('SMTP-EDC MCP Server shutdown complete');
      }
    } catch (error) {
      console.error('Error during shutdown:', error);
    }
  }
}

// Create server instance
const server = new SMTPEDCServer();

// Signal handling for graceful shutdown
const gracefulShutdown = async (signal) => {
  console.error(`Received ${signal}, shutting down gracefully...`);
  try {
    await server.shutdown();
    process.exit(0);
  } catch (error) {
    console.error('Error during shutdown:', error);
    process.exit(1);
  }
};

// Handle termination signals
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));

// Error handling
process.on('uncaughtException', async (error) => {
  console.error('Uncaught exception:', error);
  try {
    await server.shutdown();
  } catch (shutdownError) {
    console.error('Error during emergency shutdown:', shutdownError);
  }
  process.exit(1);
});

process.on('unhandledRejection', async (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason);
  try {
    await server.shutdown();
  } catch (shutdownError) {
    console.error('Error during emergency shutdown:', shutdownError);
  }
  process.exit(1);
});

// Start the server
server.start().catch(async (error) => {
  console.error('Failed to start server:', error);
  try {
    await server.shutdown();
  } catch (shutdownError) {
    console.error('Error during startup failure shutdown:', shutdownError);
  }
  process.exit(1);
});