package servicemanager

import "errors"

var (
  ErrContainerNotStarted = errors.New("Container not yet started")
)
