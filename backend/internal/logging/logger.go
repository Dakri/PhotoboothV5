package logging

import (
	"fmt"
	"sync"
	"time"
)

type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelDebug Level = "debug"
)

type Entry struct {
	Level     Level  `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
}

// BroadcastFunc is called whenever a new log entry is added
type BroadcastFunc func(entry Entry)

type Logger struct {
	mu        sync.Mutex
	entries   []Entry
	maxSize   int
	broadcast BroadcastFunc
}

var defaultLogger *Logger

func Init(maxSize int) *Logger {
	defaultLogger = &Logger{
		entries: make([]Entry, 0, maxSize),
		maxSize: maxSize,
	}
	return defaultLogger
}

func Get() *Logger {
	if defaultLogger == nil {
		Init(500)
	}
	return defaultLogger
}

func (l *Logger) SetBroadcast(fn BroadcastFunc) {
	l.broadcast = fn
}

func (l *Logger) Add(level Level, source, message string) {
	entry := Entry{
		Level:     level,
		Message:   message,
		Source:    source,
		Timestamp: time.Now().UnixMilli(),
	}

	l.mu.Lock()
	if len(l.entries) >= l.maxSize {
		// Drop oldest 10%
		drop := l.maxSize / 10
		l.entries = l.entries[drop:]
	}
	l.entries = append(l.entries, entry)
	l.mu.Unlock()

	if l.broadcast != nil {
		l.broadcast(entry)
	}
}

func (l *Logger) GetEntries(limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	// Return last N entries
	start := len(l.entries) - limit
	result := make([]Entry, limit)
	copy(result, l.entries[start:])
	return result
}

// Convenience functions
func (l *Logger) Info(source, msg string, args ...interface{}) {
	l.Add(LevelInfo, source, fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(source, msg string, args ...interface{}) {
	l.Add(LevelWarn, source, fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(source, msg string, args ...interface{}) {
	l.Add(LevelError, source, fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(source, msg string, args ...interface{}) {
	l.Add(LevelDebug, source, fmt.Sprintf(msg, args...))
}
