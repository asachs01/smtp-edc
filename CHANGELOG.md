# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## v1.0.0

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