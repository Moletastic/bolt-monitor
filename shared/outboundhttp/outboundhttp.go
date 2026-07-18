// Package outboundhttp enforces the public-network boundary for operator-supplied HTTP destinations.
package outboundhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	MaxRedirects              = 5
	MaxResponseBytes          = 1 << 20
	MaxRequestTimeout         = 30 * time.Second
	NotificationTimeout       = 10 * time.Second
	DialTimeout               = 5 * time.Second
	TLSHandshakeTimeout       = 5 * time.Second
	ResponseHeaderTimeout     = 10 * time.Second
	ianaRegistryRetrievedDate = "2026-07-17"
)

type Kind string

const (
	KindInvalidURL       Kind = "invalid_url"
	KindSchemeRejected   Kind = "scheme_rejected"
	KindAddressBlocked   Kind = "address_blocked"
	KindResolutionFailed Kind = "resolution_failed"
	KindRedirectBlocked  Kind = "redirect_blocked"
	KindRedirectLimit    Kind = "redirect_limit"
	KindTimeout          Kind = "timeout"
	KindResponseTooLarge Kind = "response_too_large"
	KindTransport        Kind = "transport_failure"
)

type Error struct {
	Kind Kind
	Host string
	err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Host == "" {
		return string(e.Kind)
	}
	return string(e.Kind) + " for destination host"
}

func (e *Error) Unwrap() error { return e.err }

func IsKind(err error, kind Kind) bool {
	var outbound *Error
	return errors.As(err, &outbound) && outbound.Kind == kind
}

func SafeMessage(err error) string {
	var outbound *Error
	if !errors.As(err, &outbound) {
		return "outbound request failed"
	}
	switch outbound.Kind {
	case KindInvalidURL:
		return "destination URL is invalid"
	case KindSchemeRejected:
		return "destination URL scheme is not allowed"
	case KindAddressBlocked:
		return "destination address is not publicly reachable"
	case KindResolutionFailed:
		return "destination could not be resolved"
	case KindRedirectBlocked:
		return "redirect destination is not allowed"
	case KindRedirectLimit:
		return "redirect limit exceeded"
	case KindTimeout:
		return "outbound request timed out"
	case KindResponseTooLarge:
		return "outbound response exceeded size limit"
	default:
		return "outbound request failed"
	}
}

type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

type Dialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type netResolver struct{ resolver *net.Resolver }

func (r netResolver) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	return r.resolver.LookupNetIP(ctx, network, host)
}

type netDialer struct{ dialer *net.Dialer }

func (d netDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, address)
}

type Executor struct {
	Resolver Resolver
	Dialer   Dialer
}

func NewExecutor() *Executor {
	return &Executor{
		Resolver: netResolver{resolver: net.DefaultResolver},
		Dialer:   netDialer{dialer: &net.Dialer{Timeout: DialTimeout}},
	}
}

func (e *Executor) resolver() Resolver {
	if e != nil && e.Resolver != nil {
		return e.Resolver
	}
	return netResolver{resolver: net.DefaultResolver}
}

func (e *Executor) dialer() Dialer {
	if e != nil && e.Dialer != nil {
		return e.Dialer
	}
	return netDialer{dialer: &net.Dialer{Timeout: DialTimeout}}
}

// ValidateURL performs pure static validation. It never resolves names or opens a connection.
func ValidateURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" {
		return nil, &Error{Kind: KindInvalidURL}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, &Error{Kind: KindSchemeRejected}
	}
	if parsed.User != nil {
		return nil, &Error{Kind: KindInvalidURL}
	}
	if parsed.Port() != "" {
		return nil, &Error{Kind: KindInvalidURL}
	}
	host := normalizeHost(parsed.Hostname())
	if isBlockedHostname(host) || isAmbiguousIPv4(host) {
		return nil, &Error{Kind: KindAddressBlocked, Host: host}
	}
	if address, err := netip.ParseAddr(host); err == nil && !isPublicAddress(address) {
		return nil, &Error{Kind: KindAddressBlocked, Host: host}
	}
	return parsed, nil
}

// ValidateDestination resolves a validated destination once and rejects any blocked answer.
func (e *Executor) ValidateDestination(ctx context.Context, raw string) (*url.URL, []netip.Addr, error) {
	parsed, err := ValidateURL(raw)
	if err != nil {
		return nil, nil, err
	}
	host := normalizeHost(parsed.Hostname())
	if address, parseErr := netip.ParseAddr(host); parseErr == nil {
		address = address.Unmap()
		if !isPublicAddress(address) {
			return nil, nil, &Error{Kind: KindAddressBlocked, Host: host}
		}
		return parsed, []netip.Addr{address}, nil
	}
	addresses, err := e.resolver().LookupNetIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return nil, nil, &Error{Kind: KindResolutionFailed, Host: host, err: err}
	}
	approved := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !isPublicAddress(address) {
			return nil, nil, &Error{Kind: KindAddressBlocked, Host: host}
		}
		approved = append(approved, address)
	}
	return parsed, approved, nil
}

type Request struct {
	Method  string
	URL     string
	Header  http.Header
	Body    io.Reader
	Timeout time.Duration
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (e *Executor) Execute(ctx context.Context, request Request) (Response, error) {
	timeout := request.Timeout
	if timeout <= 0 || timeout > MaxRequestTimeout {
		timeout = MaxRequestTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, request.Method, request.URL, request.Body)
	if err != nil {
		return Response{}, &Error{Kind: KindInvalidURL}
	}
	req.Header = request.Header.Clone()
	hadHeaders := len(req.Header) > 0
	hadBody := request.Body != nil
	client := &http.Client{
		Transport: e.transport(),
		CheckRedirect: func(next *http.Request, via []*http.Request) error {
			if len(via) >= MaxRedirects {
				return &Error{Kind: KindRedirectLimit}
			}
			previous := via[len(via)-1].URL
			if !sameOrigin(previous, next.URL) && (hadHeaders || hadBody || previous.Scheme == "https" && next.URL.Scheme != "https") {
				return &Error{Kind: KindRedirectBlocked}
			}
			return nil
		},
	}
	response, err := client.Do(req)
	if err != nil {
		var outbound *Error
		if errors.As(err, &outbound) {
			return Response{}, outbound
		}
		if ctx.Err() != nil || isTimeout(err) {
			return Response{}, &Error{Kind: KindTimeout, err: err}
		}
		return Response{}, &Error{Kind: KindTransport, err: err}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes+1))
	if err != nil {
		if ctx.Err() != nil || isTimeout(err) {
			return Response{}, &Error{Kind: KindTimeout, err: err}
		}
		return Response{}, &Error{Kind: KindTransport, err: err}
	}
	if len(body) > MaxResponseBytes {
		return Response{}, &Error{Kind: KindResponseTooLarge}
	}
	return Response{StatusCode: response.StatusCode, Header: response.Header.Clone(), Body: body}, nil
}

func (e *Executor) transport() http.RoundTripper {
	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		parsed, addresses, err := e.ValidateDestination(request.Context(), request.URL.String())
		if err != nil {
			return nil, err
		}
		port := "80"
		if parsed.Scheme == "https" {
			port = "443"
		}
		transport := &http.Transport{
			Proxy:                 nil,
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   TLSHandshakeTimeout,
			ResponseHeaderTimeout: ResponseHeaderTimeout,
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				for _, address := range addresses {
					if !isPublicAddress(address) {
						return nil, &Error{Kind: KindAddressBlocked, Host: parsed.Hostname()}
					}
					connection, dialErr := e.dialer().DialContext(ctx, network, net.JoinHostPort(address.String(), port))
					if dialErr != nil {
						continue
					}
					remote, parseErr := remoteAddress(connection.RemoteAddr())
					if parseErr != nil || !isPublicAddress(remote) || !containsAddress(addresses, remote) {
						_ = connection.Close()
						return nil, &Error{Kind: KindAddressBlocked, Host: parsed.Hostname()}
					}
					return connection, nil
				}
				return nil, &Error{Kind: KindTransport, Host: parsed.Hostname()}
			},
		}
		return transport.RoundTrip(request)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func remoteAddress(address net.Addr) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(address.String())
	if err != nil {
		return netip.Addr{}, err
	}
	return netip.ParseAddr(host)
}

func containsAddress(addresses []netip.Addr, target netip.Addr) bool {
	for _, address := range addresses {
		if address == target.Unmap() {
			return true
		}
	}
	return false
}

func isTimeout(err error) bool {
	var networkError net.Error
	return errors.As(err, &networkError) && networkError.Timeout()
}

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func isBlockedHostname(host string) bool {
	switch host {
	case "localhost", "localhost.localdomain", "instance-data.ec2.internal", "instance-data", "ip6-localhost", "ip6-loopback":
		return true
	default:
		return false
	}
}

func isAmbiguousIPv4(host string) bool {
	if strings.HasPrefix(host, "0x") || strings.HasPrefix(host, "0X") {
		return true
	}
	parts := strings.Split(host, ".")
	if len(parts) == 1 {
		for _, char := range host {
			if char < '0' || char > '9' {
				return false
			}
		}
		return host != ""
	}
	if len(parts) != 4 {
		for _, part := range parts {
			if part == "" {
				return false
			}
			for _, char := range part {
				if char < '0' || char > '9' {
					return false
				}
			}
		}
		return true
	}
	for _, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return true
		}
	}
	return false
}

// IANA IPv4/IPv6 Special-Purpose registries, retrieved 2026-07-17. Covered prefixes
// are denied unless registry properties explicitly permit forwarding and global reachability.
var deniedPrefixes = mustPrefixes(
	"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16",
	"172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.31.196.0/24", "192.52.193.0/24",
	"192.88.99.0/24", "192.168.0.0/16", "192.175.48.0/24", "198.18.0.0/15", "198.51.100.0/24",
	"203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4", "255.255.255.255/32",
	"::/128", "::1/128", "::ffff:0:0/96", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64",
	"2001::/23", "2001:2::/48", "2001:db8::/32", "2001:10::/28", "2002::/16", "3fff::/20",
	"fc00::/7", "fe80::/10", "ff00::/8",
)

func mustPrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(fmt.Sprintf("invalid outbound policy prefix %q", value))
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func isPublicAddress(address netip.Addr) bool {
	address = address.Unmap()
	if !address.IsValid() {
		return false
	}
	for _, prefix := range deniedPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}
