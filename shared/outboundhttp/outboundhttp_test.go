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
	"time"
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

type recordingResolver struct {
	mutex     sync.Mutex
	answers   map[string][]netip.Addr
	hosts     []string
	lookups   int
	lookupErr error
}

func (r *recordingResolver) LookupNetIP(_ context.Context, _, host string) ([]netip.Addr, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.hosts = append(r.hosts, host)
	r.lookups++
	return append([]netip.Addr(nil), r.answers[host]...), r.lookupErr
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

func TestValidateDestinationNormalizesAliasBeforeResolution(t *testing.T) {
	resolver := &recordingResolver{answers: map[string][]netip.Addr{
		"status.example.com": {netip.MustParseAddr("8.8.8.8")},
	}}
	_, addresses, err := (&Executor{Resolver: resolver}).ValidateDestination(context.Background(), "https://STATUS.EXAMPLE.COM./health")
	if err != nil {
		t.Fatalf("ValidateDestination error = %v", err)
	}
	if len(addresses) != 1 || addresses[0] != netip.MustParseAddr("8.8.8.8") {
		t.Fatalf("addresses = %v", addresses)
	}
	if got := strings.Join(resolver.hosts, ","); got != "status.example.com" {
		t.Fatalf("resolved hosts = %q", got)
	}
}

func TestExecutorPinsResolutionAndBypassesProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:9999")
	approved := netip.MustParseAddr("8.8.8.8")
	resolver := &recordingResolver{answers: map[string][]netip.Addr{"status.example.com": {approved}}}
	dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"}}
	executor := &Executor{Resolver: resolver, Dialer: dialer}
	if _, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"}); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if resolver.lookups != 1 || len(dialer.calls) != 1 || dialer.calls[0] != "8.8.8.8:80" {
		t.Fatalf("lookups = %d, dials = %v", resolver.lookups, dialer.calls)
	}
}

func TestExecutorFallsBackAcrossApprovedIPv4AndIPv6Candidates(t *testing.T) {
	first := netip.MustParseAddr("8.8.8.8")
	second := netip.MustParseAddr("2606:4700:4700::1111")
	dialer := &fallbackDialer{responses: map[string]string{
		"[2606:4700:4700::1111]:80": "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok",
	}, remote: second}
	executor := &Executor{Resolver: fakeResolver{addresses: []netip.Addr{first, second}}, Dialer: dialer}
	response, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if string(response.Body) != "ok" {
		t.Fatalf("response body = %q", response.Body)
	}
	if got := strings.Join(dialer.calls, ","); got != "8.8.8.8:80,[2606:4700:4700::1111]:80" {
		t.Fatalf("dial order = %q", got)
	}
}

func TestExecutorDoesNotReuseConnectionAfterDNSChanges(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	resolver := &sequenceResolver{answers: [][]netip.Addr{{approved}, {netip.MustParseAddr("10.0.0.1")}}}
	dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"}}
	executor := &Executor{Resolver: resolver, Dialer: dialer}
	if _, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"}); err != nil {
		t.Fatalf("first Execute error = %v", err)
	}
	_, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if !IsKind(err, KindAddressBlocked) {
		t.Fatalf("second Execute error = %v, want blocked address", err)
	}
	if len(dialer.calls) != 1 {
		t.Fatalf("dials = %v; blocked second request must not reuse or dial", dialer.calls)
	}
}

func TestExecutorRejectsBlockedRemoteAddressAndClosesConnection(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &closingDialer{remote: netip.MustParseAddr("10.0.0.1"), closed: make(chan struct{})}
	_, err := (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}).Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if !IsKind(err, KindAddressBlocked) {
		t.Fatalf("Execute error = %v, want blocked address", err)
	}
	dialer.assertClosed(t)
}

func TestExecutorClosesResponseBodyConnection(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	dialer := &closingDialer{remote: approved, response: "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok", closed: make(chan struct{})}
	response, err := (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}).Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if string(response.Body) != "ok" {
		t.Fatalf("response body = %q", response.Body)
	}
	dialer.assertClosed(t)
}

func TestExecutorRedirectBoundary(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	tests := []struct {
		name    string
		request Request
		want    Kind
	}{
		{name: "configured header", request: Request{Method: http.MethodGet, URL: "http://status.example.com", Header: http.Header{"X-Provider-Secret": {"secret"}}}, want: KindRedirectBlocked},
		{name: "body", request: Request{Method: http.MethodPost, URL: "http://status.example.com", Body: strings.NewReader("private")}, want: KindRedirectBlocked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialer := &scriptedDialer{remote: approved, responses: []string{"HTTP/1.1 302 Found\r\nLocation: http://other.example.com\r\nContent-Length: 0\r\n\r\n"}}
			_, err := (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: dialer}).Execute(context.Background(), test.request)
			if !IsKind(err, test.want) {
				t.Fatalf("Execute error = %v, want %q", err, test.want)
			}
			if len(dialer.calls) != 1 {
				t.Fatalf("dials = %v; redirect target must not be dialed", dialer.calls)
			}
		})
	}
}

func TestExecutorConcurrentRedirectsKeepRequestStateIsolated(t *testing.T) {
	resolver := &recordingResolver{answers: map[string][]netip.Addr{
		"one.example.com":      {netip.MustParseAddr("8.8.8.8")},
		"two.example.com":      {netip.MustParseAddr("1.1.1.1")},
		"one-next.example.com": {netip.MustParseAddr("9.9.9.9")},
		"two-next.example.com": {netip.MustParseAddr("208.67.222.222")},
	}}
	dialer := &addressResponseDialer{responses: map[string]string{
		"8.8.8.8:80":        "HTTP/1.1 302 Found\r\nLocation: http://one-next.example.com\r\nContent-Length: 0\r\n\r\n",
		"1.1.1.1:80":        "HTTP/1.1 302 Found\r\nLocation: http://two-next.example.com\r\nContent-Length: 0\r\n\r\n",
		"9.9.9.9:80":        "HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\none",
		"208.67.222.222:80": "HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\ntwo",
	}}
	executor := &Executor{Resolver: resolver, Dialer: dialer}
	results := make(chan string, 2)
	errs := make(chan error, 2)
	for _, host := range []string{"one.example.com", "two.example.com"} {
		go func(host string) {
			response, err := executor.Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://" + host})
			if err != nil {
				errs <- err
				return
			}
			results <- string(response.Body)
		}(host)
	}
	for range 2 {
		select {
		case err := <-errs:
			t.Fatalf("Execute error = %v", err)
		case <-results:
		}
	}
	if got := strings.Join(dialer.sortedCalls(), ","); got != "1.1.1.1:80,208.67.222.222:80,8.8.8.8:80,9.9.9.9:80" {
		t.Fatalf("dials = %q", got)
	}
}

func TestExecutorCancellationAndRedaction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{netip.MustParseAddr("8.8.8.8")}}, Dialer: blockingDialer{}}).Execute(ctx, Request{Method: http.MethodGet, URL: "http://status.example.com"})
	if !IsKind(err, KindTimeout) {
		t.Fatalf("Execute error = %v, want timeout", err)
	}
	if got := SafeMessage(err); got != "outbound request timed out" {
		t.Fatalf("SafeMessage = %q", got)
	}
}

func TestExecutorConfiguredTimeoutAndProviderFailureRedaction(t *testing.T) {
	approved := netip.MustParseAddr("8.8.8.8")
	_, err := (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: blockingDialer{}}).Execute(context.Background(), Request{Method: http.MethodGet, URL: "http://status.example.com", Timeout: time.Millisecond})
	if !IsKind(err, KindTimeout) {
		t.Fatalf("timeout error = %v, want timeout", err)
	}
	secretURL := "http://provider.example.com/path?token=secret"
	_, err = (&Executor{Resolver: fakeResolver{addresses: []netip.Addr{approved}}, Dialer: failingDialer{err: assertError("provider payload private")}}).Execute(context.Background(), Request{Method: http.MethodPost, URL: secretURL, Header: http.Header{"Authorization": {"Bearer private"}}, Body: strings.NewReader("provider payload")})
	if !IsKind(err, KindTransport) {
		t.Fatalf("transport error = %v, want transport failure", err)
	}
	for _, value := range []string{err.Error(), SafeMessage(err)} {
		if strings.Contains(value, "secret") || strings.Contains(value, "private") || strings.Contains(value, "payload") {
			t.Fatalf("error leaked secret: %q", value)
		}
	}
}

type fallbackDialer struct {
	mutex     sync.Mutex
	responses map[string]string
	remote    netip.Addr
	calls     []string
}

func (d *fallbackDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.mutex.Lock()
	d.calls = append(d.calls, address)
	response, ok := d.responses[address]
	d.mutex.Unlock()
	if !ok {
		return nil, errors.New("candidate unavailable")
	}
	client, server := net.Pipe()
	go serveResponse(server, response)
	return remoteConn{Conn: client, remote: &net.TCPAddr{IP: net.IP(d.remote.AsSlice()), Port: 80}}, nil
}

type closingDialer struct {
	remote   netip.Addr
	response string
	closed   chan struct{}
	once     sync.Once
}

func (d *closingDialer) DialContext(_ context.Context, _, _ string) (net.Conn, error) {
	client, server := net.Pipe()
	if d.response != "" {
		go serveResponse(server, d.response)
	}
	return trackedRemoteConn{remoteConn: remoteConn{Conn: client, remote: &net.TCPAddr{IP: net.IP(d.remote.AsSlice()), Port: 80}}, onClose: func() { d.once.Do(func() { close(d.closed) }) }, server: server}, nil
}

func (d *closingDialer) assertClosed(t *testing.T) {
	t.Helper()
	select {
	case <-d.closed:
	case <-time.After(time.Second):
		t.Fatal("response connection was not closed")
	}
}

type trackedRemoteConn struct {
	remoteConn
	onClose func()
	server  net.Conn
}

func (c trackedRemoteConn) Close() error {
	c.onClose()
	_ = c.server.Close()
	return c.remoteConn.Close()
}

type addressResponseDialer struct {
	mutex     sync.Mutex
	responses map[string]string
	calls     []string
}

func (d *addressResponseDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.mutex.Lock()
	d.calls = append(d.calls, address)
	response, ok := d.responses[address]
	d.mutex.Unlock()
	if !ok {
		return nil, errors.New("unexpected dial")
	}
	host, _, _ := net.SplitHostPort(address)
	remote := netip.MustParseAddr(host)
	client, server := net.Pipe()
	go serveResponse(server, response)
	return remoteConn{Conn: client, remote: &net.TCPAddr{IP: net.IP(remote.AsSlice()), Port: 80}}, nil
}

func (d *addressResponseDialer) sortedCalls() []string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	calls := append([]string(nil), d.calls...)
	for index := range calls {
		for next := index + 1; next < len(calls); next++ {
			if calls[next] < calls[index] {
				calls[index], calls[next] = calls[next], calls[index]
			}
		}
	}
	return calls
}

type blockingDialer struct{}

func (blockingDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

type failingDialer struct{ err error }

func (d failingDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, d.err
}

func serveResponse(connection net.Conn, response string) {
	defer connection.Close()
	request, err := http.ReadRequest(bufio.NewReader(connection))
	if err == nil && request.Body != nil {
		_ = request.Body.Close()
	}
	_, _ = connection.Write([]byte(response))
}

type assertError string

func (e assertError) Error() string { return string(e) }

func containsSecret(value string) bool {
	return value == "secret" || value == "private"
}
