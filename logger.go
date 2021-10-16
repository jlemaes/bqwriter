// Copyright 2021 OTA Insight Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bqwriter

import (
	"fmt"
	"os"
)

// Logger is the interface used by this module in order to support logging.
// By default the error messages are printed to the STDERR and debug messages are ignored, but you can
// inject any logger you wish into a Streamer.
//
// NOTE that it is assumed by this module for a Logger implementation to be safe for concurrent use.
type Logger interface {
	// Debug logs the arguments as a single debug message.
	Debug(args ...interface{})
	// Debugf logs a formatted debug message, injecting the arguments into the template string.
	Debugf(template string, args ...interface{})
	// Error logs the arguments as a single error message.
	Error(args ...interface{})
	// Errorf logs a formatted error message, injecting the arguments into the template string.
	Errorf(template string, args ...interface{})
}

// stdLogger is the builtin implementation of the Logger interface,
// and logs error messages to the STDERR, but ignores debug messages.
type stdLogger struct{}

// Debug implements Logger::Debug
func (stdl stdLogger) Debug(args ...interface{}) {} // debug messages are ignored by default

// Debugf implements Logger::Debugf
func (stdl stdLogger) Debugf(template string, args ...interface{}) {} // debug messages are ignored by default

// Error implements Logger::Error
func (stdl stdLogger) Error(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
}

// Errorf implements Logger::Errorf
func (stdl stdLogger) Errorf(template string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, template, args...)
}
