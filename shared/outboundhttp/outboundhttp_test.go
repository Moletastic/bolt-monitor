package outboundhttp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"testing"
)

type fakeResolver struct {
	addresses []netip.Addr
	err       error
}

type sequenceResolver struct {
	answers [][]netip.Addr
	index   int
}

func (r *sequenceResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	answer := r.answers[r.index]
	if r.index < len(r.answers)-1 {
		r.index++
	}
	return answer, nil
}

type scriptedDialer struct {
	mutex     sync.Mutex
	responses []string
	remote    netip.Addr
	calls     []string
}

func (d *scriptedDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.calls = append(d.calls, address)
	if len(d.responses) == 0 {
		return nil, errors.New("unexpected dial")
	}
	response := d.responses[0]
	d.responses = d.responses[1:]
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		request, err := http.ReadRequest(bufio.NewReader(server))
		if err == nil && request.Body != nil {
			_ = request.Body.Close()
		}
		_, _ = server.Write([]byte(response))
	}()
	return remoteConn{Conn: client, remote: &net.TCPAddr{IP: net.IP(d.remote.AsSlice()), Port: 443}}, nil
}

type remoteConn struct {
	net.Conn
	remote net.Addr
}

func (c remoteConn) RemoteAddr() net.Addr { return c.remote }

func (r fakeResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.addresses, r.err
}

func TestValidateURLStaticPolicy(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		kind Kind
	}{
		{name: "public https", raw: "https://status.example.com/health"},
		{name: "public http", raw: "http://status.example.com/health"},
		{name: "unsupported scheme", raw: "ftp://status.example.com", kind: KindSchemeRejected},
		{name: "userinfo", raw: "https://token:secret@status.example.com", kind: KindInvalidURL},
		{name: "explicit port", raw: "https://status.example.com:8443", kind: KindInvalidURL},
		{name: "cgnat", raw: "http://100.64.0.1", kind: KindAddressBlocked},
		{name: "benchmark", raw: "http://198.18.0.1", kind: KindAddressBlocked},
		{name: "documentation ipv4", raw: "http://203.0.113.1", kind: KindAddressBlocked},
		{name: "documentation ipv6", raw: "http://[2001:db8::1]", kind: KindAddressBlocked},
		{name: "mapped loopback", raw: "http://[::ffff:127.0.0.1]", kind: KindAddressBlocked},
		{name: "integer bypass", raw: "http://2130706433", kind: KindAddressBlocked},
		{name: "shortened bypass", raw: "http://127.1", kind: KindAddressBlocked},
		{name: "octal bypass", raw: "http://0177.0.0.1", kind: KindAddressBlocked},
		{name: "hex bypass", raw: "http://0x7f000001", kind: KindAddressBlocked},
		{name: "localhost alias", raw: "http://LOCALHOST./", kind: KindAddressBlocked},
		{name: "aws metadata alias", raw: "http://instance-data.ec2.internal/latest", kind: KindAddressBlocked},
		{name: "ec2 metadata", raw: "http://169.254.169.254/latest", kind: KindAddressBlocked},
		{name: "ecs credentials", raw: "http://169.254.170.2", kind: KindAddressBlocked},
		{name: "eks pod identity", raw: "http://169.254.170.23", kind: KindAddressBlocked},
		{name: "vpc dns", raw: "http://169.254.169.253", kind: KindAddressBlocked},
		{name: "time sync", raw: "http://169.254.169.123", kind: KindAddressBlocked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ValidateURL(test.raw)
			if test.kind == "" {
				if err != nil {
					t.Fatalf("ValidateURL(%q) error = %v", test.raw, err)
				}
				return
			}
			if !IsKind(err, test.kind) {
				t.Fatalf("ValidateURL(%q) error = %v, want kind %q", test.raw, err, test.kind)
			}
			if got := err.Error(); got == "" || containsSecret(got) {
				t.Fatalf("unsafe error = %q", got)
			}
		})
	}
}

func TestSafeMessageNeverContainsSecrets(t *testing.T) {
	err := &Error{Kind: KindTransport, Host: "status.example.com", err: assertError("https://token:secret@example.com?key=private")}
	if got := SafeMessage(err); got != "outbound request failed" {
		t.Fatalf("SafeMessage = %q", got)
	}
}

func TestValidateDestinationRejectsUnsafeDNSAnswers(t *testing.T) {
	tests := []struct {
		name      string
		resolver  fakeResolver
		wantKind  Kind
		wantCount int
	}{
		{name: "public ipv4 and ipv6", resolver: fakeResolver{addresses: []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("2606:4700:4700::1111")}}, wantCount: 2},
		{name: "empty", resolver: fakeResolver{}, wantKind: KindResolutionFailed},
		{name: "failure", resolver: fakeResolver{err: errors.New("resolver secret")}, wantKind: KindResolutionFailed},
		{name: "blocked", resolver: fakeResolver{addresses: []netip.Addr{netip.MustParseAddr("127.0.0.1")}}, wantKind: KindAddressBlocked},
		{name: "mixed", resolver: fakeResolver{addresses: []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("10.0.0.1")}}, wantKind: KindAddressBlocked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, addresses, err := (&Executor{Resolver: test.resolver}).ValidateDestination(context.Background(), "https://status.example.com")
			if test.wantKind != "" {
				if !IsKind(err, test.wantKind) {
					t.Fatalf("ValidateDestination error = %v, want %q", err, test.wantKind)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateDestination error = %v", err)
			}
			if len(addresses) != test.wantCount {
				t.Fatalf("addresses = %v, want %d", addresses, test.wantCount)
			}
		})
	}
}

func TestExecutorPinsApprovedAddressAndPreservesHost(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"}}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}
	response, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com/health", Timeout: MaxRequestTimeout})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if response.StatusCode != http.StatusOK || string(response.Body) != "ok" {
		t.Fatalf("response = %#v", response)
	}
	if len(dialer.calls) != 1 || dialer.calls[0] != "8.8.8.8:80" {
		t.Fatalf("dial calls = %v", dialer.calls)
	}
}

func TestExecutorRejectsUnapprovedRemoteAddress(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &scriptedDialer{remote: netip.MustParseAddr("1.1.1.1"), responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"}}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}
	_, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if !IsKind(err, KindAddressBlocked) {
		t.Fatalf("Execute error = %v, want blocked address", err)
	}
}

func TestExecutorRejectsSensitiveCrossOriginRedirect(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 302 Found\r\nLocation: http://other.example.com\r\nContent-Length: 0\r\n\r\n"}}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}
	_, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com", Header: http.Header{"Authorization": {"secret"}}})
	if !IsKind(err, KindRedirectBlocked) {
		t.Fatalf("Execute error = %v, want blocked redirect", err)
	}
	if len(dialer.calls) != 1 {
		t.Fatalf("dial calls = %v, redirect target must not be dialed", dialer.calls)
	}
}

func TestExecutorFollowsCredentialFreeRedirectWithFreshDial(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &scriptedDialer{remote: approved, responses: []string{
		"HTTP/1.1 302 Found\r\nLocation: http://other.example.com\r\nContent-Length: 0\r\n\r\n",
		"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok",
	}}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}
	response, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if response.StatusCode != http.StatusOK || len(dialer.calls) != 2 {
		t.Fatalf("response = %#v, calls = %v", response, dialer.calls)
	}
	if strings.Contains(strings.Join(dialer.calls, ","), "status.example.com") {
		t.Fatalf("dial must use pinned address: %v", dialer.calls)
	}
}

func TestValidateDestinationRejectsDNSChangeBetweenRequests(t *testing.T) {
	resolver := &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}, {netip.MustParseAddr("10.0.0.1")}}}
	executor := &Executor{Resolver: resolver}
	if _, _, err := executor.ValidateDestination(context.Background(), "https://status.example.com"); err != nil {
		t.Fatalf("first validation error = %v", err)
	}
	if _, _, err := executor.ValidateDestination(context.Background(), "https://status.example.com"); !IsKind(err, KindAddressBlocked) {
		t.Fatalf("second validation error = %v, want blocked", err)
	}
}

func TestExecutorRejectsOversizedResponse(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	body := strings.Repeat("x", MaxResponseBytes+1)
	dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: " + fmt.Sprint(len(body)) + "\r\n\r\n" + body}}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}
	_, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if !IsKind(err, KindResponseTooLarge) {
		t.Fatalf("Execute error = %v, want response too large", err)
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }

func containsSecret(value string) bool {
	return value == "secret" || value == "private"
}
