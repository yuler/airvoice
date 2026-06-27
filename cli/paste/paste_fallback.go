//go:build !darwin && !linux

package paste

import (
	"fmt"
)

func New() (Paster, error) {
	return nil, fmt.Errorf("unsupported platform")
}
