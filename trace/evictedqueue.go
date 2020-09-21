// Copyright 2019, OpenCensus Authors
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

package trace

type evictedQueue struct {
	ringQueue    []interface{}
	capacity     int
	droppedCount int
	writeIdx     int
	readIdx      int
	startRead    bool
}

func newEvictedQueue(capacity int) *evictedQueue {
	eq := &evictedQueue{
		capacity:  capacity,
		ringQueue: make([]interface{}, 0),
	}

	return eq
}

func (eq *evictedQueue) add(value interface{}) {
	if len(eq.ringQueue) < eq.capacity {
		eq.ringQueue = append(eq.ringQueue, value)
		return
	}

	eq.ringQueue[eq.writeIdx] = value
	eq.droppedCount++
	eq.writeIdx++
	eq.writeIdx %= eq.capacity
	eq.readIdx = eq.writeIdx
}

// Do not add more item after use readNext
func (eq *evictedQueue) readNext() interface{} {
	if eq.startRead && eq.readIdx == eq.writeIdx {
		return nil
	}

	eq.startRead = true
	res := eq.ringQueue[eq.readIdx]
	eq.readIdx++
	eq.readIdx %= eq.capacity
	return res
}
