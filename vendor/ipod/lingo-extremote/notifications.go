package extremote

import (
	"encoding/binary"
	"github.com/oandrew/ipod"
	"github.com/oandrew/ipod/media"
)

func (n *PlayStatusChangeNotification) MarshalBinary() ([]byte, error) {
	return append([]byte{n.Status}, n.Data...), nil
}

// Notifications is session-local. A new identification clears subscriptions.
type Notifications struct {
	Mask     uint32
	Previous media.Snapshot
}

func (n *Notifications) Configure(payload interface{}, current media.Snapshot) {
	switch p := payload.(type) {
	case *SetPlayStatusChangeNotification:
		n.Mask = p.EventMask
		n.Previous = current
	case *SetPlayStatusChangeNotificationShort:
		n.Mask = 0
		if p.Enabled {
			n.Mask = ^uint32(0)
		}
		n.Previous = current
	}
}
func (n *Notifications) Poll(current media.Snapshot, out ipod.CommandWriter) {
	emit := func(status byte, value *uint32) {
		if n.Mask&(1<<status) == 0 {
			return
		}
		p := &PlayStatusChangeNotification{Status: status}
		if value != nil {
			p.Data = make([]byte, 4)
			binary.BigEndian.PutUint32(p.Data, *value)
		}
		ipod.Send(out, p)
	}
	if current.Identity() != n.Previous.Identity() || current.Connected != n.Previous.Connected {
		index := uint32(0)
		emit(1, &index)
	}
	if current.State() == 0 && n.Previous.State() != 0 {
		emit(0, nil)
	}
	if current.Position != n.Previous.Position || current.Status != n.Previous.Status {
		emit(4, &current.Position)
	}
	n.Previous = current
}
