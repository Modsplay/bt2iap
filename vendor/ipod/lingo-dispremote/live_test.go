package dispremote

import (
	"bytes"
	"github.com/oandrew/ipod"
	"github.com/oandrew/ipod/media"
	"testing"
)

type fakeMedia struct{}

func (fakeMedia) Snapshot() media.Snapshot {
	return media.Snapshot{Connected: true, Title: "Track", Artist: "Artist", Album: "Album", Genre: "Pop", Status: "playing", Duration: 120000, Position: 5000}
}
func (fakeMedia) Control(string) error { return nil }
func TestDisplayMetadata(t *testing.T) {
	old := media.Current
	defer func() { media.Current = old }()
	media.Current = fakeMedia{}
	for _, typ := range []TrackInfoType{TrackInfoTypeArtist, TrackInfoTypeAlbum, TrackInfoTypeGenre, TrackInfoTypeTrack} {
		req, _ := ipod.BuildCommand(&GetIndexedPlayingTrackInfo{InfoType: typ})
		out := &ipod.CmdBuffer{}
		if err := HandleDispRemote(req, out, nil); err != nil {
			t.Fatal(err)
		}
		data, err := (&ipod.CommandSerde{}).MarshalCmd(out.Commands[0])
		if err != nil {
			t.Fatal(err)
		}
		expected := map[TrackInfoType]string{TrackInfoTypeArtist: "Artist", TrackInfoTypeAlbum: "Album", TrackInfoTypeGenre: "Pop", TrackInfoTypeTrack: "Track"}[typ]
		if !bytes.Contains(data, []byte(expected)) {
			t.Fatalf("missing metadata %q in %x", expected, data)
		}
	}
}
