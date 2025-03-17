package terasu

import (
	"context"
	"crypto"
	"crypto/ecdh"
	"crypto/tls"
	"errors"
	"hash"
	"io"
	"unsafe"
)

//go:linkname defaultConfig crypto/tls.defaultConfig
func defaultConfig() *tls.Config

type clientHelloMsg struct {
	raw                              []byte
	vers                             uint16
	random                           []byte
	sessionId                        []byte
	cipherSuites                     []uint16
	compressionMethods               []uint8
	serverName                       string
	ocspStapling                     bool
	supportedCurves                  []tls.CurveID
	supportedPoints                  []uint8
	ticketSupported                  bool
	sessionTicket                    []uint8
	supportedSignatureAlgorithms     []tls.SignatureScheme
	supportedSignatureAlgorithmsCert []tls.SignatureScheme
	secureRenegotiationSupported     bool
	secureRenegotiation              []byte
	extendedMasterSecret             bool
	alpnProtocols                    []string
	scts                             bool
	supportedVersions                []uint16
	cookie                           []byte
	keyShares                        []byte
	earlyData                        bool
}

//go:linkname marshal crypto/tls.(*clientHelloMsg).marshal
func marshal(m *clientHelloMsg) ([]byte, error)

func (m *clientHelloMsg) marshal() ([]byte, error) {
	return marshal(m)
}

//go:linkname unmarshal crypto/tls.(*clientHelloMsg).unmarshal
func unmarshal(m *clientHelloMsg, data []byte) bool

func (m *clientHelloMsg) unmarshal(data []byte) bool {
	return unmarshal(m, data)
}

//go:linkname makeClientHello crypto/tls.(*Conn).makeClientHello
func makeClientHello(c *_trsconn) (*clientHelloMsg, *ecdh.PrivateKey, error)

func (c *_trsconn) makeClientHello() (*clientHelloMsg, *ecdh.PrivateKey, error) {
	return makeClientHello(c)
}

// A sessionState is a resumable session.
type sessionState struct {
	// Encoded as a SessionState (in the language of RFC 8446, Section 3).
	//
	//   enum { server(1), client(2) } SessionStateType;
	//
	//   opaque Certificate<1..2^24-1>;
	//
	//   Certificate CertificateChain<0..2^24-1>;
	//
	//   opaque Extra<0..2^24-1>;
	//
	//   struct {
	//       uint16 version;
	//       SessionStateType type;
	//       uint16 cipher_suite;
	//       uint64 created_at;
	//       opaque secret<1..2^8-1>;
	//       Extra extra<0..2^24-1>;
	//       uint8 ext_master_secret = { 0, 1 };
	//       uint8 early_data = { 0, 1 };
	//       CertificateEntry certificate_list<0..2^24-1>;
	//       CertificateChain verified_chains<0..2^24-1>; /* excluding leaf */
	//       select (SessionState.early_data) {
	//           case 0: Empty;
	//           case 1: opaque alpn<1..2^8-1>;
	//       };
	//       select (SessionState.type) {
	//           case server: Empty;
	//           case client: struct {
	//               select (SessionState.version) {
	//                   case VersionTLS10..VersionTLS12: Empty;
	//                   case VersionTLS13: struct {
	//                       uint64 use_by;
	//                       uint32 age_add;
	//                   };
	//               };
	//           };
	//       };
	//   } SessionState;
	//

	// Extra is ignored by crypto/tls, but is encoded by [SessionState.Bytes]
	// and parsed by [ParseSessionState].
	//
	// This allows [Config.UnwrapSession]/[Config.WrapSession] and
	// [ClientSessionCache] implementations to store and retrieve additional
	// data alongside this session.
	//
	// To allow different layers in a protocol stack to share this field,
	// applications must only append to it, not replace it, and must use entries
	// that can be recognized even if out of order (for example, by starting
	// with an id and version prefix).
	Extra [][]byte

	// EarlyData indicates whether the ticket can be used for 0-RTT in a QUIC
	// connection. The application may set this to false if it is true to
	// decline to offer 0-RTT even if supported.
	EarlyData bool

	version     uint16
	isClient    bool
	cipherSuite uint16
}

//go:linkname loadSession crypto/tls.(*Conn).loadSession
func loadSession(c *_trsconn, hello *clientHelloMsg) (
	session *sessionState, earlySecret *EarlySecret, binderKey []byte, err error,
)

func (c *_trsconn) loadSession(hello *clientHelloMsg) (
	session *sessionState, earlySecret *EarlySecret, binderKey []byte, err error,
) {
	return loadSession(c, hello)
}

//go:linkname clientSessionCacheKey crypto/tls.(*Conn).clientSessionCacheKey
func clientSessionCacheKey(c *_trsconn) string

func (c *_trsconn) clientSessionCacheKey() string {
	return clientSessionCacheKey(c)
}

// A cipherSuiteTLS13 defines only the pair of the AEAD algorithm and hash
// algorithm to be used with HKDF. See RFC 8446, Appendix B.4.
type cipherSuiteTLS13 struct {
	id     uint16
	keyLen int
	aead   func(key, fixedNonce []byte) any
	hash   crypto.Hash
}

type EarlySecret struct {
	secret []byte
	hash   func() fips140Hash
}

type fips140Hash interface {
	// Write (via the embedded io.Writer interface) adds more data to the
	// running hash. It never returns an error.
	io.Writer

	// Sum appends the current hash to b and returns the resulting slice.
	// It does not change the underlying hash state.
	Sum(b []byte) []byte

	// Reset resets the Hash to its initial state.
	Reset()

	// Size returns the number of bytes Sum will return.
	Size() int

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int
}

//go:linkname ClientEarlyTrafficSecret crypto/internal/fips140/tls13.(*EarlySecret).ClientEarlyTrafficSecret
func ClientEarlyTrafficSecret(s *EarlySecret, transcript fips140Hash) []byte

//go:linkname cipherSuiteTLS13ByID crypto/tls.cipherSuiteTLS13ByID
func cipherSuiteTLS13ByID(id uint16) *cipherSuiteTLS13

type handshakeMessage interface {
	marshal() ([]byte, error)
	unmarshal([]byte) bool
}

type transcriptHash interface {
	Write([]byte) (int, error)
}

//go:linkname transcriptMsg crypto/tls.transcriptMsg
func transcriptMsg(msg handshakeMessage, h transcriptHash) error

const clientEarlyTrafficLabel = "c e traffic"

//go:linkname quicSetWriteSecret crypto/tls.(*Conn).quicSetWriteSecret
func quicSetWriteSecret(c *_trsconn, level tls.QUICEncryptionLevel, suite uint16, secret []byte)

//go:linkname readHandshake crypto/tls.(*Conn).readHandshake
func readHandshake(c *_trsconn, transcript transcriptHash) (any, error)

func (c *_trsconn) readHandshake(transcript transcriptHash) (any, error) {
	return readHandshake(c, transcript)
}

type serverHelloMsg struct {
	raw    []byte
	vers   uint16
	random []byte
}

//go:linkname sendAlert crypto/tls.(*Conn).sendAlert
func sendAlert(c *_trsconn, err alert) error

func (c *_trsconn) sendAlert(err alert) error {
	return sendAlert(c, err)
}

//go:linkname unexpectedMessageError crypto/tls.unexpectedMessageError
func unexpectedMessageError(wanted, got any) error

const (
	alertUnexpectedMessage alert = 10
	alertIllegalParameter  alert = 47
)

//go:linkname pickTLSVersion crypto/tls.(*Conn).pickTLSVersion
func pickTLSVersion(c *_trsconn, serverHello *serverHelloMsg) error

func (c *_trsconn) pickTLSVersion(serverHello *serverHelloMsg) error {
	return pickTLSVersion(c, serverHello)
}

//go:linkname maxSupportedVersion crypto/tls.(*Config).maxSupportedVersion
func maxSupportedVersion(c *tls.Config, isClient bool) uint16

const roleClient = true

const (
	// downgradeCanaryTLS12 or downgradeCanaryTLS11 is embedded in the server
	// random as a downgrade protection if the server would be capable of
	// negotiating a higher version. See RFC 8446, Section 4.1.3.
	downgradeCanaryTLS12 = "DOWNGRD\x01"
	downgradeCanaryTLS11 = "DOWNGRD\x00"
)

type clientHandshakeStateTLS13 struct {
	c           *Conn
	ctx         context.Context
	serverHello *serverHelloMsg
	hello       *clientHelloMsg
	ecdheKey    *ecdh.PrivateKey

	session     *sessionState
	earlySecret *EarlySecret
	binderKey   []byte

	certReq       *uintptr
	usingPSK      bool
	sentDummyCCS  bool
	suite         *cipherSuiteTLS13
	transcript    hash.Hash
	masterSecret  []byte
	trafficSecret []byte // client_application_traffic_secret_0
}

//go:linkname handshake13 crypto/tls.(*clientHandshakeStateTLS13).handshake
func handshake13(hs *clientHandshakeStateTLS13) error

func (hs *clientHandshakeStateTLS13) handshake() error {
	return handshake13(hs)
}

// A finishedHash calculates the hash of a set of handshake messages suitable
// for including in a Finished message.
type finishedHash struct {
	client hash.Hash
	server hash.Hash

	// Prior to TLS 1.2, an additional MD5 hash is required.
	clientMD5 hash.Hash
	serverMD5 hash.Hash

	// In TLS 1.2, a full buffer is sadly required.
	buffer []byte

	version uint16
	prf     func(result, secret, label, seed []byte)
}

type clientHandshakeState struct {
	c            *Conn
	ctx          context.Context
	serverHello  *serverHelloMsg
	hello        *clientHelloMsg
	suite        *uintptr
	finishedHash finishedHash
	masterSecret []byte
	session      *sessionState // the session being resumed
	ticket       []byte        // a fresh ticket received during this handshake
}

//go:linkname handshake crypto/tls.(*clientHandshakeState).handshake
func handshake(hs *clientHandshakeState) error

func (hs *clientHandshakeState) handshake() error {
	return handshake(hs)
}

// writeHandshakeRecord writes a handshake message to the connection and updates
// the record layer state. If transcript is non-nil the marshalled message is
// written to it.
func (c *_trsconn) writeHandshakeRecord(msg handshakeMessage, transcript transcriptHash, firstFragmentLen uint8) (int, error) {
	c.out.Lock()
	defer c.out.Unlock()

	data, err := msg.marshal()
	if err != nil {
		return 0, err
	}
	if transcript != nil {
		transcript.Write(data)
	}

	return c.writeRecordLocked(recordTypeHandshake, firstFragmentLen, data)
}

func (cout *Conn) clientHandshake(firstFragmentLen uint8) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		c := (*_trsconn)(unsafe.Pointer(cout))

		if c.config == nil {
			c.config = defaultConfig()
		}

		// This may be a renegotiation handshake, in which case some fields
		// need to be reset.
		c.didResume = false

		hello, ecdheKey, err := c.makeClientHello()
		if err != nil {
			return err
		}
		c.serverName = hello.serverName

		session, earlySecret, binderKey, err := c.loadSession(hello)
		if err != nil {
			return err
		}
		if session != nil {
			defer func() {
				// If we got a handshake failure when resuming a session, throw away
				// the session ticket. See RFC 5077, Section 3.2.
				//
				// RFC 8446 makes no mention of dropping tickets on failure, but it
				// does require servers to abort on invalid binders, so we need to
				// delete tickets to recover from a corrupted PSK.
				if err != nil {
					if cacheKey := c.clientSessionCacheKey(); cacheKey != "" {
						c.config.ClientSessionCache.Put(cacheKey, nil)
					}
				}
			}()
		}

		if _, err := c.writeHandshakeRecord(hello, nil, firstFragmentLen); err != nil {
			return err
		}

		if hello.earlyData {
			suite := cipherSuiteTLS13ByID(session.cipherSuite)
			transcript := suite.hash.New()
			if err := transcriptMsg(hello, transcript); err != nil {
				return err
			}
			earlyTrafficSecret := ClientEarlyTrafficSecret(earlySecret, transcript)
			quicSetWriteSecret(c, tls.QUICEncryptionLevelEarly, suite.id, earlyTrafficSecret)
		}

		// serverHelloMsg is not included in the transcript
		msg, err := c.readHandshake(nil)
		if err != nil {
			return err
		}

		var serverHello *serverHelloMsg
		if !isTypeEqual(msg, "*tls.serverHelloMsg") {
			c.sendAlert(alertUnexpectedMessage)
			return unexpectedMessageError(serverHello, msg)
		}
		serverHello = (*serverHelloMsg)(*(*unsafe.Pointer)(
			unsafe.Add(unsafe.Pointer(&msg), unsafe.Sizeof(uintptr(0))),
		))

		if err := c.pickTLSVersion(serverHello); err != nil {
			return err
		}

		// If we are negotiating a protocol version that's lower than what we
		// support, check for the server downgrade canaries.
		// See RFC 8446, Section 4.1.3.
		maxVers := maxSupportedVersion(c.config, roleClient)
		tls12Downgrade := string(serverHello.random[24:]) == downgradeCanaryTLS12
		tls11Downgrade := string(serverHello.random[24:]) == downgradeCanaryTLS11
		if maxVers == tls.VersionTLS13 && c.vers <= tls.VersionTLS12 && (tls12Downgrade || tls11Downgrade) ||
			maxVers == tls.VersionTLS12 && c.vers <= tls.VersionTLS11 && tls11Downgrade {
			c.sendAlert(alertIllegalParameter)
			return errors.New("tls: downgrade attempt detected, possibly due to a MitM attack or a broken middlebox")
		}

		if c.vers == tls.VersionTLS13 {
			hs := &clientHandshakeStateTLS13{
				c:           cout,
				ctx:         ctx,
				serverHello: serverHello,
				hello:       hello,
				ecdheKey:    ecdheKey,
				session:     session,
				earlySecret: earlySecret,
				binderKey:   binderKey,
			}

			// In TLS 1.3, session tickets are delivered after the handshake.
			return hs.handshake()
		}

		hs := &clientHandshakeState{
			c:           cout,
			ctx:         ctx,
			serverHello: serverHello,
			hello:       hello,
			session:     session,
		}

		if err := hs.handshake(); err != nil {
			return err
		}

		return nil
	}
}
