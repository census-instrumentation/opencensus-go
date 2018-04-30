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
//

package zpages

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"go.opencensus.io/trace"
	"go.opencensus.io/zpages/internal"
)

var (
	fs                = internal.FS(false)
	templateFunctions = template.FuncMap{
		"count":         count,
		"totalCount":    totalCount,
		"ms":            latency,
		"totalMs":       totalLatency,
		"velocity":      velocity,
		"totalVelocity": totalVelocity,
		"dataRate":      dataRate,
		"totalDataRate": totalDataRate,
		"even":          even,
		"traceid":       traceIDFormatter,
	}
	headerTemplate       = parseTemplate("header")
	summaryTableTemplate = parseTemplate("summary")
	statsTemplate        = parseTemplate("rpcz")
	tracesTableTemplate  = parseTemplate("traces")
	footerTemplate       = parseTemplate("footer")
)

func parseTemplate(name string) *template.Template {
	f, err := fs.Open("/templates/" + name + ".html")
	if err != nil {
		log.Panicf("%v: %v", name, err)
	}
	defer f.Close()
	text, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panicf("%v: %v", name, err)
	}
	return template.Must(template.New(name).Funcs(templateFunctions).Parse(string(text)))
}

func countFormatter(num int64) string {
	var floatVal float64
	var suffix string
	if num >= 1e12 {
		floatVal = float64(num) / 1e9
		suffix = " T "
	} else if num >= 1e9 {
		floatVal = float64(num) / 1e9
		suffix = " G "
	} else if num >= 1e6 {
		floatVal = float64(num) / 1e6
		suffix = " M "
	}

	if floatVal != 0 {
		return fmt.Sprintf("%1.3f%s", floatVal, suffix)
	}
	return fmt.Sprint(num)
}

func msFormatter(latency float64) string {
	if latency < 10.0 {
		return fmt.Sprintf("%.2f", latency)
	}
	return strconv.Itoa(int(latency))
}

func rateFormatter(r float64) string {
	return fmt.Sprintf("%.3f", r)
}

func dataRateFormatter(b float64) string {
	return fmt.Sprintf("%.3f", b/1e3)
}

func traceIDFormatter(r traceRow) template.HTML {
	sc := r.SpanContext
	if sc == (trace.SpanContext{}) {
		return ""
	}
	col := "black"
	if sc.TraceOptions.IsSampled() {
		col = "blue"
	}
	if r.ParentSpanID != (trace.SpanID{}) {
		return template.HTML(fmt.Sprintf(`trace_id: <b style="color:%s">%s</b> span_id: %s parent_span_id: %s`, col, sc.TraceID, sc.SpanID, r.ParentSpanID))
	}
	return template.HTML(fmt.Sprintf(`trace_id: <b style="color:%s">%s</b> span_id: %s`, col, sc.TraceID, sc.SpanID))
}

func even(x int) bool {
	return x%2 == 0
}

func latency(ws *windowStat) string {
	_, _, diff := ws.read()
	if diff == nil {
		return "-"
	}
	return msFormatter(diff.Mean)
}

func totalLatency(ws *windowStat) string {
	if ws.lastUpdate == -1 {
		return "-"
	}
	return msFormatter(ws.intervals[ws.lastUpdate].distribution.Mean)
}

func count(ws *windowStat) string {
	_, _, diff := ws.read()
	if diff == nil {
		return countFormatter(0)
	}
	return countFormatter(diff.Count)
}

func totalCount(ws *windowStat) string {
	if ws.lastUpdate == -1 {
		return countFormatter(0)
	}
	return countFormatter(ws.intervals[ws.lastUpdate].distribution.Count)
}

func dataRate(ws *windowStat) string {
	start, end, diff := ws.read()
	if diff == nil {
		return dataRateFormatter(0)
	}
	seconds := float64(end.Sub(start) / time.Second)
	if seconds == 0 {
		return dataRateFormatter(0)
	}
	return dataRateFormatter(float64(diff.Sum()) / seconds)
}

func totalDataRate(ws *windowStat) string {
	if ws.lastUpdate == -1 {
		return dataRateFormatter(0)
	}
	last := ws.intervals[ws.lastUpdate]
	seconds := float64(last.updateTime.Sub(ws.startTime) * time.Second)
	if seconds == 0 {
		return dataRateFormatter(0)
	}
	return dataRateFormatter(float64(last.distribution.Sum()) / seconds)
}

func velocity(ws *windowStat) string {
	start, end, diff := ws.read()
	if diff == nil {
		return "-"
	}
	seconds := float64(end.Sub(start)) / float64(time.Second)
	if seconds == 0 {
		return rateFormatter(0)
	}
	rate := float64(diff.Count) / seconds
	return rateFormatter(rate)
}

func totalVelocity(ws *windowStat) string {
	if ws.lastUpdate == -1 {
		return "-"
	}
	last := ws.intervals[ws.lastUpdate]
	seconds := float64(last.updateTime.Sub(ws.startTime)) / float64(time.Second)
	if seconds == 0 {
		return rateFormatter(0)
	}
	rate := float64(last.distribution.Count) / seconds
	return dataRateFormatter(rate)
}
