package di

import "github.com/planetary-social/go-ssb/logging"

type Config struct {
	LoggingLevel  logging.Level
	DataDirectory string
	ListenAddress string
}
