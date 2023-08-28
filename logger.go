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
	"sync/atomic"

	"github.com/casbin/casbin/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = &Logger{}

// Logger is the implementation for a Logger using zap.
type Logger struct {
	enabled int32
	logger  *zap.Logger
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
		logger: zapLogger,
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

func (l *Logger) LogModel(model [][]string) {
	if !l.IsEnabled() {
		return
	}

	l.logger.Info("LogModel", zap.Array("model", stringMatrix(model)))
}

func (l *Logger) LogEnforce(matcher string, request []interface{}, result bool, explains [][]string) {
	if !l.IsEnabled() {
		return
	}

	l.logger.Info(
		"LogEnforce",
		zap.String("matcher", matcher),
		zap.Array("request", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
			for _, v := range request {
				if err := enc.AppendReflected(v); err != nil {
					return err
				}
			}
			return nil
		})),
		zap.Bool("result", result),
		zap.Array("explains", stringMatrix(explains)),
	)
}

func (l *Logger) LogPolicy(policy map[string][][]string) {
	if !l.IsEnabled() {
		return
	}

	l.logger.Info("LogPolicy", zap.Object("policy", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range policy {
			if err := enc.AddArray(k, stringMatrix(v)); err != nil {
				return err
			}
		}
		return nil
	})))
}

func (l *Logger) LogRole(roles []string) {
	if !l.IsEnabled() {
		return
	}

	l.logger.Info("LogRole", zap.Strings("roles", roles))
}

func (l *Logger) LogError(err error, msg ...string) {
	if !l.IsEnabled() {
		return
	}

	l.logger.Error("LogError", zap.Error(err), zap.Strings("msg", msg))
}
