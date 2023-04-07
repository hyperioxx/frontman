package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	logger, err := NewDefaultLogger("debug")
	if err != nil {
		t.Error(err)
	}
	// For illustrative purposes
	logger.Info("basic ", "logging")
	logger.Infof("info formatted log: %s to %d", "I can count", 123)
	logger.Debug("basic ", "logging")
	logger.Debugf("debug formatted log: %s to %d", "I can count", 123)
	logger.Error("basic ", "logging")
	logger.Errorf("error formatted log: %s to %d", "I could not count", 123)
	logger.WithFields(ErrorLevel, "unexpected traffic received", Error("terrible error message"))
	logger.WithFields(InfoLevel, "ingress traffic received", String("url", "https://github.com/Frontman-Labs/frontman"), String("host", "162.1.3.2"), Int("port", 32133), Bool("tls_enabled", true))
}
