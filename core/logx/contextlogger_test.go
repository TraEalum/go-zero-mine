package logx

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceLog(t *testing.T) {
	SetLevel(InfoLevel)
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, span := tp.Tracer("trace-id").Start(context.Background(), "span-id")
	defer span.End()

	WithContext(ctx).Info(testlog)
	validate(t, w.String(), true, true)
}

func TestTraceError(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, span := tp.Tracer("trace-id").Start(context.Background(), "span-id")
	defer span.End()

	var nilCtx context.Context
	l := WithContext(context.Background())
	l = l.WithContext(nilCtx)
	l = l.WithContext(ctx)
	SetLevel(ErrorLevel)
	l.WithDuration(time.Second).Error(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Errorf(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Errorv(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Errorw(testlog, Field("basket", "ball"))
	validate(t, w.String(), true, true)
	assert.True(t, strings.Contains(w.String(), "basket"), w.String())
	assert.True(t, strings.Contains(w.String(), "ball"), w.String())
}

func TestTraceInfo(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, span := tp.Tracer("trace-id").Start(context.Background(), "span-id")
	defer span.End()

	SetLevel(InfoLevel)
	l := WithContext(ctx)
	l.WithDuration(time.Second).Info(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Infof(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Infov(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Infow(testlog, Field("basket", "ball"))
	validate(t, w.String(), true, true)
	assert.True(t, strings.Contains(w.String(), "basket"), w.String())
	assert.True(t, strings.Contains(w.String(), "ball"), w.String())
}

func TestTraceInfoConsole(t *testing.T) {
	old := atomic.SwapUint32(&encoding, jsonEncodingType)
	defer atomic.StoreUint32(&encoding, old)

	w := new(mockWriter)
	o := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(o)
	}()

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, span := tp.Tracer("trace-id").Start(context.Background(), "span-id")
	defer span.End()

	l := WithContext(ctx)
	SetLevel(InfoLevel)
	l.WithDuration(time.Second).Info(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Infof(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Infov(testlog)
	validate(t, w.String(), true, true)
}

func TestTraceSlow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, span := tp.Tracer("trace-id").Start(context.Background(), "span-id")
	defer span.End()

	l := WithContext(ctx)
	SetLevel(InfoLevel)
	l.WithDuration(time.Second).Slow(testlog)
	assert.True(t, strings.Contains(w.String(), traceKey))
	assert.True(t, strings.Contains(w.String(), spanKey))
	w.Reset()
	l.WithDuration(time.Second).Slowf(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Slowv(testlog)
	validate(t, w.String(), true, true)
	w.Reset()
	l.WithDuration(time.Second).Sloww(testlog, Field("basket", "ball"))
	validate(t, w.String(), true, true)
	assert.True(t, strings.Contains(w.String(), "basket"), w.String())
	assert.True(t, strings.Contains(w.String(), "ball"), w.String())
}

func TestTraceWithoutContext(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	l := WithContext(context.Background())
	SetLevel(InfoLevel)
	l.WithDuration(time.Second).Info(testlog)
	validate(t, w.String(), false, false)
	w.Reset()
	l.WithDuration(time.Second).Infof(testlog)
	validate(t, w.String(), false, false)
}

func TestLogWithFields(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	writer.lock.RLock()
	defer func() {
		writer.lock.RUnlock()
		writer.Store(old)
	}()

	ctx := WithFields(context.Background(), Field("foo", "bar"))
	l := WithContext(ctx)
	SetLevel(InfoLevel)
	l.Info(testlog)

	var val mockValue
	assert.Nil(t, json.Unmarshal([]byte(w.String()), &val))
	assert.Equal(t, "bar", val.Foo)
}

func validate(t *testing.T, body string, expectedTrace, expectedSpan bool) {
	var val mockValue
	dec := json.NewDecoder(strings.NewReader(body))

	for {
		var doc mockValue
		err := dec.Decode(&doc)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			continue
		}

		val = doc
	}

	assert.Equal(t, expectedTrace, len(val.Trace) > 0, body)
	assert.Equal(t, expectedSpan, len(val.Span) > 0, body)
}

type mockValue struct {
	Trace string `json:"trace"`
	Span  string `json:"span"`
	Foo   string `json:"foo"`
}
