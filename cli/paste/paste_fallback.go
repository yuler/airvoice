//go:build !darwin && !linux && !windows

package paste

import (
	"fmt"
)

func New() (Paster, error) {
	return nil, fmt.Errorf("unsupported platform")
}
