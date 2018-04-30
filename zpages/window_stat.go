package zpages

import (
	"math"
	"time"

	"go.opencensus.io/stats/view"
)

type interval struct {
	distribution *view.DistributionData
	index        int64
	updateTime   time.Time
}

type windowStat struct {
	intervals      []interval
	lastUpdate     int
	granularitySec int64
	startUnix      int64
	startTime      time.Time
	window         time.Duration
}

var clock = func() time.Time {
	return time.Now()
}

func (ws *windowStat) init(window time.Duration) {
	historySec := int64(window.Seconds())
	granularitySec := int64(historySec / 60)
	ws.intervals = make([]interval, int(historySec/granularitySec+1))
	for i := range ws.intervals {
		ws.intervals[i].index = -1
	}
	ws.granularitySec = granularitySec
	ws.lastUpdate = -1
	ws.window = window
}

func (ws *windowStat) update(t time.Time, dist *view.DistributionData) {
	if ws.lastUpdate == -1 {
		ws.startTime = t
		ws.startUnix = t.Unix() - (t.Unix() % ws.granularitySec)
	}
	sliceIndex, intv := ws.intervalFor(t)
	if intv != nil {
		intv.distribution = dist
		intv.updateTime = t
		ws.lastUpdate = sliceIndex
	}
}

func (ws *windowStat) intervalFor(t time.Time) (int, *interval) {
	deltaSec := t.Unix() - ws.startUnix
	if deltaSec < 0 {
		return -1, nil
	}
	ind := deltaSec / ws.granularitySec
	n := int64(len(ws.intervals))
	sliceIndex := int(ind % n)
	intv := &ws.intervals[sliceIndex]
	if intv.index == ind {
		return sliceIndex, intv
	}
	if intv.index < ind {
		*intv = interval{index: ind}
		return sliceIndex, intv
	}
	// we are attempting to get an interval for something that has already
	// fallen out of the circular buffer
	return -1, nil
}

func (ws *windowStat) read() (actualStart, actualEnd time.Time, diff *view.DistributionData) {
	if ws.lastUpdate == -1 {
		return
	}

	now := clock()
	start := now.Add(-ws.window)

	// Try to find the interval closest to start.
	minDiff := time.Duration(math.MaxInt64)
	var startInterval *interval
	n := len(ws.intervals)
	for i := (ws.lastUpdate + 1) % n; i != ws.lastUpdate; i = (i + 1) % n {
		if ws.intervals[i].index < 0 {
			continue
		}
		diff := ws.intervals[i].updateTime.Sub(start)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			startInterval = &ws.intervals[i]
		}
	}

	endInterval := ws.intervals[ws.lastUpdate]

	if startInterval == nil {
		diff = endInterval.distribution
		actualStart = endInterval.updateTime
		actualEnd = endInterval.updateTime
		return
	}

	startDist, endDist := startInterval.distribution, endInterval.distribution
	actualStart, actualEnd = startInterval.updateTime, endInterval.updateTime

	buckets := len(endDist.CountPerBucket)
	if buckets != len(startDist.CountPerBucket) {
		panic("can't subtract distributions with different number of buckets")
	}
	diff = &view.DistributionData{}
	diff.Count = endDist.Count - startDist.Count
	diff.Mean = (endDist.Sum() - startDist.Sum()) / float64(diff.Count)
	for i := 0; i < buckets; i++ {
		diff.CountPerBucket = append(diff.CountPerBucket, endDist.CountPerBucket[i]-startDist.CountPerBucket[i])
	}
	return
}
