# Email Provider Configuration Guide

This guide provides SMTP configuration details for major email service providers to use with SMTP-EDC Desktop Extension.

## Table of Contents
- [Gmail](#gmail)
- [Microsoft 365 / Outlook](#microsoft-365--outlook)
- [Yahoo Mail](#yahoo-mail)
- [iCloud Mail](#icloud-mail)
- [Amazon SES](#amazon-ses)
- [SendGrid](#sendgrid)
- [Mailgun](#mailgun)
- [Postmark](#postmark)
- [Custom/Corporate Servers](#customcorporate-servers)

---

## Gmail

### Configuration
- **Server**: `smtp.gmail.com`
- **Port**: `587` (STARTTLS) or `465` (SSL/TLS)
- **Authentication**: OAuth2 (recommended) or App Password
- **Username**: Your full Gmail address (e.g., `user@gmail.com`)

### Setup Instructions
1. Enable 2-Step Verification in your Google Account
2. Generate an App Password:
   - Go to [Google Account Settings](https://myaccount.google.com/security)
   - Select "2-Step Verification"
   - Click on "App passwords"
   - Generate a new app password for "Mail"
3. Use the generated app password instead of your regular password

### Important Notes
- Google no longer supports "Less secure app access"
- App passwords are required for SMTP authentication
- Consider using OAuth2 for production applications

### Reference
- [Gmail SMTP Settings](https://support.google.com/mail/answer/7126229)
- [App Passwords Guide](https://support.google.com/accounts/answer/185833)

---

## Microsoft 365 / Outlook

### Configuration
- **Server**: `smtp.office365.com` or `smtp-mail.outlook.com`
- **Port**: `587` (STARTTLS)
- **Authentication**: OAuth2 (recommended) or Basic Auth
- **Username**: Your full email address

### Setup Instructions
1. For personal Outlook.com accounts:
   - Enable two-factor authentication
   - Create an app password from [Account Security](https://account.microsoft.com/security)
2. For Microsoft 365 business accounts:
   - Ensure SMTP AUTH is enabled for your account
   - Contact your administrator if authentication fails

### Important Notes
- Microsoft is deprecating Basic Authentication
- OAuth2 is recommended for new implementations
- Some organizations may have additional security policies

### Reference
- [Microsoft 365 SMTP Settings](https://docs.microsoft.com/en-us/exchange/mail-flow-best-practices/how-to-set-up-a-multifunction-device-or-application-to-send-email-using-microsoft-365-or-office-365)
- [Outlook.com SMTP Settings](https://support.microsoft.com/en-us/office/pop-imap-and-smtp-settings-for-outlook-com-d088b986-291d-42b8-9564-9c414e2aa040)

---

## Yahoo Mail

### Configuration
- **Server**: `smtp.mail.yahoo.com`
- **Port**: `587` (STARTTLS) or `465` (SSL/TLS)
- **Authentication**: App Password required
- **Username**: Your Yahoo email address (without @yahoo.com)

### Setup Instructions
1. Enable two-step verification
2. Generate an app password:
   - Go to [Yahoo Account Security](https://login.yahoo.com/account/security)
   - Click "Generate app password"
   - Select "Other App" and name it
3. Use the generated app password for SMTP authentication

### Reference
- [Yahoo Mail SMTP Settings](https://help.yahoo.com/kb/SLN4724.html)

---

## iCloud Mail

### Configuration
- **Server**: `smtp.mail.me.com`
- **Port**: `587` (STARTTLS)
- **Authentication**: App-specific password required
- **Username**: Your iCloud email address

### Setup Instructions
1. Enable two-factor authentication for your Apple ID
2. Generate an app-specific password:
   - Sign in to [Apple ID](https://appleid.apple.com)
   - Go to Security section
   - Generate an app-specific password
3. Use the app-specific password for SMTP authentication

### Reference
- [iCloud Mail Server Settings](https://support.apple.com/en-us/HT202304)
- [App-Specific Passwords](https://support.apple.com/en-us/HT204397)

---

## Amazon SES

### Configuration
- **Server**: `email-smtp.[region].amazonaws.com`
  - Example: `email-smtp.us-east-1.amazonaws.com`
- **Port**: `587` (STARTTLS) or `465` (SSL/TLS)
- **Authentication**: SMTP credentials (not AWS credentials)
- **Username**: SMTP username from AWS Console

### Setup Instructions
1. Verify your domain or email address in SES
2. Create SMTP credentials:
   - Go to Amazon SES Console
   - Navigate to "SMTP Settings"
   - Create "My SMTP Credentials"
3. Note: SMTP credentials are different from AWS IAM credentials

### Important Notes
- Start in sandbox mode (limited sending)
- Request production access when ready
- Monitor sending quotas and reputation

### Reference
- [Amazon SES SMTP Settings](https://docs.aws.amazon.com/ses/latest/dg/smtp-connect.html)
- [Getting Started with SES](https://docs.aws.amazon.com/ses/latest/dg/getting-started.html)

---

## SendGrid

### Configuration
- **Server**: `smtp.sendgrid.net`
- **Port**: `587` (STARTTLS) or `465` (SSL/TLS)
- **Authentication**: API Key as password
- **Username**: `apikey` (literal string)
- **Password**: Your SendGrid API key

### Setup Instructions
1. Create a SendGrid account
2. Generate an API key:
   - Go to Settings → API Keys
   - Create a new API key with "Mail Send" permissions
3. Use "apikey" as username and the API key as password

### Reference
- [SendGrid SMTP Documentation](https://docs.sendgrid.com/for-developers/sending-email/integrating-with-the-smtp-api)

---

## Mailgun

### Configuration
- **Server**: `smtp.mailgun.org` (US) or `smtp.eu.mailgun.org` (EU)
- **Port**: `587` (STARTTLS) or `465` (SSL/TLS)
- **Authentication**: SMTP credentials
- **Username**: Default SMTP login from Mailgun dashboard
- **Password**: Default password from Mailgun dashboard

### Setup Instructions
1. Add and verify your domain
2. Find SMTP credentials:
   - Go to Sending → Domain settings
   - Click on your domain
   - Find SMTP credentials section

### Reference
- [Mailgun SMTP Documentation](https://documentation.mailgun.com/en/latest/user_manual.html#sending-via-smtp)

---

## Postmark

### Configuration
- **Server**: `smtp.postmarkapp.com`
- **Port**: `587` (STARTTLS) or `2525` (alternative)
- **Authentication**: Server API Token
- **Username**: Server API Token
- **Password**: Server API Token

### Setup Instructions
1. Create a Postmark server
2. Get your Server API Token:
   - Go to Servers → Your Server
   - API Tokens tab
3. Use the same token for both username and password

### Reference
- [Postmark SMTP Documentation](https://postmarkapp.com/developer/user-guide/sending-email/sending-with-smtp)

---

## Custom/Corporate Servers

### Common Configurations

#### Exchange Server (On-Premise)
- **Port**: Usually `587` (STARTTLS) or `25` (internal)
- **Authentication**: NTLM or Basic
- **Username**: Domain\Username or email address

#### Postfix/Sendmail
- **Port**: `25` (internal) or `587` (submission)
- **Authentication**: PLAIN, LOGIN, or CRAM-MD5
- **Username**: System username or email address

### Troubleshooting Tips

1. **Connection Issues**
   - Verify firewall rules allow outbound SMTP
   - Check if ISP blocks port 25
   - Try alternative ports (587, 2525, 465)

2. **Authentication Failures**
   - Confirm username format (email vs. username)
   - Check if app passwords are required
   - Verify account has SMTP permissions

3. **TLS/SSL Problems**
   - Try both STARTTLS (587) and SSL/TLS (465)
   - Check certificate validity
   - Consider `skip_tls_verify` for testing only

4. **Rate Limiting**
   - Check provider's sending limits
   - Implement proper rate limiting in configuration
   - Monitor for throttling responses

### Security Best Practices

1. **Always use encrypted connections** (STARTTLS or SSL/TLS)
2. **Never store passwords in plain text** - use environment variables or secure vaults
3. **Use app-specific passwords** instead of account passwords
4. **Implement rate limiting** to avoid being flagged as spam
5. **Monitor bounce rates** and maintain good sender reputation
6. **Verify SPF, DKIM, and DMARC** records for your domain

---

## Quick Reference Table

| Provider | Server | Port | Auth Method |
|----------|--------|------|-------------|
| Gmail | smtp.gmail.com | 587/465 | App Password |
| Outlook | smtp.office365.com | 587 | App Password/OAuth2 |
| Yahoo | smtp.mail.yahoo.com | 587/465 | App Password |
| iCloud | smtp.mail.me.com | 587 | App-Specific Password |
| Amazon SES | email-smtp.[region].amazonaws.com | 587/465 | SMTP Credentials |
| SendGrid | smtp.sendgrid.net | 587/465 | API Key |
| Mailgun | smtp.mailgun.org | 587/465 | SMTP Credentials |
| Postmark | smtp.postmarkapp.com | 587/2525 | Server Token |

---

## Need Help?

If you encounter issues with a specific provider:

1. Check the provider's official SMTP documentation
2. Verify your account has SMTP access enabled
3. Ensure you're using the correct authentication method
4. Test with `debug_mode: true` in the extension configuration
5. Check the [SMTP-EDC GitHub Issues](https://github.com/asachs01/smtp-edc/issues) for similar problems

For providers not listed here, consult their official documentation or contact their support team for SMTP configuration details.