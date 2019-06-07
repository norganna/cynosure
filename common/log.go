package common

import "google.golang.org/grpc/grpclog"

var logger grpclog.LoggerV2

// SetLogger sets the global logger.
func SetLogger(log grpclog.LoggerV2) {
	logger = log
}

// Logger returns the global logger.
func Logger() grpclog.LoggerV2 {
	return logger
}
