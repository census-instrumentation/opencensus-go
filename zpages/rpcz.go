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
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
)

var (
	mu    sync.Mutex // protects snaps
	snaps = make(map[methodKey]*statSnapshot)

	// viewType lists the views we are interested in for RPC stats.
	// A view's map value indicates whether that view contains data for received
	// RPCs.
	viewType = map[*view.View]bool{
		ocgrpc.ClientCompletedRPCsView:       false,
		ocgrpc.ClientSentBytesPerRPCView:     false,
		ocgrpc.ClientReceivedBytesPerRPCView: false,
		ocgrpc.ClientRoundtripLatencyView:    false,
		ocgrpc.ServerCompletedRPCsView:       true,
		ocgrpc.ServerReceivedBytesPerRPCView: true,
		ocgrpc.ServerSentBytesPerRPCView:     true,
		ocgrpc.ServerLatencyView:             true,
	}
)

func init() {
	views := make([]*view.View, 0, len(viewType))
	for v := range viewType {
		views = append(views, v)
	}
	if err := view.Register(views...); err != nil {
		log.Printf("Error registering views: %v", err)
	}
	view.RegisterExporter(snapExporter{})
}

func rpczHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	WriteHTMLRpczPage(w)
}

// WriteHTMLRpczPage writes an HTML document to w containing per-method RPC stats.
func WriteHTMLRpczPage(w io.Writer) {
	if err := headerTemplate.Execute(w, headerData{Title: "RPC Stats"}); err != nil {
		log.Printf("zpages: executing template: %v", err)
	}
	WriteHTMLRpczSummary(w)
	if err := footerTemplate.Execute(w, nil); err != nil {
		log.Printf("zpages: executing template: %v", err)
	}
}

// WriteHTMLRpczSummary writes HTML to w containing per-method RPC stats.
//
// It includes neither a header nor footer, so you can embed this data in other pages.
func WriteHTMLRpczSummary(w io.Writer) {
	mu.Lock()
	if err := statsTemplate.Execute(w, getStatsPage()); err != nil {
		log.Printf("zpages: executing template: %v", err)
	}
	mu.Unlock()
}

// WriteTextRpczPage writes formatted text to w containing per-method RPC stats.
func WriteTextRpczPage(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	page := getStatsPage()

	for i, sg := range page.StatGroups {
		switch i {
		case 0:
			fmt.Fprint(w, "Sent:\n")
		case 1:
			fmt.Fprint(w, "\nReceived:\n")
		}
		tw := tabwriter.NewWriter(w, 6, 8, 1, ' ', 0)
		fmt.Fprint(tw, "Method\tCount\t\t\tAvgLat\t\t\tRate\t\t\tIn (KiB/s)\t\t\tOut (KiB/s)\t\t\tErrors\t\t\n")
		fmt.Fprint(tw, "\tMin\tHr\tTot\tMin\tHr\tTot\tMin\tHr\tTot\tMin\tHr\tTot\tMin\tHr\tTot\tMin\tHr\tTot\n")
		for _, s := range sg.Snapshots {
			fmt.Fprintln(tw,
				strings.Join([]string{
					s.Method,
					count(&s.LatencyMinute),
					count(&s.LatencyHour),
					totalCount(&s.LatencyHour),
					latency(&s.LatencyMinute),
					latency(&s.LatencyHour),
					totalLatency(&s.LatencyHour),
					velocity(&s.LatencyMinute),
					velocity(&s.LatencyHour),
					totalVelocity(&s.LatencyHour),
					dataRate(&s.InputMinute),
					dataRate(&s.InputHour),
					totalDataRate(&s.InputHour),
					dataRate(&s.OutputMinute),
					dataRate(&s.OutputHour),
					totalDataRate(&s.OutputHour),
					velocity(&s.ErrorsMinute),
					velocity(&s.ErrorsHour),
					totalVelocity(&s.ErrorsHour),
				}, "\t"))
		}
		tw.Flush()
	}
}

// headerData contains data for the header template.
type headerData struct {
	Title string
}

// statsPage aggregates stats on the page for 'sent' and 'received' categories
type statsPage struct {
	StatGroups []*statGroup
}

// statGroup aggregates snapshots for a directional category
type statGroup struct {
	Direction string
	Snapshots []*statSnapshot
}

func (s *statGroup) Len() int {
	return len(s.Snapshots)
}

func (s *statGroup) Swap(i, j int) {
	s.Snapshots[i], s.Snapshots[j] = s.Snapshots[j], s.Snapshots[i]
}

func (s *statGroup) Less(i, j int) bool {
	return s.Snapshots[i].Method < s.Snapshots[j].Method
}

// statSnapshot holds the data items that are presented in a single row of RPC
// stat information.
type statSnapshot struct {
	Method        string
	Received      bool
	LatencyMinute windowStat
	LatencyHour   windowStat
	InputMinute   windowStat
	InputHour     windowStat
	OutputMinute  windowStat
	OutputHour    windowStat
	ErrorsMinute  windowStat
	ErrorsHour    windowStat
}

type methodKey struct {
	method   string
	received bool
}

type snapExporter struct{}

func (s snapExporter) ExportView(vd *view.Data) {
	received, ok := viewType[vd.View]
	if !ok {
		return
	}
	if len(vd.Rows) == 0 {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if vd.View == ocgrpc.ServerCompletedRPCsView || vd.View == ocgrpc.ClientCompletedRPCsView {
		errors := make(map[string]int64)
		for _, row := range vd.Rows {
			method := getMethod(row)
			if method == "" {
				continue
			}
			status := getStatus(row)
			if status != "" && status != "OK" {
				errors[method]++
			}
		}
		for method, errorCount := range errors {
			s := snapshot(method, received)
			dist := &view.DistributionData{
				Mean:  1.0,
				Count: errorCount,
			}
			s.ErrorsHour.update(vd.End, dist)
			s.ErrorsMinute.update(vd.End, dist)
		}
		return
	}

	for _, row := range vd.Rows {
		method := getMethod(row)
		if method == "" {
			continue
		}

		s := snapshot(method, received)

		var (
			dist *view.DistributionData
			ok   bool
		)
		if dist, ok = row.Data.(*view.DistributionData); !ok {
			continue
		}

		var hour, minute *windowStat
		switch vd.View {
		case ocgrpc.ClientRoundtripLatencyView:
			hour = &s.LatencyHour
			minute = &s.LatencyMinute
		case ocgrpc.ClientSentBytesPerRPCView:
			hour = &s.OutputHour
			minute = &s.OutputMinute
		case ocgrpc.ClientReceivedBytesPerRPCView:
			hour = &s.InputHour
			minute = &s.InputMinute
		case ocgrpc.ServerLatencyView:
			hour = &s.LatencyHour
			minute = &s.LatencyMinute
		case ocgrpc.ServerSentBytesPerRPCView:
			hour = &s.OutputHour
			minute = &s.OutputMinute
		case ocgrpc.ServerReceivedBytesPerRPCView:
			hour = &s.InputHour
			minute = &s.InputMinute
		}
		if hour == nil || minute == nil {
			continue
		}
		hour.update(vd.End, dist)
		minute.update(vd.End, dist)
	}

}

func getStatus(row *view.Row) string {
	for _, tag := range row.Tags {
		if tag.Key == ocgrpc.KeyServerStatus || tag.Key == ocgrpc.KeyClientStatus {
			return tag.Value
		}
	}
	return ""
}

func getMethod(row *view.Row) string {
	for _, tag := range row.Tags {
		if tag.Key == ocgrpc.KeyClientMethod || tag.Key == ocgrpc.KeyServerMethod {
			return tag.Value
		}
	}
	return ""
}

func snapshot(method string, received bool) *statSnapshot {
	key := methodKey{method: method, received: received}
	if s := snaps[key]; s != nil {
		return s
	}
	s := &statSnapshot{Method: method, Received: received}
	for _, hourWindow := range []*windowStat{&s.OutputHour, &s.LatencyHour, &s.ErrorsHour, &s.InputHour} {
		hourWindow.init(time.Hour)
	}
	for _, minuteWindow := range []*windowStat{&s.OutputMinute, &s.LatencyMinute, &s.ErrorsMinute, &s.InputMinute} {
		minuteWindow.init(time.Minute)
	}
	snaps[key] = s
	return s
}

func getStatsPage() *statsPage {
	sentStats := statGroup{Direction: "Sent"}
	receivedStats := statGroup{Direction: "Received"}
	for key, sg := range snaps {
		if key.received {
			receivedStats.Snapshots = append(receivedStats.Snapshots, sg)
		} else {
			sentStats.Snapshots = append(sentStats.Snapshots, sg)
		}
	}
	sort.Sort(&sentStats)
	sort.Sort(&receivedStats)

	return &statsPage{
		StatGroups: []*statGroup{&sentStats, &receivedStats},
	}
}
