package logger

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLoggingBootstrap(t *testing.T) {
	defaultLog.Init()
	std := logrus.StandardLogger()
	if std.Level != logrus.InfoLevel {
		t.Errorf("Expected INFO level logs, but got: %s", std.Level)
	}
	if std.Out != os.Stdout {
		t.Errorf("Expected output to Stdout")
	}
	if _, ok := std.Formatter.(*DefaultLogFormatter); !ok {
		t.Errorf("Expected *containerpilot.DefaultLogFormatter got: %v", reflect.TypeOf(std.Formatter))
	}
}

func TestLoggingConfigInit(t *testing.T) {
	testLog := &Config{
		Level:  "DEBUG",
		Format: "text",
		Output: "stderr",
	}
	testLog.Init()
	std := logrus.StandardLogger()
	if std.Level != logrus.DebugLevel {
		t.Errorf("Expected 'debug' level logs, but got: %s", std.Level)
	}
	if std.Out != os.Stderr {
		t.Errorf("Expected output to Stderr")
	}
	if _, ok := std.Formatter.(*logrus.TextFormatter); !ok {
		t.Errorf("Expected *logrus.TextFormatter got: %v", reflect.TypeOf(std.Formatter))
	}
	// Reset to defaults
	defaultLog.Init()
}

func TestDefaultFormatterEmptyMessage(t *testing.T) {
	formatter := &DefaultLogFormatter{}
	_, err := formatter.Format(logrus.WithFields(
		logrus.Fields{
			"level": "info",
			"msg":   "something",
		},
	))
	if err != nil {
		t.Errorf("Did not expect error: %v", err)
	}
}

func TestDefaultFormatterPanic(t *testing.T) {
	defaultLog.Init()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic but did not")
		}
	}()
	logrus.Panicln("Panic Test")
}

func TestFileLogger(t *testing.T) {
	// initialize logger
	filename := "/tmp/test_log"
	testLog := &Config{
		Level:  "DEBUG",
		Format: "text",
		Output: filename,
	}
	err := testLog.Init()
	if err != nil {
		t.Errorf("Did not expect error: %v", err)
	}

	// write a log message
	logMsg := "this is a test"
	logrus.Info(logMsg)
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Did not expect error: %v", err)
	}
	if len(content) == 0 {
		t.Error("could not write log to file")
	}
	logs := string(content)
	if !strings.Contains(logs, logMsg) {
		t.Errorf("expected log file to contain '%s', got '%s'", logMsg, logs)
	}
}
