//go:build wasip1

package wasip1_test

import (
	"net"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/musaprg/dispatchrunnet/wasip1"
	"golang.org/x/net/nettest"
)

func TestConn(t *testing.T) {
	tests := []struct {
		network string
		address string
	}{
		{
			network: "tcp",
			address: ":0",
		},

		{
			network: "tcp4",
			address: ":0",
		},

		{
			network: "tcp6",
			address: ":0",
		},

		{
			network: "unix",
			address: "wasip1.sock",
		},
	}

	for _, test := range tests {
		t.Run(test.network, func(t *testing.T) {
			nettest.TestConn(t, func() (c1, c2 net.Conn, stop func(), err error) {
				network := test.network
				address := test.address
				switch network {
				case "unix":
					address = filepath.Join(t.TempDir(), address)
				}

				l, err := wasip1.Listen(network, address)
				if err != nil {
					return nil, nil, nil, err
				}
				defer l.Close()

				conns := make(chan net.Conn, 1)
				errch := make(chan error, 1)
				go func() {
					c, err := l.Accept()
					if err != nil {
						errch <- err
					} else {
						conns <- c
					}
				}()

				dialer := &wasip1.Dialer{}
				dialer.Deadline, _ = t.Deadline()

				laddr := l.Addr()
				c1, err = dialer.Dial(laddr.Network(), laddr.String())
				if err != nil {
					return nil, nil, nil, err
				}

				select {
				case c2 := <-conns:
					return c1, c2, func() { c1.Close(); c2.Close() }, nil
				case err := <-errch:
					c1.Close()
					return nil, nil, nil, err
				}
			})
		})
	}
}

func TestPacketConn(t *testing.T) {
	// Note: this is not as thorough of a test as TestConn because UDP is lossy
	// and building a net.Conn on top of a net.PacketConn causes tests to fail
	// due to packet losses.
	tests := []struct {
		network string
		address string
	}{
		{
			network: "udp",
			address: "127.0.0.1:0",
		},

		{
			network: "udp4",
			address: "127.0.0.1:0",
		},

		{
			network: "udp6",
			address: "[::1]:0",
		},

		{
			network: "unixgram",
			address: filepath.Join(t.TempDir(), "wasip1.sock"),
		},
	}

	for _, test := range tests {
		t.Run(test.network, func(t *testing.T) {
			network := test.network
			address := test.address

			c1, err := wasip1.ListenPacket(network, address)
			if err != nil {
				t.Fatal(err)
			}
			defer c1.Close()
			addr := c1.LocalAddr()

			c, err := wasip1.Dial(addr.Network(), addr.String())
			if err != nil {
				t.Fatal(err)
			}
			c2 := c.(net.PacketConn)
			defer c2.Close()

			rb2 := make([]byte, 128)
			wb := []byte("PACKETCONN TEST")

			// On unix the domain, the local address of connected sockets is
			// empty which makes it impossible for the listening connection
			// to send it a packet.
			if network != "unixgram" {
				if n, err := c1.WriteTo(wb, c2.LocalAddr()); err != nil {
					t.Fatal(err)
				} else if n != len(wb) {
					t.Fatalf("write with wrong number of bytes: want=%d got=%d", len(wb), n)
				}

				if n, addr, err := c2.ReadFrom(rb2); err != nil {
					t.Fatal(err)
				} else if n != len(wb) {
					t.Fatalf("read with wrong number of bytes: want=%d got=%d", len(wb), n)
				} else if !reflect.DeepEqual(addr, c1.LocalAddr()) {
					t.Fatalf("read from wrong address: want=%s got=%s", c1.LocalAddr(), addr)
				}
			}

			if n, err := c.Write(wb); err != nil {
				t.Fatal(err)
			} else if n != len(wb) {
				t.Fatalf("write with wrong number of bytes: want=%d got=%d", len(wb), n)
			}

			rb1 := make([]byte, 128)
			if n, addr, err := c1.ReadFrom(rb1); err != nil {
				t.Fatal(err)
			} else if n != len(wb) {
				t.Fatalf("read with wrong number of bytes: want=%d got=%d", len(wb), n)
			} else if !reflect.DeepEqual(addr, c2.LocalAddr()) {
				t.Fatalf("read from wrong address: want=%s got=%s", c2.LocalAddr(), addr)
			}
		})
	}
}
