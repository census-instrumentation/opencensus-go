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

package prometheus

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/prometheus/client_golang/prometheus"
)

type mSlice []*stats.Int64Measure

func (measures *mSlice) createAndAppend(name, desc, unit string) {
	m := stats.Int64(name, desc, unit)
	*measures = append(*measures, m)
}

type vCreator []*view.View

func (vc *vCreator) createAndAppend(name, description string, keys []tag.Key, measure stats.Measure, agg *view.Aggregation) {
	v := &view.View{
		Name:        name,
		Description: description,
		TagKeys:     keys,
		Measure:     measure,
		Aggregation: agg,
	}
	*vc = append(*vc, v)
}

func TestMetricsEndpointOutput(t *testing.T) {
	exporter, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}

	names := []string{"foo", "bar", "baz"}

	var measures mSlice
	for _, name := range names {
		measures.createAndAppend("tests/"+name, name, "")
	}

	var vc vCreator
	for _, m := range measures {
		vc.createAndAppend(m.Name(), m.Description(), nil, m, view.Count())
	}

	if err := view.Register(vc...); err != nil {
		t.Fatalf("failed to create views: %v", err)
	}
	defer view.Unregister(vc...)

	view.SetReportingPeriod(time.Millisecond)

	for _, m := range measures {
		stats.Record(context.Background(), m.M(1))
	}

	srv := httptest.NewServer(exporter)
	defer srv.Close()

	var i int
	var output string
	for {
		time.Sleep(10 * time.Millisecond)
		if i == 1000 {
			t.Fatal("no output at /metrics (10s wait)")
		}
		i++

		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatalf("failed to get /metrics: %v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		resp.Body.Close()

		output = string(body)
		if output != "" {
			break
		}
	}

	if strings.Contains(output, "collected before with the same name and label values") {
		t.Fatal("metric name and labels being duplicated but must be unique")
	}

	if strings.Contains(output, "error(s) occurred") {
		t.Fatal("error reported by prometheus registry")
	}

	for _, name := range names {
		if !strings.Contains(output, "tests_"+name+" 1") {
			t.Fatalf("measurement missing in output: %v", name)
		}
	}
}

func TestCumulativenessFromHistograms(t *testing.T) {
	exporter, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}

	m := stats.Float64("tests/bills", "payments by denomination", stats.UnitDimensionless)
	v := &view.View{
		Name:        "cash/register",
		Description: "this is a test",
		Measure:     m,

		// Intentionally used repeated elements in the ascending distribution.
		// to ensure duplicate distribution items are handles.
		Aggregation: view.Distribution(1, 5, 5, 5, 5, 10, 20, 50, 100, 250),
	}

	if err := view.Register(v); err != nil {
		t.Fatalf("Register error: %v", err)
	}
	defer view.Unregister(v)

	// Give the reporter ample time to process registration
	//<-time.After(10 * reportPeriod)

	values := []float64{0.25, 245.67, 12, 1.45, 199.9, 7.69, 187.12}
	// We want the results that look like this:
	// 1:   [0.25]      		| 1 + prev(i) = 1 + 0 = 1
	// 5:   [1.45]			| 1 + prev(i) = 1 + 1 = 2
	// 10:	[7.69]			| 1 + prev(i) = 1 + 2 = 3
	// 20:  [12]			| 1 + prev(i) = 1 + 3 = 4
	// 50:  []			| 0 + prev(i) = 0 + 4 = 4
	// 100: []			| 0 + prev(i) = 0 + 4 = 4
	// 250: [187.12, 199.9, 245.67]	| 3 + prev(i) = 3 + 4 = 7
	wantLines := []string{
		`cash_register_bucket{le="1"} 1`,
		`cash_register_bucket{le="5"} 2`,
		`cash_register_bucket{le="10"} 3`,
		`cash_register_bucket{le="20"} 4`,
		`cash_register_bucket{le="50"} 4`,
		`cash_register_bucket{le="100"} 4`,
		`cash_register_bucket{le="250"} 7`,
		`cash_register_bucket{le="+Inf"} 7`,
		`cash_register_sum 654.0799999999999`, // Summation of the input values
		`cash_register_count 7`,
	}

	ctx := context.Background()
	ms := make([]stats.Measurement, 0, len(values))
	for _, value := range values {
		mx := m.M(value)
		ms = append(ms, mx)
	}
	stats.Record(ctx, ms...)

	// Give the recorder ample time to process recording
	//<-time.After(10 * reportPeriod)

	cst := httptest.NewServer(exporter)
	defer cst.Close()
	res, err := http.Get(cst.URL)
	if err != nil {
		t.Fatalf("http.Get error: %v", err)
	}
	blob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Read body error: %v", err)
	}
	str := strings.Trim(string(blob), "\n")
	lines := strings.Split(str, "\n")
	nonComments := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, "#") {
			nonComments = append(nonComments, line)
		}
	}

	got := strings.Join(nonComments, "\n")
	want := strings.Join(wantLines, "\n")
	if got != want {
		t.Fatalf("\ngot:\n%s\n\nwant:\n%s\n", got, want)
	}
}

func TestHistogramUnorderedBucketBounds(t *testing.T) {
	exporter, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}

	m := stats.Float64("tests/bills", "payments by denomination", stats.UnitDimensionless)
	v := &view.View{
		Name:        "cash/register",
		Description: "this is a test",
		Measure:     m,

		// Intentionally used unordered and duplicated elements in the distribution
		// to ensure unordered bucket bounds are handled.
		Aggregation: view.Distribution(10, 5, 1, 1, 50, 5, 20, 100, 250),
	}

	if err := view.Register(v); err != nil {
		t.Fatalf("Register error: %v", err)
	}
	defer view.Unregister(v)

	// Give the reporter ample time to process registration
	//<-time.After(10 * reportPeriod)

	values := []float64{0.25, 245.67, 12, 1.45, 199.9, 7.69, 187.12}
	// We want the results that look like this:
	// 1:   [0.25]      		| 1 + prev(i) = 1 + 0 = 1
	// 5:   [1.45]			| 1 + prev(i) = 1 + 1 = 2
	// 10:	[7.69]			| 1 + prev(i) = 1 + 2 = 3
	// 20:  [12]			| 1 + prev(i) = 1 + 3 = 4
	// 50:  []			| 0 + prev(i) = 0 + 4 = 4
	// 100: []			| 0 + prev(i) = 0 + 4 = 4
	// 250: [187.12, 199.9, 245.67]	| 3 + prev(i) = 3 + 4 = 7
	wantLines := []string{
		`cash_register_bucket{le="1"} 1`,
		`cash_register_bucket{le="5"} 2`,
		`cash_register_bucket{le="10"} 3`,
		`cash_register_bucket{le="20"} 4`,
		`cash_register_bucket{le="50"} 4`,
		`cash_register_bucket{le="100"} 4`,
		`cash_register_bucket{le="250"} 7`,
		`cash_register_bucket{le="+Inf"} 7`,
		`cash_register_sum 654.0799999999999`, // Summation of the input values
		`cash_register_count 7`,
	}

	ctx := context.Background()
	ms := make([]stats.Measurement, 0, len(values))
	for _, value := range values {
		mx := m.M(value)
		ms = append(ms, mx)
	}
	stats.Record(ctx, ms...)

	// Give the recorder ample time to process recording
	//<-time.After(10 * reportPeriod)

	cst := httptest.NewServer(exporter)
	defer cst.Close()
	res, err := http.Get(cst.URL)
	if err != nil {
		t.Fatalf("http.Get error: %v", err)
	}
	blob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Read body error: %v", err)
	}
	str := strings.Trim(string(blob), "\n")
	lines := strings.Split(str, "\n")
	nonComments := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, "#") {
			nonComments = append(nonComments, line)
		}
	}

	got := strings.Join(nonComments, "\n")
	want := strings.Join(wantLines, "\n")
	if got != want {
		t.Fatalf("\ngot:\n%s\n\nwant:\n%s\n", got, want)
	}
}

func TestConstLabelsIncluded(t *testing.T) {
	constLabels := prometheus.Labels{
		"service": "spanner",
	}
	measureLabel, _ := tag.NewKey("method")

	exporter, err := NewExporter(Options{
		ConstLabels: constLabels,
	})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}

	names := []string{"foo", "bar", "baz"}

	var measures mSlice
	for _, name := range names {
		measures.createAndAppend("tests/"+name, name, "")
	}

	var vc vCreator
	for _, m := range measures {
		vc.createAndAppend(m.Name(), m.Description(), []tag.Key{measureLabel}, m, view.Count())
	}

	if err := view.Register(vc...); err != nil {
		t.Fatalf("failed to create views: %v", err)
	}
	defer view.Unregister(vc...)

	view.SetReportingPeriod(time.Millisecond)

	ctx, _ := tag.New(context.Background(), tag.Upsert(measureLabel, "issue961"))
	for _, m := range measures {
		stats.Record(ctx, m.M(1))
	}

	srv := httptest.NewServer(exporter)
	defer srv.Close()

	var i int
	var output string
	for {
		time.Sleep(10 * time.Millisecond)
		if i == 1000 {
			t.Fatal("no output at /metrics (10s wait)")
		}
		i++

		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatalf("failed to get /metrics: %v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		resp.Body.Close()

		output = string(body)
		if output != "" {
			break
		}
	}

	if strings.Contains(output, "collected before with the same name and label values") {
		t.Fatal("metric name and labels being duplicated but must be unique")
	}

	if strings.Contains(output, "error(s) occurred") {
		t.Fatal("error reported by prometheus registry")
	}

	want := `# HELP tests_bar bar
# TYPE tests_bar counter
tests_bar{method="issue961",service="spanner"} 1
# HELP tests_baz baz
# TYPE tests_baz counter
tests_baz{method="issue961",service="spanner"} 1
# HELP tests_foo foo
# TYPE tests_foo counter
tests_foo{method="issue961",service="spanner"} 1
`
	if output != want {
		t.Fatal("output differed from expected")
	}
}

func TestViewMeasureWithoutTag(t *testing.T) {
	exporter, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}
	m := stats.Int64("tests/foo", "foo", stats.UnitDimensionless)
	k1, _ := tag.NewKey("key/1")
	k2, _ := tag.NewKey("key/2")
	k3, _ := tag.NewKey("key/3")
	k4, _ := tag.NewKey("key/4")
	k5, _ := tag.NewKey("key/5")
	randomKey, _ := tag.NewKey("issue659")
	v := &view.View{
		Name:        m.Name(),
		Description: m.Description(),
		TagKeys:     []tag.Key{k2, k5, k3, k1, k4}, // Ensure view has a tag
		Measure:     m,
		Aggregation: view.Count(),
	}
	if err := view.Register(v); err != nil {
		t.Fatalf("failed to create views: %v", err)
	}
	defer view.Unregister(v)
	view.SetReportingPeriod(time.Millisecond)
	// Make a measure without some tags in the view.
	ctx1, _ := tag.New(context.Background(), tag.Upsert(k4, "issue659"), tag.Upsert(randomKey, "value"), tag.Upsert(k2, "issue659"))
	stats.Record(ctx1, m.M(1))
	ctx2, _ := tag.New(context.Background(), tag.Upsert(k5, "issue659"), tag.Upsert(k3, "issue659"), tag.Upsert(k1, "issue659"))
	stats.Record(ctx2, m.M(2))
	srv := httptest.NewServer(exporter)
	defer srv.Close()
	var i int
	var output string
	for {
		time.Sleep(10 * time.Millisecond)
		if i == 1000 {
			t.Fatal("no output at /metrics (10s wait)")
		}
		i++
		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatalf("failed to get /metrics: %v", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		resp.Body.Close()
		output = string(body)
		if output != "" {
			break
		}
	}
	if strings.Contains(output, "collected before with the same name and label values") {
		t.Fatal("metric name and labels being duplicated but must be unique")
	}
	if strings.Contains(output, "error(s) occurred") {
		t.Fatal("error reported by prometheus registry")
	}
	want := `# HELP tests_foo foo
# TYPE tests_foo counter
tests_foo{key_1="",key_2="issue659",key_3="",key_4="issue659",key_5=""} 1
tests_foo{key_1="issue659",key_2="",key_3="issue659",key_4="",key_5="issue659"} 1
`
	if output != want {
		t.Fatalf("output differed from expected output: %s want: %s", output, want)
	}
}
