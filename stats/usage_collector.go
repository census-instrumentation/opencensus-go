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
	"fmt"
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
)

type usageCollector struct {
	mDescriptors  map[string]MeasureDesc
	vDescriptors  map[string]ViewDesc
	subscriptions map[Subscription]bool
}

func newUsageCollector() *usageCollector {
	return &usageCollector{
		mDescriptors:  make(map[string]MeasureDesc),
		vDescriptors:  make(map[string]ViewDesc),
		subscriptions: make(map[Subscription]bool),
	}
}

func (uc *usageCollector) registerMeasureDesc(md MeasureDesc) error {
	meta := md.Meta()
	if _, ok := uc.mDescriptors[meta.name]; ok {
		return fmt.Errorf("a measure descriptor with the same name %s is already registered", meta.name)
	}

	for n, d := range uc.mDescriptors {
		if md == d {
			return fmt.Errorf("the measure descriptor %v was already registered under a different name %s", md, n)
		}
	}

	uc.mDescriptors[meta.name] = md
	return nil
}

func (uc *usageCollector) unregisterMeasureDesc(mName string) error {
	_, ok := uc.mDescriptors[mName]
	if !ok {
		return fmt.Errorf("no measure descriptor with the name %s is registered", mName)
	}

	delete(uc.mDescriptors, mName)
	return nil
}

func (uc *usageCollector) registerViewDesc(vd ViewDesc, now time.Time) error {
	vdc := vd.ViewDescCommon()
	md, ok := uc.mDescriptors[vdc.MeasureDescName]
	if !ok {
		return fmt.Errorf("registerViewDesc(_) failed. ViewDesc %v cannot be regsitered. It has MeasureDescName=%v which is not registered", *vdc, vdc.MeasureDescName)
	}

	if tmp, ok := uc.vDescriptors[vdc.Name]; ok {
		return fmt.Errorf("registerViewDesc(_) failed. ViewDesc %v cannot be regsitered. A different ViewDesc has already been registered with a Name=%v", *vdc, tmp.ViewDescCommon().Name)
	}

	for vwName, vwDesc := range uc.vDescriptors {
		if vwDesc == vd {
			return fmt.Errorf("registerViewDesc(_) failed. ViewDesc %v is already registered under a different name %v", vdc, vwName)
		}
	}

	if err := vd.isValid(); err != nil {
		return fmt.Errorf("registerViewDesc(_) failed. %v", err)
	}

	vdc.start = now
	vdc.signatures = make(map[string]aggregator)

	uc.vDescriptors[vdc.Name] = vd
	vdc.subscriptions = make(map[Subscription]bool)
	md.Meta().aggViewDescs[vd] = struct{}{}

	return nil
}

func (uc *usageCollector) unregisterViewDesc(vwName string) error {
	avd, ok := uc.vDescriptors[vwName]
	if !ok {
		return fmt.Errorf("no view descriptor with the name %s is registered", vwName)
	}

	vd := avd.ViewDescCommon()
	md, ok := uc.mDescriptors[vd.MeasureDescName]
	if !ok {
		return fmt.Errorf("no measure descriptor with the name %s is registered", vd.MeasureDescName)
	}

	delete(uc.vDescriptors, vwName)
	delete(md.Meta().aggViewDescs, avd)
	return nil
}

func (uc *usageCollector) addSubscription(s Subscription) error {
	if uc.subscriptions[s] {
		return fmt.Errorf("addSubscription(_) failed. Subscription %v already used to subscribe", s)
	}

	uc.subscriptions[s] = true
	for _, desc := range uc.vDescriptors {
		if s.contains(desc) {
			s.addViewDesc(desc)
			desc.ViewDescCommon().subscriptions[s] = true
		}
	}
	return nil
}

func (uc *usageCollector) unsubscribe(s Subscription) error {
	if !uc.subscriptions[s] {
		return fmt.Errorf("removeSubscription(_) failed. Subscription %v not used to subscribe", s)
	}

	for _, desc := range uc.vDescriptors {
		delete(desc.ViewDescCommon().subscriptions, s)
	}
	delete(uc.subscriptions, s)
	return nil
}

func (uc *usageCollector) recordMeasurement(now time.Time, ts tagging.TagsSet, m Measurement) error {
	md := m.measureDesc()
	meta := md.Meta()
	tmp, ok := uc.mDescriptors[meta.name]
	if !ok || tmp != md {
		return fmt.Errorf("error recording measurement. %v was not registered or its name was modified after registration", md)
	}

	for avd := range meta.aggViewDescs {
		var sig []byte
		vdc := avd.ViewDescCommon()
		if len(vdc.TagKeys) == 0 {
			// This is a "don't care about keys" view. sig is empty for all
			// records. Aggregates all records in the same view aggregation.
		} else {
			sig = tagging.EncodeToValuesSignature(ts, vdc.TagKeys)
		}

		if err := uc.add(vdc.start, now, vdc.signatures, string(sig), avd, m); err != nil {
			return fmt.Errorf("error recording measurement %v", err)
		}
	}
	return nil
}

func (uc *usageCollector) recordManyMeasurement(now time.Time, ts tagging.TagsSet, ms []Measurement) error {
	for _, m := range ms {
		md := m.measureDesc()
		meta := md.Meta()
		tmp, ok := uc.mDescriptors[meta.name]
		if !ok || tmp != md {
			return fmt.Errorf("error recording measurement. %v was not registered or its name was modified after registration", md)
		}
	}

	// TODO(iamm2): optimize this to avoid calling recordMeasurement multiple
	// times. Reuse fullSignature on as many "all tags views" as possible.
	for _, md := range ms {
		err := uc.recordMeasurement(now, ts, md)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uc *usageCollector) add(start, now time.Time, signatures map[string]aggregator, sig string, vd ViewDesc, m Measurement) error {
	agg, found := signatures[sig]
	if !found {
		var err error
		if agg, err = vd.createAggregator(start); err != nil {
			return err
		}
		signatures[sig] = agg
	}

	agg.addSample(m, now)
	return nil
}

func (uc *usageCollector) retrieveViews(now time.Time) []*View {
	var views []*View
	for _, vd := range uc.vDescriptors {
		vw, err := vd.retrieveView(now)
		if err != nil {
			//// TODO(iamm2) log error fmt.Errorf("error retrieving view for view description %v. %v", *vd, err)
			continue
		}

		views = append(views, vw)
	}
	return views
}

func (uc *usageCollector) retrieveView(name string, now time.Time, vd ViewDesc) (*View, error) {
	vdc := vd.ViewDescCommon()

	tmp, ok := uc.vDescriptors[vdc.Name]
	if !ok {
		return nil, fmt.Errorf("no view descriptor with the name %s is registered", vdc.MeasureDescName)
	}

	if tmp != vd {
		return nil, fmt.Errorf("a different view %v was registered with this name %v", tmp, vdc.Name)
	}

	vw, err := vd.retrieveView(now)
	if err != nil {
		return nil, fmt.Errorf("error retrieving view for view description %v. %v", vd, err)
	}

	return vw, nil
}

func (uc *usageCollector) retrieveViewByName(name string, now time.Time) (*View, error) {
	vd, ok := uc.vDescriptors[name]
	if !ok {
		return nil, fmt.Errorf("no view descriptor with the name %s is registered", name)
	}

	vw, err := vd.retrieveView(now)
	if err != nil {
		return nil, fmt.Errorf("error retrieving view for view description %v. %v", vd, err)
	}

	return vw, nil
}
