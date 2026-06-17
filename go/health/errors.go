package health

import "errors"

var ErrNatsDisconnected = errors.New("nats is not connected")
