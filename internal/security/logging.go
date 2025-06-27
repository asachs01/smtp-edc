package security

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
	LogFatal
)

// LogEntry represents a log entry with security context
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Component string
	SessionID string
	UserID    string
	RemoteIP  string
	Metadata  map[string]interface{}
	Sensitive bool
}

// SecureLogger provides secure logging with data masking
type SecureLogger struct {
	level         LogLevel
	output        *log.Logger
	maskSensitive bool
	auditMode     bool
	patterns      []*SensitivePattern
	maxLogSize    int64
	retentionDays int
	rotateDaily   bool
	logFile       *os.File
	logPath       string
}

// SensitivePattern defines patterns for sensitive data detection
type SensitivePattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
	Category    string
}

// LogConfig configures the secure logger
type LogConfig struct {
	Level          LogLevel
	MaskSensitive  bool
	AuditMode      bool
	MaxLogSize     int64
	RetentionDays  int
	RotateDaily    bool
	LogPath        string
	CustomPatterns []SensitivePattern
}

// NewSecureLogger creates a new secure logger
func NewSecureLogger(config *LogConfig) (*SecureLogger, error) {
	sl := &SecureLogger{
		level:         config.Level,
		maskSensitive: config.MaskSensitive,
		auditMode:     config.AuditMode,
		maxLogSize:    config.MaxLogSize,
		retentionDays: config.RetentionDays,
		rotateDaily:   config.RotateDaily,
		logPath:       config.LogPath,
		patterns:      getDefaultSensitivePatterns(),
	}

	// Add custom patterns
	for _, pattern := range config.CustomPatterns {
		sl.patterns = append(sl.patterns, &pattern)
	}

	// Initialize log output
	if err := sl.initializeOutput(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}

	return sl, nil
}

// initializeOutput initializes the log output destination
func (sl *SecureLogger) initializeOutput() error {
	if sl.logPath == "" {
		// Log to stdout if no path specified
		sl.output = log.New(os.Stdout, "", 0)
		return nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(sl.logPath, 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create log file
	logFileName := fmt.Sprintf("%s/smtp-edc-%s.log", sl.logPath, time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	sl.logFile = file
	sl.output = log.New(file, "", 0)
	return nil
}

// Debug logs a debug message
func (sl *SecureLogger) Debug(component, message string, metadata ...map[string]interface{}) {
	if sl.level <= LogDebug {
		sl.log(LogDebug, component, message, metadata...)
	}
}

// Info logs an info message
func (sl *SecureLogger) Info(component, message string, metadata ...map[string]interface{}) {
	if sl.level <= LogInfo {
		sl.log(LogInfo, component, message, metadata...)
	}
}

// Warn logs a warning message
func (sl *SecureLogger) Warn(component, message string, metadata ...map[string]interface{}) {
	if sl.level <= LogWarn {
		sl.log(LogWarn, component, message, metadata...)
	}
}

// Error logs an error message
func (sl *SecureLogger) Error(component, message string, metadata ...map[string]interface{}) {
	if sl.level <= LogError {
		sl.log(LogError, component, message, metadata...)
	}
}

// Fatal logs a fatal message and exits
func (sl *SecureLogger) Fatal(component, message string, metadata ...map[string]interface{}) {
	sl.log(LogFatal, component, message, metadata...)
	os.Exit(1)
}

// Audit logs an audit message (always logged regardless of level)
func (sl *SecureLogger) Audit(component, message string, metadata ...map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogInfo,
		Message:   message,
		Component: component,
		Sensitive: false,
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	// Audit logs are always written
	sl.writeLog(entry, true)
}

// AuthLog logs authentication-related events
func (sl *SecureLogger) AuthLog(username, event, result string, metadata ...map[string]interface{}) {
	message := fmt.Sprintf("AUTH: user=%s event=%s result=%s", username, event, result)

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogInfo,
		Message:   message,
		Component: "auth",
		UserID:    username,
		Sensitive: true,
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	sl.writeLog(entry, true)
}

// SecurityLog logs security-related events
func (sl *SecureLogger) SecurityLog(event, severity, description string, metadata ...map[string]interface{}) {
	message := fmt.Sprintf("SECURITY: event=%s severity=%s desc=%s", event, severity, description)

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogWarn,
		Message:   message,
		Component: "security",
		Sensitive: true,
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	sl.writeLog(entry, true)
}

// ConnectionLog logs connection events
func (sl *SecureLogger) ConnectionLog(action, server string, port int, success bool, metadata ...map[string]interface{}) {
	result := "SUCCESS"
	if !success {
		result = "FAILED"
	}

	message := fmt.Sprintf("CONNECTION: action=%s server=%s port=%d result=%s", action, server, port, result)

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogInfo,
		Message:   message,
		Component: "connection",
		Sensitive: false,
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	sl.writeLog(entry, false)
}

// TLSLog logs TLS-related events
func (sl *SecureLogger) TLSLog(event, server string, tlsVersion, cipherSuite string, metadata ...map[string]interface{}) {
	message := fmt.Sprintf("TLS: event=%s server=%s version=%s cipher=%s", event, server, tlsVersion, cipherSuite)

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogInfo,
		Message:   message,
		Component: "tls",
		Sensitive: false,
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	sl.writeLog(entry, false)
}

// log is the internal logging method
func (sl *SecureLogger) log(level LogLevel, component, message string, metadata ...map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: component,
		Sensitive: sl.containsSensitiveData(message),
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
		sl.extractContextFromMetadata(entry)
	}

	sl.writeLog(entry, false)
}

// writeLog writes the log entry to the output
func (sl *SecureLogger) writeLog(entry *LogEntry, forceWrite bool) {
	if !forceWrite && sl.level > entry.Level {
		return
	}

	// Mask sensitive data if enabled
	message := entry.Message
	if sl.maskSensitive {
		message = sl.maskSensitiveData(message)
	}

	// Format log entry
	logLine := sl.formatLogEntry(entry, message)

	// Write to output
	sl.output.Print(logLine)

	// Rotate logs if necessary
	if sl.rotateDaily {
		sl.rotateLogsIfNeeded()
	}
}

// formatLogEntry formats a log entry for output
func (sl *SecureLogger) formatLogEntry(entry *LogEntry, message string) string {
	levelStr := sl.levelToString(entry.Level)
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	// Base format
	logLine := fmt.Sprintf("[%s] %s [%s] %s", timestamp, levelStr, entry.Component, message)

	// Add context information
	if entry.SessionID != "" {
		logLine += fmt.Sprintf(" session=%s", entry.SessionID)
	}
	if entry.UserID != "" {
		logLine += fmt.Sprintf(" user=%s", sl.maskSensitiveData(entry.UserID))
	}
	if entry.RemoteIP != "" {
		logLine += fmt.Sprintf(" ip=%s", entry.RemoteIP)
	}

	// Add metadata
	if entry.Metadata != nil {
		for key, value := range entry.Metadata {
			if key != "session_id" && key != "user_id" && key != "remote_ip" {
				valueStr := fmt.Sprintf("%v", value)
				if sl.maskSensitive {
					valueStr = sl.maskSensitiveData(valueStr)
				}
				logLine += fmt.Sprintf(" %s=%s", key, valueStr)
			}
		}
	}

	return logLine + "\n"
}

// maskSensitiveData masks sensitive data in log messages
func (sl *SecureLogger) maskSensitiveData(message string) string {
	maskedMessage := message

	for _, pattern := range sl.patterns {
		maskedMessage = pattern.Pattern.ReplaceAllString(maskedMessage, pattern.Replacement)
	}

	return maskedMessage
}

// containsSensitiveData checks if a message contains sensitive data
func (sl *SecureLogger) containsSensitiveData(message string) bool {
	for _, pattern := range sl.patterns {
		if pattern.Pattern.MatchString(message) {
			return true
		}
	}
	return false
}

// extractContextFromMetadata extracts context information from metadata
func (sl *SecureLogger) extractContextFromMetadata(entry *LogEntry) {
	if entry.Metadata == nil {
		return
	}

	if sessionID, ok := entry.Metadata["session_id"].(string); ok {
		entry.SessionID = sessionID
	}
	if userID, ok := entry.Metadata["user_id"].(string); ok {
		entry.UserID = userID
	}
	if remoteIP, ok := entry.Metadata["remote_ip"].(string); ok {
		entry.RemoteIP = remoteIP
	}
}

// levelToString converts log level to string
func (sl *SecureLogger) levelToString(level LogLevel) string {
	switch level {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	case LogFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// rotateLogsIfNeeded rotates logs if needed
func (sl *SecureLogger) rotateLogsIfNeeded() {
	if sl.logFile == nil {
		return
	}

	// Check if we need to rotate based on date
	now := time.Now()
	// currentLogFile := fmt.Sprintf("%s/smtp-edc-%s.log", sl.logPath, now.Format("2006-01-02"))

	// Get current log file info
	fileInfo, err := sl.logFile.Stat()
	if err != nil {
		return
	}

	// Check if log file is from a different day
	if !strings.Contains(fileInfo.Name(), now.Format("2006-01-02")) {
		sl.rotateLog()
	}

	// Check if log file exceeds max size
	if sl.maxLogSize > 0 && fileInfo.Size() > sl.maxLogSize {
		sl.rotateLog()
	}
}

// rotateLog rotates the current log file
func (sl *SecureLogger) rotateLog() {
	if sl.logFile != nil {
		sl.logFile.Close()
	}

	// Create new log file
	if err := sl.initializeOutput(); err != nil {
		// Fall back to stdout if rotation fails
		sl.output = log.New(os.Stdout, "", 0)
	}
}

// Close closes the logger and cleans up resources
func (sl *SecureLogger) Close() error {
	if sl.logFile != nil {
		return sl.logFile.Close()
	}
	return nil
}

// AddSensitivePattern adds a custom sensitive data pattern
func (sl *SecureLogger) AddSensitivePattern(name, pattern, replacement, category string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}

	sl.patterns = append(sl.patterns, &SensitivePattern{
		Name:        name,
		Pattern:     regex,
		Replacement: replacement,
		Category:    category,
	})

	return nil
}

// GetStats returns logging statistics
func (sl *SecureLogger) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["level"] = sl.levelToString(sl.level)
	stats["mask_sensitive"] = sl.maskSensitive
	stats["audit_mode"] = sl.auditMode
	stats["patterns_count"] = len(sl.patterns)

	if sl.logFile != nil {
		if fileInfo, err := sl.logFile.Stat(); err == nil {
			stats["log_file_size"] = fileInfo.Size()
			stats["log_file_name"] = fileInfo.Name()
		}
	}

	return stats
}

// getDefaultSensitivePatterns returns default patterns for sensitive data detection
func getDefaultSensitivePatterns() []*SensitivePattern {
	return []*SensitivePattern{
		{
			Name:        "password",
			Pattern:     regexp.MustCompile(`(?i)(password|passwd|pwd)[=:\s]+([^\s]+)`),
			Replacement: "${1}=***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "token",
			Pattern:     regexp.MustCompile(`(?i)(token|auth|bearer)[=:\s]+([^\s]+)`),
			Replacement: "${1}=***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "api_key",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey)[=:\s]+([^\s]+)`),
			Replacement: "${1}=***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "email",
			Pattern:     regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			Replacement: "***EMAIL***",
			Category:    "pii",
		},
		{
			Name:        "base64_auth",
			Pattern:     regexp.MustCompile(`(?i)(auth|authorization)[=:\s]+([A-Za-z0-9+/]{20,}={0,2})`),
			Replacement: "${1}=***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "smtp_auth",
			Pattern:     regexp.MustCompile(`(?i)(C:|S:)\s*(AUTH\s+[A-Za-z0-9+/]{8,}={0,2})`),
			Replacement: "${1} AUTH ***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "plain_auth",
			Pattern:     regexp.MustCompile(`(?i)(C:|S:)\s*([A-Za-z0-9+/]{20,}={0,2})`),
			Replacement: "${1} ***MASKED***",
			Category:    "credentials",
		},
		{
			Name:        "ip_address",
			Pattern:     regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`),
			Replacement: "***IP***",
			Category:    "network",
		},
		{
			Name:        "session_id",
			Pattern:     regexp.MustCompile(`(?i)(session[_-]?id|sessionid)[=:\s]+([a-f0-9-]{8,})`),
			Replacement: "${1}=***MASKED***",
			Category:    "session",
		},
	}
}
