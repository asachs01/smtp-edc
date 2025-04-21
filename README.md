# SMTP-EDC

A cross-platform SMTP testing tool written in Go, equivalent to Swaks.

## Features

- Basic SMTP transaction testing
- Support for multiple authentication methods (PLAIN, LOGIN, CRAM-MD5)
- TLS/STARTTLS support
- Debug mode for detailed protocol interaction
- Cross-platform compatibility

## Installation

```bash
go install github.com/asachs/smtp-edc/cmd/smtp-edc@latest
```

## Usage

Basic usage:

```bash
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com
```

With authentication:

```bash
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com \
    --auth plain --username user --password pass
```

With TLS:

```bash
smtp-edc --server smtp.example.com --port 587 --from sender@example.com --to recipient@example.com
```

## Command Line Options

- `--server`: SMTP server hostname or IP address (required)
- `--port`: SMTP server port (default: 25)
- `--from`: Sender email address (required)
- `--to`: Recipient email address (required)
- `--subject`: Email subject
- `--body`: Email body
- `--verbose`: Enable verbose output
- `--debug`: Enable debug mode
- `--auth`: Authentication type (plain, login, cram-md5)
- `--username`: Authentication username
- `--password`: Authentication password

## Development

### Building from Source

```bash
git clone https://github.com/asachs/smtp-edc.git
cd smtp-edc
go build -o smtp-edc cmd/smtp-edc/main.go
```

### Project Structure

```
smtp-edc/
├── cmd/
│   └── smtp-edc/
│       └── main.go
├── internal/
│   ├── client/
│   ├── message/
│   ├── auth/
│   └── transport/
├── pkg/
│   ├── smtp/
│   └── utils/
└── go.mod
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 