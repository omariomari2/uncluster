package safehttp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

const (
	MaxResponseBytes = 25 << 20
	BrowserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	maxRedirects = 10
)

var ErrBlockedAddress = errors.New("target resolves to a non-public address")

var guardedDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
	Control:   guardAddress,
}

var guardedTransport = &http.Transport{
	DialContext:           guardedDialer.DialContext,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 30 * time.Second,
	MaxIdleConns:          32,
	MaxIdleConnsPerHost:   8,
	IdleConnTimeout:       30 * time.Second,
}

var reservedRanges = parseCIDRs([]string{
	"0.0.0.0/8",
	"100.64.0.0/10",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
	"2001:db8::/32",
})

func parseCIDRs(cidrs []string) []*net.IPNet {
	blocks := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		if _, block, err := net.ParseCIDR(cidr); err == nil {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: must be http or https", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL %q: missing host", rawURL)
	}
	return nil
}

func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if mapped := ip.To4(); mapped != nil {
		ip = mapped
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return true
	}
	for _, block := range reservedRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func guardAddress(_ string, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: cannot parse dial address %q", ErrBlockedAddress, address)
	}
	if ip := net.ParseIP(host); IsBlockedIP(ip) {
		return fmt.Errorf("%w: %s", ErrBlockedAddress, host)
	}
	return nil
}

func Client(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: guardedTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return ValidateURL(req.URL.String())
		},
	}
}

func ReadBody(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxResponseBytes {
		return nil, fmt.Errorf("response exceeds %d byte limit", MaxResponseBytes)
	}
	return data, nil
}
