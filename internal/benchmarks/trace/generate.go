package trace

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.opencensus.io/internal/benchmarks/proto"
	"go.opencensus.io/plugin/ocgrpc"

	"github.com/odeke-em/cli-spinner"
)

var addr = ":8796"

var tmpl = template.Must(template.New("chartIt.html").
	Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			blob, _ := json.Marshal(v)
			return string(blob)
		},
	}).Parse(chartHTML))

func Generate(dir string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listening on addr %q err: %v", addr, err)
	}
	srv := grpc.NewServer(grpc.StatsHandler(ocgrpc.NewServerStatsHandler()))
	proto.RegisterPingServer(srv, new(server))
	go srv.Serve(ln)
	defer srv.Stop()

	metrics := runTestsAndCollectMetrics()
	fileCreator := map[string]bool{
		"allocsPerOp": true,
		"bytesPerOp":  false,
	}

	for key, graphingAllocs := range fileCreator {
		outFile := filepath.Join(dir, fmt.Sprintf("%s.html", key))
		f, err := os.Create(outFile)
		if err != nil {
			log.Fatalf("Creating graph display file: %q err: %v", key, err)
		}
		metrics.GraphingAllocs = graphingAllocs
		if err := tmpl.Execute(f, metrics); err != nil {
			log.Fatalf("Creating template: %v", err)
		}
		f.Close()
		log.Printf("Generated file: %q", outFile)
	}
}

type server int

func (s *server) Ping(ctx context.Context, p *proto.Payload) (*proto.Payload, error) {
	return &proto.Payload{"Pong"}, nil
}

type graphData struct {
	QPS        []string
	MemAllocs  []float64
	Throughput []float64
}

type graphTemplate struct {
	GraphingAllocs bool
	Traced         *graphData
	Untraced       *graphData
}

func runTestsAndCollectMetrics() *graphTemplate {
	log.Printf("Running tests to collect the metrics")
	spin := spinner.New(10)
	spin.Start()
	defer spin.Stop()

	graphingMap := make(map[string]*graphData)

	funcsMap := map[string]func(*testing.B, int){
		"Traced":   benchmarkTraced,
		"Untraced": benchmarkUntraced,
	}

	qpsL := []int{1, 10, 100, 1000}
	for key, fn := range funcsMap {
		log10QPSLabels := make([]string, 0, len(qpsL))
		memAllocs := make([]float64, 0, len(qpsL))
		throughput := make([]float64, 0, len(qpsL))
		for _, qps := range qpsL {
			br := testing.Benchmark(func(b *testing.B) {
				fn(b, qps)
			})
			log10 := math.Log10(float64(qps))
			log10QPSLabels = append(log10QPSLabels, fmt.Sprintf("%.2f (%d QPS)", log10, qps))
			memAllocs = append(memAllocs, float64(br.MemAllocs)/float64(br.N))
			throughput = append(throughput, float64(br.MemBytes)/float64(br.N))
		}

		graphingMap[key] = &graphData{
			QPS:        log10QPSLabels,
			MemAllocs:  memAllocs,
			Throughput: throughput,
		}
	}

	return &graphTemplate{Traced: graphingMap["Traced"], Untraced: graphingMap["Untraced"]}
}

const chartHTML = `
<HTML>
  <head>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.7.1/Chart.bundle.min.js"></script>
    <script>
    </script>
  </head>

  <body>
    <canvas id="benchesChart" width="300" height="300%"></canvas>
    <script>
      var ctx = document.getElementById('benchesChart').getContext('2d');
      var chart = new Chart(ctx, {
	type: 'line',
	data: {
	  labels: {{ json .Traced.QPS }},
	  datasets: [{
	    label: 'Traced',
	    backgroundColor: '#00FF00',
	    borderColor: '#00FF00',
	    fill: false,
	    data: {{if .GraphingAllocs}} {{ json .Traced.MemAllocs }} {{else}} {{ json .Traced.Throughput }} {{end}},
	  }, {
	    label: 'Untraced',
	    backgroundColor: '#0000FF',
	    borderColor: '#0000FF',
	    fill: false,
	    data: {{if .GraphingAllocs}} {{ json .Untraced.MemAllocs }} {{else}} {{ json .Untraced.Throughput }} {{end}},
	  }]
	},
	options: {
	  responsive: true,
	  title:{
	    display: true,
	    text: '{{if .GraphingAllocs}}Allocs/Op vs QPS log10(n){{else}}Bytes/Op vs QPS log10(n){{end}}',
	  }
	},
      });
    </script>
  </body>
</HTML>
`
