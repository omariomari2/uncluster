package safehttp

import (
	"net"
	"testing"
	"time"
)

func TestIsBlockedIPRejectsNonPublicAddresses(t *testing.T) {
	blocked := []string{
		"127.0.0.1",
		"169.254.169.254",
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"0.0.0.0",
		"100.64.0.1",
		"224.0.0.1",
		"::1",
		"fc00::1",
		"fe80::1",
	}

	for _, addr := range blocked {
		ip := net.ParseIP(addr)
		if ip == nil {
			t.Fatalf("test fixture %q is not a valid IP", addr)
		}
		if !IsBlockedIP(ip) {
			t.Errorf("expected %s to be blocked", addr)
		}
	}
}

func TestIsBlockedIPAllowsPublicAddresses(t *testing.T) {
	allowed := []string{
		"1.1.1.1",
		"8.8.8.8",
		"93.184.216.34",
		"2606:4700:4700::1111",
	}

	for _, addr := range allowed {
		ip := net.ParseIP(addr)
		if ip == nil {
			t.Fatalf("test fixture %q is not a valid IP", addr)
		}
		if IsBlockedIP(ip) {
			t.Errorf("expected %s to be allowed", addr)
		}
	}
}

func TestIsBlockedIPRejectsNil(t *testing.T) {
	if !IsBlockedIP(nil) {
		t.Error("expected a nil IP to be blocked")
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://example.com/style.css", false},
		{"http", "http://example.com/style.css", false},
		{"file scheme", "file:///etc/passwd", true},
		{"gopher scheme", "gopher://example.com/", true},
		{"no scheme", "//example.com/style.css", true},
		{"no host", "https:///style.css", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateURL(%q) = nil, want error", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateURL(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}

func TestClientReusesGuardedTransport(t *testing.T) {
	first := Client(10 * time.Second)
	second := Client(30 * time.Second)

	if first == second {
		t.Fatal("Client() returned the same http.Client; per-call timeouts should remain independent")
	}
	if first.Transport != second.Transport {
		t.Fatal("Client() created distinct transports; callers cannot reuse guarded connection pools")
	}
}
