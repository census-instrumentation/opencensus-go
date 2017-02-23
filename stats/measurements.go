package stats

import "fmt"

type Measurement interface {
	measureDesc() MeasureDesc
	float64() float64
	int64() int64
}

type measurementFloat64 struct {
	md MeasureDesc
	v  float64
}

func (mf *measurementFloat64) measureDesc() MeasureDesc {
	return mf.md
}

func (mf *measurementFloat64) float64() float64 {
	return mf.v
}

func (mf *measurementFloat64) int64() int64 {
	panic(fmt.Sprintf("called int64() on %v", mf))
}

type measurementInt64 struct {
	md MeasureDesc
	v  int64
}

func (mi *measurementInt64) measureDesc() MeasureDesc {
	return mi.md
}

func (mi *measurementInt64) float64() float64 {
	panic(fmt.Sprintf("called float64() on %v", mi))
}

func (mf *measurementInt64) int64() int64 {
	return mf.v
}

type measurementBool struct {
	md MeasureDesc
	v  bool
}

func (mb *measurementBool) measureDesc() MeasureDesc {
	return mb.md
}

type measurementString struct {
	md MeasureDesc
	v  string
}

func (ms *measurementString) measureDesc() MeasureDesc {
	return ms.md
}
