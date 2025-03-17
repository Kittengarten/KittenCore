package ip

import (
	"context"
	"net/http"
	"time"

	"github.com/RomiChan/syncx"
)

var IsIPv6Available = syncx.Lazy[bool]{Init: func() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://v6.ipv6-test.com/json/widgetdata.php?callback=?", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}}

func init() {
	go IsIPv6Available.Get()
}
