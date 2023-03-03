package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	fatihcolor "github.com/fatih/color"
	"github.com/zeromicro/go-zero/core/color"
)

type (
	Writer interface {
		Alert(v interface{})
		Close() error
		Error(v interface{}, fields ...LogField)
		Info(v interface{}, fields ...LogField)
		Severe(v interface{})
		Slow(v interface{}, fields ...LogField)
		Stack(v interface{})
		Stat(v interface{}, fields ...LogField)
	}

	atomicWriter struct {
		writer Writer
		lock   sync.RWMutex
	}

	concreteWriter struct {
		infoLog   io.WriteCloser
		errorLog  io.WriteCloser
		severeLog io.WriteCloser
		slowLog   io.WriteCloser
		statLog   io.WriteCloser
		stackLog  io.Writer
	}
)

// NewWriter creates a new Writer with the given io.Writer.
func NewWriter(w io.Writer) Writer {
	lw := newLogWriter(log.New(w, "", flags))

	return &concreteWriter{
		infoLog:   lw,
		errorLog:  lw,
		severeLog: lw,
		slowLog:   lw,
		statLog:   lw,
		stackLog:  lw,
	}
}

func (w *atomicWriter) Load() Writer {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.writer
}

func (w *atomicWriter) Store(v Writer) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.writer = v
}

func (w *atomicWriter) Swap(v Writer) Writer {
	w.lock.Lock()
	defer w.lock.Unlock()
	old := w.writer
	w.writer = v
	return old
}

func newConsoleWriter() Writer {
	outLog := newLogWriter(log.New(fatihcolor.Output, "", flags))
	errLog := newLogWriter(log.New(fatihcolor.Error, "", flags))
	return &concreteWriter{
		infoLog:   outLog,
		errorLog:  errLog,
		severeLog: errLog,
		slowLog:   errLog,
		stackLog:  newLessWriter(errLog, options.logStackCooldownMills),
		statLog:   outLog,
	}
}

func newFileWriter(c LogConf) (Writer, error) {
	var err error
	var opts []LogOption
	var infoLog io.WriteCloser
	var errorLog io.WriteCloser
	var severeLog io.WriteCloser
	var slowLog io.WriteCloser
	var statLog io.WriteCloser
	var stackLog io.Writer

	if len(c.Path) == 0 {
		return nil, ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}
	if c.MaxBackups > 0 {
		opts = append(opts, WithMaxBackups(c.MaxBackups))
	}
	if c.MaxSize > 0 {
		opts = append(opts, WithMaxSize(c.MaxSize))
	}

	opts = append(opts, WithRotation(c.Rotation))

	accessFile := path.Join(c.Path, accessFilename)
	errorFile := path.Join(c.Path, errorFilename)
	severeFile := path.Join(c.Path, severeFilename)
	slowFile := path.Join(c.Path, slowFilename)
	statFile := path.Join(c.Path, statFilename)

	handleOptions(opts)
	setupLogLevel(c)

	if infoLog, err = createOutput(accessFile); err != nil {
		return nil, err
	}

	if errorLog, err = createOutput(errorFile); err != nil {
		return nil, err
	}

	if severeLog, err = createOutput(severeFile); err != nil {
		return nil, err
	}

	if slowLog, err = createOutput(slowFile); err != nil {
		return nil, err
	}

	if statLog, err = createOutput(statFile); err != nil {
		return nil, err
	}

	stackLog = newLessWriter(errorLog, options.logStackCooldownMills)

	return &concreteWriter{
		infoLog:   infoLog,
		errorLog:  errorLog,
		severeLog: severeLog,
		slowLog:   slowLog,
		statLog:   statLog,
		stackLog:  stackLog,
	}, nil
}

func (w *concreteWriter) Alert(v interface{}) {
	output(w.errorLog, levelAlert, v)
}

func (w *concreteWriter) Close() error {
	if err := w.infoLog.Close(); err != nil {
		return err
	}

	if err := w.errorLog.Close(); err != nil {
		return err
	}

	if err := w.severeLog.Close(); err != nil {
		return err
	}

	if err := w.slowLog.Close(); err != nil {
		return err
	}

	return w.statLog.Close()
}

func (w *concreteWriter) Error(v interface{}, fields ...LogField) {
	output(w.errorLog, levelError, v, fields...)
}

func (w *concreteWriter) Info(v interface{}, fields ...LogField) {
	output(w.infoLog, levelInfo, v, fields...)
}

func (w *concreteWriter) Severe(v interface{}) {
	output(w.severeLog, levelFatal, v)
}

func (w *concreteWriter) Slow(v interface{}, fields ...LogField) {
	output(w.slowLog, levelSlow, v, fields...)
}

func (w *concreteWriter) Stack(v interface{}) {
	output(w.stackLog, levelError, v)
}

func (w *concreteWriter) Stat(v interface{}, fields ...LogField) {
	output(w.statLog, levelStat, v, fields...)
}

type nopWriter struct{}

func (n nopWriter) Alert(_ interface{}) {
}

func (n nopWriter) Close() error {
	return nil
}

func (n nopWriter) Error(_ interface{}, _ ...LogField) {
}

func (n nopWriter) Info(_ interface{}, _ ...LogField) {
}

func (n nopWriter) Severe(_ interface{}) {
}

func (n nopWriter) Slow(_ interface{}, _ ...LogField) {
}

func (n nopWriter) Stack(_ interface{}) {
}

func (n nopWriter) Stat(_ interface{}, _ ...LogField) {
}

func buildFields(fields ...LogField) []string {
	var items []string

	for _, field := range fields {
		items = append(items, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}

	return items
}

func output(writer io.Writer, level string, val interface{}, fields ...LogField) {
	fields = append(fields, Field(callerKey, getCaller(callerDepth)))

	switch atomic.LoadUint32(&encoding) {
	case plainEncodingType:
		writePlainAny(writer, level, val, buildFields(fields...)...)
	default:
		entry := make(logEntryWithFields)
		for _, field := range fields {
			entry[field.Key] = field.Value
		}
		entry[timestampKey] = getTimestamp()
		entry[levelKey] = level
		entry[contentKey] = val
		writeJson(writer, entry)
	}
}

func wrapLevelWithColor(level string) string {
	var colour color.Color
	switch level {
	case levelAlert:
		colour = color.FgRed
	case levelError:
		colour = color.FgRed
	case levelFatal:
		colour = color.FgRed
	case levelInfo:
		colour = color.FgBlue
	case levelSlow:
		colour = color.FgYellow
	case levelStat:
		colour = color.FgGreen
	}

	if colour == color.NoColor {
		return level
	}

	return color.WithColorPadding(level, colour)
}

func writeJson(writer io.Writer, info interface{}) {
	if content, err := json.Marshal(info); err != nil {
		log.Println(err.Error())
	} else if writer == nil {
		log.Println(string(content))
	} else {
		writer.Write(append(content, '\n'))
	}
}

func writePlainAny(writer io.Writer, level string, val interface{}, fields ...string) {
	level = wrapLevelWithColor(level)

	switch v := val.(type) {
	case string:
		writePlainText(writer, level, v, fields...)
	case error:
		writePlainText(writer, level, v.Error(), fields...)
	case fmt.Stringer:
		writePlainText(writer, level, v.String(), fields...)
	default:
		writePlainValue(writer, level, v, fields...)
	}
}

func writePlainText(writer io.Writer, level, msg string, fields ...string) {
	var buf strings.Builder
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(level)
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(msg)
	for _, item := range fields {
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(item)
	}
	buf.WriteByte('\n')
	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := fmt.Fprint(writer, buf.String()); err != nil {
		log.Println(err.Error())
	}
}

func writePlainValue(writer io.Writer, level string, val interface{}, fields ...string) {
	var buf strings.Builder
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(level)
	buf.WriteByte(plainEncodingSep)
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		log.Println(err.Error())
		return
	}

	for _, item := range fields {
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(item)
	}
	buf.WriteByte('\n')
	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := fmt.Fprint(writer, buf.String()); err != nil {
		log.Println(err.Error())
	}
}
