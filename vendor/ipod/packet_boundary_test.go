package ipod

import (
	"bytes"
	"testing"
)

func TestMetadataPacketBoundary(t *testing.T) {
	for _, n := range []int{255, 256, 257} {
		p := bytes.Repeat([]byte{0x41}, n)
		w := NewPacketWriter()
		w.WritePacket(p)
		got, e := NewPacketReader(w.Bytes()).ReadPacket()
		if e != nil || !bytes.Equal(p, got) {
			t.Fatalf("packet length %d: %v", n, e)
		}
	}
}
