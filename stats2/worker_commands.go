// Copyright 2017 Google Inc.
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

// Package stats defines the stats collection API and its native Go
// implementation.

package stats2

import (
	"fmt"
	"time"

	"github.com/google/working-instrumentation-go/tags"
)

type command interface {
	handleCommand(w *worker)
}

type measureByNameReq struct {
	name string
	c    chan *measureByNameResp
}

type measureByNameResp struct {
	m   Measure
	err error
}

func (cmd *measureByNameReq) handleCommand(w *worker) {
	if m, ok := w.measuresByName[cmd.name]; ok {
		cmd.c <- &measureByNameResp{
			m,
			nil,
		}
		return
	}
	cmd.c <- &measureByNameResp{
		nil,
		fmt.Errorf(""),
	}
}

type measureRegistrationReq struct {
	m   Measure
	err chan error
}

type measureUnregistrationReq struct {
	m   Measure
	err chan error
}

type viewByNameReq struct {
	name string
	c    chan *viewByNameResp
}

type viewByNameResp struct {
	v   View
	err error
}

type viewRegistrationReq struct {
	v   View
	err chan error
}

type viewUnregistrationReq struct {
	v   View
	err chan error
}

type viewSubscriptionReq struct {
	v   View
	c   chan *ViewData
	err chan error
}

type viewUnsubscriptionReq struct {
	v   View
	c   chan *ViewData
	err chan error
}

type viewStartCollectionReq struct {
	v   View
	err chan error
}

type viewStopCollectionReq struct {
	v   View
	err chan error
}

type viewDataRetrievalReq struct {
	v View
	c chan *viewDataRetrievalResp
}

type viewDataRetrievalResp struct {
	rows []*Rows
	err  error
}

type recordingFloat64Req struct {
	ts *tags.TagSet
	mf MeasureFloat64
	v  float64
}

type recordingInt64Req struct {
	ts *tags.TagSet
	mf MeasureInt64
	v  int64
}

type recordingManyReq struct {
	ts          *tags.TagSet
	measurement []Measurement
}

type reportingPeriod struct {
	d time.Duration
}
