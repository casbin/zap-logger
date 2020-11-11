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
	"os"
	"testing"

	"github.com/casbin/casbin/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testNewLogger(t *testing.T) *Logger {
	logger := NewLogger(false, true)

	if logger == nil {
		t.Error("initialize logger failed")
	}

	return logger
}

func testNewLoggerByZap(t *testing.T) *Logger {
	var encoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "event",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		zap.NewAtomicLevel(),
	)
	logger := NewLoggerByZap(zap.New(core), false)

	if logger == nil {
		t.Error("initialize logger failed")
	}

	return logger
}

func testLoggerLog(t *testing.T, logger log.Logger) {
	for i := 0; i < 10000; i++ {
		logger.LogEnforce(1, "2", &[]interface{}{"3"}, &[]string{"4"}, &[]interface{}{"5"})
		logger.LogPolicy(2, "3", []string{"4"}, []string{"5"}, &[]interface{}{"6"}, &[]interface{}{"7"})
		logger.LogModel(3, []string{"4"}, [][]string{{"5"}, {"6"}})
		logger.LogRole(5, "6", []string{"7"})
	}
}

func TestLogger(t *testing.T) {
	loggerByDefault := testNewLogger(t)
	loggerByZap := testNewLoggerByZap(t)

	if loggerByDefault.IsEnabled() || loggerByZap.IsEnabled() {
		t.Error("IsEnabled logger error")
	}

	loggerByDefault.EnableLog(true)
	loggerByZap.EnableLog(true)

	if !loggerByDefault.IsEnabled() || !loggerByZap.IsEnabled() {
		t.Error("Enable logger error")
	}

	testLoggerLog(t, loggerByDefault)
	testLoggerLog(t, loggerByZap)
}
