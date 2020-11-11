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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the implementation for a Logger using zap.
type Logger struct {
	enabled bool
	logger  *zap.Logger
}

// Mapping of event field.
var logEventMap = map[int]string{
	0: "LogTypeGrantedAccessRequest",
	1: "LogTypeRejectedAccessRequest",
	2: "LogTypeLoadPolicy",
	3: "LogTypePrintModel",
	4: "LogTypePrintPolicy",
	5: "LogTypePrintRole",
	6: "LogTypeLinkRole",
}

// NewLogger is the default constructor for Logger.
// Params : enabled, jsonEncode
// enabled initialize recording state, jsonEncode initialize log whether structured as json.
func NewLogger(enabled, jsonEncode bool) *Logger {
	var encoderConfig = zapcore.EncoderConfig{
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

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{enabled: enabled, logger: logger}
}

// NewLoggerByZap creates zap-logger by an existing zap instance.
func NewLoggerByZap(zapLogger *zap.Logger, enabled bool) *Logger {
	return &Logger{
		enabled: enabled,
		logger:  zapLogger,
	}
}

func (l *Logger) EnableLog(enable bool) {
	l.enabled = enable
}

func (l *Logger) IsEnabled() bool {
	return l.enabled
}

func (l *Logger) LogModel(event int, line []string, model [][]string) {
	if !l.enabled {
		return
	}

	l.logger.Info(logEventMap[event], zap.Strings("line", line))
}

func (l *Logger) LogEnforce(event int, line string, request *[]interface{}, policies *[]string, result *[]interface{}) {
	if !l.enabled {
		return
	}

	l.logger.Info(logEventMap[event], zap.String("line", line))
}

func (l *Logger) LogPolicy(event int, line string, pPolicyFormat []string, gPolicyFormat []string, pPolicy *[]interface{}, gPolicy *[]interface{}) {
	if !l.enabled {
		return
	}

	if pPolicy != nil {
		for k, v := range *pPolicy {
			l.logger.Info(logEventMap[event], zap.String("line", line), zap.Any("pPolicyFormat", pPolicyFormat[k]), zap.Any("pPolicy", v))
		}
	}
	if gPolicy != nil {
		for k, v := range *gPolicy {
			l.logger.Info(logEventMap[event], zap.String("line", line), zap.Any("gPolicyFormat", gPolicyFormat[k]), zap.Any("gPolicy", v))
		}
	}
}

func (l *Logger) LogRole(event int, line string, role []string) {
	if !l.enabled {
		return
	}

	l.logger.Info(logEventMap[event], zap.String("line", line))
}
