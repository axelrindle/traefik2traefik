package internal

import (
	traefik "github.com/traefik/traefik/v3/pkg/config/runtime"
)

type RouterRepresentation struct {
	*traefik.RouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}
