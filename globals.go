package dht

import (
	"golang.org/x/time/rate"
)

var defaultSendLimiter = rate.NewLimiter(rate.Inf, 0)
