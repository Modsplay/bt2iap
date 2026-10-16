package extremote

import (
	"github.com/oandrew/ipod"
	"github.com/oandrew/ipod/media"
)

type DeviceExtRemote interface {
	PlaybackStatus() (trackLength, trackPos uint32, state PlayerState)
}

func ackSuccess(req *ipod.Command) *ACK {
	return &ACK{Status: ACKStatusSuccess, CmdID: req.ID.CmdID()}
}

// func ackPending(req ipod.Packet, maxWait uint32) ACKPending {
// 	return ACKPending{Status: ACKStatusPending, CmdID: uint8(req.ID.CmdID()), MaxWait: maxWait}
// }

func HandleExtRemote(req *ipod.Command, tr ipod.CommandWriter, dev DeviceExtRemote) error {
	live := media.Current.Snapshot()
	switch msg := req.Payload.(type) {

	case *GetCurrentPlayingTrackChapterInfo:
		ipod.Respond(req, tr, &ReturnCurrentPlayingTrackChapterInfo{
			CurrentChapterIndex: 0,
			ChapterCount:        1,
		})
	case *SetCurrentPlayingTrackChapter:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetCurrentPlayingTrackChapterPlayStatus:
		ipod.Respond(req, tr, &ReturnCurrentPlayingTrackChapterPlayStatus{
			ChapterPosition: 0,
			ChapterLength:   0,
		})
	case *GetCurrentPlayingTrackChapterName:
		ipod.Respond(req, tr, &ReturnCurrentPlayingTrackChapterName{
			ChapterName: ipod.StringToBytes("chapter"),
		})
	case *GetAudiobookSpeed:
		ipod.Respond(req, tr, &ReturnAudiobookSpeed{
			Speed: 0,
		})
	case *SetAudiobookSpeed:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetIndexedPlayingTrackInfo:
		var info interface{}
		switch msg.InfoType {
		case TrackInfoCaps:
			info = &TrackCaps{
				Caps:         0x0,
				TrackLength:  live.Duration,
				ChapterCount: 1,
			}
		case TrackInfoDescription, TrackInfoLyrics:
			info = &TrackLongText{
				Flags:       0x0,
				PacketIndex: 0,
				Text:        0x00,
			}
		case TrackInfoArtworkCount:
			info = struct{}{}
		default:
			info = []byte{0x00}

		}
		ipod.Respond(req, tr, &ReturnIndexedPlayingTrackInfo{
			InfoType: msg.InfoType,
			Info:     info,
		})
	case *GetArtworkFormats:
		ipod.Respond(req, tr, &RetArtworkFormats{})
	case *GetTrackArtworkData:
		ipod.Respond(req, tr, &ACK{
			Status: ACKStatusFailed,
			CmdID:  req.ID.CmdID(),
		})
	case *ResetDBSelection:
		ipod.Respond(req, tr, ackSuccess(req))
	case *SelectDBRecord:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetNumberCategorizedDBRecords:
		ipod.Respond(req, tr, &ReturnNumberCategorizedDBRecords{
			RecordCount: 1,
		})
	case *RetrieveCategorizedDatabaseRecords:
		ipod.Respond(req, tr, &ReturnCategorizedDatabaseRecord{})
	case *GetPlayStatus:
		ipod.Respond(req, tr, &ReturnPlayStatus{
			TrackLength:   live.Duration,
			TrackPosition: live.Position,
			State:         PlayerState(live.State()),
		})
	case *GetCurrentPlayingTrackIndex:
		ipod.Respond(req, tr, &ReturnCurrentPlayingTrackIndex{
			TrackIndex: 0,
		})
	case *GetIndexedPlayingTrackTitle:
		ipod.Respond(req, tr, &ReturnIndexedPlayingTrackTitle{
			Title: ipod.StringToBytes(live.Title),
		})
	case *GetIndexedPlayingTrackArtistName:
		ipod.Respond(req, tr, &ReturnIndexedPlayingTrackArtistName{
			ArtistName: ipod.StringToBytes(live.Artist),
		})
	case *GetIndexedPlayingTrackAlbumName:
		ipod.Respond(req, tr, &ReturnIndexedPlayingTrackAlbumName{
			AlbumName: ipod.StringToBytes(live.Album),
		})
	case *SetPlayStatusChangeNotification:
		ipod.Respond(req, tr, ackSuccess(req))
	case *SetPlayStatusChangeNotificationShort:
		ipod.Respond(req, tr, ackSuccess(req))
	case *PlayCurrentSelection:
		respondControl(req, tr, "Play")
	case *PlayControl:
		actions := map[PlayControlCmd]string{PlayControlToggle: "Toggle", PlayControlStop: "Stop", PlayControlNextTrack: "Next", PlayControlPrevTrack: "Previous", PlayControlNext: "Next", PlayControlPrev: "Previous", PlayControlPlay: "Play", PlayControlPause: "Pause", PlayControlStartFF: "FastForward", PlayControlStartRew: "Rewind", PlayControlEndFFRew: "Play"}
		respondControl(req, tr, actions[msg.Cmd])
	case *GetTrackArtworkTimes:
		ipod.Respond(req, tr, &RetTrackArtworkTimes{})
	case *GetShuffle:
		ipod.Respond(req, tr, &ReturnShuffle{Mode: ShuffleMode(live.ShuffleMode())})
	case *SetShuffle:
		respondControl(req, tr, media.ShuffleAction(byte(msg.Mode)))

	case *GetRepeat:
		ipod.Respond(req, tr, &ReturnRepeat{Mode: RepeatMode(live.RepeatMode())})
	case *SetRepeat:
		respondControl(req, tr, media.RepeatAction(byte(msg.Mode)))

	case *SetDisplayImage:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetMonoDisplayImageLimits:
		ipod.Respond(req, tr, &ReturnMonoDisplayImageLimits{
			MaxWidth:    640,
			MaxHeight:   960,
			PixelFormat: 0x01,
		})
	case *GetNumPlayingTracks:
		ipod.Respond(req, tr, &ReturnNumPlayingTracks{
			NumTracks: live.Count(),
		})
	case *SetCurrentPlayingTrack:
	case *SelectSortDBRecord:
	case *GetColorDisplayImageLimits:
		ipod.Respond(req, tr, &ReturnColorDisplayImageLimits{
			MaxWidth:    640,
			MaxHeight:   960,
			PixelFormat: 0x01,
		})
	case *ResetDBSelectionHierarchy:
		ipod.Respond(req, tr, &ACK{Status: ACKStatusFailed, CmdID: req.ID.CmdID()})

	case *GetDBiTunesInfo:
	// RetDBiTunesInfo:
	case *GetUIDTrackInfo:
	// RetUIDTrackInfo:
	case *GetDBTrackInfo:
	// RetDBTrackInfo:
	case *GetPBTrackInfo:
	// RetPBTrackInfo:

	default:
		_ = msg
	}
	return nil
}

func respondControl(req *ipod.Command, tr ipod.CommandWriter, action string) {
	status := ACKStatusSuccess
	if media.Current.Control(action) != nil {
		status = ACKStatusFailed
	}
	ipod.Respond(req, tr, &ACK{Status: status, CmdID: req.ID.CmdID()})
}
