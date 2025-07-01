package security

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
)

// TLSConfig represents TLS configuration with security enhancements
type TLSConfig struct {
	// Basic TLS settings
	MinVersion         uint16
	MaxVersion         uint16
	InsecureSkipVerify bool
	ServerName         string

	// Certificate validation
	RootCAs               *x509.CertPool
	ClientCertificates    []tls.Certificate
	VerifyPeerCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

	// Security features
	RequireCertificate  bool
	VerifyConnection    func(cs tls.ConnectionState) error
	PreferServerCiphers bool
	CipherSuites        []uint16

	// Custom validation
	AllowedSANs        []string
	AllowedCNs         []string
	PinCertificates    []string // SHA-256 fingerprints
	MaxCertChainLength int

	// Debugging and troubleshooting
	Debug          bool
	LogTLSErrors   bool
	CollectMetrics bool
}

// TLSInfo contains information about a TLS connection
type TLSInfo struct {
	Version                    string
	CipherSuite                string
	ServerCertificates         []*x509.Certificate
	VerifiedChains             [][]*x509.Certificate
	ConnectionState            tls.ConnectionState
	HandshakeComplete          bool
	DidResume                  bool
	NegotiatedProtocol         string
	NegotiatedProtocolIsMutual bool
	PeerCertificatesValid      bool
	CertificateErrors          []error
	SecurityScore              int
	Recommendations            []string
}

// TLSVerifier provides comprehensive TLS verification and troubleshooting
type TLSVerifier struct {
	config  *TLSConfig
	metrics *TLSMetrics
}

// TLSMetrics tracks TLS connection metrics
type TLSMetrics struct {
	TotalConnections      int64
	SuccessfulConnections int64
	FailedConnections     int64
	CertificateErrors     int64
	HandshakeErrors       int64
	VersionDistribution   map[string]int64
	CipherDistribution    map[string]int64
	LastUpdated           time.Time
}

// NewTLSConfig creates a secure TLS configuration
func NewTLSConfig(serverName string) *TLSConfig {
	return &TLSConfig{
		MinVersion:          tls.VersionTLS12,
		MaxVersion:          tls.VersionTLS13,
		InsecureSkipVerify:  false,
		ServerName:          serverName,
		RequireCertificate:  true,
		PreferServerCiphers: false, // Prefer client cipher order for security
		MaxCertChainLength:  5,
		Debug:               false,
		LogTLSErrors:        true,
		CollectMetrics:      true,
		CipherSuites:        getSecureCipherSuites(),
	}
}

// NewTLSVerifier creates a new TLS verifier
func NewTLSVerifier(config *TLSConfig) *TLSVerifier {
	return &TLSVerifier{
		config: config,
		metrics: &TLSMetrics{
			VersionDistribution: make(map[string]int64),
			CipherDistribution:  make(map[string]int64),
			LastUpdated:         time.Now(),
		},
	}
}

// BuildTLSConfig converts our TLS config to Go's tls.Config
func (tc *TLSConfig) BuildTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		MinVersion:               tc.MinVersion,
		MaxVersion:               tc.MaxVersion,
		InsecureSkipVerify:       tc.InsecureSkipVerify,
		ServerName:               tc.ServerName,
		RootCAs:                  tc.RootCAs,
		Certificates:             tc.ClientCertificates,
		PreferServerCipherSuites: tc.PreferServerCiphers,
		CipherSuites:             tc.CipherSuites,
	}

	// Add custom certificate verification
	if !tc.InsecureSkipVerify {
		tlsConfig.VerifyPeerCertificate = tc.buildVerifyPeerCertificate()
		tlsConfig.VerifyConnection = tc.buildVerifyConnection()
	}

	return tlsConfig
}

// VerifyTLSConnection performs comprehensive TLS verification
func (tv *TLSVerifier) VerifyTLSConnection(serverAddr string, port int) (*TLSInfo, error) {
	tv.metrics.TotalConnections++

	address := fmt.Sprintf("%s:%d", serverAddr, port)

	// Create TLS configuration
	tlsConfig := tv.config.BuildTLSConfig()

	// Establish connection with timeout
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		tv.metrics.FailedConnections++
		return nil, fmt.Errorf("TLS connection failed: %v", err)
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()

	// Build TLS info
	info := &TLSInfo{
		Version:                    tlsVersionString(state.Version),
		CipherSuite:                tls.CipherSuiteName(state.CipherSuite),
		ServerCertificates:         state.PeerCertificates,
		VerifiedChains:             state.VerifiedChains,
		ConnectionState:            state,
		HandshakeComplete:          state.HandshakeComplete,
		DidResume:                  state.DidResume,
		NegotiatedProtocol:         state.NegotiatedProtocol,
		NegotiatedProtocolIsMutual: true, // Always true in Go 1.16+
	}

	// Verify certificates
	certErrors := tv.verifyCertificates(state.PeerCertificates, state.VerifiedChains)
	info.CertificateErrors = certErrors
	info.PeerCertificatesValid = len(certErrors) == 0

	// Calculate security score and recommendations
	info.SecurityScore = tv.calculateSecurityScore(&state)
	info.Recommendations = tv.generateRecommendations(&state, certErrors)

	// Update metrics
	tv.updateMetrics(&state, len(certErrors) == 0)

	if info.PeerCertificatesValid {
		tv.metrics.SuccessfulConnections++
	} else {
		tv.metrics.FailedConnections++
		tv.metrics.CertificateErrors++
	}

	return info, nil
}

// TestSMTPSTARTTLS tests STARTTLS capability for SMTP servers
func (tv *TLSVerifier) TestSMTPSTARTTLS(serverAddr string, port int) (*TLSInfo, error) {
	address := fmt.Sprintf("%s:%d", serverAddr, port)

	// Connect to SMTP server
	conn, err := net.DialTimeout("tcp", address, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Read server greeting
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read server greeting: %v", err)
	}

	// Send EHLO
	hostname := "smtp-edc-client"
	_, err = conn.Write([]byte(fmt.Sprintf("EHLO %s\r\n", hostname)))
	if err != nil {
		return nil, fmt.Errorf("failed to send EHLO: %v", err)
	}

	// Read EHLO response
	_, err = conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read EHLO response: %v", err)
	}

	// Check if STARTTLS is supported
	response := string(buffer)
	if !strings.Contains(response, "STARTTLS") {
		return nil, fmt.Errorf("server does not support STARTTLS")
	}

	// Send STARTTLS command
	_, err = conn.Write([]byte("STARTTLS\r\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to send STARTTLS: %v", err)
	}

	// Read STARTTLS response
	_, err = conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read STARTTLS response: %v", err)
	}

	// Check if STARTTLS was accepted
	startTLSResponse := string(buffer)
	if !strings.HasPrefix(startTLSResponse, "220") {
		return nil, fmt.Errorf("server rejected STARTTLS: %s", startTLSResponse)
	}

	// Upgrade to TLS
	tlsConfig := tv.config.BuildTLSConfig()
	tlsConn := tls.Client(conn, tlsConfig)

	err = tlsConn.Handshake()
	if err != nil {
		return nil, fmt.Errorf("TLS handshake failed: %v", err)
	}

	// Get connection state and build info
	state := tlsConn.ConnectionState()
	info := &TLSInfo{
		Version:                    tlsVersionString(state.Version),
		CipherSuite:                tls.CipherSuiteName(state.CipherSuite),
		ServerCertificates:         state.PeerCertificates,
		VerifiedChains:             state.VerifiedChains,
		ConnectionState:            state,
		HandshakeComplete:          state.HandshakeComplete,
		DidResume:                  state.DidResume,
		NegotiatedProtocol:         state.NegotiatedProtocol,
		NegotiatedProtocolIsMutual: true, // Always true in Go 1.16+
	}

	// Verify certificates
	certErrors := tv.verifyCertificates(state.PeerCertificates, state.VerifiedChains)
	info.CertificateErrors = certErrors
	info.PeerCertificatesValid = len(certErrors) == 0

	// Calculate security score and recommendations
	info.SecurityScore = tv.calculateSecurityScore(&state)
	info.Recommendations = tv.generateRecommendations(&state, certErrors)

	return info, nil
}

// verifyCertificates performs comprehensive certificate verification
func (tv *TLSVerifier) verifyCertificates(certs []*x509.Certificate, chains [][]*x509.Certificate) []error {
	var errors []error

	if len(certs) == 0 {
		errors = append(errors, fmt.Errorf("no certificates provided"))
		return errors
	}

	serverCert := certs[0]

	// Check certificate chain length
	if tv.config.MaxCertChainLength > 0 && len(certs) > tv.config.MaxCertChainLength {
		errors = append(errors, fmt.Errorf("certificate chain too long: %d (max: %d)", len(certs), tv.config.MaxCertChainLength))
	}

	// Check certificate expiration
	now := time.Now()
	if now.Before(serverCert.NotBefore) {
		errors = append(errors, fmt.Errorf("certificate not yet valid (valid from: %v)", serverCert.NotBefore))
	}
	if now.After(serverCert.NotAfter) {
		errors = append(errors, fmt.Errorf("certificate expired (expired on: %v)", serverCert.NotAfter))
	}

	// Check if certificate expires soon (within 30 days)
	if serverCert.NotAfter.Sub(now) < 30*24*time.Hour {
		errors = append(errors, fmt.Errorf("certificate expires soon: %v", serverCert.NotAfter))
	}

	// Verify server name
	if tv.config.ServerName != "" {
		if err := serverCert.VerifyHostname(tv.config.ServerName); err != nil {
			errors = append(errors, fmt.Errorf("hostname verification failed: %v", err))
		}
	}

	// Check allowed CNs
	if len(tv.config.AllowedCNs) > 0 {
		allowed := false
		for _, allowedCN := range tv.config.AllowedCNs {
			if serverCert.Subject.CommonName == allowedCN {
				allowed = true
				break
			}
		}
		if !allowed {
			errors = append(errors, fmt.Errorf("certificate CN not in allowed list: %s", serverCert.Subject.CommonName))
		}
	}

	// Check allowed SANs
	if len(tv.config.AllowedSANs) > 0 {
		allowed := false
		for _, allowedSAN := range tv.config.AllowedSANs {
			for _, san := range serverCert.DNSNames {
				if san == allowedSAN {
					allowed = true
					break
				}
			}
			if allowed {
				break
			}
		}
		if !allowed {
			errors = append(errors, fmt.Errorf("certificate SANs not in allowed list"))
		}
	}

	// Certificate pinning
	if len(tv.config.PinCertificates) > 0 {
		fingerprint := calculateCertFingerprint(serverCert)
		pinned := false
		for _, pin := range tv.config.PinCertificates {
			if fingerprint == pin {
				pinned = true
				break
			}
		}
		if !pinned {
			errors = append(errors, fmt.Errorf("certificate not pinned (fingerprint: %s)", fingerprint))
		}
	}

	// Check certificate algorithm strength
	if serverCert.PublicKeyAlgorithm == x509.RSA {
		if rsaKey, ok := serverCert.PublicKey.(*rsa.PublicKey); ok {
			if rsaKey.Size() < 256 { // Less than 2048 bits
				errors = append(errors, fmt.Errorf("weak RSA key size: %d bits", rsaKey.Size()*8))
			}
		}
	}

	return errors
}

// calculateSecurityScore calculates a security score based on TLS connection properties
func (tv *TLSVerifier) calculateSecurityScore(state *tls.ConnectionState) int {
	score := 0

	// TLS version scoring
	switch state.Version {
	case tls.VersionTLS13:
		score += 40
	case tls.VersionTLS12:
		score += 30
	case tls.VersionTLS11:
		score += 10
	case tls.VersionTLS10:
		score += 5
	}

	// Cipher suite scoring
	cipherScore := getCipherSuiteScore(state.CipherSuite)
	score += cipherScore

	// Certificate scoring
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]

		// Key algorithm and size
		switch cert.PublicKeyAlgorithm {
		case x509.ECDSA:
			score += 20
		case x509.RSA:
			if rsaKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
				keySize := rsaKey.Size() * 8
				if keySize >= 4096 {
					score += 15
				} else if keySize >= 2048 {
					score += 10
				} else {
					score += 0 // Weak key
				}
			}
		case x509.Ed25519:
			score += 25
		}

		// Certificate validity period
		validityPeriod := cert.NotAfter.Sub(cert.NotBefore)
		if validityPeriod <= 90*24*time.Hour {
			score += 10 // Short-lived certificates are better
		} else if validityPeriod <= 365*24*time.Hour {
			score += 5
		}
	}

	// Perfect Forward Secrecy
	if isPFSCipherSuite(state.CipherSuite) {
		score += 10
	}

	return score
}

// generateRecommendations generates security recommendations based on TLS analysis
func (tv *TLSVerifier) generateRecommendations(state *tls.ConnectionState, certErrors []error) []string {
	var recommendations []string

	// TLS version recommendations
	if state.Version < tls.VersionTLS12 {
		recommendations = append(recommendations, "Upgrade to TLS 1.2 or higher")
	}
	if state.Version < tls.VersionTLS13 {
		recommendations = append(recommendations, "Consider upgrading to TLS 1.3 for better security")
	}

	// Cipher suite recommendations
	if !isSecureCipherSuite(state.CipherSuite) {
		recommendations = append(recommendations, "Use a more secure cipher suite")
	}
	if !isPFSCipherSuite(state.CipherSuite) {
		recommendations = append(recommendations, "Use cipher suites with Perfect Forward Secrecy")
	}

	// Certificate recommendations
	if len(certErrors) > 0 {
		recommendations = append(recommendations, "Fix certificate validation errors")
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]

		// Check certificate expiration
		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
		if daysUntilExpiry < 30 {
			recommendations = append(recommendations, fmt.Sprintf("Certificate expires in %d days - renew soon", daysUntilExpiry))
		}

		// Check key size
		if cert.PublicKeyAlgorithm == x509.RSA {
			if rsaKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
				keySize := rsaKey.Size() * 8
				if keySize < 2048 {
					recommendations = append(recommendations, "Use RSA keys of at least 2048 bits")
				}
			}
		}
	}

	return recommendations
}

// updateMetrics updates TLS connection metrics
func (tv *TLSVerifier) updateMetrics(state *tls.ConnectionState, success bool) {
	if !tv.config.CollectMetrics {
		return
	}

	tv.metrics.VersionDistribution[tlsVersionString(state.Version)]++
	tv.metrics.CipherDistribution[tls.CipherSuiteName(state.CipherSuite)]++
	tv.metrics.LastUpdated = time.Now()
}

// Helper functions

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

func getSecureCipherSuites() []uint16 {
	// Return only secure cipher suites
	return []uint16{
		// TLS 1.3 cipher suites (preferred)
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_AES_128_GCM_SHA256,

		// TLS 1.2 cipher suites with PFS
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
}

func getCipherSuiteScore(cipherSuite uint16) int {
	// Scoring based on cipher suite security
	switch cipherSuite {
	case tls.TLS_AES_256_GCM_SHA384, tls.TLS_CHACHA20_POLY1305_SHA256:
		return 30 // TLS 1.3 with strong encryption
	case tls.TLS_AES_128_GCM_SHA256:
		return 25 // TLS 1.3 with adequate encryption
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return 20 // Strong encryption with PFS
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:
		return 20 // ChaCha20 with PFS
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return 15 // Adequate encryption with PFS
	default:
		return 5 // Other cipher suites get minimal score
	}
}

func isSecureCipherSuite(cipherSuite uint16) bool {
	secureSuites := getSecureCipherSuites()
	for _, secure := range secureSuites {
		if cipherSuite == secure {
			return true
		}
	}
	return false
}

func isPFSCipherSuite(cipherSuite uint16) bool {
	// Check if cipher suite supports Perfect Forward Secrecy
	switch cipherSuite {
	case tls.TLS_AES_256_GCM_SHA384, tls.TLS_CHACHA20_POLY1305_SHA256, tls.TLS_AES_128_GCM_SHA256:
		return true // TLS 1.3 always has PFS
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return true // ECDHE provides PFS
	default:
		return false
	}
}

func calculateCertFingerprint(cert *x509.Certificate) string {
	// Calculate SHA-256 fingerprint of certificate
	hash := sha256.Sum256(cert.Raw)
	return fmt.Sprintf("%x", hash)
}

func (tc *TLSConfig) buildVerifyPeerCertificate() func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// Custom certificate verification logic
		if tc.RequireCertificate && len(rawCerts) == 0 {
			return fmt.Errorf("no certificates provided")
		}

		// Additional custom verification can be added here
		if tc.VerifyPeerCertificate != nil {
			return tc.VerifyPeerCertificate(rawCerts, verifiedChains)
		}

		return nil
	}
}

func (tc *TLSConfig) buildVerifyConnection() func(cs tls.ConnectionState) error {
	return func(cs tls.ConnectionState) error {
		// Custom connection verification logic
		if tc.MinVersion > 0 && cs.Version < tc.MinVersion {
			return fmt.Errorf("TLS version too low: %s", tlsVersionString(cs.Version))
		}

		if !isSecureCipherSuite(cs.CipherSuite) {
			return fmt.Errorf("insecure cipher suite: %s", tls.CipherSuiteName(cs.CipherSuite))
		}

		// Additional custom verification can be added here
		if tc.VerifyConnection != nil {
			return tc.VerifyConnection(cs)
		}

		return nil
	}
}
