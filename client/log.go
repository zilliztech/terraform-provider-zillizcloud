package client

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelInfo
	LogLevelDebug
)

type LoggerWrapper struct {
	logger   *log.Logger
	minLevel LogLevel
}

func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	default:
		return LogLevelError
	}
}

func NewLoggerWrapper(base *log.Logger) *LoggerWrapper {
	level := parseLogLevel(os.Getenv("ZILLIZ_LOG_LEVEL"))
	return &LoggerWrapper{
		logger:   base,
		minLevel: level,
	}
}
func (l *LoggerWrapper) logf(level LogLevel, skip int, format string, v ...any) {
	if level > l.minLevel {
		return
	}
	prefix := findExportedCaller()
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s", prefix, msg)
}

func findExportedCaller() string {
	pcs := make([]uintptr, 20)
	n := runtime.Callers(3, pcs) // skip runtime.Callers, findExportedCaller, logf
	if n == 0 {
		return "unknown"
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		fn := frame.Function
		if fn == "" {
			continue
		}
		parts := strings.Split(fn, ".")
		base := parts[len(parts)-1]

		// Skip logging functions
		if isLogFunc(base) {
			continue
		}

		if len(base) > 0 && isUpper(base[0]) {
			file := frame.File
			if idx := strings.LastIndex(file, "/"); idx != -1 {
				file = file[idx+1:]
			}
			return fmt.Sprintf("%s:%d: %s", file, frame.Line, base)
		}
		if !more {
			break
		}
	}
	return "unknown"
}

func isLogFunc(name string) bool {
	switch name {
	case "Printf", "Println", "Debugf", "Infof", "Errorf", "logf":
		return true
	default:
		return false
	}
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func (l *LoggerWrapper) Debugf(format string, v ...any) {
	l.logf(LogLevelDebug, 4, format, v...)
}

func (l *LoggerWrapper) Infof(format string, v ...any) {
	l.logf(LogLevelInfo, 4, format, v...)
}

func (l *LoggerWrapper) Errorf(format string, v ...any) {
	l.logf(LogLevelError, 4, format, v...)
}

func (l *LoggerWrapper) Printf(format string, v ...any) {
	l.Infof(format, v...)
}

func (l *LoggerWrapper) Println(v ...any) {
	l.Infof(strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
}

func generateShortID() string {
	now := time.Now().UnixNano()
	pid := os.Getpid()

	randBytes := make([]byte, 8)
	_, err := rand.Read(randBytes)
	if err != nil {
		raw := []byte(strconv.FormatInt(now, 10) + strconv.Itoa(pid))
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:])[:8]
	}

	seed := append([]byte(strconv.FormatInt(now, 10)+strconv.Itoa(pid)), randBytes...)
	sum := sha256.Sum256(seed)
	return hex.EncodeToString(sum[:])[:8]
}

var sensitiveKeys = []string{
	"password", "secret", "token", "access_token", "authorization", "api_key",
}

func (l *LoggerWrapper) logResponseBody(body []byte) {
	if os.Getenv("ZILLIZ_LOG_RAW_JSON") == "true" {
		return
	}

	var parsed zillizResponse[any]
	if err := json.Unmarshal(body, &parsed); err != nil {
		l.Printf("Response Body: [non-json] (len=%d)", len(body))
		return
	}

	bytes, _ := json.Marshal(parsed)
	l.Printf("Response JSON: %s", maskSensitiveFields(string(bytes)))
}
func (l *LoggerWrapper) logRequestBody(buf io.ReadWriter) {
	if os.Getenv("ZILLIZ_LOG_RAW_JSON") == "true" {
		return
	}
	l.Printf("Request: %s", maskSensitiveFields(buf.(*bytes.Buffer).String()))
}

func maskSensitiveFields(s string) string {
	fields := []string{"password", "access_token", "secret", "authorization", "api_key"}

	for _, key := range fields {
		pattern := `"` + regexp.QuoteMeta(key) + `"\s*:\s*".*?"`
		replacement := `"` + key + `":"***"`
		re := regexp.MustCompile(pattern)
		s = re.ReplaceAllString(s, replacement)
	}
	return s
}

func describeField(key string, val any) string {
	if isSensitiveKey(key) {
		return "*** (masked)"
	}
	switch v := val.(type) {
	case string:
		if len(v) > 80 {
			return fmt.Sprintf("string(len=%d)", len(v))
		}
		return fmt.Sprintf("string: \"%s\"", v)
	case float64:
		return fmt.Sprintf("number: %.2f", v)
	case bool:
		return fmt.Sprintf("bool: %v", v)
	case []any:
		return fmt.Sprintf("array(len=%d)", len(v))
	case map[string]any:
		return fmt.Sprintf("object(fields=%d)", len(v))
	case nil:
		return "null"
	default:
		return fmt.Sprintf("type=%T", v)
	}
}

func isSensitiveKey(key string) bool {
	k := normalizeKey(key)
	for _, sk := range sensitiveKeys {
		if k == sk {
			return true
		}
	}
	return false
}

func normalizeKey(k string) string {
	k = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(k, "")
	return string(bytes.ToLower([]byte(k)))
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
