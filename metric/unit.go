package metric

// Unit is a string encoded according to the case-sensitive abbreviations from the
// Unified Code for Units of Measure: http://unitsofmeasure.org/ucum.html
type Unit string

const (
	UnitDimensionless Unit = "1"
	UnitBytes         Unit = "By"
	UnitMilliseconds  Unit = "ms"
)
