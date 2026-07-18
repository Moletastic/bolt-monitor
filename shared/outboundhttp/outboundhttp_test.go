package outboundhttp

import (
	"context"
	"errors"
	"net/netip"
	"testing"
)

type fakeResolver struct {
	addresses []netip.Addr
	err       error
}

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

type assertError string

func (e assertError) Error() string { return string(e) }

func containsSecret(value string) bool {
	return value == "secret" || value == "private"
}
