package media

import (
	"github.com/godbus/dbus/v5"
	"testing"
)

func TestSelectConnectedAddressedPlayer(t *testing.T) {
	device := dbus.ObjectPath("/device")
	player := dbus.ObjectPath("/device/player0")
	objects := Objects{device: {"org.bluez.Device1": {"Connected": dbus.MakeVariant(true)}, "org.bluez.MediaControl1": {"Player": dbus.MakeVariant(player)}}, player: {"org.bluez.MediaPlayer1": {"Device": dbus.MakeVariant(device), "Status": dbus.MakeVariant("playing"), "Position": dbus.MakeVariant(uint32(120)), "Track": dbus.MakeVariant(map[string]dbus.Variant{"Title": dbus.MakeVariant("Café 東京"), "Artist": dbus.MakeVariant("Artist"), "Duration": dbus.MakeVariant(uint32(100))})}}}
	path, s := selectPlayer(objects)
	if path != player || s.Title != "Café 東京" || s.Position != 100 || s.State() != 1 {
		t.Fatalf("bad snapshot: %+v", s)
	}
	objects[device]["org.bluez.Device1"]["Connected"] = dbus.MakeVariant(false)
	path, s = selectPlayer(objects)
	if path != "" || s.Connected || s.Title != "" {
		t.Fatal("stale data retained after disconnect")
	}
}
func TestSettingsAndStates(t *testing.T) {
	if ShuffleAction(7) != "" || RepeatAction(7) != "" {
		t.Fatal("invalid settings accepted")
	}
	if (Snapshot{Status: "paused"}).State() != 2 || (Snapshot{}).State() != 0 {
		t.Fatal("bad status")
	}
}
