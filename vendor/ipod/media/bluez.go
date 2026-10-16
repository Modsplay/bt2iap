// Package media connects iAP to the currently addressed Bluetooth player.
package media

import (
	"context"
	"errors"
	"github.com/godbus/dbus/v5"
	"strings"
	"sync"
	"time"
)

type Snapshot struct {
	Connected                           bool
	Title, Artist, Album, Genre, Status string
	Shuffle, Repeat                     string
	Duration, Position                  uint32
}

func (s Snapshot) State() byte {
	switch s.Status {
	case "playing", "forward-seek", "reverse-seek":
		return 1
	case "paused":
		return 2
	default:
		return 0
	}
}
func (s Snapshot) Count() uint32 {
	if s.Connected {
		return 1
	}
	return 0
}
func (s Snapshot) Identity() string { return s.Title + "\x00" + s.Artist + "\x00" + s.Album }

// A single virtual track represents the phone's current item; no library is fabricated.
type Backend interface {
	Snapshot() Snapshot
	Control(string) error
}

var Current Backend = &BlueZ{}

type BlueZ struct {
	mu   sync.Mutex
	conn *dbus.Conn
}
type Objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant

func text(m map[string]dbus.Variant, k string) string {
	v, _ := m[k].Value().(string)
	return strings.ReplaceAll(v, "\x00", "")
}
func number(m map[string]dbus.Variant, k string) uint32 {
	v, _ := m[k].Value().(uint32)
	if v == 0xffffffff {
		return 0
	}
	return v
}
func selectPlayer(objects Objects) (dbus.ObjectPath, Snapshot) {
	var best dbus.ObjectPath
	var result Snapshot
	bestScore := -1
	for path, ifs := range objects {
		p, ok := ifs["org.bluez.MediaPlayer1"]
		if !ok {
			continue
		}
		device, _ := p["Device"].Value().(dbus.ObjectPath)
		d := objects[device]
		connected, _ := d["org.bluez.Device1"]["Connected"].Value().(bool)
		if !connected {
			continue
		}
		addressed, _ := d["org.bluez.MediaControl1"]["Player"].Value().(dbus.ObjectPath)
		if addressed != "" && addressed != path {
			continue
		}
		score := 0
		if addressed == path {
			score++
		}
		if text(p, "Status") == "playing" {
			score += 2
		}
		if score < bestScore || (score == bestScore && string(path) > string(best)) {
			continue
		}
		track, _ := p["Track"].Value().(map[string]dbus.Variant)
		result = Snapshot{Connected: true, Title: text(track, "Title"), Artist: text(track, "Artist"), Album: text(track, "Album"), Genre: text(track, "Genre"), Status: text(p, "Status"), Shuffle: text(p, "Shuffle"), Repeat: text(p, "Repeat"), Duration: number(track, "Duration"), Position: number(p, "Position")}
		if result.Duration > 0 && result.Position > result.Duration {
			result.Position = result.Duration
		}
		best, bestScore = path, score
	}
	return best, result
}
func (b *BlueZ) objects(ctx context.Context) (Objects, error) {
	if b.conn == nil || !b.conn.Connected() {
		c, err := dbus.ConnectSystemBus()
		if err != nil {
			return nil, err
		}
		b.conn = c
	}
	var objects Objects
	err := b.conn.Object("org.bluez", "/").CallWithContext(ctx, "org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects)
	return objects, err
}
func (b *BlueZ) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	objects, err := b.objects(ctx)
	if err != nil {
		return Snapshot{}
	}
	_, s := selectPlayer(objects)
	return s
}
func (b *BlueZ) Control(action string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	objects, err := b.objects(ctx)
	if err != nil {
		return err
	}
	path, s := selectPlayer(objects)
	if path == "" {
		return errors.New("no connected Bluetooth player")
	}
	if action == "Toggle" {
		if s.State() == 1 {
			action = "Pause"
		} else {
			action = "Play"
		}
	}
	if parts := strings.SplitN(action, ":", 2); len(parts) == 2 {
		valid := false
		if parts[0] == "Shuffle" {
			valid = parts[1] == "off" || parts[1] == "alltracks" || parts[1] == "group"
		}
		if parts[0] == "Repeat" {
			valid = parts[1] == "off" || parts[1] == "singletrack" || parts[1] == "alltracks"
		}
		if !valid {
			return errors.New("unsupported player setting")
		}
		return b.conn.Object("org.bluez", path).CallWithContext(ctx, "org.freedesktop.DBus.Properties.Set", 0, "org.bluez.MediaPlayer1", parts[0], dbus.MakeVariant(parts[1])).Err
	}
	switch action {
	case "Play", "Pause", "Stop", "Next", "Previous", "FastForward", "Rewind":
	default:
		return errors.New("unsupported player control")
	}
	return b.conn.Object("org.bluez", path).CallWithContext(ctx, "org.bluez.MediaPlayer1."+action, 0).Err
}

func (s Snapshot) ShuffleMode() byte {
	switch s.Shuffle {
	case "alltracks":
		return 1
	case "group":
		return 2
	}
	return 0
}
func (s Snapshot) RepeatMode() byte {
	switch s.Repeat {
	case "singletrack":
		return 1
	case "alltracks", "group":
		return 2
	}
	return 0
}
func ShuffleAction(mode byte) string {
	switch mode {
	case 0:
		return "Shuffle:off"
	case 1:
		return "Shuffle:alltracks"
	case 2:
		return "Shuffle:group"
	}
	return ""
}
func RepeatAction(mode byte) string {
	switch mode {
	case 0:
		return "Repeat:off"
	case 1:
		return "Repeat:singletrack"
	case 2:
		return "Repeat:alltracks"
	}
	return ""
}
