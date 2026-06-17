package ipc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func extractError(resp *http.Response) error {
	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(resp.Body)
	return fmt.Errorf("[%d] %s", resp.StatusCode, buf.String())
}

func renderJSON(w http.ResponseWriter, data any) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(data)
}
