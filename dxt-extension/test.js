#!/usr/bin/env node

import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs/promises';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Test suite for SMTP-EDC DXT Extension
 */
class DXTTester {
  constructor() {
    this.serverPath = path.join(__dirname, 'server', 'index.js');
    this.testsPassed = 0;
    this.testsFailed = 0;
  }

  async runTest(name, request, expectedFields = []) {
    console.log(`\n📝 Testing: ${name}`);
    
    return new Promise((resolve) => {
      const child = spawn('node', [this.serverPath], {
        stdio: ['pipe', 'pipe', 'pipe']
      });

      let stdout = '';
      let stderr = '';
      let timeout;

      // Set timeout
      timeout = setTimeout(() => {
        child.kill();
        console.log(`❌ Test timed out`);
        this.testsFailed++;
        resolve(false);
      }, 5000);

      child.stdout.on('data', (data) => {
        stdout += data.toString();
      });

      child.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      child.on('close', (code) => {
        clearTimeout(timeout);
        
        try {
          // Parse JSONRPC response
          const lines = stdout.split('\n').filter(line => line.trim());
          const response = lines.find(line => {
            try {
              const parsed = JSON.parse(line);
              return parsed.jsonrpc === '2.0';
            } catch {
              return false;
            }
          });

          if (!response) {
            console.log(`❌ No valid JSONRPC response received`);
            console.log('stdout:', stdout);
            console.log('stderr:', stderr);
            this.testsFailed++;
            resolve(false);
            return;
          }

          const parsed = JSON.parse(response);
          
          // Check for expected fields
          let success = true;
          for (const field of expectedFields) {
            if (!this.hasNestedField(parsed, field)) {
              console.log(`❌ Missing expected field: ${field}`);
              success = false;
            }
          }

          if (success) {
            console.log(`✅ Test passed`);
            this.testsPassed++;
          } else {
            console.log(`❌ Test failed`);
            console.log('Response:', JSON.stringify(parsed, null, 2));
            this.testsFailed++;
          }

          resolve(success);
        } catch (error) {
          console.log(`❌ Error parsing response: ${error.message}`);
          console.log('stdout:', stdout);
          console.log('stderr:', stderr);
          this.testsFailed++;
          resolve(false);
        }
      });

      // Send the request
      const requestStr = JSON.stringify(request) + '\n';
      child.stdin.write(requestStr);
      child.stdin.end();
    });
  }

  hasNestedField(obj, path) {
    const parts = path.split('.');
    let current = obj;
    
    // Handle MCP content format if present
    if (current.result && current.result.content && Array.isArray(current.result.content)) {
      const content = current.result.content[0];
      if (content && content.type === 'text' && content.text) {
        try {
          // Parse the JSON content
          const parsed = JSON.parse(content.text);
          // Replace the result with the parsed content
          current.result = parsed;
        } catch {
          // If not JSON, keep as is
        }
      }
    }
    
    for (const part of parts) {
      if (part.includes('[')) {
        // Handle array notation
        const [field, index] = part.split('[');
        const idx = parseInt(index.replace(']', ''));
        if (!current[field] || !Array.isArray(current[field]) || !current[field][idx]) {
          return false;
        }
        current = current[field][idx];
      } else {
        if (!current[part]) {
          return false;
        }
        current = current[part];
      }
    }
    
    return true;
  }

  async testInitialization() {
    console.log('\n🔧 Testing Server Initialization...');
    
    const request = {
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '0.1.0',
        capabilities: {},
        clientInfo: {
          name: 'test-client',
          version: '1.0.0'
        }
      }
    };

    await this.runTest('Server initialization', request, [
      'result.protocolVersion',
      'result.capabilities',
      'result.serverInfo.name'
    ]);
  }

  async testListTools() {
    console.log('\n🔧 Testing List Tools...');
    
    const request = {
      jsonrpc: '2.0',
      id: 2,
      method: 'tools/list',
      params: {}
    };

    await this.runTest('List available tools', request, [
      'result.tools[0].name',
      'result.tools[0].description',
      'result.tools[0].inputSchema'
    ]);
  }

  async testListResources() {
    console.log('\n📚 Testing List Resources...');
    
    const request = {
      jsonrpc: '2.0',
      id: 3,
      method: 'resources/list',
      params: {}
    };

    await this.runTest('List available resources', request, [
      'result.resources[0].uri',
      'result.resources[0].name',
      'result.resources[0].mimeType'
    ]);
  }

  async testValidateAddresses() {
    console.log('\n✉️ Testing Email Validation...');
    
    const request = {
      jsonrpc: '2.0',
      id: 4,
      method: 'tools/call',
      params: {
        name: 'smtp_validate_addresses',
        arguments: {
          addresses: ['test@example.com', 'invalid-email', 'user@domain.org'],
          checkMX: false
        }
      }
    };

    await this.runTest('Validate email addresses', request, [
      'result.total',
      'result.valid',
      'result.invalid',
      'result.results'
    ]);
  }

  async testConnectionValidation() {
    console.log('\n🔒 Testing Security Validation...');
    
    // This should fail due to localhost restriction
    const request = {
      jsonrpc: '2.0',
      id: 5,
      method: 'tools/call',
      params: {
        name: 'smtp_test_connection',
        arguments: {
          server: 'localhost',
          port: 25
        }
      }
    };

    // We expect an error for this test
    await this.runTest('Security validation (should fail)', request, [
      'error.code',
      'error.message'
    ]);
  }

  async testManifestValidation() {
    console.log('\n📋 Validating Manifest...');
    
    try {
      const manifestPath = path.join(__dirname, 'manifest.json');
      const content = await fs.readFile(manifestPath, 'utf8');
      const manifest = JSON.parse(content);
      
      // Check required fields
      const required = ['dxt_version', 'name', 'version', 'description', 'author', 'server'];
      let valid = true;
      
      for (const field of required) {
        if (!manifest[field]) {
          console.log(`❌ Missing required field: ${field}`);
          valid = false;
        }
      }
      
      if (valid) {
        console.log('✅ Manifest validation passed');
        this.testsPassed++;
      } else {
        this.testsFailed++;
      }
    } catch (error) {
      console.log(`❌ Manifest validation failed: ${error.message}`);
      this.testsFailed++;
    }
  }

  async runAllTests() {
    console.log('🚀 Starting SMTP-EDC DXT Extension Tests\n');
    console.log('=' .repeat(50));

    await this.testManifestValidation();
    await this.testInitialization();
    await this.testListTools();
    await this.testListResources();
    await this.testValidateAddresses();
    await this.testConnectionValidation();

    console.log('\n' + '='.repeat(50));
    console.log('\n📊 Test Results:');
    console.log(`✅ Passed: ${this.testsPassed}`);
    console.log(`❌ Failed: ${this.testsFailed}`);
    console.log(`📈 Total: ${this.testsPassed + this.testsFailed}`);
    
    const successRate = ((this.testsPassed / (this.testsPassed + this.testsFailed)) * 100).toFixed(1);
    console.log(`🎯 Success Rate: ${successRate}%`);

    if (this.testsFailed === 0) {
      console.log('\n🎉 All tests passed!');
      process.exit(0);
    } else {
      console.log('\n⚠️  Some tests failed. Please review the output above.');
      process.exit(1);
    }
  }
}

// Run tests
const tester = new DXTTester();
tester.runAllTests().catch(error => {
  console.error('❌ Test suite failed:', error);
  process.exit(1);
});