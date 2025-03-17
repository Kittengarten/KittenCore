package tls

import (
	"github.com/Kittengarten/KittenCore/kitten/core/http"

	trshttp "github.com/fumiama/terasu/http"
	trshttp2 "github.com/fumiama/terasu/http2"
)

func init() {
	http.TLSClient = trshttp.DefaultClient
	http.TLSHTTP2Client = trshttp2.DefaultClient
}
