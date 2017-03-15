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

package stats

import (
	"log"
)

type Subscription interface {
	contains(vw ViewDesc) bool
	addViewDesc(desc ViewDesc)
	reportUsage()
}

type SingleSubscription struct {
	C        chan *View
	ViewName string
	vwDesc   ViewDesc
	vw       *View
}

func (ss *SingleSubscription) contains(vw ViewDesc) bool {
	if ss.ViewName == vw.ViewDescCommon().Name {
		return true
	}
	return false
}

func (ss *SingleSubscription) addViewDesc(desc ViewDesc) {
	ss.vwDesc = desc
}

func (ss *SingleSubscription) reportUsage() {
	if ss.vw == nil {
		return
	}
	select {
	case ss.C <- ss.vw:
	default:
		log.Printf("*SingleSubscription.reportUsage(_) dropped view %v. Receiver channel not ready.", ss.vw)
	}
	ss.vw = nil
}

type MultiSubscription struct {
	C            chan []*View
	ViewNames    []string
	MeasureNames []string
	vwDescs      []ViewDesc
	vws          []*View
}

func (ms *MultiSubscription) containsViewName(n string) bool {
	if len(ms.ViewNames) == 0 {
		return true
	}
	for _, name := range ms.ViewNames {
		if name == n {
			return true
		}
	}
	return false
}

func (ms *MultiSubscription) containsMeasureName(n string) bool {
	if len(ms.MeasureNames) == 0 {
		return true
	}
	for _, name := range ms.MeasureNames {
		if name == n {
			return true
		}
	}
	return false
}

func (ms *MultiSubscription) contains(vw ViewDesc) bool {
	if ms.containsMeasureName(vw.ViewDescCommon().MeasureDescName) {
		if ms.containsViewName(vw.ViewDescCommon().Name) {
			return true
		}
	}
	return false
}

func (ms *MultiSubscription) addViewDesc(desc ViewDesc) {
	ms.vwDescs = append(ms.vwDescs, desc)
}

func (ms *MultiSubscription) reportUsage() {
	if len(ms.vws) == 0 {
		return
	}
	select {
	case ms.C <- ms.vws:
	default:
		log.Printf("*MultiSubscription.reportUsage(_) dropped views %v. Receiver channel not ready.", ms.vws)
	}
	ms.vws = nil
}
