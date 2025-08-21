import crypto from 'crypto';
import { McpError, ErrorCode } from '@modelcontextprotocol/sdk/types.js';

/**
 * Security utilities for SMTP-EDC MCP Server
 */
export class SecurityManager {
  constructor() {
    // Maximum allowed sizes
    this.MAX_EMAIL_SIZE = 25 * 1024 * 1024; // 25MB
    this.MAX_ATTACHMENT_SIZE = 10 * 1024 * 1024; // 10MB
    this.MAX_RECIPIENTS = 100;
    this.MAX_SUBJECT_LENGTH = 998; // RFC 5322
    
    // Blocked patterns for security
    this.BLOCKED_DOMAINS = new Set([
      'example.com',
      'test.com',
      'localhost'
    ]);
    
    // Sensitive data patterns
    this.SENSITIVE_PATTERNS = [
      /\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12})\b/, // Credit cards
      /\b[A-Z0-9]{20}\b/, // API keys (generic pattern)
      /-----BEGIN (?:RSA )?PRIVATE KEY-----/, // Private keys
    ];
  }

  /**
   * Sanitize input to prevent injection attacks
   */
  sanitizeInput(input) {
    if (typeof input !== 'string') return input;
    
    // Remove null bytes
    input = input.replace(/\0/g, '');
    
    // Remove control characters except common ones
    input = input.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '');
    
    // Limit length
    if (input.length > 10000) {
      input = input.substring(0, 10000);
    }
    
    return input;
  }

  /**
   * Validate email content for security issues
   */
  validateEmailContent(params) {
    const { subject, body, attachments } = params;
    
    // Check subject length
    if (subject && subject.length > this.MAX_SUBJECT_LENGTH) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Subject exceeds maximum length of ${this.MAX_SUBJECT_LENGTH} characters`
      );
    }
    
    // Check body size
    const bodySize = Buffer.byteLength(body || '', 'utf8');
    if (bodySize > this.MAX_EMAIL_SIZE) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Email body exceeds maximum size of ${this.MAX_EMAIL_SIZE} bytes`
      );
    }
    
    // Check for sensitive data
    this.checkSensitiveData(body);
    this.checkSensitiveData(subject);
    
    // Validate attachments
    if (attachments && Array.isArray(attachments)) {
      for (const attachment of attachments) {
        this.validateAttachment(attachment);
      }
    }
    
    return true;
  }

  /**
   * Check for sensitive data patterns
   */
  checkSensitiveData(text) {
    if (!text) return;
    
    for (const pattern of this.SENSITIVE_PATTERNS) {
      if (pattern.test(text)) {
        throw new McpError(
          ErrorCode.InvalidParams,
          'Email content appears to contain sensitive data (API keys, credit cards, or private keys). Please remove sensitive information before sending.'
        );
      }
    }
  }

  /**
   * Validate attachment security
   */
  validateAttachment(attachment) {
    const BLOCKED_EXTENSIONS = [
      '.exe', '.dll', '.bat', '.cmd', '.com', '.pif', '.scr',
      '.vbs', '.js', '.jar', '.zip', '.rar'
    ];
    
    if (attachment.filename) {
      const ext = attachment.filename.toLowerCase().substr(attachment.filename.lastIndexOf('.'));
      if (BLOCKED_EXTENSIONS.includes(ext)) {
        throw new McpError(
          ErrorCode.InvalidParams,
          `Attachment type '${ext}' is not allowed for security reasons`
        );
      }
    }
    
    if (attachment.content) {
      const size = Buffer.byteLength(attachment.content, 'base64');
      if (size > this.MAX_ATTACHMENT_SIZE) {
        throw new McpError(
          ErrorCode.InvalidParams,
          `Attachment exceeds maximum size of ${this.MAX_ATTACHMENT_SIZE} bytes`
        );
      }
    }
  }

  /**
   * Validate recipient list
   */
  validateRecipients(to, cc, bcc) {
    const allRecipients = [
      ...(Array.isArray(to) ? to : [to]),
      ...(cc ? (Array.isArray(cc) ? cc : [cc]) : []),
      ...(bcc ? (Array.isArray(bcc) ? bcc : [bcc]) : [])
    ].filter(Boolean);
    
    if (allRecipients.length > this.MAX_RECIPIENTS) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Total recipients (${allRecipients.length}) exceeds maximum of ${this.MAX_RECIPIENTS}`
      );
    }
    
    // Check for blocked domains
    for (const email of allRecipients) {
      const domain = email.split('@')[1];
      if (this.BLOCKED_DOMAINS.has(domain)) {
        throw new McpError(
          ErrorCode.InvalidParams,
          `Domain '${domain}' is blocked for security reasons`
        );
      }
    }
    
    return true;
  }

  /**
   * Validate server connection parameters
   */
  validateServerParams(server, port) {
    // Block localhost and private IPs
    const BLOCKED_HOSTS = [
      'localhost',
      '127.0.0.1',
      '0.0.0.0',
      '::1'
    ];
    
    if (BLOCKED_HOSTS.includes(server)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Connections to localhost are not allowed for security reasons'
      );
    }
    
    // Check for private IP ranges
    const privateIPRanges = [
      /^10\./,
      /^172\.(1[6-9]|2[0-9]|3[0-1])\./,
      /^192\.168\./
    ];
    
    for (const range of privateIPRanges) {
      if (range.test(server)) {
        throw new McpError(
          ErrorCode.InvalidParams,
          'Connections to private IP addresses are not allowed'
        );
      }
    }
    
    // Validate port
    const BLOCKED_PORTS = [22, 23, 135, 139, 445, 3389]; // SSH, Telnet, SMB, RDP
    if (BLOCKED_PORTS.includes(port)) {
      throw new McpError(
        ErrorCode.InvalidParams,
        `Port ${port} is blocked for security reasons`
      );
    }
    
    if (port < 1 || port > 65535) {
      throw new McpError(
        ErrorCode.InvalidParams,
        'Port must be between 1 and 65535'
      );
    }
    
    return true;
  }

  /**
   * Mask sensitive data in logs
   */
  maskSensitiveData(data) {
    if (typeof data !== 'string') return data;
    
    // Mask email addresses (keep domain)
    data = data.replace(/([a-zA-Z0-9._%+-]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})/g, 
      (match, local, domain) => {
        const masked = local.charAt(0) + '*'.repeat(Math.min(local.length - 1, 5));
        return `${masked}@${domain}`;
      });
    
    // Mask potential passwords
    data = data.replace(/"password"\s*:\s*"[^"]+"/gi, '"password":"***"');
    
    // Mask authorization headers
    data = data.replace(/authorization:\s*[^\s,]+/gi, 'authorization: ***');
    
    return data;
  }

  /**
   * Generate secure message ID
   */
  generateMessageId(domain = 'smtp-edc.local') {
    const timestamp = Date.now();
    const random = crypto.randomBytes(8).toString('hex');
    return `<${timestamp}.${random}@${domain}>`;
  }

  /**
   * Validate authentication credentials
   */
  validateCredentials(username, password) {
    if (!username || !password) {
      return true; // Allow anonymous SMTP
    }
    
    // Check for common weak passwords
    const WEAK_PASSWORDS = [
      'password', '123456', 'password123', 'admin', 'letmein',
      'qwerty', '111111', 'abc123', 'monkey', 'master'
    ];
    
    if (WEAK_PASSWORDS.includes(password.toLowerCase())) {
      console.warn('Warning: Weak password detected. Consider using a stronger password.');
    }
    
    // Check password length
    if (password.length < 8) {
      console.warn('Warning: Password is less than 8 characters. Consider using a longer password.');
    }
    
    return true;
  }
}

/**
 * Audit logger for security events
 */
export class AuditLogger {
  constructor(debugMode = false) {
    this.debugMode = debugMode;
    this.events = [];
    this.MAX_EVENTS = 1000;
  }

  log(event, details) {
    const entry = {
      timestamp: new Date().toISOString(),
      event: event,
      details: details
    };
    
    this.events.push(entry);
    
    // Maintain size limit
    if (this.events.length > this.MAX_EVENTS) {
      this.events.shift();
    }
    
    if (this.debugMode) {
      console.error(`[AUDIT] ${event}:`, details);
    }
  }

  getEvents(limit = 100) {
    return this.events.slice(-limit);
  }
}