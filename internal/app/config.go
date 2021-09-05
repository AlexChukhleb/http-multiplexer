package app

import "time"

type Config struct {
	MaxOutgoingConnections int
	MaxLinksPerRequest     int
	RequestTimeout         time.Duration
}
