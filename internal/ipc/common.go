package ipc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func extractError(resp *http.Response) error {
	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("[%d] %s", resp.StatusCode, buf.String())
}

func renderJSON(w http.ResponseWriter, data any) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "Application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("failed to render json to client", "error", err)
	}
}
