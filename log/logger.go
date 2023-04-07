package log

import (
	"fmt"
)

type logLevel string

const (
	InfoLevel    logLevel  = "info"
	DebugLevel   logLevel  = "debug"
	WarnLevel    logLevel  = "warn"
	ErrorLevel   logLevel  = "error"
	boolField    fieldType = "bool"
	stringField  fieldType = "string"
	integerField fieldType = "integer"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithFields(level logLevel, msg string, fields ...Field)
}

// Field used for structured logging
type Field struct {
	key          string
	stringValue  string
	integerValue int64
	value        interface{}
	fieldType    fieldType
}

type fieldType string

func String(key string, value string) Field {
	return Field{
		key:         key,
		stringValue: value,
		fieldType:   stringField,
	}
}

func Bool(key string, value bool) Field {
	var val int64
	if value {
		val = 1
	}
	return Field{
		key:          key,
		integerValue: val,
		fieldType:    boolField,
	}
}

func Int(key string, value int64) Field {
	return Field{
		key:          key,
		integerValue: value,
		fieldType:    integerField,
	}
}

func Error(value string) Field {
	return Field{
		key:   "err",
		value: value,
	}
}

func ParseLevel(str string) logLevel {
	var lvl logLevel
	lvl.unmarshalString(str)
	return lvl
}

func (l *logLevel) unmarshalString(str string) {
	switch str {
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "INFO", "": // make the zero value useful
		*l = InfoLevel
	case "warn", "WARN":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	default:
		fmt.Println("unknown log level ", str, " proceeding with log levle Info")
		*l = InfoLevel
	}
}
