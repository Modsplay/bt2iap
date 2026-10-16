package extremote

import (
	"bytes"
	"errors"
	"github.com/oandrew/ipod"
	"github.com/oandrew/ipod/media"
	"testing"
)

type fakeMedia struct {
	s       media.Snapshot
	actions []string
	err     error
}

func (f *fakeMedia) Snapshot() media.Snapshot { return f.s }
func (f *fakeMedia) Control(a string) error   { f.actions = append(f.actions, a); return f.err }
func query(t *testing.T, p interface{}) interface{} {
	t.Helper()
	cmd, e := ipod.BuildCommand(p)
	if e != nil {
		t.Fatal(e)
	}
	out := &ipod.CmdBuffer{}
	if e = HandleExtRemote(cmd, out, nil); e != nil {
		t.Fatal(e)
	}
	if len(out.Commands) != 1 {
		t.Fatalf("responses %d", len(out.Commands))
	}
	return out.Commands[0].Payload
}
func TestLiveMetadataAndControls(t *testing.T) {
	old := media.Current
	defer func() { media.Current = old }()
	f := &fakeMedia{s: media.Snapshot{Connected: true, Title: "Heart For Brains", Artist: "Roar", Album: "EP", Duration: 169000, Position: 1260, Status: "paused"}}
	media.Current = f
	if !bytes.Equal(query(t, &GetIndexedPlayingTrackTitle{}).(*ReturnIndexedPlayingTrackTitle).Title, ipod.StringToBytes(f.s.Title)) {
		t.Fatal("title not live")
	}
	if query(t, &GetPlayStatus{}).(*ReturnPlayStatus).TrackPosition != 1260 {
		t.Fatal("position not live")
	}
	for _, x := range []struct {
		cmd  PlayControlCmd
		want string
	}{{PlayControlNextTrack, "Next"}, {PlayControlPrevTrack, "Previous"}, {PlayControlToggle, "Toggle"}, {PlayControlPause, "Pause"}, {PlayControlPlay, "Play"}, {PlayControlStop, "Stop"}} {
		if query(t, &PlayControl{Cmd: x.cmd}).(*ACK).Status != ACKStatusSuccess || f.actions[len(f.actions)-1] != x.want {
			t.Fatal("bad control routing")
		}
	}
	f.err = errors.New("disconnected")
	if query(t, &PlayControl{Cmd: PlayControlNextTrack}).(*ACK).Status != ACKStatusFailed {
		t.Fatal("false success on control failure")
	}
	f.s = media.Snapshot{}
	if query(t, &GetNumPlayingTracks{}).(*ReturnNumPlayingTracks).NumTracks != 0 {
		t.Fatal("stale track count")
	}
}
func TestNotificationWireFormatAndSubscription(t *testing.T) {
	n := Notifications{}
	out := &ipod.CmdBuffer{}
	s := media.Snapshot{Connected: true, Title: "new", Position: 1000, Status: "playing"}
	n.Poll(s, out)
	if len(out.Commands) != 0 {
		t.Fatal("unsolicited notification")
	}
	n.Configure(&SetPlayStatusChangeNotificationShort{Enabled: true}, media.Snapshot{})
	n.Poll(s, out)
	if len(out.Commands) != 2 {
		t.Fatalf("want track and position, got %d", len(out.Commands))
	}
	raw, err := (&ipod.CommandSerde{}).MarshalCmd(out.Commands[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, []byte{4, 0, 0x27, 1, 0, 0, 0, 0}) {
		t.Fatalf("bad wire packet %x", raw)
	}
	n.Configure(&SetPlayStatusChangeNotificationShort{Enabled: false}, s)
	out = &ipod.CmdBuffer{}
	n.Poll(media.Snapshot{}, out)
	if len(out.Commands) != 0 {
		t.Fatal("disable ignored")
	}
}
