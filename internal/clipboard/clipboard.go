package clipboard

import (
	"log/slog"
	"sync/atomic"

	"github.com/Nadim147c/yankd/internal/models"
	"github.com/neurlang/wayland/wl"
)

// InterfaceName is the wlr-data-control-unstable-v1 interface name.
const InterfaceName = "zwlr_data_control_manager_v1"

// Client is wayland that handle wayland clipboard protocol.
type Client struct {
	display  *wl.Display
	registry *wl.Registry
	// manager       *protocol.ZwlrDataControlManagerV1
	clips         chan<- models.Event
	seatGlobals   map[uint32]uint32
	deviceName    uint32
	deviceVersion uint32
	closed        atomic.Bool
}

// NewClient creates a new wayland client.
func NewClient() *Client {
	c := new(Client)
	c.seatGlobals = make(map[uint32]uint32)
	slog.Debug("clipboard client created")
	return c
}

// Close closes the underlying socket connection.
func (h *Client) Close() error {
	h.closed.Store(true)
	return h.display.Context().Close()
}
