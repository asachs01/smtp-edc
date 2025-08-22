# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Model Context Protocol (MCP) server support for AI assistant integration
- MCP tools for SMTP operations (test_connection, send_email, validate_addresses, load_template)
- MCP resources for accessing configuration, templates, and statistics
- MCP server configuration management
- Support for both STDIO and HTTP transports for MCP
- Comprehensive MCP integration documentation
- Example MCP client configurations
- Tests for MCP functionality
- Desktop Extension (DXT) packaging for SMTP-EDC
  - Full DXT v0.1 specification compliance
  - MCP server implementation with @modelcontextprotocol/sdk
  - Comprehensive security features (input validation, rate limiting, network security)
  - Build and packaging scripts for DXT distribution
  - Test suite with 100% pass rate
  - Icon and branding assets
- Enhanced user configuration options for DXT extension
  - Default username and from address settings
  - Authentication type selection (plain, login, cram-md5, oauth2)
  - STARTTLS and TLS verification controls
  - Attachment size and type restrictions
  - Template directory configuration
  - MX record validation toggle
- Email Provider Configuration Guide
  - Setup instructions for Gmail, Microsoft 365/Outlook, Yahoo Mail, iCloud
  - Configuration for Amazon SES, SendGrid, Mailgun, Postmark
  - Custom/corporate server guidelines
  - Security best practices and troubleshooting tips
  - Quick reference table for all providers
- CI/CD integration for DXT packaging
  - Automated DXT build in release workflow
  - Validation with official dxt pack command
  - Automatic upload of smtp-edc.dxt to GitHub releases
  - SHA256 checksum generation and distribution

## [v1.3.0] - 2025-07-16

### Added
- Wails v2 integration for desktop UI application
- React/TypeScript frontend with modern build tooling
- Integrated release workflow with UI builds
- Frontend development environment with ESLint and Prettier
- API documentation with OpenAPI specification
- Desktop application binary (smtp-edc-ui)
- Automated UI asset uploads to GitHub releases
- Release verification scripts

### Changed
- Enhanced release workflow to support multiple build artifacts
- Reorganized project structure to support UI components
- Updated build system to handle both CLI and UI builds

### Removed
- Duplicate configuration directories (.clinerules, .serena, .windsurf, .trae)

## [v1.2.0] - 2025-06-25

### Added
- Homebrew formula integrated into main repository
- GitHub Pages documentation site with Jekyll
- MIT License file
- Comprehensive README updates with installation instructions

### Changed
- GoReleaser configuration to specify main path correctly
- Release workflow improvements for better automation
- Updated documentation structure

### Fixed
- Homebrew formula test to use --help instead of --version
- Build path specifications in GoReleaser

## [v1.1.0] - 2025-04-24

### Added
- Comprehensive test suite for core functionality
- Project logo and branding assets
- Initial documentation improvements

### Changed
- README updates with better project description
- Minor code improvements and cleanup

## [v1.0.0] - 2025-04-22

### Added
- Initial project setup
- Basic SMTP client implementation
- Command-line interface
- Support for basic SMTP commands (HELO, EHLO, MAIL FROM, RCPT TO, DATA, QUIT)
- Debug mode for protocol interaction logging
- Signal handling for graceful shutdown
- Authentication support (PLAIN, LOGIN, CRAM-MD5)
- TLS/STARTTLS support
- Command-line options for authentication and TLS

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- Added support for secure authentication methods
- Added TLS/STARTTLS support for encrypted connections

[Unreleased]: https://github.com/asachs01/smtp-edc/compare/v1.3.0...HEAD
[v1.3.0]: https://github.com/asachs01/smtp-edc/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/asachs01/smtp-edc/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/asachs01/smtp-edc/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/asachs01/smtp-edc/releases/tag/v1.0.0
