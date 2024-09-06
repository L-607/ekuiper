// Copyright 2024 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracer

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/lf-edge/ekuiper/v2/internal/conf"
)

type SpanExporter struct {
	remoteSpanExport *otlptrace.Exporter
	LocalSpanStorage LocalSpanStorage
}

func NewSpanExporter(remoteCollector bool) (*SpanExporter, error) {
	s := &SpanExporter{}
	if remoteCollector {
		exporter, err := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(conf.Config.OpenTelemetry.RemoteEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		s.remoteSpanExport = exporter
	}
	s.LocalSpanStorage = newLocalSpanMemoryStorage(conf.Config.OpenTelemetry.LocalTraceCapacity)
	return s, nil
}

func (l *SpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if l == nil {
		return nil
	}
	if l.remoteSpanExport != nil {
		err := l.remoteSpanExport.ExportSpans(ctx, spans)
		if err != nil {
			conf.Log.Warnf("export remote span err: %v", err)
		}
	}
	for _, span := range spans {
		l.LocalSpanStorage.SaveSpan(span)
	}
	return nil
}

func (l *SpanExporter) Shutdown(ctx context.Context) error {
	if l == nil {
		return nil
	}
	if l.remoteSpanExport != nil {
		err := l.remoteSpanExport.Shutdown(ctx)
		if err != nil {
			conf.Log.Warnf("shutdown remote span exporter err: %v", err)
		}
	}
	return nil
}

func (l *SpanExporter) GetTraceById(traceID string) *LocalSpan {
	return l.LocalSpanStorage.GetTraceById(traceID)
}

func (l *SpanExporter) GetTraceByRuleID(ruleID string, limit int) []string {
	if l.LocalSpanStorage == nil {
		return nil
	}
	return l.LocalSpanStorage.GetTraceByRuleID(ruleID, limit)
}

type LocalSpanStorage interface {
	SaveSpan(span sdktrace.ReadOnlySpan) error
	GetTraceById(traceID string) *LocalSpan
	GetTraceByRuleID(ruleID string, limit int) []string
}

type LocalSpanMemoryStorage struct {
	sync.RWMutex
	queue *Queue
	// traceid -> spanid -> span
	m map[string]map[string]*LocalSpan
	// rule -> traceID
	ruleTraceMap map[string]map[string]struct{}
}

func newLocalSpanMemoryStorage(capacity int) *LocalSpanMemoryStorage {
	return &LocalSpanMemoryStorage{
		queue:        NewQueue(capacity),
		ruleTraceMap: make(map[string]map[string]struct{}),
		m:            map[string]map[string]*LocalSpan{},
	}
}

func (l *LocalSpanMemoryStorage) SaveSpan(span sdktrace.ReadOnlySpan) error {
	l.Lock()
	defer l.Unlock()
	localSpan := FromReadonlySpan(span)
	return l.saveSpan(localSpan)
}

func (l *LocalSpanMemoryStorage) saveSpan(localSpan *LocalSpan) error {
	droppedTraceID := l.queue.Enqueue(localSpan)
	if droppedTraceID != "" {
		delete(l.m, droppedTraceID)
	}
	spanMap, ok := l.m[localSpan.TraceID]
	if !ok {
		spanMap = make(map[string]*LocalSpan)
		l.m[localSpan.TraceID] = spanMap
	}
	if len(localSpan.RuleID) > 0 {
		traceMap, ok := l.ruleTraceMap[localSpan.RuleID]
		if !ok {
			traceMap = make(map[string]struct{})
			l.ruleTraceMap[localSpan.RuleID] = traceMap
		}
		traceMap[localSpan.TraceID] = struct{}{}
	}

	spanMap[localSpan.SpanID] = localSpan
	return nil
}

func (l *LocalSpanMemoryStorage) GetTraceById(traceID string) *LocalSpan {
	l.RLock()
	defer l.RUnlock()
	allSpans := l.m[traceID]
	if len(allSpans) < 1 {
		return nil
	}
	rootSpan := findRootSpan(allSpans)
	if rootSpan == nil {
		return nil
	}
	copySpan := make(map[string]*LocalSpan)
	for k, s := range allSpans {
		copySpan[k] = s
	}
	buildSpanLink(rootSpan, copySpan)
	return rootSpan
}

func (l *LocalSpanMemoryStorage) GetTraceByRuleID(ruleID string, limit int) []string {
	l.RLock()
	defer l.RUnlock()
	traceMap := l.ruleTraceMap[ruleID]
	r := make([]string, 0)
	if limit < 1 {
		limit = len(traceMap)
	}
	count := 0
	for traceID := range traceMap {
		r = append(r, traceID)
		count++
		if count >= limit {
			break
		}
	}
	return r
}

func findRootSpan(allSpans map[string]*LocalSpan) *LocalSpan {
	for id1, span1 := range allSpans {
		if span1.ParentSpanID == "" {
			return span1
		}
		isRoot := true
		for id2, span2 := range allSpans {
			if id1 == id2 {
				continue
			}
			if span1.ParentSpanID == span2.SpanID {
				isRoot = false
				break
			}
		}
		if isRoot {
			return span1
		}
	}
	return nil
}

func buildSpanLink(cur *LocalSpan, OtherSpans map[string]*LocalSpan) {
	for k, otherSpan := range OtherSpans {
		if cur.SpanID == otherSpan.ParentSpanID {
			cur.ChildSpan = append(cur.ChildSpan, otherSpan)
			delete(OtherSpans, k)
		}
	}
	for _, span := range cur.ChildSpan {
		buildSpanLink(span, OtherSpans)
	}
}

// Queue is traceID FIFO queue with sized capacity
type Queue struct {
	m        map[string]struct{}
	items    []string
	capacity int
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		m:        make(map[string]struct{}),
		items:    make([]string, 0),
		capacity: capacity,
	}
}

func (q *Queue) Enqueue(item *LocalSpan) string {
	_, ok := q.m[item.TraceID]
	if ok {
		return ""
	}
	dropped := ""
	if len(q.items) >= q.capacity {
		dropped = q.Dequeue()
	}
	q.items = append(q.items, item.TraceID)
	return dropped
}

func (q *Queue) Dequeue() string {
	if len(q.items) == 0 {
		return ""
	}
	traceID := q.items[0]
	q.items = q.items[1:]
	delete(q.m, traceID)
	return traceID
}

func (q *Queue) Len() int {
	return len(q.items)
}