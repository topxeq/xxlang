// Unit tests for ftp_client.go using a lightweight fake connection.
package objects

import (
	"bytes"
	"net"
	"net/textproto"
	"testing"
	"time"
)

// fakeConn is a minimal net.Conn implementation used to test FtpClient logic
// without a real FTP server. It records written commands and serves predefined
// responses to drives ReadResponse calls from textproto.Conn.
type fakeConn struct {
	writes []string
	reads  *bytes.Buffer
	closed bool
}

func newFakeConn() *fakeConn {
	return &fakeConn{reads: &bytes.Buffer{}}
}

func (f *fakeConn) Read(p []byte) (n int, err error) {
	return f.reads.Read(p)
}

func (f *fakeConn) Write(p []byte) (n int, err error) {
	f.writes = append(f.writes, string(p))
	return len(p), nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func (f *fakeConn) LocalAddr() net.Addr                { return nil }
func (f *fakeConn) RemoteAddr() net.Addr               { return nil }
func (f *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (f *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

func (f *fakeConn) PushResponse(s string) {
	// Ensure responses end with CRLF as FTP replies usually do.
	if len(s) == 0 {
		return
	}
	if s[len(s)-2:] != "\r\n" {
		s = s + "\r\n"
	}
	f.reads.WriteString(s)
}

// Test parsePasvResponse success case
func TestParsePasvResponse_Success(t *testing.T) {
	host, port, err := parsePasvResponse("227 Entering Passive Mode (192,168,1,2,195,15)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != "192.168.1.2" {
		t.Fatalf("unexpected host: %s", host)
	}
	if port != (195<<8 | 15) {
		t.Fatalf("unexpected port: %d", port)
	}
}

func TestParsePasvResponse_Invalid(t *testing.T) {
	if _, _, err := parsePasvResponse("invalid"); err == nil {
		t.Fatalf("expected error for invalid PASV response")
	}
}

func TestParseListLine_FileAndDir(t *testing.T) {
	fi := parseListLine("-rw-r--r--  1 user group 1234 Jan 01 00:00 file.txt")
	if fi.Name != "file.txt" || fi.IsDir {
		t.Fatalf("unexpected file info: %+v", fi)
	}
	if fi.Size != 1234 {
		t.Fatalf("unexpected size: %d", fi.Size)
	}

	di := parseListLine("drwxr-xr-x  2 user group 0 Jan 01 00:00 dir")
	if !di.IsDir || di.Name != "dir" {
		t.Fatalf("unexpected dir info: %+v", di)
	}
}

func TestCurrentDir_Success(t *testing.T) {
	fc := newFakeConn()
	// expect: PWD -> 257 "/home/user"
	fc.PushResponse("257 \"/home/user\"")

	c := &FtpClient{connected: true}
	c.conn = fc
	c.text = textproto.NewConn(fc)

	dir, err := c.CurrentDir()
	if err != nil {
		t.Fatalf("CurrentDir error: %v", err)
	}
	if dir != "/home/user" {
		t.Fatalf("unexpected dir: %s", dir)
	}
}

func TestCurrentDir_NotConnected(t *testing.T) {
	c := &FtpClient{connected: false}
	if _, err := c.CurrentDir(); err == nil {
		t.Fatalf("expected error when not connected")
	}
}

func TestExists_IsDir_IsFile_Scenarios(t *testing.T) {
	// Exists: SIZE returns 213 -> true
	fc1 := newFakeConn()
	fc1.PushResponse("213 1234")
	c1 := &FtpClient{connected: true, timeout: 1 * time.Second}
	c1.conn = fc1
	c1.text = textproto.NewConn(fc1)
	if !c1.Exists("path/file.txt") {
		t.Fatalf("Exists should be true when SIZE returns 213")
	}

	// IsDir: CWD returns 250 -> true, then CDUP returns 200
	fc2 := newFakeConn()
	fc2.PushResponse("250 directory ok")
	fc2.PushResponse("200 ok")
	c2 := &FtpClient{connected: true}
	c2.conn = fc2
	c2.text = textproto.NewConn(fc2)
	if !c2.IsDir("remote/dir") {
		t.Fatalf("IsDir should be true when CWD returns 250")
	}

	// IsFile: SIZE returns 213
	fc3 := newFakeConn()
	fc3.PushResponse("213 4567")
	c3 := &FtpClient{connected: true}
	c3.conn = fc3
	c3.text = textproto.NewConn(fc3)
	if !c3.IsFile("remote/file.txt") {
		t.Fatalf("IsFile should be true when SIZE returns 213")
	}
}

func TestSetType_NotConnected(t *testing.T) {
	c := NewFtpClient()
	if err := c.SetType("binary"); err == nil {
		t.Fatalf("expected error when not connected")
	}
}

func TestMkdirAll_Succeeds(t *testing.T) {
	fc := newFakeConn()
	// Three components -> three MKD attempts
	fc.PushResponse("200 OK")
	fc.PushResponse("200 OK")
	fc.PushResponse("200 OK")
	c := &FtpClient{connected: true}
	c.conn = fc
	c.text = textproto.NewConn(fc)
	if err := c.MkdirAll("a/b/c"); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
}

func TestClose_NotConnected(t *testing.T) {
	c := &FtpClient{connected: false}
	if err := c.Close(); err != nil {
		t.Fatalf("Close should not error when not connected: %v", err)
	}
}

func TestClose_Connected(t *testing.T) {
	fc := newFakeConn()
	c := &FtpClient{connected: true}
	c.conn = fc
	c.text = textproto.NewConn(fc)
	if err := c.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
