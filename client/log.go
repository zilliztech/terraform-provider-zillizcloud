package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
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

var sensitiveKeys = []string{
	"password", "secret", "token", "access_token", "authorization", "api_key",
}

func maskSensitiveFields(s string) string {
	for _, key := range sensitiveKeys {
		pattern := `"` + regexp.QuoteMeta(key) + `"\s*:\s*".*?"`
		replacement := `"` + key + `":"***"`
		re := regexp.MustCompile(pattern)
		s = re.ReplaceAllString(s, replacement)
	}
	return s
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

func RequestToCurl(req *http.Request) (string, error) {
	var sb strings.Builder

	sb.WriteString("curl -i")

	// Method
	if req.Method != http.MethodGet {
		sb.WriteString(fmt.Sprintf(" -X %s", req.Method))
	}

	// Headers
	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys) // keep output deterministic

	for _, k := range keys {
		for _, v := range req.Header[k] {
			sb.WriteString(fmt.Sprintf(` -H "%s: %s"`, k, v))
		}
	}

	// Body (if repeatable)
	if req.Body != nil && req.GetBody != nil {
		bodyReader, err := req.GetBody()
		if err != nil {
			return "", fmt.Errorf("could not clone body: %w", err)
		}
		defer bodyReader.Close()
		bodyBytes := make([]byte, req.ContentLength)
		_, err = bodyReader.Read(bodyBytes)
		if err == nil && len(bodyBytes) > 0 {
			bodyStr := string(bodyBytes)
			bodyStr = strings.TrimRight(bodyStr, "\r\n\t ")
			sb.WriteString(fmt.Sprintf(` --data-binary '%s'`, escapeSingleQuote(maskSensitiveFields(bodyStr))))
		}
	}

	// URL
	sb.WriteString(fmt.Sprintf(" '%s'", req.URL.String()))

	return sb.String(), nil
}

func escapeSingleQuote(s string) string {
	// Escapes single quotes in curl string using the correct bash-safe pattern
	return strings.ReplaceAll(s, `'`, `'\''`)
}
