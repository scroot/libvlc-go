package vlc

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L.
#cgo linux LDFLAGS: -lvlc
#cgo windows LDFLAGS: -lvlc.x64.dll
#include <vlc/vlc.h>
#include <stdlib.h>
extern void go_callback_samples(void *samples, int length);
static void Audio_Play_Cb(void *data, const void *samples, unsigned count, int64_t pts){
	go_callback_samples(samples, count);
}
static void setAudioCallback(libvlc_media_player_t *p_mi) {
	libvlc_audio_set_callbacks(p_mi, Audio_Play_Cb, NULL,NULL,NULL,NULL,0);
}
*/
import "C"
import (
	"errors"
	"unsafe"
	"math"
)

// Player is a media player used to play a single media file.
// For playing media lists (playlists) use ListPlayer instead.
type Player struct {
	player *C.libvlc_media_player_t
}

// NewPlayer creates an instance of a single-media player.
func NewPlayer() (*Player, error) {
	if inst == nil {
		return nil, errors.New("module must be initialized first")
	}

	if player := C.libvlc_media_player_new(inst.handle); player != nil {
		return &Player{player: player}, nil
	}

	return nil, getError()
}

//export go_callback_samples
func go_callback_samples(samples unsafe.Pointer, length C.int) {

	//Sum dbVolume
	size := int(length)
	buf := C.GoBytes(samples, length)
	var dbl,dbr int
	var vall, valr int16
	var suml, sumr float64
	for i:=0;i<size; i+=4 {
		a ,b ,c,d:= buf[i],buf[i+1],buf[i+2],buf[i+3]
		vall, valr = int16(a<<16+b), int16(c<<16+d)
		suml,sumr = suml+math.Abs( float64(vall)), sumr+math.Abs( float64( valr))
	}
	// Left
	suml = suml/float64( size/2)
	if suml > 0{
		dbl = int( 20.0*math.Log10(suml))
	}
	// Rigth
	sumr = sumr/float64(size/2)
	if sumr > 0{
		dbr = int( 20*math.Log10(sumr))
	}


	data:=DbVolume{dbl,dbr}
	select {
	case dbChan <- data:
	default: //drop
	}
}


// Release destroys the media player instance.
func (p *Player) Release() error {
	if p.player == nil {
		return nil
	}

	C.libvlc_media_player_release(p.player)
	p.player = nil

	return getError()
}

func (p *Player) EnabledAudioPlayCallback() error {
	if p.player == nil {
		return nil
	}
	C.setAudioCallback(p.player)
	return getError()
}

// Play plays the current media.
func (p *Player) Play() error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}
	if p.IsPlaying() {
		return nil
	}

	if C.libvlc_media_player_play(p.player) < 0 {
		return getError()
	}

	return nil
}

// IsPlaying returns a boolean value specifying if the player is currently
// playing.
func (p *Player) IsPlaying() bool {
	if p.player == nil {
		return false
	}

	return C.libvlc_media_player_is_playing(p.player) != 0
}

// Stop cancels the currently playing media, if there is one.
func (p *Player) Stop() error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_stop(p.player)
	return getError()
}

// SetPause sets the pause state of the media player.
// Pass in true to pause the current media, or false to resume it.
func (p *Player) SetPause(pause bool) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_set_pause(p.player, C.int(boolToInt(pause)))
	return getError()
}

// TogglePause pauses/resumes the player.
// Calling this method has no effect if there is no media.
func (p *Player) TogglePause() error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_pause(p.player)
	return getError()
}

// SetFullScreen sets the fullscreen state of the media player.
// Pass in true to enable fullscreen, or false to disable it.
func (p *Player) SetFullScreen(fullscreen bool) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_set_fullscreen(p.player, C.int(boolToInt(fullscreen)))
	return getError()
}

// ToggleFullScreen toggles the fullscreen status of the player,
// on non-embedded video outputs.
func (p *Player) ToggleFullScreen() error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_toggle_fullscreen(p.player)
	return getError()
}

// IsFullScreen gets the fullscreen status of the current player.
func (p *Player) IsFullScreen() (bool, error) {
	if p.player == nil {
		return false, errors.New("player must be initialized first")
	}

	return (C.libvlc_get_fullscreen(p.player) != C.int(0)), getError()
}

// Volume returns the volume of the player.
func (p *Player) Volume() (int, error) {
	if p.player == nil {
		return 0, errors.New("player must be initialized first")
	}

	return int(C.libvlc_audio_get_volume(p.player)), getError()
}

// SetVolume sets the volume of the player.
func (p *Player) SetVolume(volume int) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_audio_set_volume(p.player, C.int(volume))
	return getError()
}

// Media returns the current media of the player, if one exists.
func (p *Player) Media() (*Media, error) {
	if p.player == nil {
		return nil, errors.New("player must be initialized first")
	}

	media := C.libvlc_media_player_get_media(p.player)
	if media == nil {
		return nil, nil
	}

	// This call will not release the media. Instead, it will decrement
	// the reference count increased by libvlc_media_player_get_media.
	C.libvlc_media_release(media)

	return &Media{media}, nil
}

// SetMedia sets the provided media as the current media of the player.
func (p *Player) SetMedia(m *Media) error {
	return p.setMedia(m)
}

// LoadMediaFromPath loads the media from the specified path and adds it as the
// current media of the player.
func (p *Player) LoadMediaFromPath(path string) (*Media, error) {
	return p.loadMedia(path, true)
}

// LoadMediaFromURL loads the media from the specified URL and adds it as the
// current media of the player.
func (p *Player) LoadMediaFromURL(url string) (*Media, error) {
	return p.loadMedia(url, false)
}

// SetAudioOutput selects an audio output module.
// Any change will take be effect only after playback is stopped and restarted.
// Audio output cannot be changed while playing.
func (p *Player) SetAudioOutput(output string) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	cOutput := C.CString(output)
	defer C.free(unsafe.Pointer(cOutput))

	if C.libvlc_audio_output_set(p.player, cOutput) != 0 {
		return getError()
	}

	return nil
}

// MediaLength returns media length in milliseconds.
func (p *Player) MediaLength() (int, error) {
	if p.player == nil {
		return 0, errors.New("player must be initialized first")
	}

	return int(C.libvlc_media_player_get_length(p.player)), getError()
}

// MediaState returns the state of the current media.
func (p *Player) MediaState() (MediaState, error) {
	if p.player == nil {
		return 0, errors.New("player must be initialized first")
	}

	state := int(C.libvlc_media_player_get_state(p.player))
	return MediaState(state), getError()
}

// MediaPosition returns media position as a
// float percentage between 0.0 and 1.0.
func (p *Player) MediaPosition() (float32, error) {
	if p.player == nil {
		return 0, errors.New("player must be initialized first")
	}

	return float32(C.libvlc_media_player_get_position(p.player)), getError()
}

// SetMediaPosition sets media position as percentage between 0.0 and 1.0.
// Some formats and protocols do not support this.
func (p *Player) SetMediaPosition(pos float32) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_set_position(p.player, C.float(pos))
	return getError()
}

// MediaTime returns media time in milliseconds.
func (p *Player) MediaTime() (int, error) {
	if p.player == nil {
		return 0, errors.New("player must be initialized first")
	}

	return int(C.libvlc_media_player_get_time(p.player)), getError()
}

// SetMediaTime sets the media time in milliseconds.
// Some formats and protocals do not support this.
func (p *Player) SetMediaTime(t int) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_set_time(p.player, C.libvlc_time_t(int64(t)))
	return getError()
}

// WillPlay returns true if the current media is not in a
// finished or error state.
func (p *Player) WillPlay() bool {
	if p.player == nil {
		return false
	}

	return C.libvlc_media_player_will_play(p.player) != 0
}

// SetXWindow sets the X window to play on.
func (p *Player) SetXWindow(windowID uint32) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_set_xwindow(p.player, C.uint(windowID))
	return getError()
}

// SetWindowHwnd sets the Windows Hwnd to play on
func (p *Player) SetWindowHwnd(hwnd uintptr) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}

	C.libvlc_media_player_set_hwnd(p.player, unsafe.Pointer(hwnd))
	return getError()
}

// EventManager returns the event manager responsible for the media list.
func (p *Player) EventManager() (*EventManager, error) {
	if p.player == nil {
		return nil, errors.New("player must be initialized first")
	}

	manager := C.libvlc_media_player_event_manager(p.player)
	if manager == nil {
		return nil, errors.New("could not retrieve player event manager")
	}

	return newEventManager(manager), nil
}

func (p *Player) loadMedia(path string, local bool) (*Media, error) {
	m, err := newMedia(path, local)
	if err != nil {
		return nil, err
	}

	if err = p.setMedia(m); err != nil {
		m.Release()
		return nil, err
	}

	return m, nil
}

func (p *Player) setMedia(m *Media) error {
	if p.player == nil {
		return errors.New("player must be initialized first")
	}
	if m.media == nil {
		return errors.New("media must be initialized first")
	}

	C.libvlc_media_player_set_media(p.player, m.media)
	return getError()
}

