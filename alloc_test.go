package pfilter

import (
	"net"
	"testing"
	"time"
)

// Returning a buffer to the pool must not allocate.
func TestBufPoolPutDoesNotAllocate(t *testing.T) {
	p := newBufPool(1500)
	b := p.get()
	if n := testing.AllocsPerRun(1000, func() { p.put(b); b = p.get() }); n != 0 {
		t.Fatalf("%v allocations per put and get", n)
	}
}

func BenchmarkPushedPacket(b *testing.B) {
	sock, _ := net.ListenPacket("udp", "127.0.0.1:0")
	f := NewPacketFilter(sock)
	c := f.NewConn(1, nil)
	got := make(chan struct{}, 1)
	c.(interface {
		SetPacketReceiver(func([]byte, net.Addr, error))
	}).SetPacketReceiver(func([]byte, net.Addr, error) {
		select {
		case got <- struct{}{}:
		default:
		}
	})
	f.Start()
	peer, _ := net.ListenPacket("udp", "127.0.0.1:0")
	defer peer.Close()
	defer sock.Close()
	msg := []byte("x")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		peer.WriteTo(msg, sock.LocalAddr())
		select {
		case <-got:
		case <-time.After(time.Second):
			b.Fatal("no packet")
		}
	}
}
