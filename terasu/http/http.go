package http

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/FloatTech/ttl"

	"github.com/fumiama/terasu"
	"github.com/fumiama/terasu/dns"
	"github.com/fumiama/terasu/ip"
)

var (
	ErrNoTLSConnection  = errors.New("no tls connection")
	ErrEmptyHostAddress = errors.New("empty host addr")
)

var defaultDialer = net.Dialer{
	Timeout: time.Minute,
}

func SetDefaultClientTimeout(t time.Duration) {
	defaultDialer.Timeout = t
}

var lookupTable = ttl.NewCache[string, []string](time.Hour)

var DefaultClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if defaultDialer.Timeout != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, defaultDialer.Timeout)
				defer cancel()
			}

			if !defaultDialer.Deadline.IsZero() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, defaultDialer.Deadline)
				defer cancel()
			}

			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			addrs := lookupTable.Get(host)
			if len(addrs) == 0 {
				addrs, err = dns.DefaultResolver.LookupHost(ctx, host)
				if err != nil {
					if ip.IsIPv6Available.Get() {
						addrs, err = dns.IPv6Servers.LookupHostFallback(ctx, host)
					} else {
						addrs, err = dns.IPv4Servers.LookupHostFallback(ctx, host)
					}
					if err != nil {
						return nil, err
					}
				}
				lookupTable.Set(host, addrs)
			}
			if len(addr) == 0 {
				return nil, ErrEmptyHostAddress
			}
			var conn net.Conn
			var tlsConn *tls.Conn
			for _, a := range addrs {
				conn, err = defaultDialer.DialContext(ctx, network, net.JoinHostPort(a, port))
				if err != nil {
					continue
				}
				tlsConn = tls.Client(conn, &tls.Config{
					ServerName: host,
				})
				err = terasu.Use(tlsConn).HandshakeContext(ctx, terasu.DefaultFirstFragmentLen)
				if err == nil {
					break
				}
				_ = tlsConn.Close()
				tlsConn = nil
				conn, err = defaultDialer.DialContext(ctx, network, net.JoinHostPort(a, port))
				if err != nil {
					continue
				}
				tlsConn = tls.Client(conn, &tls.Config{
					ServerName: host,
				})
				err = tlsConn.HandshakeContext(ctx)
				if err == nil {
					break
				}
				_ = tlsConn.Close()
				tlsConn = nil
			}
			return tlsConn, err
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

func Head(url string) (resp *http.Response, err error) {
	return DefaultClient.Head(url)
}

func Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, contentType, body)
}

func PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return DefaultClient.PostForm(url, data)
}
