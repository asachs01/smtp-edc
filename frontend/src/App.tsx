import React, { useState, useEffect } from 'react';
import logo from './assets/images/smtp-edc.png';
import './App.css';
import {
  TestConnection,
  SendEmail,
  LoadConfig,
  SaveConfig,
} from '../wailsjs/go/main/App';
import { services, type EmailComposition } from './types/api';

function App() {
  // Connection state
  const [config, setConfig] = useState<services.ConfigData>({
    server: '',
    port: 587,
    username: '',
    password: '',
    authType: 'plain',
    startTLS: true,
    skipVerify: false,
    templates: {},
  });

  // Email composition state
  const [email, setEmail] = useState<EmailComposition>({
    from: '',
    to: [''],
    cc: [],
    bcc: [],
    subject: '',
    body: '',
    htmlBody: '',
    attachments: [],
    headers: {},
  });

  // UI state
  const [connectionResult, setConnectionResult] = useState<string>('');
  const [emailResult, setEmailResult] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Load config on startup
  useEffect(() => {
    LoadConfig('')
      .then((configData: services.ConfigData) => {
        if (configData && configData.server) {
          setConfig((prev: services.ConfigData) => ({
            ...prev,
            ...configData,
          }));
        }
      })
      .catch((error: unknown) => {
        // eslint-disable-next-line no-console
        console.error('Failed to load config:', error);
      });
  }, []);

  // Helper functions
  const updateConfig = (field: keyof services.ConfigData, value: unknown) => {
    setConfig((prev: services.ConfigData) => ({ ...prev, [field]: value }));
  };

  const updateEmail = (field: keyof EmailComposition, value: unknown) => {
    setEmail((prev: EmailComposition) => ({ ...prev, [field]: value }));
  };

  // Helper to ensure array fields are always arrays
  const ensureArray = (field: string | string[] | undefined): string[] => {
    if (Array.isArray(field)) return field;
    if (typeof field === 'string') return [field];
    return [];
  };

  const addEmailAddress = (field: 'to' | 'cc' | 'bcc') => {
    setEmail((prev: EmailComposition) => {
      const currentArray = ensureArray(prev[field]);
      return {
        ...prev,
        [field]: [...currentArray, ''],
      };
    });
  };

  const removeEmailAddress = (
    _: unknown,
    i: number,
    field: 'to' | 'cc' | 'bcc'
  ) => {
    setEmail((prev: EmailComposition) => {
      const currentArray = ensureArray(prev[field]);
      return {
        ...prev,
        [field]: currentArray.filter((_, index) => index !== i),
      };
    });
  };

  const updateEmailAddress = (
    field: 'to' | 'cc' | 'bcc',
    index: number,
    value: string
  ) => {
    setEmail((prev: EmailComposition) => {
      const currentArray = ensureArray(prev[field]);
      return {
        ...prev,
        [field]: currentArray.map((addr: string, i: number) =>
          i === index ? value : addr
        ),
      };
    });
  };

  // Connection testing
  const testConnection = async () => {
    setIsLoading(true);
    try {
      const result = await TestConnection(config);
      setConnectionResult(
        result.success ? '✅ Connection successful' : `❌ ${result.message}`
      );
    } catch (error) {
      setConnectionResult(`❌ Connection failed: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const saveConfiguration = async () => {
    try {
      await SaveConfig(config, '');
      setConnectionResult('✅ Configuration saved');
    } catch (error) {
      setConnectionResult(`❌ Save failed: ${error}`);
    }
  };

  // Email sending
  const sendTestEmail = async () => {
    setIsLoading(true);
    try {
      const toArray = ensureArray(email.to);
      const ccArray = ensureArray(email.cc);
      const bccArray = ensureArray(email.bcc);

      const emailRequest: services.EmailRequest = {
        from: email.from || '',
        to: toArray.filter((addr: string) => addr.trim() !== ''),
        cc: ccArray.filter((addr: string) => addr.trim() !== ''),
        bcc: bccArray.filter((addr: string) => addr.trim() !== ''),
        subject: email.subject || '',
        body: email.body || '',
        htmlBody: email.htmlBody || '',
        attachments: email.attachments || [],
        headers: email.headers || {},
        convertValues: () => ({}),
      };

      const result = await SendEmail(emailRequest);
      setEmailResult(
        result.success ? '✅ Email sent successfully' : `❌ ${result.message}`
      );
    } catch (error) {
      setEmailResult(`❌ Send failed: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div id='App'>
      {/* Header */}
      <header className='app-header'>
        <img src={logo} className='app-logo' alt='SMTP-EDC' />
        <h1>SMTP Email Delivery Checker</h1>
        <p>Test SMTP connections and send emails</p>
      </header>

      {/* Main Content */}
      <div className='main-content'>
        {/* Connection Settings Section */}
        <section className='settings-section'>
          <h2>SMTP Connection Settings</h2>
          <div className='form-grid'>
            <div className='input-group'>
              <label htmlFor='server'>SMTP Server:</label>
              <input
                id='server'
                type='text'
                value={config.server}
                onChange={e => updateConfig('server', e.target.value)}
                placeholder='smtp.gmail.com'
              />
            </div>

            <div className='input-group'>
              <label htmlFor='port'>Port:</label>
              <input
                id='port'
                type='number'
                value={config.port}
                onChange={e =>
                  updateConfig('port', parseInt(e.target.value) || 587)
                }
                placeholder='587'
              />
            </div>

            <div className='input-group'>
              <label htmlFor='username'>Username:</label>
              <input
                id='username'
                type='text'
                value={config.username}
                onChange={e => updateConfig('username', e.target.value)}
                placeholder='your-email@example.com'
              />
            </div>

            <div className='input-group'>
              <label htmlFor='password'>Password:</label>
              <input
                id='password'
                type='password'
                value={config.password}
                onChange={e => updateConfig('password', e.target.value)}
                placeholder='Your password or app password'
              />
            </div>

            <div className='input-group'>
              <label htmlFor='authType'>Authentication:</label>
              <select
                id='authType'
                value={config.authType}
                onChange={e => updateConfig('authType', e.target.value)}
              >
                <option value='plain'>Plain</option>
                <option value='login'>Login</option>
                <option value='cram-md5'>CRAM-MD5</option>
                <option value='oauth2'>OAuth2</option>
              </select>
            </div>

            <div className='input-group checkbox-group'>
              <label>
                <input
                  type='checkbox'
                  checked={config.startTLS}
                  onChange={e => updateConfig('startTLS', e.target.checked)}
                />
                Use TLS/SSL
              </label>
            </div>

            <div className='input-group checkbox-group'>
              <label>
                <input
                  type='checkbox'
                  checked={config.skipVerify}
                  onChange={e => updateConfig('skipVerify', e.target.checked)}
                />
                Skip Certificate Verification
              </label>
            </div>
          </div>

          <div className='button-group'>
            <button
              onClick={testConnection}
              disabled={isLoading}
              className='btn-primary'
            >
              {isLoading ? 'Testing...' : 'Test Connection'}
            </button>
            <button onClick={saveConfiguration} className='btn-secondary'>
              Save Configuration
            </button>
          </div>

          {connectionResult && (
            <div className='result-display'>{connectionResult}</div>
          )}
        </section>

        {/* Email Composition Section */}
        <section className='settings-section'>
          <h2>Email Composition</h2>
          <div className='form-grid'>
            <div className='input-group'>
              <label htmlFor='from'>From:</label>
              <input
                id='from'
                type='email'
                value={email.from}
                onChange={e => updateEmail('from', e.target.value)}
                placeholder='sender@example.com'
              />
            </div>

            <div className='input-group'>
              <label>To Recipients:</label>
              {ensureArray(email.to).map((addr: string, index: number) => (
                <div key={index} className='email-address-row'>
                  <input
                    type='email'
                    value={addr}
                    onChange={e =>
                      updateEmailAddress('to', index, e.target.value)
                    }
                    placeholder='recipient@example.com'
                  />
                  {ensureArray(email.to).length > 1 && (
                    <button
                      type='button'
                      onClick={() => removeEmailAddress(null, index, 'to')}
                      className='btn-remove'
                    >
                      ×
                    </button>
                  )}
                </div>
              ))}
              <button
                type='button'
                onClick={() => addEmailAddress('to')}
                className='btn-add'
              >
                + Add Recipient
              </button>
            </div>

            <div className='input-group'>
              <label>CC Recipients:</label>
              {ensureArray(email.cc).map((addr: string, index: number) => (
                <div key={index} className='email-address-row'>
                  <input
                    type='email'
                    value={addr}
                    onChange={e =>
                      updateEmailAddress('cc', index, e.target.value)
                    }
                    placeholder='cc@example.com'
                  />
                  <button
                    type='button'
                    onClick={() => removeEmailAddress(null, index, 'cc')}
                    className='btn-remove'
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                type='button'
                onClick={() => addEmailAddress('cc')}
                className='btn-add'
              >
                + Add CC
              </button>
            </div>

            <div className='input-group'>
              <label htmlFor='subject'>Subject:</label>
              <input
                id='subject'
                type='text'
                value={email.subject}
                onChange={e => updateEmail('subject', e.target.value)}
                placeholder='Test Email Subject'
              />
            </div>

            <div className='input-group'>
              <label htmlFor='body'>Body (Text):</label>
              <textarea
                id='body'
                value={email.body}
                onChange={e => updateEmail('body', e.target.value)}
                placeholder='Your email message here...'
                rows={4}
              />
            </div>

            <div className='input-group'>
              <label htmlFor='htmlBody'>Body (HTML):</label>
              <textarea
                id='htmlBody'
                value={email.htmlBody}
                onChange={e => updateEmail('htmlBody', e.target.value)}
                placeholder='<h1>HTML content here...</h1>'
                rows={4}
              />
            </div>
          </div>

          <div className='button-group'>
            <button
              onClick={sendTestEmail}
              disabled={
                isLoading ||
                !email.from ||
                ensureArray(email.to).every((addr: string) => !addr.trim())
              }
              className='btn-primary'
            >
              {isLoading ? 'Sending...' : 'Send Test Email'}
            </button>
          </div>

          {emailResult && (
            <div className='result-display'>{emailResult}</div>
          )}
        </section>
      </div>
    </div>
  );
}

export default App;
