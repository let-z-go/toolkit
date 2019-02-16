package logger

import (
	"fmt"
	"io"
	"log"
)

const (
	SeverityInfo SeverityLevel = iota
	SeverityWarning
	SeverityError
	SeverityFatal

	numberOfSeverityLevels = iota
)

type Logger struct {
	bases [numberOfSeverityLevels]*log.Logger
}

func (self *Logger) Initialize(name string, severityLevel SeverityLevel, writer1 io.Writer, writers ...io.Writer) *Logger {
	for severityLevel < numberOfSeverityLevels {
		var writer io.Writer

		if i := int(severityLevel) - 1; i >= 0 && i < len(writers) {
			writer = writers[i]
		} else {
			writer = writer1
		}

		var logPrefix string

		if name == "" {
			logPrefix = ""
		} else {
			logPrefix = name + " - "
		}

		logPrefix += severityLevel.String() + " "
		self.bases[severityLevel] = log.New(writer, logPrefix, log.LstdFlags)
		severityLevel++
	}

	return self
}

func (self *Logger) Info(args ...interface{}) {
	if base := self.bases[SeverityInfo]; base != nil {
		base.Print(args...)
	}
}

func (self *Logger) Infoln(args ...interface{}) {
	if base := self.bases[SeverityInfo]; base != nil {
		base.Println(args...)
	}
}

func (self *Logger) Infof(format string, args ...interface{}) {
	if base := self.bases[SeverityInfo]; base != nil {
		base.Printf(format, args...)
	}
}

func (self *Logger) Warning(args ...interface{}) {
	if base := self.bases[SeverityWarning]; base != nil {
		base.Print(args...)
	}
}

func (self *Logger) Warningln(args ...interface{}) {
	if base := self.bases[SeverityWarning]; base != nil {
		base.Println(args...)
	}
}

func (self *Logger) Warningf(format string, args ...interface{}) {
	if base := self.bases[SeverityWarning]; base != nil {
		base.Printf(format, args...)
	}
}

func (self *Logger) Error(args ...interface{}) {
	if base := self.bases[SeverityError]; base != nil {
		base.Print(args...)
	}
}

func (self *Logger) Errorln(args ...interface{}) {
	if base := self.bases[SeverityError]; base != nil {
		base.Println(args...)
	}
}

func (self *Logger) Errorf(format string, args ...interface{}) {
	if base := self.bases[SeverityError]; base != nil {
		base.Printf(format, args...)
	}
}

func (self *Logger) Fatal(args ...interface{}) {
	if base := self.bases[SeverityFatal]; base != nil {
		base.Fatal(args...)
	}
}

func (self *Logger) Fatalln(args ...interface{}) {
	if base := self.bases[SeverityFatal]; base != nil {
		base.Fatalln(args...)
	}
}

func (self *Logger) Fatalf(format string, args ...interface{}) {
	if base := self.bases[SeverityFatal]; base != nil {
		base.Fatalf(format, args...)
	}
}

type SeverityLevel int

func (self SeverityLevel) String() string {
	switch self {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityFatal:
		return "FATAL"
	default:
		return ""
	}
}

func (self SeverityLevel) GoString() string {
	switch self {
	case SeverityInfo:
		return "<SeverityInfo>"
	case SeverityWarning:
		return "<SeverityWarning>"
	case SeverityError:
		return "<SeverityError>"
	case SeverityFatal:
		return "<SeverityFatal>"
	default:
		return fmt.Sprintf("<SeverityLevel:%d>", self)
	}
}
