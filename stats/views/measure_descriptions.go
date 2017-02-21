package views

type MeasureDesc interface {
	Meta() *measureDesc
}

type measureDescFloat64 struct {
	*measureDesc
}

func NewMeasureDescFloat64(name string, description string, unit *MeasurementUnit) *measureDescFloat64 {
	return &measureDescFloat64{
		&measureDesc{
			name:         name,
			description:  description,
			unit:         unit,
			aggViewDescs: make(map[AggregationViewDesc]struct{}),
		},
	}
}

func (md *measureDescFloat64) Meta() *measureDesc {
	return md.measureDesc
}

func (md *measureDescFloat64) CreateMeasurement(v float64) Measurement {
	return &measurementFloat64{
		md: md,
		v:  v,
	}
}

type measureDescInt64 struct {
	*measureDesc
}

func NewMeasureDescInt64(name string, description string, unit *MeasurementUnit) *measureDescInt64 {
	return &measureDescInt64{
		&measureDesc{
			name:         name,
			description:  description,
			unit:         unit,
			aggViewDescs: make(map[AggregationViewDesc]struct{}),
		},
	}
}

func (md *measureDescInt64) Meta() *measureDesc {
	return md.measureDesc
}

// measureDesc describes a data point (measurement) type accounted
// for by the stats library, such as RAM or CPU time.
type measureDesc struct {
	// The name must be unique. Used to link the MeasureDesc to a ViewDesc.
	// Examples are cpu:tickCount, diskio:time...
	name string
	// The description is used for display purposes only. It is meant to be
	// human readable and is used to show the resource in dashboards.
	// Example are CPU profile ticks, Disk I/O, Disk usage in usecs...
	description  string
	unit         *MeasurementUnit
	aggViewDescs map[AggregationViewDesc]struct{}
}

func (md *measureDesc) Name() string {
	return md.name
}

func (md *measureDesc) Description() string {
	return md.description
}

func (md *measureDesc) Unit() *MeasurementUnit {
	return md.unit
}

// MeasurementUnit is the unit of measurement for a resource.
type MeasurementUnit struct {
	Power10      int
	Numerators   []BasicUnit
	Denominators []BasicUnit
}

// BasicUnit is used for representing the basic units used to construct
// MeasurementUnits.
type BasicUnit byte

// These constants are the type of basic units allowed.
const (
	UnknownUnit BasicUnit = iota
	ScalarUnit
	BitsUnit
	BytesUnit
	SecsUnit
	CoresUnit
)
