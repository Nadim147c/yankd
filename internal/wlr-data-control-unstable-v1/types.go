package wlr

import "github.com/neurlang/wayland/wl"

type (
	BaseProxy = wl.BaseProxy
	Context   = wl.Context
	Event     = wl.Event
	Seat      = wl.Seat
	Surface   = wl.Surface
	Keyboard  = wl.Keyboard
	Output    = wl.Output
)

var NewKeyboard = wl.NewKeyboard

func SafeCast[T any](p wl.Proxy) T {
	return wl.SafeCast[T](p)
}
