// Copyright 2020 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zaplogger

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v3/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = &Logger{}

// Logger is the implementation for a Logger using zap.
type Logger struct {
	enabled     int32
	logger      *zap.Logger
	eventTypes  map[log.EventType]bool
	logCallback func(entry *log.LogEntry) error
}

type stringMatrix [][]string

func (matrix stringMatrix) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, vector := range matrix {
		if err := enc.AppendArray(zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
			for _, item := range vector {
				enc.AppendString(item)
			}
			return nil
		})); err != nil {
			return err
		}
	}
	return nil
}

// NewLogger is the default constructor for Logger.
// Params : enabled, jsonEncode
// enabled initialize recording state, jsonEncode initialize log whether structured as json.
func NewLogger(enabled, jsonEncode bool) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "event",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.InfoLevel) // default log level: info

	encodeMethod := "json"
	if !jsonEncode {
		encodeMethod = "console"
	}

	config := zap.Config{
		Level:            atomicLevel,
		Development:      true,
		Encoding:         encodeMethod,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return NewLoggerByZap(zapLogger, enabled)
}

// NewLoggerByZap creates zap-logger by an existing zap instance.
func NewLoggerByZap(zapLogger *zap.Logger, enabled bool) *Logger {
	logger := &Logger{
		logger:     zapLogger,
		eventTypes: make(map[log.EventType]bool),
	}
	logger.EnableLog(enabled)
	return logger
}

func (l *Logger) EnableLog(enable bool) {
	var enab int32
	if enable {
		enab = 1
	}
	atomic.StoreInt32(&l.enabled, enab)
}

func (l *Logger) IsEnabled() bool {
	return atomic.LoadInt32(&l.enabled) == 1
}

// SetEventTypes sets the event types that should be logged.
// Only events matching these types will have IsActive set to true.
func (l *Logger) SetEventTypes(eventTypes []log.EventType) error {
	l.eventTypes = make(map[log.EventType]bool)
	for _, et := range eventTypes {
		l.eventTypes[et] = true
	}
	return nil
}

// OnBeforeEvent is called before an event occurs.
// It sets the StartTime and determines if the event should be active based on configured event types.
func (l *Logger) OnBeforeEvent(entry *log.LogEntry) error {
	if entry == nil {
		return fmt.Errorf("log entry is nil")
	}

	entry.StartTime = time.Now()

	// Set IsActive based on whether this event type is enabled
	// If no event types are configured, all events are considered active
	if len(l.eventTypes) == 0 {
		entry.IsActive = true
	} else {
		entry.IsActive = l.eventTypes[entry.EventType]
	}

	return nil
}

// OnAfterEvent is called after an event completes.
// It calculates the duration, logs the entry if active, and calls the user callback if set.
func (l *Logger) OnAfterEvent(entry *log.LogEntry) error {
	if entry == nil {
		return fmt.Errorf("log entry is nil")
	}

	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.StartTime)

	// Only log if the event is active
	if entry.IsActive && l.IsEnabled() {
		// Build zap fields from log entry
		fields := []zap.Field{
			zap.String("event_type", string(entry.EventType)),
			zap.Duration("duration", entry.Duration),
		}

		// Add event-specific fields
		switch entry.EventType {
		case log.EventEnforce:
			fields = append(fields,
				zap.String("subject", entry.Subject),
				zap.String("object", entry.Object),
				zap.String("action", entry.Action),
				zap.String("domain", entry.Domain),
				zap.Bool("allowed", entry.Allowed),
			)

		case log.EventAddPolicy, log.EventRemovePolicy, log.EventLoadPolicy, log.EventSavePolicy:
			fields = append(fields, zap.Int("rule_count", entry.RuleCount))
			if len(entry.Rules) > 0 {
				fields = append(fields, zap.Array("rules", stringMatrix(entry.Rules)))
			}
		}

		// Log at appropriate level
		message := string(entry.EventType)
		if entry.Error != nil {
			fields = append(fields, zap.Error(entry.Error))
			l.logger.Error(message, fields...)
		} else {
			l.logger.Info(message, fields...)
		}
	}

	// Call user-provided callback if set
	if l.logCallback != nil {
		return l.logCallback(entry)
	}

	return nil
}

// SetLogCallback sets a user-provided callback function.
// The callback is called at the end of OnAfterEvent.
func (l *Logger) SetLogCallback(callback func(entry *log.LogEntry) error) error {
	l.logCallback = callback
	return nil
}
