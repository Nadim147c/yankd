package clipboard

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/wlclient"
)

// InterfaceName is the wlr-data-control-unstable-v1 interface name
const InterfaceName = "zwlr_data_control_manager_v1"

type mimeHandler struct {
	mu    sync.Mutex
	mimes []string
}

func (h *mimeHandler) HandleZwlrDataControlOfferV1Offer(
	e protocol.ZwlrDataControlOfferV1OfferEvent,
) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mimes = append(h.mimes, e.MimeType)
	slog.Debug("mime type added", "mime", e.MimeType, "total", len(h.mimes))
}

// Client is wayland that handle wayland clipboard protocol
type Client struct {
	display       *wl.Display
	registry      *wl.Registry
	manager       *protocol.ZwlrDataControlManagerV1
	clips         chan<- Clip
	seatGlobals   map[uint32]uint32
	deviceName    uint32
	deviceVersion uint32
	closed        atomic.Bool
}

// NewClient creates a new wayland client
func NewClient(clips chan<- Clip) *Client {
	c := new(Client)
	c.seatGlobals = make(map[uint32]uint32)
	c.clips = clips
	slog.Debug("clipboard client created")
	return c
}

// Close closes the underlying socket connection.
func (h *Client) Close() error {
	h.closed.Store(true)
	return h.display.Context().Close()
}

// HandleZwlrDataControlDeviceV1DataOffer handles whenever new clipboard is
// offered.
func (h *Client) HandleZwlrDataControlDeviceV1DataOffer(
	e protocol.ZwlrDataControlDeviceV1DataOfferEvent,
) {
	slog.Debug("data offer received", "offer_id", e.Id.Id())

	collector := &mimeHandler{}
	e.Id.AddOfferHandler(collector)

	if err := wlclient.DisplayRoundtrip(h.display); err != nil {
		if !h.closed.Load() {
			slog.Error(
				"registry roundtrip failed",
				"error", err,
				"closed-attempt", h.closed.Load(),
			)
		}
		return
	}

	slog.Info(
		"mime types collected",
		"offer_id", e.Id.Id(),
		"count", len(collector.mimes),
		"mimes", collector.mimes,
	)

	parser := newClipboardParser(e.Id, collector.mimes)
	clip, err := parser.Parse()
	if err != nil {
		slog.Error(
			"failed to parse clipboard content",
			"offer_id", e.Id.Id(),
			"error", err,
		)
		return
	}

	slog.Debug("clipboard content parsed successfully", "offer_id", e.Id.Id())
	h.clips <- clip
}

// HandleZwlrDataControlDeviceV1Selection handles selection changes. Currently
// does nothing!
func (h *Client) HandleZwlrDataControlDeviceV1Selection(
	protocol.ZwlrDataControlDeviceV1SelectionEvent,
) {
	slog.Debug("selection changed")
}

// HandleZwlrDataControlDeviceV1PrimarySelection handles primary selection
// changes. Currently does nothing!
func (h *Client) HandleZwlrDataControlDeviceV1PrimarySelection(
	protocol.ZwlrDataControlDeviceV1PrimarySelectionEvent,
) {
	slog.Debug("primary selection changed")
}

// HandleRegistryGlobal handles wl_seat and zwlr_data_control_manager_v1
// added.
func (h *Client) HandleRegistryGlobal(ev wl.RegistryGlobalEvent) {
	if ev.Interface == "wl_seat" {
		h.seatGlobals[ev.Name] = ev.Version
		slog.Debug(
			"wl_seat global registered",
			"name", ev.Name,
			"version", ev.Version,
		)
	}

	if ev.Interface == InterfaceName {
		h.deviceName = ev.Name
		h.deviceVersion = ev.Version
		slog.Debug(
			"zwlr_data_control_manager_v1 global registered",
			"name", ev.Name,
			"version", ev.Version,
		)
	}
}

// HandleRegistryGlobalRemove handles remove of globals.
func (h *Client) HandleRegistryGlobalRemove(ev wl.RegistryGlobalRemoveEvent) {
	if _, exists := h.seatGlobals[ev.Name]; exists {
		delete(h.seatGlobals, ev.Name)
		slog.Debug("wl_seat global removed", "name", ev.Name)
	}
}

// Watch watches for clipboard changes and send new clips to given channel.
func Watch(ctx context.Context, clips chan<- Clip) error {
	slog.Info("starting clipboard watch")

	client := NewClient(clips)

	display, err := wlclient.DisplayConnect(nil)
	if err != nil {
		slog.Error("failed to connect to wayland display", "error", err)
		return err
	}
	defer display.Context().Close()
	client.display = display
	slog.Debug("connected to wayland display")

	registry, err := display.GetRegistry()
	if err != nil {
		slog.Error("failed to get registry", "error", err)
		return err
	}
	defer registry.Context().Close()
	client.registry = registry
	slog.Debug("got wayland registry")

	wlclient.RegistryAddListener(registry, client)
	if err := wlclient.DisplayRoundtrip(display); err != nil {
		slog.Error("registry roundtrip failed", "error", err)
		return fmt.Errorf("registry roundtrip failed: %w", err)
	}

	var seat *wl.Seat
	for id, ver := range client.seatGlobals {
		seat = wlclient.RegistryBindSeatInterface(registry, id, ver)
		slog.Debug("bound to wl_seat", "id", id, "version", ver)
		break
	}

	if seat == nil {
		slog.Error("no wl_seat global found")
		return errors.New("no wl_seat global found")
	}
	defer seat.Context().Close()

	manager := protocol.NewZwlrDataControlManagerV1(display.Context())
	err = registry.Bind(
		client.deviceName,
		"zwlr_data_control_manager_v1",
		client.deviceVersion,
		manager,
	)
	if err != nil {
		slog.Error("failed to bind zwlr_data_control_manager_v1", "error", err)
		return err
	}
	slog.Debug("bound to zwlr_data_control_manager_v1")

	if err := wlclient.DisplayRoundtrip(display); err != nil {
		slog.Error("registry roundtrip failed", "error", err)
		return fmt.Errorf("registry roundtrip failed: %w", err)
	}

	if manager == nil {
		slog.Error("zwlr_data_control_manager_v1 is nil after binding")
		return errors.New("no zwlr_data_control_manager_v1 global found")
	}

	device, err := manager.GetDataDevice(seat)
	if err != nil {
		slog.Error("failed to get data device", "error", err)
		return err
	}
	slog.Debug("got data device")

	device.AddDataOfferHandler(client)
	device.AddSelectionHandler(client)
	device.AddPrimarySelectionHandler(client)
	slog.Debug("event handlers registered")

	slog.Info("clipboard watch initialized, listening for changes")

	go func() {
		<-ctx.Done()
		slog.Info("context cancelled → attempting clean close")
		client.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("clipboard watch context cancelled")
			return ctx.Err()
		default:
			err := wlclient.DisplayDispatch(display)
			if err != nil && !client.closed.Load() {
				slog.Error("dispatch failed", "error", err)
				return fmt.Errorf("dispatch failed: %w", err)
			}
		}
	}
}
