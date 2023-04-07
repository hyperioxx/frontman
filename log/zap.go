package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	zap      *zap.Logger
	sugarZap *zap.SugaredLogger
}

func (l ZapLogger) Debugf(format string, args ...interface{}) {
	l.sugarZap.Debugf(format, args...)
}

func (l ZapLogger) Debug(args ...interface{}) {
	l.sugarZap.Debug(args...)
}

func (l ZapLogger) Fatalf(format string, args ...interface{}) {
	l.sugarZap.Fatalf(format, args...)
}

func (l ZapLogger) Fatal(args ...interface{}) {
	l.sugarZap.Fatal(args...)
}

func (l ZapLogger) Infof(format string, args ...interface{}) {
	l.sugarZap.Infof(format, args...)
}

func (l ZapLogger) Info(args ...interface{}) {
	l.sugarZap.Info(args...)
}

func (l ZapLogger) Warnf(format string, args ...interface{}) {
	l.sugarZap.Warnf(format, args...)
}

func (l ZapLogger) Warn(args ...interface{}) {
	l.sugarZap.Warn(args...)
}
func (l ZapLogger) Errorf(format string, args ...interface{}) {
	l.sugarZap.Errorf(format, args...)
}

func (l ZapLogger) Error(args ...interface{}) {
	l.sugarZap.Error(args...)
}

func fieldsToZap(fields ...Field) (zfields []zapcore.Field) {
	for _, field := range fields {
		zfields = append(zfields, zap.Field{
			Key:       field.key,
			String:    field.stringValue,
			Integer:   field.integerValue,
			Interface: field.value,
			Type:      field.GetZapType(),
		})
	}

	return zfields
}

func (f *Field) GetZapType() zapcore.FieldType {
	switch f.fieldType {
	case stringField:
		return zapcore.StringType
	case boolField:
		return zapcore.BoolType
	case integerField:
		return zapcore.Int64Type
	default:
		return zapcore.StringType
	}

}

func (l ZapLogger) WithFields(level logLevel, msg string, fields ...Field) {
	lvl, err := zapcore.ParseLevel(string(level))
	if err != nil {
		l.Error("Unknown log level: %s", level)
		lvl = zap.InfoLevel
	}
	l.zap.Log(lvl, msg, fieldsToZap(fields...)...)
}

func NewZapLogger(level logLevel) (Logger, error) {
	cfg := zap.NewProductionConfig()
	lvl, err := zapcore.ParseLevel(string(level))
	if err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	zap, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	logger := &ZapLogger{
		zap:      zap,
		sugarZap: zap.Sugar(),
	}
	return logger, nil
}

func NewDefaultLogger(level logLevel) (Logger, error) {
	return NewZapLogger(level)
}
