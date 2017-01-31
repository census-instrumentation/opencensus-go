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
)

type usageCollector struct {
	mDescriptors map[string]*MeasureDesc
	vDescriptors map[string]AggregationViewDesc
}

func (uc *usageCollector) registerMeasureDesc(md *MeasureDesc) error {
	if _, ok := uc.mDescriptors[md.Name]; ok {
		return fmt.Errorf("a measure descriptor with the same name %s is already registered", md.Name)
	}

	for n, d := range uc.mDescriptors {
		if md == d {
			return fmt.Errorf("the measure descriptor %v was already registered under a different name %s", md, n)
		}
	}

	md.aggViewDescs = make(map[AggregationViewDesc]struct{})
	uc.mDescriptors[md.Name] = md
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

func (uc *usageCollector) registerViewDesc(avd AggregationViewDesc, now time.Time) error {
	vd := avd.viewDesc()
	md, ok := uc.mDescriptors[vd.MeasureDescName]
	if !ok {
		return fmt.Errorf("view contains a resource %s that is not registered", vd.MeasureDescName)
	}

	if v, ok := uc.vDescriptors[vd.Name]; ok {
		return fmt.Errorf("a view %v with the same name %s is already registered", v, v.viewDesc().Name)
	}

	for vwName, vwDesc := range uc.vDescriptors {
		if vwDesc == avd {
			return fmt.Errorf("view %v is already registered under a different name %s", vd, vwName)
		}
	}

	if err := avd.isValid(); err != nil {
		return err
	}

	vd.start = now
	vd.signatures = make(map[string]aggregator)

	uc.vDescriptors[vd.Name] = avd
	md.aggViewDescs[avd] = struct{}{}

	return nil
}

func (uc *usageCollector) unregisterViewDesc(vwName string) error {
	avd, ok := uc.vDescriptors[vwName]
	if !ok {
		return fmt.Errorf("no view descriptor with the name %s is registered", vwName)
	}

	vd := avd.viewDesc()
	md, ok := uc.mDescriptors[vd.MeasureDescName]
	if !ok {
		return fmt.Errorf("no measure descriptor with the name %s is registered", vd.MeasureDescName)
	}

	delete(uc.vDescriptors, vwName)
	delete(md.aggViewDescs, avd)
	return nil
}

func (uc *usageCollector) subscribeToViewDesc(vwName string, c chan *View) error {
	avd, ok := uc.vDescriptors[vwName]
	if !ok {
		return fmt.Errorf("no view descriptor with the name %s is registered", vwName)
	}

	vd := avd.viewDesc()
	if _, ok := vd.vChans[c]; ok {
		return fmt.Errorf("channel is already used to subscribe to this viewDesc %s", vwName)
	}

	vd.vChans[c] = struct{}{}
	return nil
}

func (uc *usageCollector) unsubscribeFromViewDesc(vwName string, c chan *View) error {
	avd, ok := uc.vDescriptors[vwName]
	if !ok {
		return fmt.Errorf("no view descriptor with the name %s is registered", vwName)
	}

	vd := avd.viewDesc()
	if _, ok := vd.vChans[c]; !ok {
		return fmt.Errorf("channel is not used to subscribe to this viewDesc %s", vwName)
	}

	delete(vd.vChans, c)
	return nil
}

func (uc *usageCollector) recordMeasurement(now time.Time, ct contextTags, md *MeasureDesc, v float64) error {
	tmp, ok := uc.mDescriptors[md.Name]
	if !ok || tmp != md {
		return fmt.Errorf("error recording measurement. %v was not registered or its name was modified after registration", md)
	}

	for avd := range md.aggViewDescs {
		var sig string
		vd := avd.viewDesc()
		if len(vd.TagKeys) == 0 {
			// This is the all keys view.
			sig = ct.encodeToFullSignature()
		} else {
			sig = ct.encodeToValuesSignature(vd.TagKeys)
		}

		if err := uc.add(vd.start, now, vd.signatures, sig, avd, v); err != nil {
			return fmt.Errorf("error recording measurement %v", err)
		}
	}
	return nil
}

func (uc *usageCollector) recordManyMeasurement(now time.Time, ct contextTags, mds []*MeasureDesc, vs []float64) error {
	for _, tmp := range mds {
		md, ok := uc.mDescriptors[tmp.Name]
		if !ok || md != tmp {
			return fmt.Errorf("error recording measurement. %v was not registered or its name was modified after registration", md)
		}
	}

	if len(mds) != len(vs) {
		return fmt.Errorf("len([]*MeasureDesc)=%v different than len(vs)=%v", len(mds), len(vs))
	}

	// TODO(iamm2): optimize this to avoid calling recordMeasurement multiple
	// times. Reuse fullSignature on as many "all tags views" as possible.
	for i, md := range mds {
		err := uc.recordMeasurement(now, ct, md, vs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (uc *usageCollector) add(start, now time.Time, signatures map[string]aggregator, sig string, avd AggregationViewDesc, val float64) error {
	agg, found := signatures[sig]
	if !found {
		var err error
		if agg, err = avd.createAggregator(start); err != nil {
			return err
		}
		signatures[sig] = agg
	}

	agg.addSample(val, now)
	return nil
}

func (uc *usageCollector) retrieveViews(now time.Time) []*View {
	var views []*View
	for _, avd := range uc.vDescriptors {
		vw, err := avd.retrieveView(now)
		if err != nil {
			//// TODO(iamm2) log error fmt.Errorf("error retrieving view for view description %v. %v", *vd, err)
		}

		views = append(views, vw)
	}
	return views
}

func (uc *usageCollector) retrieveView(now time.Time, avd AggregationViewDesc) (*View, error) {
	vd := avd.viewDesc()

	tmp, ok := uc.vDescriptors[vd.Name]
	if !ok {
		return nil, fmt.Errorf("no view descriptor with the name %s is registered", vd.MeasureDescName)
	}

	if tmp != avd {
		return nil, fmt.Errorf("a different view %v was registered with this name %v", tmp, vd.Name)
	}

	vw, err := avd.retrieveView(now)
	if err != nil {
		return nil, fmt.Errorf("error retrieving view for view description %v. %v", avd, err)
	}

	return vw, nil
}
