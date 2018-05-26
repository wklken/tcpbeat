package tcp

import (
	"bufio"
	"bytes"
	"net"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/pkg/errors"
)

type CallbackFunc = func(data []byte)

// Client is a remote client.
type Client struct {
	conn           net.Conn
	log            *logp.Logger
	callback       CallbackFunc
	done           chan struct{}
	splitFunc      bufio.SplitFunc
	maxMessageSize uint64
	timeout        time.Duration
}

func NewClient(
	conn net.Conn,
	log *logp.Logger,
	callback CallbackFunc,
	splitFunc bufio.SplitFunc,
	maxReadMessage uint64,
	timeout time.Duration,
) *Client {
	client := &Client{
		conn:           conn,
		log:            log.With("remote_address", conn.RemoteAddr()),
		callback:       callback,
		done:           make(chan struct{}),
		splitFunc:      splitFunc,
		maxMessageSize: maxReadMessage,
		timeout:        timeout,
		//metadata: inputsource.NetworkMetadata{
		//	RemoteAddr: conn.RemoteAddr(),
		//	TLS:        extractSSLInformation(conn),
		//},
	}
	//extractSSLInformation(conn)
	return client
}

func (c *Client) Handle() error {
	r := NewResetableLimitedReader(NewDeadlineReader(c.conn, c.timeout), c.maxMessageSize)
	buf := bufio.NewReader(r)
	scanner := bufio.NewScanner(buf)
	scanner.Split(c.splitFunc)

	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			// we are forcing a close on the socket, lets ignore any error that could happen.
			select {
			case <-c.done:
				break
			default:
			}
			// This is a user defined limit and we should notify the user.
			if IsMaxReadBufferErr(err) {
				c.log.Errorw("client error", "error", err)
			}
			return errors.Wrap(err, "tcp client error")
		}
		r.Reset()
		//c.callback(scanner.Bytes(), c.metadata)
		c.callback(scanner.Bytes())
	}

	// We are out of the scanner, either we reached EOF or another fatal error occured.
	// like we failed to complete the TLS handshake or we are missing the client certificate when
	// mutual auth is on, which is the default.
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Close() {
	close(c.done)
	c.conn.Close()
}

//func extractSSLInformation(c net.Conn) *inputsource.TLSMetadata {
//	if tls, ok := c.(*tls.Conn); ok {
//		state := tls.ConnectionState()
//		return &inputsource.TLSMetadata{
//			TLSVersion:       tlscommon.ResolveTLSVersion(state.Version),
//			CipherSuite:      tlscommon.ResolveCipherSuite(state.CipherSuite),
//			ServerName:       state.ServerName,
//			PeerCertificates: extractCertificate(state.PeerCertificates),
//		}
//	}
//	return nil
//}
//
//func extractCertificate(certificates []*x509.Certificate) []string {
//	strCertificate := make([]string, len(certificates))
//	for idx, c := range certificates {
//		// Ignore errors here, problematics cert have failed
//		//the handshake at this point.
//		b, _ := x509.MarshalPKIXPublicKey(c.PublicKey)
//		strCertificate[idx] = string(b)
//	}
//	return strCertificate
//}

func SplitFunc(lineDelimiter []byte) bufio.SplitFunc {
	ld := []byte(lineDelimiter)
	if bytes.Equal(ld, []byte("\n")) {
		// This will work for most usecases and will also strip \r if present.
		// CustomDelimiter, need to match completely and the delimiter will be completely removed from
		// the returned byte slice
		return bufio.ScanLines
	}
	return factoryDelimiter(ld)
}
