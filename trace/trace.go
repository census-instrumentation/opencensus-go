// Copyright 2017, OpenCensus Authors
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

package trace

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// Span represents a span of a trace.  It has an associated SpanContext, and
// stores data accumulated while the span is active.
//
// Ideally users should interact with Spans by calling the functions in this
// package that take a Context parameter.
type Span struct {
	// data contains information recorded about the span.
	//
	// It will be non-nil if we are exporting the span or recording events for it.
	// Otherwise, data is nil, and the Span is simply a carrier for the
	// SpanContext, so that the trace ID is propagated.
	data        *SpanData
	mu          sync.Mutex // protects the contents of *data (but not the pointer value.)
	spanContext SpanContext
	// spanStore is the spanStore this span belongs to, if any, otherwise it is nil.
	*spanStore
}

// IsRecordingEvents returns true if events are being recorded for the current span.
func IsRecordingEvents(ctx context.Context) bool {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return false
	}
	return s.IsRecordingEvents()
}

// IsRecordingEvents returns true if events are being recorded for this span.
func (s *Span) IsRecordingEvents() bool {
	if s == nil {
		return false
	}
	return s.data != nil
}

// TraceOptions contains options associated with a trace span.
type TraceOptions uint32

// IsSampled returns true if the current span will be exported.
func IsSampled(ctx context.Context) bool {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return false
	}
	return s.IsSampled()
}

// IsSampled returns true if this span will be exported.
func (s *Span) IsSampled() bool {
	if s == nil {
		return false
	}
	return s.spanContext.IsSampled()
}

// IsSampled returns true if the span will be exported.
func (sc SpanContext) IsSampled() bool {
	return sc.TraceOptions.IsSampled()
}

// setIsSampled sets the TraceOptions bit that determines whether the span will be exported.
func (sc *SpanContext) setIsSampled(sampled bool) {
	if sampled {
		sc.TraceOptions |= 1
	} else {
		sc.TraceOptions &= ^TraceOptions(1)
	}
}

// IsSampled returns true if the span will be exported.
func (t TraceOptions) IsSampled() bool {
	return t&1 == 1
}

// SpanContext contains the state that must propagate across process boundaries.
//
// SpanContext is not an implementation of context.Context.
// TODO: add reference to external Census docs for SpanContext.
type SpanContext struct {
	TraceID
	SpanID
	TraceOptions
}

type contextKey struct{}

// FromContext returns the Span stored in a context, or nil if there isn't one.
func FromContext(ctx context.Context) *Span {
	s, _ := ctx.Value(contextKey{}).(*Span)
	return s
}

// WithSpan returns a new context with the given Span attached.
func WithSpan(parent context.Context, s *Span) context.Context {
	return context.WithValue(parent, contextKey{}, s)
}

// StartSpanOptions contains options concerning how a span is started.
type StartSpanOptions struct {
	// RecordEvents indicates whether to record data for this span, and include
	// the span in a local span store.
	// Events will also be recorded if the span will be exported.
	RecordEvents bool
	Sampler      // if non-nil, the Sampler to consult for this span.
	// RegisterNameForLocalSpanStore indicates that a local span store for spans
	// of this name should be created, if one does not exist.
	// If RecordEvents is false, this option has no effect.
	RegisterNameForLocalSpanStore bool
}

// StartSpan starts a new child span of the current span in the context.
//
// If there is no span in the context, creates a new trace and span.
func StartSpan(ctx context.Context, name string) context.Context {
	parentSpan, _ := ctx.Value(contextKey{}).(*Span)
	return WithSpan(ctx, parentSpan.StartSpanWithOptions(name, StartSpanOptions{}))
}

// StartSpanWithOptions starts a new child span of the current span in the context.
//
// If there is no span in the context, creates a new trace and span.
func StartSpanWithOptions(ctx context.Context, name string, o StartSpanOptions) context.Context {
	parentSpan, _ := ctx.Value(contextKey{}).(*Span)
	return WithSpan(ctx, parentSpan.StartSpanWithOptions(name, o))
}

// StartSpanWithRemoteParent starts a new child span with the given parent SpanContext.
//
// If there is an existing span in ctx, it is ignored -- the returned Span is a
// child of the span specified by parent.
func StartSpanWithRemoteParent(ctx context.Context, name string, parent SpanContext, o StartSpanOptions) context.Context {
	return WithSpan(ctx, NewSpanWithRemoteParent(name, parent, o))
}

// StartSpan starts a new child span.
//
// If s is nil, creates a new trace and span, like the function NewSpan.
func (s *Span) StartSpan(name string) *Span {
	return s.StartSpanWithOptions(name, StartSpanOptions{})
}

// StartSpanWithOptions starts a new child span using the given options.
//
// If s is nil, creates a new trace and span, like the function NewSpan.
func (s *Span) StartSpanWithOptions(name string, o StartSpanOptions) *Span {
	if s != nil {
		return startSpanInternal(name, true, s.spanContext, false, o)
	}
	return startSpanInternal(name, false, SpanContext{}, false, o)
}

// NewSpan returns a new span.
//
// The returned span has no parent span; a new trace ID will be created for it.
func NewSpan(name string, o StartSpanOptions) *Span {
	return startSpanInternal(name, false, SpanContext{}, false, o)
}

// NewSpanWithRemoteParent returns a new span with the given parent SpanContext.
func NewSpanWithRemoteParent(name string, parent SpanContext, o StartSpanOptions) *Span {
	return startSpanInternal(name, true, parent, true, o)
}

func startSpanInternal(name string, hasParent bool, parent SpanContext, remoteParent bool, o StartSpanOptions) *Span {
	span := &Span{}
	span.spanContext = parent
	mu.Lock()
	if !hasParent {
		span.spanContext.TraceID = newTraceIDLocked()
	}
	span.spanContext.SpanID = newSpanIDLocked()
	sampler := defaultSampler
	mu.Unlock()

	if !hasParent || remoteParent || o.Sampler != nil {
		// If this span is the child of a local span and no Sampler is set in the
		// options, keep the parent's TraceOptions.
		//
		// Otherwise, consult the Sampler in the options if it is non-nil, otherwise
		// the default sampler.
		if o.Sampler != nil {
			sampler = o.Sampler
		}
		span.spanContext.setIsSampled(sampler.Sample(SamplingParameters{
			ParentContext:   parent,
			TraceID:         span.spanContext.TraceID,
			SpanID:          span.spanContext.SpanID,
			Name:            name,
			HasRemoteParent: remoteParent}).Sample)
	}

	if !o.RecordEvents && !span.spanContext.IsSampled() {
		return span
	}

	span.data = &SpanData{
		SpanContext:     span.spanContext,
		StartTime:       time.Now(),
		Name:            name,
		HasRemoteParent: remoteParent,
	}
	if hasParent {
		span.data.ParentSpanID = parent.SpanID
	}
	if o.RecordEvents {
		var ss *spanStore
		if o.RegisterNameForLocalSpanStore {
			ss = spanStoreForNameCreateIfNew(name)
		} else {
			ss = spanStoreForName(name)
		}
		if ss != nil {
			span.spanStore = ss
			ss.add(span)
		}
	}

	return span
}

// EndSpan ends the current span.
//
// The context passed to EndSpan will still refer to the now-ended span, so any
// code that adds more information to it, like SetSpanStatus, should be called
// before EndSpan.
func EndSpan(ctx context.Context) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.End()
}

// End ends the span.
func (s *Span) End() {
	if !s.IsRecordingEvents() {
		return
	}
	// TODO: optimize to avoid this call if sd won't be used.
	sd := s.makeSpanData()
	sd.EndTime = time.Now()
	if s.spanStore != nil {
		s.spanStore.finished(s, sd)
	}
	if s.spanContext.IsSampled() {
		// TODO: consider holding exportersMu for less time.
		exportersMu.Lock()
		for e := range exporters {
			e.Export(sd)
		}
		exportersMu.Unlock()
	}
}

// makeSpanData produces a SpanData representing the current state of the Span.
// It requires that s.data is non-nil.
func (s *Span) makeSpanData() *SpanData {
	var sd SpanData
	s.mu.Lock()
	sd = *s.data
	if s.data.Attributes != nil {
		sd.Attributes = make(map[string]interface{})
		for k, v := range s.data.Attributes {
			sd.Attributes[k] = v
		}
	}
	s.mu.Unlock()
	return &sd
}

// SpanContextFromContext returns the SpanContext of the current span, if there
// is one.
func SpanContextFromContext(ctx context.Context) (SpanContext, bool) {
	s, _ := ctx.Value(contextKey{}).(*Span)
	if s == nil {
		return SpanContext{}, false
	}
	return s.SpanContext(), true
}

// SpanContext returns the SpanContext of the span.
func (s *Span) SpanContext() SpanContext {
	if s == nil {
		return SpanContext{}
	}
	return s.spanContext
}

// SetSpanStatus sets the status of the current span, if it is recording events.
func SetSpanStatus(ctx context.Context, status Status) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.SetStatus(status)
}

// SetStatus sets the status of the span, if it is recording events.
func (s *Span) SetStatus(status Status) {
	if !s.IsRecordingEvents() {
		return
	}
	s.mu.Lock()
	s.data.Status = status
	s.mu.Unlock()
}

// SetSpanAttributes sets attributes in the current span.
//
// Existing attributes whose keys appear in the attributes parameter are overwritten.
func SetSpanAttributes(ctx context.Context, attributes ...Attribute) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.SetAttributes(attributes...)
}

// SetAttributes sets attributes in the span.
//
// Existing attributes whose keys appear in the attributes parameter are overwritten.
func (s *Span) SetAttributes(attributes ...Attribute) {
	if !s.IsRecordingEvents() {
		return
	}
	s.mu.Lock()
	if s.data.Attributes == nil {
		s.data.Attributes = make(map[string]interface{})
	}
	copyAttributes(s.data.Attributes, attributes)
	s.mu.Unlock()
}

// copyAttributes copies a slice of Attributes into a map.
func copyAttributes(m map[string]interface{}, attributes []Attribute) {
	for _, a := range attributes {
		switch a := a.(type) {
		case BoolAttribute:
			m[a.Key] = a.Value
		case Int64Attribute:
			m[a.Key] = a.Value
		case StringAttribute:
			m[a.Key] = a.Value
		}
	}
}

// LazyPrint adds an annotation to the current span using a fmt.Stringer.
//
// str.String is called only when the annotation text is needed.
func LazyPrint(ctx context.Context, str fmt.Stringer) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.LazyPrint(str)
}

// LazyPrint adds an annotation using a fmt.Stringer.
//
// str.String is called only when the annotation text is needed.
func (s *Span) LazyPrint(str fmt.Stringer) {
	if !s.IsRecordingEvents() {
		return
	}
	s.lazyPrintInternal(nil, str)
}

// LazyPrintWithAttributes adds an annotation with attributes to the current span using a fmt.Stringer.
//
// str.String is called only when the annotation text is needed.
func LazyPrintWithAttributes(ctx context.Context, attributes []Attribute, str fmt.Stringer) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.LazyPrintWithAttributes(attributes, str)
}

// LazyPrintWithAttributes adds an annotation with attributes using a fmt.Stringer.
//
// str.String is called only when the annotation text is needed.
func (s *Span) LazyPrintWithAttributes(attributes []Attribute, str fmt.Stringer) {
	if !s.IsRecordingEvents() {
		return
	}
	s.lazyPrintInternal(attributes, str)
}

func (s *Span) lazyPrintInternal(attributes []Attribute, str fmt.Stringer) {
	now := time.Now()
	msg := str.String()
	var a map[string]interface{}
	s.mu.Lock()
	if len(attributes) != 0 {
		a = make(map[string]interface{})
		copyAttributes(a, attributes)
	}
	s.data.Annotations = append(s.data.Annotations, Annotation{
		Time:       now,
		Message:    msg,
		Attributes: a,
	})
	s.mu.Unlock()
}

// LazyPrintf adds an annotation to the current span.
//
// The format string is evaluated with its arguments only when the annotation text is needed.
func LazyPrintf(ctx context.Context, format string, a ...interface{}) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.LazyPrintf(format, a...)
}

// LazyPrintf adds an annotation.
//
// The format string is evaluated with its arguments only when the annotation text is needed.
func (s *Span) LazyPrintf(format string, a ...interface{}) {
	if !s.IsRecordingEvents() {
		return
	}
	s.lazyPrintfInternal(nil, format, a...)
}

func (s *Span) lazyPrintfInternal(attributes []Attribute, format string, a ...interface{}) {
	now := time.Now()
	msg := fmt.Sprintf(format, a...)
	var m map[string]interface{}
	s.mu.Lock()
	if len(attributes) != 0 {
		m = make(map[string]interface{})
		copyAttributes(m, attributes)
	}
	s.data.Annotations = append(s.data.Annotations, Annotation{
		Time:       now,
		Message:    msg,
		Attributes: m,
	})
	s.mu.Unlock()
}

// LazyPrintfWithAttributes adds an annotation with attributes to the current span.
//
// The format string is evaluated with its arguments only when the annotation text is needed.
func LazyPrintfWithAttributes(ctx context.Context, attributes []Attribute, format string, a ...interface{}) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.LazyPrintfWithAttributes(attributes, format, a...)
}

// LazyPrintfWithAttributes adds an annotation with attributes.
//
// The format string is evaluated with its arguments only when the annotation text is needed.
func (s *Span) LazyPrintfWithAttributes(attributes []Attribute, format string, a ...interface{}) {
	if !s.IsRecordingEvents() {
		return
	}
	s.lazyPrintfInternal(attributes, format, a...)
}

// Print adds an annotation to the current span.
func Print(ctx context.Context, str string) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.Print(str)
}

// Print adds an annotation.
func (s *Span) Print(str string) {
	if !s.IsRecordingEvents() {
		return
	}
	s.printStringInternal(nil, str)
}

func (s *Span) printStringInternal(attributes []Attribute, str string) {
	now := time.Now()
	var a map[string]interface{}
	s.mu.Lock()
	if len(attributes) != 0 {
		a = make(map[string]interface{})
		copyAttributes(a, attributes)
	}
	s.data.Annotations = append(s.data.Annotations, Annotation{
		Time:       now,
		Message:    str,
		Attributes: a,
	})
	s.mu.Unlock()
}

// PrintWithAttributes adds an annotation with attributes to the current span.
func PrintWithAttributes(ctx context.Context, attributes []Attribute, str string) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.PrintWithAttributes(attributes, str)
}

// PrintWithAttributes adds an annotation with attributes.
func (s *Span) PrintWithAttributes(attributes []Attribute, str string) {
	if !s.IsRecordingEvents() {
		return
	}
	s.printStringInternal(attributes, str)
}

// SetStackTrace adds a stack trace to the current span.
func SetStackTrace(ctx context.Context) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.SetStackTrace()
}

// SetStackTrace adds a stack trace to the span.
func (s *Span) SetStackTrace() {
	if !s.IsRecordingEvents() {
		return
	}
	pcs := make([]uintptr, 20 /* TODO: configurable stack size? */)
	_ = runtime.Callers(1, pcs[:])
	s.mu.Lock()
	s.data.StackTrace = pcs
	s.mu.Unlock()
}

// AddMessageSendEvent adds a message send event to the current span.
//
// messageID is an identifier for the message, which is recommended to be
// unique in this span and the same between the send event and the receive
// event (this allows to identify a message between the sender and receiver).
// For example, this could be a sequence id.
func AddMessageSendEvent(ctx context.Context, messageID, uncompressedByteSize, compressedByteSize int64) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize)
}

// AddMessageSendEvent adds a message send event to the span.
//
// messageID is an identifier for the message, which is recommended to be
// unique in this span and the same between the send event and the receive
// event (this allows to identify a message between the sender and receiver).
// For example, this could be a sequence id.
func (s *Span) AddMessageSendEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	if !s.IsRecordingEvents() {
		return
	}
	now := time.Now()
	s.mu.Lock()
	s.data.MessageEvents = append(s.data.MessageEvents, MessageEvent{
		Time:                 now,
		EventType:            MessageEventTypeSent,
		MessageID:            messageID,
		UncompressedByteSize: uncompressedByteSize,
		CompressedByteSize:   compressedByteSize,
	})
	s.mu.Unlock()
}

// AddMessageReceiveEvent adds a message receive event to the current span.
//
// messageID is an identifier for the message, which is recommended to be
// unique in this span and the same between the send event and the receive
// event (this allows to identify a message between the sender and receiver).
// For example, this could be a sequence id.
func AddMessageReceiveEvent(ctx context.Context, messageID, uncompressedByteSize, compressedByteSize int64) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize)
}

// AddMessageReceiveEvent adds a message receive event to the span.
//
// messageID is an identifier for the message, which is recommended to be
// unique in this span and the same between the send event and the receive
// event (this allows to identify a message between the sender and receiver).
// For example, this could be a sequence id.
func (s *Span) AddMessageReceiveEvent(messageID, uncompressedByteSize, compressedByteSize int64) {
	if !s.IsRecordingEvents() {
		return
	}
	now := time.Now()
	s.mu.Lock()
	s.data.MessageEvents = append(s.data.MessageEvents, MessageEvent{
		Time:                 now,
		EventType:            MessageEventTypeRecv,
		MessageID:            messageID,
		UncompressedByteSize: uncompressedByteSize,
		CompressedByteSize:   compressedByteSize,
	})
	s.mu.Unlock()
}

// AddLink adds a link to the current span.
func AddLink(ctx context.Context, l Link) {
	s, ok := ctx.Value(contextKey{}).(*Span)
	if !ok {
		return
	}
	s.AddLink(l)
}

// AddLink adds a link to the span.
func (s *Span) AddLink(l Link) {
	if !s.IsRecordingEvents() {
		return
	}
	s.mu.Lock()
	s.data.Links = append(s.data.Links, l)
	s.mu.Unlock()
}

func (s *Span) String() string {
	if s == nil {
		return "<nil>"
	}
	if s.data == nil {
		return fmt.Sprintf("span %s", s.spanContext.SpanID)
	}
	s.mu.Lock()
	str := fmt.Sprintf("span %s %q", s.spanContext.SpanID, s.data.Name)
	s.mu.Unlock()
	return str
}

var (
	mu          sync.Mutex // protects the variables below
	traceIDRand *rand.Rand
	traceIDAdd  [2]uint64
	nextSpanID  uint64
	spanIDInc   uint64
)

func init() {
	// initialize traceID and spanID generators.
	var rngSeed int64
	for _, p := range []interface{}{
		&rngSeed, &traceIDAdd, &nextSpanID, &spanIDInc,
	} {
		binary.Read(crand.Reader, binary.LittleEndian, p)
	}
	traceIDRand = rand.New(rand.NewSource(rngSeed))
	spanIDInc |= 1
}

// newSpanIDLocked returns a non-zero SpanID from a randomly-chosen sequence.
// mu should be held while this function is called.
func newSpanIDLocked() SpanID {
	id := nextSpanID
	nextSpanID += spanIDInc
	if nextSpanID == 0 {
		nextSpanID += spanIDInc
	}
	var sid SpanID
	binary.LittleEndian.PutUint64(sid[:], id)
	return sid
}

// newTraceIDLocked returns a non-zero TraceID from a randomly-chosen sequence.
// mu should be held while this function is called.
func newTraceIDLocked() TraceID {
	var tid TraceID
	// Construct the trace ID from two outputs of traceIDRand, with a constant
	// added to each half for additional entropy.
	binary.LittleEndian.PutUint64(tid[0:8], traceIDRand.Uint64()+traceIDAdd[0])
	binary.LittleEndian.PutUint64(tid[8:16], traceIDRand.Uint64()+traceIDAdd[1])
	return tid
}
