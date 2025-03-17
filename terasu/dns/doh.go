package dns

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/ttl"
	"golang.org/x/net/http2"

	"github.com/fumiama/terasu"
	"github.com/fumiama/terasu/ip"
)

var ErrEmptyHostAddress = errors.New("empty host addr")

type recordType uint16

const (
	recordTypeNone recordType = 0
	recordTypeA    recordType = 1
	recordTypeAAAA recordType = 28
)

type dohjsonresponse struct {
	Status   uint32
	TC       bool
	RD       bool
	RA       bool
	AD       bool
	CD       bool
	Question []struct {
		Name string     `json:"name"`
		Type recordType `json:"type"`
	}
	Answer []struct {
		Name string     `json:"name"`
		Type recordType `json:"type"`
		TTL  uint16
		Data string `json:"data"`
	}
	EdnsClientSubnet string `json:"edns_client_subnet"`
	Comment          string
}

func (jr *dohjsonresponse) hosts() []string {
	if len(jr.Answer) == 0 {
		return nil
	}
	hosts := make([]string, 0, len(jr.Answer))
	for _, ans := range jr.Answer {
		if ans.Type == recordTypeA || ans.Type == recordTypeAAAA {
			hosts = append(hosts, ans.Data)
		}
	}
	return hosts
}

var lookupTable = ttl.NewCache[string, []string](time.Hour)

var trsHTTP2ClientWithSystemDNS = http.Client{
	Transport: &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
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
				addrs, err = net.DefaultResolver.LookupHost(ctx, host)
				if err != nil {
					return nil, err
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
				tlsConn = tls.Client(conn, cfg)
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
				tlsConn = tls.Client(conn, cfg)
				err = tlsConn.HandshakeContext(ctx)
				if err == nil {
					break
				}
				_ = tlsConn.Close()
				tlsConn = nil
			}
			return tlsConn, err
		},
	},
}

func lookupdoh(server, u string) (jr dohjsonresponse, err error) {
	jr, err = lookupdohwithtype(server, u, preferreddohtype())
	if err == nil {
		return
	}
	if ip.IsIPv6Available.Get() {
		jr, err = lookupdohwithtype(server, u, recordTypeA)
	}
	return
}

func lookupdohwithtype(server, u string, typ recordType) (jr dohjsonresponse, err error) {
	sb := strings.Builder{}
	sb.WriteString(server)
	sb.WriteString("?name=")
	sb.WriteString(url.QueryEscape(u))
	if typ != recordTypeNone {
		sb.WriteString("&type=")
		sb.WriteString(strconv.Itoa(int(typ)))
	}
	req, err := http.NewRequest("GET", sb.String(), nil)
	if err != nil {
		return
	}
	req.Header.Add("accept", "application/dns-json")
	resp, err := trsHTTP2ClientWithSystemDNS.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&jr)
	if err != nil {
		return
	}
	if jr.Status != 0 {
		err = errors.New("comment: " + jr.Comment)
	}
	return
}

func preferreddohtype() recordType {
	if ip.IsIPv6Available.Get() {
		return recordTypeAAAA
	}
	return recordTypeA
}
