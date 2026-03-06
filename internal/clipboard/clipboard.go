package clipboard

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/Nadim147c/yankd/internal/models"
	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
	"github.com/neurlang/wayland/wl"
)

// InterfaceName is the wlr-data-control-unstable-v1 interface name.
const InterfaceName = "zwlr_data_control_manager_v1"

// Client is wayland that handle wayland clipboard protocol.
type Client struct {
	mu       sync.Mutex
	display  *wl.Display
	registry *wl.Registry

	// set types
	mimes map[string][]byte
	event *models.ClipboardEvent

	manager *protocol.ZwlrDataControlManagerV1
	device  *protocol.ZwlrDataControlDeviceV1

	eventChan     chan<- models.ClipboardEvent
	seatGlobals   map[uint32]uint32
	deviceName    uint32
	deviceVersion uint32

	connected atomic.Bool
	closed    atomic.Bool
	paused    atomic.Bool
}

// NewClient creates a new wayland client.
func NewClient() *Client {
	c := new(Client)
	c.seatGlobals = make(map[uint32]uint32)
	slog.Debug("clipboard client created")
	return c
}

func (c *Client) SetPaused(b bool) bool {
	if b {
		slog.Info("Saving history has been paused")
	} else {
		slog.Info("Saving history has been resumed")
	}
	c.paused.Store(b)
	return b
}

func (c *Client) TogglePaused() bool {
	currenState := c.paused.Load()
	return c.SetPaused(!currenState)
}

// Close closes the underlying socket connection.
func (c *Client) Close() error {
	c.closed.Store(true)
	return c.display.Context().Close()
}
