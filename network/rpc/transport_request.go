package rpc

import "encoding/json"

type RequestBody struct {
	Name []string        `json:"name"`
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}
