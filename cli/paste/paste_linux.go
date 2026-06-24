//go:build linux

package paste

import (
	"fmt"
)

func New() (Paster, error) {
	session := DetectSession()
	switch session {
	case SessionWayland:
		return &waylandPaster{}, nil
	case SessionX11:
		return &x11Paster{}, nil
	default:
		return nil, fmt.Errorf("unsupported or undetected session type: %s", session)
	}
}
