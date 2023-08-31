/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import "log"

type Logger struct {
	Level LogLevel
}

func (l *Logger) Debug(msg string) {
	if l.Level >= LogLevelDebug {
		log.Println("[DEBUG]", msg)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Level >= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) Info(msg string) {
	if l.Level >= LogLevelInfo {
		log.Println("[INFO]", msg)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.Level >= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

func (l *Logger) Warn(msg string) {
	if l.Level >= LogLevelWarn {
		log.Println("[WARN]", msg)
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.Level >= LogLevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

func (l *Logger) Error(msg string) {
	if l.Level >= LogLevelError {
		log.Println("[ERROR]", msg)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.Level >= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}
