package clipboard

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/wlclient"
)

type MimeHandler struct {
	mu    sync.Mutex
	mimes []string
}

func (h *MimeHandler) HandleZwlrDataControlOfferV1Offer(
	e protocol.ZwlrDataControlOfferV1OfferEvent,
) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mimes = append(h.mimes, e.MimeType)
}

type Client struct {
	display  *wl.Display
	registry *wl.Registry
	manager  *protocol.ZwlrDataControlManagerV1

	clips chan Clip

	seatGlobals map[uint32]uint32

	deviceName    uint32
	deviceVersion uint32
}

func NewClient(clips chan Clip) *Client {
	c := new(Client)
	c.seatGlobals = make(map[uint32]uint32)
	c.clips = clips
	return c
}

func (h *Client) HandleZwlrDataControlDeviceV1DataOffer(
	e protocol.ZwlrDataControlDeviceV1DataOfferEvent,
) {
	collector := &MimeHandler{}
	e.Id.AddOfferHandler(collector)
	if err := wlclient.DisplayRoundtrip(h.display); err != nil {
		slog.Error("registry roundtrip failed", "error", err)
		return
	}

	slog.Info("Mime types collected", "id", e.Id.Id(), "mimes", collector.mimes)

	parser := NewClipboardParser(e.Id, collector.mimes)
	clip, err := parser.Parse()
	if err != nil {
		slog.Error("Failed to parse clipboard content", "error", err)
		return
	}
	h.clips <- clip
}

func (h *Client) HandleZwlrDataControlDeviceV1Selection(
	ev protocol.ZwlrDataControlDeviceV1SelectionEvent,
) {
}

func (h *Client) HandleZwlrDataControlDeviceV1PrimarySelection(
	ev protocol.ZwlrDataControlDeviceV1PrimarySelectionEvent,
) {
}

func (h *Client) HandleRegistryGlobal(ev wl.RegistryGlobalEvent) {
	if ev.Interface == "wl_seat" {
		h.seatGlobals[ev.Name] = ev.Version
	}
	if ev.Interface == "zwlr_data_control_manager_v1" {
		h.deviceName = ev.Name
		h.deviceVersion = ev.Version
	}
}

func (h *Client) HandleRegistryGlobalRemove(ev wl.RegistryGlobalRemoveEvent) {
	delete(h.seatGlobals, ev.Name)
}

func Watch(ctx context.Context, clips chan Clip) error {
	client := NewClient(clips)

	display, err := wlclient.DisplayConnect(nil)
	if err != nil {
		return err
	}
	defer display.Context().Close()
	client.display = display

	registry, err := display.GetRegistry()
	if err != nil {
		return err
	}
	defer registry.Context().Close()
	client.registry = registry

	wlclient.RegistryAddListener(registry, client)

	if err := wlclient.DisplayRoundtrip(display); err != nil {
		return fmt.Errorf("registry roundtrip failed: %v", err)
	}

	var seat *wl.Seat
	for id, ver := range client.seatGlobals {
		seat = wlclient.RegistryBindSeatInterface(registry, id, ver)
		break
	}

	if seat == nil {
		return fmt.Errorf("no wl_seat global found")
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
		return err
	}

	if err := wlclient.DisplayRoundtrip(display); err != nil {
		return fmt.Errorf("registry roundtrip failed: %v", err)
	}

	if manager == nil {
		return errors.New("no zwlr_data_control_manager_v1 global found")
	}

	device, err := manager.GetDataDevice(seat)
	if err != nil {
		return err
	}

	device.AddDataOfferHandler(client)
	device.AddSelectionHandler(client)
	device.AddPrimarySelectionHandler(client)

	fmt.Println("Watching clipboard changes...")

	for {
		if err := wlclient.DisplayDispatch(display); err != nil {
			return fmt.Errorf("dispatch failed: %v", err)
		}
	}
}
