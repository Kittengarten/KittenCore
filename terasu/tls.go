package terasu

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"hash"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"
	_ "unsafe"
)

type recordType uint8

const (
	recordTypeChangeCipherSpec recordType = 20
	recordTypeAlert            recordType = 21
	recordTypeHandshake        recordType = 22
	recordTypeApplicationData  recordType = 23
)

const (
	recordHeaderLen = 5 // record header length
)

type alert uint8

//go:linkname alertError tls.(tls.alert).Error
func alertError(e alert) string

func (e alert) Error() string {
	return alertError(e)
}

// A halfConn represents one direction of the record layer
// connection, either sending or receiving.
type halfConn struct {
	sync.Mutex

	err     error  // first permanent error
	version uint16 // protocol version
	cipher  any    // cipher algorithm
	mac     hash.Hash
	seq     [8]byte // 64-bit sequence number

	scratchBuf [13]byte // to avoid allocs; interface method args escape

	nextCipher any       // next encryption state
	nextMac    hash.Hash // next MAC algorithm

	level         tls.QUICEncryptionLevel // current QUIC encryption level
	trafficSecret []byte                  // current TLS 1.3 traffic secret
}

type Conn tls.Conn

// A _trsconn represents a secured connection.
// It implements the net._trsconn interface.
type _trsconn struct {
	// constant
	conn        net.Conn
	isClient    bool
	handshakeFn func(context.Context) error // (*Conn).clientHandshake or serverHandshake
	quic        *uintptr                    // nil for non-QUIC connections

	// isHandshakeComplete is true if the connection is currently transferring
	// application data (i.e. is not currently processing a handshake).
	// isHandshakeComplete is true implies handshakeErr == nil.
	isHandshakeComplete atomic.Bool
	// constant after handshake; protected by handshakeMutex
	handshakeMutex sync.Mutex
	handshakeErr   error       // error resulting from handshake
	vers           uint16      // TLS version
	haveVers       bool        // version has been negotiated
	config         *tls.Config // configuration passed to constructor
	// handshakes counts the number of handshakes performed on the
	// connection so far. If renegotiation is disabled then this is either
	// zero or one.
	handshakes       int
	extMasterSecret  bool
	didResume        bool // whether this connection was a session resumption
	didHRR           bool // whether a HelloRetryRequest was sent/received
	cipherSuite      uint16
	curveID          tls.CurveID
	ocspResponse     []byte   // stapled OCSP response
	scts             [][]byte // signed certificate timestamps from server
	peerCertificates []*x509.Certificate
	// activeCertHandles contains the cache handles to certificates in
	// peerCertificates that are used to track active references.
	activeCertHandles []*uintptr
	// verifiedChains contains the certificate chains that we built, as
	// opposed to the ones presented by the server.
	verifiedChains [][]*x509.Certificate
	// serverName contains the server name indicated by the client, if any.
	serverName string
	// secureRenegotiation is true if the server echoed the secure
	// renegotiation extension. (This is meaningless as a server because
	// renegotiation is not supported in that case.)
	secureRenegotiation bool
	// ekm is a closure for exporting keying material.
	ekm func(label string, context []byte, length int) ([]byte, error)
	// resumptionSecret is the resumption_master_secret for handling
	// or sending NewSessionTicket messages.
	resumptionSecret []byte
	echAccepted      bool

	// ticketKeys is the set of active session ticket keys for this
	// connection. The first one is used to encrypt new tickets and
	// all are tried to decrypt tickets.
	ticketKeys []byte

	// clientFinishedIsFirst is true if the client sent the first Finished
	// message during the most recent handshake. This is recorded because
	// the first transmitted Finished message is the tls-unique
	// channel-binding value.
	clientFinishedIsFirst bool

	// closeNotifyErr is any error from sending the alertCloseNotify record.
	closeNotifyErr error
	// closeNotifySent is true if the Conn attempted to send an
	// alertCloseNotify record.
	closeNotifySent bool

	// clientFinished and serverFinished contain the Finished message sent
	// by the client or server in the most recent handshake. This is
	// retained to support the renegotiation extension and tls-unique
	// channel-binding.
	clientFinished [12]byte
	serverFinished [12]byte

	// clientProtocol is the negotiated ALPN protocol.
	clientProtocol string

	// input/output
	in, out halfConn
}

//go:linkname outBufPool crypto/tls.outBufPool
var outBufPool sync.Pool

//go:linkname tlsWriteRecordLocked crypto/tls.(*Conn).writeRecordLocked
func tlsWriteRecordLocked(c *_trsconn, typ recordType, data []byte) (int, error)

//go:linkname maxPayloadSizeForWrite crypto/tls.(*Conn).maxPayloadSizeForWrite
func maxPayloadSizeForWrite(c *_trsconn, typ recordType) int

func (c *_trsconn) maxPayloadSizeForWrite(typ recordType) int {
	return maxPayloadSizeForWrite(c, typ)
}

//go:linkname sliceForAppend crypto/tls.sliceForAppend
func sliceForAppend(in []byte, n int) (head, tail []byte)

//go:linkname encrypt crypto/tls.(*halfConn).encrypt
func encrypt(hc *halfConn, record, payload []byte, rand io.Reader) ([]byte, error)

func (hc *halfConn) encrypt(record, payload []byte, rand io.Reader) ([]byte, error) {
	return encrypt(hc, record, payload, rand)
}

//go:linkname rand crypto/tls.(*Config).rand
func rand(c *tls.Config) io.Reader

//go:linkname write crypto/tls.(*Conn).write
func write(c *_trsconn, data []byte) (int, error)

func (c *_trsconn) write(data []byte) (int, error) {
	return write(c, data)
}

//go:linkname flush crypto/tls.(*Conn).flush
func flush(c *_trsconn) (int, error)

func (c *_trsconn) flush() (int, error) {
	return flush(c)
}

//go:linkname changeCipherSpec crypto/tls.(*halfConn).changeCipherSpec
func changeCipherSpec(hc *halfConn) error

func (hc *halfConn) changeCipherSpec() error {
	return changeCipherSpec(hc)
}

//go:linkname sendAlertLocked crypto/tls.(*Conn).sendAlertLocked
func sendAlertLocked(c *_trsconn, err alert) error

func (c *_trsconn) sendAlertLocked(err alert) error {
	return sendAlertLocked(c, err)
}

// writeRecordLocked writes a TLS record with the given type and payload to the
// connection and updates the record layer state.
func (c *_trsconn) writeRecordLocked(typ recordType, firstFragmentLen uint8, data []byte) (int, error) {
	if c.quic != nil {
		return tlsWriteRecordLocked(c, typ, data)
	}

	outBufPtr := outBufPool.Get().(*[]byte)
	outBuf := *outBufPtr
	defer func() {
		// You might be tempted to simplify this by just passing &outBuf to Put,
		// but that would make the local copy of the outBuf slice header escape
		// to the heap, causing an allocation. Instead, we keep around the
		// pointer to the slice header returned by Get, which is already on the
		// heap, and overwrite and return that.
		*outBufPtr = outBuf
		outBufPool.Put(outBufPtr)
	}()

	var n int
	isFirstLoop := true
	for len(data) > 0 {
		m := len(data)
		if !isFirstLoop {
			if maxPayload := c.maxPayloadSizeForWrite(typ); m > maxPayload {
				m = maxPayload
			}
		} else {
			m = int(firstFragmentLen)
		}

		_, outBuf = sliceForAppend(outBuf[:0], recordHeaderLen)
		outBuf[0] = byte(typ)
		vers := c.vers
		if vers == 0 {
			// Some TLS servers fail if the record version is
			// greater than TLS 1.0 for the initial ClientHello.
			vers = tls.VersionTLS10
		} else if vers == tls.VersionTLS13 {
			// TLS 1.3 froze the record layer version to 1.2.
			// See RFC 8446, Section 5.1.
			vers = tls.VersionTLS12
		}
		outBuf[1] = byte(vers >> 8)
		outBuf[2] = byte(vers)
		outBuf[3] = byte(m >> 8)
		outBuf[4] = byte(m)

		var err error
		outBuf, err = c.out.encrypt(outBuf, data[:m], rand(c.config))
		if err != nil {
			return n, err
		}
		if _, err := c.write(outBuf); err != nil {
			return n, err
		}
		n += m
		data = data[m:]
		if isFirstLoop {
			isFirstLoop = false
			if _, err := c.flush(); err != nil {
				return n, err
			}
		}
	}

	if typ == recordTypeChangeCipherSpec && c.vers != tls.VersionTLS13 {
		if err := c.out.changeCipherSpec(); err != nil {
			return n, c.sendAlertLocked(alert(
				*(*uintptr)(
					unsafe.Add(unsafe.Pointer(&err), unsafe.Sizeof(uintptr(0))),
				),
			))
		}
	}

	return n, nil
}
