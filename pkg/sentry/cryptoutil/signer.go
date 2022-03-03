package cryptoutil

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	defaultCertValidity = time.Hour * 24 * 365 * 10
)

// Signer is the interface for signing pem encoded CSR
type Signer interface {
	Sign(csr []byte) ([]byte, error)
}

// Options is the options used to construct signer
type options struct {
	IsClient           bool
	IsServer           bool
	CAKeyDecrypt       PasswordFunc
	CSRSubjectValidate []SubjectValidateFunc
	CertValidity       time.Duration
	AltNames           []string
	IPAddress          []string
}

// Option is the functional arg for setting options
type Option func(*options)

// WithClient is used to sign client certs
func WithClient() Option {
	return func(o *options) {
		o.IsClient = true
	}
}

// WithServer is used to sign server certs
func WithServer() Option {
	return func(o *options) {
		o.IsServer = true
	}
}

// WithCAKeyDecrypt passes the password function to decrypt ca key
func WithCAKeyDecrypt(pf PasswordFunc) Option {
	return func(o *options) {
		o.CAKeyDecrypt = pf
	}
}

// WithCSRSubjectValidate is used to validate subject of CSR
func WithCSRSubjectValidate(svf SubjectValidateFunc) Option {
	return func(o *options) {
		o.CSRSubjectValidate = append(o.CSRSubjectValidate, svf)
	}
}

// WithCertValidity makes the issued certificate expire after the duration
func WithCertValidity(d time.Duration) Option {
	return func(o *options) {
		o.CertValidity = d
	}
}

// WithAltName adds subject alt name to the signed certificate
func WithAltName(dns string) Option {
	return func(o *options) {
		o.AltNames = append(o.AltNames, dns)
	}
}

// WithIPAddress adds ip address to the signed certificate
func WithIPAddress(ip string) Option {
	return func(o *options) {
		o.IPAddress = append(o.IPAddress, ip)
	}
}

type signer struct {
	ca   *x509.Certificate
	key  crypto.PrivateKey
	opts *options
}

func (s *signer) Sign(csr []byte) ([]byte, error) {
	cr, err := DecodeCSR(csr)
	if err != nil {
		return nil, err
	}

	for _, svf := range s.opts.CSRSubjectValidate {
		if err = svf(cr.Subject); err != nil {
			return nil, err
		}
	}
	template := &x509.Certificate{
		SerialNumber: getSerialNumber(),
		Issuer:       s.ca.Subject,
		Subject:      cr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(s.opts.CertValidity),
		//ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	if s.opts.IsClient {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	if s.opts.IsServer {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	for _, altName := range s.opts.AltNames {
		template.DNSNames = append(template.DNSNames, altName)
	}

	for _, altIP := range s.opts.IPAddress {
		ip := net.ParseIP(altIP)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip address %s", altIP)
		}
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	// sign the certificate
	b, err := x509.CreateCertificate(rand.Reader, template, s.ca, cr.PublicKey, s.key)
	if err != nil {
		err = fmt.Errorf("unable to create certificate %s", err.Error())
		return nil, err
	}

	return EncodeCert(b), nil
}

// SubjectValidateFunc validates the subject of CSR before signing the request
type SubjectValidateFunc func(subject pkix.Name) error

// NoSubjectValidate ignores subject validation of CSR
var NoSubjectValidate = func(subject pkix.Name) error {
	return nil
}

// CNShouldBe validates if CommonName of CSR is same as the passed CN
var CNShouldBe = func(cn string) SubjectValidateFunc {
	return func(subject pkix.Name) error {
		if subject.CommonName != cn {
			return fmt.Errorf("expected CN %s got %s", cn, subject.CommonName)
		}
		return nil
	}
}

// CNShouldBeStar validates if CommonName of CSR is same as the passed CN *.domain
var CNShouldBeStar = func(cn string) SubjectValidateFunc {
	return func(subject pkix.Name) error {
		if subject.CommonName[0] != '*' && subject.CommonName != cn {
			return fmt.Errorf("expected CN %s got %s", cn, subject.CommonName)
		}

		sfx := subject.CommonName[1:]
		if !strings.HasSuffix(cn, sfx) {
			return fmt.Errorf("expected CN %s got %s", cn, subject.CommonName)
		}

		return nil
	}
}

// NewSigner return a CSR signer for given PEM encoded CA cert and key
func NewSigner(cert, key []byte, opts ...Option) (Signer, error) {

	signerOpts := &options{}
	for _, opt := range opts {
		opt(signerOpts)
	}

	if signerOpts.CAKeyDecrypt == nil {
		signerOpts.CAKeyDecrypt = NoPassword
	}

	// if cert validit
	if signerOpts.CertValidity == 0 {
		signerOpts.CertValidity = defaultCertValidity
	}

	ca, err := DecodeCert(cert)
	if err != nil {
		return nil, err
	}

	privKey, err := DecodePrivateKey(key, signerOpts.CAKeyDecrypt)
	if err != nil {
		return nil, err
	}

	return &signer{ca, privKey, signerOpts}, nil
}
