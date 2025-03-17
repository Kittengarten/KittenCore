package terasu

import (
	"context"
	"crypto/tls"
	"unsafe"
)

var DefaultFirstFragmentLen uint8 = 4

// Use terasu in this TLS conn
func Use(conn *tls.Conn) *Conn {
	return (*Conn)(conn)
}

// Handshake do terasu handshake in this TLS conn
func (conn *Conn) Handshake(firstFragmentLen uint8) error {
	expose := (*_trsconn)(unsafe.Pointer(conn))
	fnbak := expose.handshakeFn
	expose.handshakeFn = conn.clientHandshake(firstFragmentLen)
	defer func() { expose.handshakeFn = fnbak }()
	return (*tls.Conn)(conn).Handshake()
}

// Handshake do terasu handshake with ctx in this TLS conn
func (conn *Conn) HandshakeContext(ctx context.Context, firstFragmentLen uint8) error {
	expose := (*_trsconn)(unsafe.Pointer(conn))
	fnbak := expose.handshakeFn
	expose.handshakeFn = conn.clientHandshake(firstFragmentLen)
	defer func() { expose.handshakeFn = fnbak }()
	return (*tls.Conn)(conn).HandshakeContext(ctx)
}
