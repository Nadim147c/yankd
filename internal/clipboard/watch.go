package clipboard

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Nadim147c/yankd/internal/models"
	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/wlclient"
)

type mimeHandler struct {
	mu        sync.Mutex
	fromYankd bool
	mimes     []string
}

func (h *mimeHandler) HandleZwlrDataControlOfferV1Offer(
	e protocol.ZwlrDataControlOfferV1OfferEvent,
) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fromYankd = h.fromYankd || e.MimeType == YankdMimeType
	h.mimes = append(h.mimes, e.MimeType)
	slog.Debug("mime type added", "mime", e.MimeType, "total", len(h.mimes))
}

// HandleZwlrDataControlDeviceV1DataOffer handles whenever new clipboard is
// offered.
func (c *Client) HandleZwlrDataControlDeviceV1DataOffer(e protocol.ZwlrDataControlDeviceV1DataOfferEvent) {
	slog.Debug("data offer received", "offer_id", e.Id.Id())

	if c.paused.Load() {
		slog.Debug("Skipping clipboard", "reason", "paused")
		return
	}

	collector := &mimeHandler{}
	e.Id.AddOfferHandler(collector)

	if err := wlclient.DisplayRoundtrip(c.display); err != nil {
		if !c.closed.Load() {
			slog.Error("registry roundtrip failed", "error", err)
		}
		return
	}

	slog.Info(
		"mime types collected",
		"offer_id", e.Id.Id(),
		"count", len(collector.mimes),
		"mimes", collector.mimes,
		"from_yankd", collector.fromYankd,
	)

	if collector.fromYankd {
		return
	}

	parser := newClipboardParser(e.Id, collector.mimes)
	event, err := parser.parse()
	if err != nil {
		slog.Error("failed to parse clipboard content", "offer_id", e.Id.Id(), "error", err)
		return
	}

	slog.Debug("clipboard content parsed successfully", "offer_id", e.Id.Id())
	c.eventChan <- event
}

// HandleZwlrDataControlDeviceV1Selection handles selection changes. Currently does nothing!
func (c *Client) HandleZwlrDataControlDeviceV1Selection(protocol.ZwlrDataControlDeviceV1SelectionEvent) {
	slog.Debug("selection changed")
}

// HandleZwlrDataControlDeviceV1PrimarySelection handles primary selection
// changes. Currently does nothing!
func (c *Client) HandleZwlrDataControlDeviceV1PrimarySelection(protocol.ZwlrDataControlDeviceV1PrimarySelectionEvent) {
	slog.Debug("primary selection changed")
}

// HandleRegistryGlobal handles wl_seat and zwlr_data_control_manager_v1 added.
func (c *Client) HandleRegistryGlobal(ev wl.RegistryGlobalEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ev.Interface == "wl_seat" {
		c.seatGlobals[ev.Name] = ev.Version
		slog.Debug("wl_seat global registered", "name", ev.Name, "version", ev.Version)
	}

	if ev.Interface == InterfaceName {
		c.deviceName = ev.Name
		c.deviceVersion = ev.Version
		slog.Debug(
			"zwlr_data_control_manager_v1 global registered",
			"name", ev.Name,
			"version", ev.Version,
		)
	}
}

// HandleRegistryGlobalRemove handles remove of globals.
func (c *Client) HandleRegistryGlobalRemove(ev wl.RegistryGlobalRemoveEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.seatGlobals[ev.Name]; exists {
		delete(c.seatGlobals, ev.Name)
		slog.Debug("wl_seat global removed", "name", ev.Name)
	}
}

// Watch watches for clipboard changes and send new clips to given channel.
func (c *Client) Listen(ctx context.Context, events chan<- models.ClipboardEvent) error {
	slog.Info("starting clipboard watch")
	c.eventChan = events

	display, err := wlclient.DisplayConnect(nil)
	if err != nil {
		slog.Error("failed to connect to wayland display", "error", err)
		return err
	}
	c.display = display
	defer c.Close()

	slog.Debug("connected to wayland display")

	registry, err := display.GetRegistry()
	if err != nil {
		slog.Error("failed to get registry", "error", err)
		return err
	}

	c.registry = registry
	slog.Debug("got wayland registry")

	wlclient.RegistryAddListener(registry, c)
	if err := wlclient.DisplayRoundtrip(display); err != nil {
		slog.Error("registry roundtrip failed", "error", err)
		return fmt.Errorf("registry roundtrip failed: %w", err)
	}

	c.connected.Store(true)

	var seat *wl.Seat
	for id, ver := range c.seatGlobals {
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

	err = registry.Bind(c.deviceName, "zwlr_data_control_manager_v1", c.deviceVersion, manager)
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

	c.manager = manager
	c.device = device

	device.AddDataOfferHandler(c)
	device.AddSelectionHandler(c)
	device.AddPrimarySelectionHandler(c)
	slog.Debug("event handlers registered")

	slog.Info("clipboard watch initialized, listening for changes")

	context.AfterFunc(ctx, func() { c.Close() }) //nolint

	for {
		select {
		case <-ctx.Done():
			slog.Info("clipboard watch context cancelled")
			return ctx.Err()
		default:
			err := wlclient.DisplayDispatch(display)
			if err != nil {
				slog.Error("dispatch failed", "error", err)
				return fmt.Errorf("dispatch failed: %w", err)
			}
		}
	}
}
