package message

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	// Regular expression for basic email format validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidateEmail performs basic validation of an email address
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// ValidateEmailWithMX performs email validation including MX record lookup
func ValidateEmailWithMX(email string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	domain := parts[1]

	// Look up MX records
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup MX records for %s: %v", domain, err)
	}

	if len(mxRecords) == 0 {
		return fmt.Errorf("no MX records found for domain %s", domain)
	}

	return nil
}

// ValidateAddressList validates a list of email addresses
func ValidateAddressList(addresses []string, checkMX bool) error {
	for _, addr := range addresses {
		if checkMX {
			if err := ValidateEmailWithMX(addr); err != nil {
				return fmt.Errorf("invalid email address %s: %v", addr, err)
			}
		} else {
			if err := ValidateEmail(addr); err != nil {
				return fmt.Errorf("invalid email address %s: %v", addr, err)
			}
		}
	}
	return nil
}

// ValidateMessage validates all email addresses in a message
func ValidateMessage(msg *Message, checkMX bool) error {
	// Validate sender
	if err := ValidateEmail(msg.From); err != nil {
		return fmt.Errorf("invalid sender address: %v", err)
	}

	// Validate recipients
	if err := ValidateAddressList(msg.To, checkMX); err != nil {
		return fmt.Errorf("invalid To address: %v", err)
	}

	if err := ValidateAddressList(msg.Cc, checkMX); err != nil {
		return fmt.Errorf("invalid Cc address: %v", err)
	}

	if err := ValidateAddressList(msg.Bcc, checkMX); err != nil {
		return fmt.Errorf("invalid Bcc address: %v", err)
	}

	return nil
}
