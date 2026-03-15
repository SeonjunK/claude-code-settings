// Package logger provides session-scoped file logging with rotation.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// MaxLogSize is the max size per log file before rotation (5 MB).
	MaxLogSize = 5 * 1024 * 1024
	// MaxLogAge is how long to keep old log files (7 days).
	MaxLogAge = 7 * 24 * time.Hour
	// MaxRotated is the max number of rotated files per session.
	MaxRotated = 3
)

// Logger wraps slog with file-based, session-scoped output.
type Logger struct {
	file      *os.File
	slogger   *slog.Logger
	logPath   string
	logsDir   string
	sessionID string
}

// New creates a Logger that writes to {logsDir}/{sessionID}.log.
// If sessionID is empty, a timestamp-based name is used.
// Old log files beyond MaxLogAge are pruned on creation.
func New(logsDir, sessionID string) (*Logger, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("create logs dir: %w", err)
	}

	if sessionID == "" {
		sessionID = time.Now().Format("20060102-150405")
	}

	logPath := filepath.Join(logsDir, sessionID+".log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	l := &Logger{
		file:      f,
		slogger:   slog.New(handler),
		logPath:   logPath,
		logsDir:   logsDir,
		sessionID: sessionID,
	}

	// Prune old logs (best-effort, non-blocking)
	go l.prune()

	return l, nil
}

// Nop returns a no-op Logger that discards all output.
func Nop() *Logger {
	return &Logger{
		slogger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.rotateIfNeeded()
	l.slogger.Debug(msg, args...)
}

// Info logs at info level.
func (l *Logger) Info(msg string, args ...any) {
	l.rotateIfNeeded()
	l.slogger.Info(msg, args...)
}

// Warn logs at warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.rotateIfNeeded()
	l.slogger.Warn(msg, args...)
}

// Error logs at error level.
func (l *Logger) Error(msg string, args ...any) {
	l.rotateIfNeeded()
	l.slogger.Error(msg, args...)
}

// Close flushes and closes the log file.
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// rotateIfNeeded rotates the current log file if it exceeds MaxLogSize.
func (l *Logger) rotateIfNeeded() {
	if l.file == nil {
		return
	}
	info, err := l.file.Stat()
	if err != nil || info.Size() < MaxLogSize {
		return
	}

	l.file.Close()

	// Shift existing rotated files: .2 → .3, .1 → .2
	base := l.logPath
	for i := MaxRotated - 1; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d", base, i)
		dst := fmt.Sprintf("%s.%d", base, i+1)
		os.Rename(src, dst)
	}
	os.Rename(base, base+".1")

	// Remove oldest if over limit
	oldest := fmt.Sprintf("%s.%d", base, MaxRotated+1)
	os.Remove(oldest)

	// Reopen fresh file
	f, err := os.OpenFile(base, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	l.file = f
	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
	l.slogger = slog.New(handler)
}

// prune removes log files older than MaxLogAge.
func (l *Logger) prune() {
	entries, err := os.ReadDir(l.logsDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-MaxLogAge)

	// Collect log files with their info
	type logFile struct {
		path    string
		modTime time.Time
	}
	var logs []logFile

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			// Also match rotated files like .log.1, .log.2
			if !isRotatedLog(e.Name()) {
				continue
			}
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(l.logsDir, e.Name())

		// Don't delete current session's log
		if path == l.logPath {
			continue
		}

		logs = append(logs, logFile{path: path, modTime: info.ModTime()})
	}

	// Sort by mod time, newest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].modTime.After(logs[j].modTime)
	})

	for _, lf := range logs {
		if lf.modTime.Before(cutoff) {
			os.Remove(lf.path)
		}
	}
}

func isRotatedLog(name string) bool {
	// Match patterns like session-id.log.1, session-id.log.2
	for i := 1; i <= MaxRotated+1; i++ {
		if strings.HasSuffix(name, fmt.Sprintf(".log.%d", i)) {
			return true
		}
	}
	return false
}
