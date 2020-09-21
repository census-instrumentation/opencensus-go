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

import (
	"reflect"
	"testing"
)

func init() {
}

func TestAddAndReadNext(t *testing.T) {
	t.Run("len(ringQueue) < capacity", func(t *testing.T) {
		values := []string{"value1", "value2"}
		capacity := 3
		q := newEvictedQueue(capacity)

		for _, value := range values {
			q.add(value)
		}

		gotValues := make([]string, len(q.ringQueue))
		for i := 0; i < len(gotValues); i++ {
			gotValues[i] = q.readNext().(string)
		}

		if !reflect.DeepEqual(values, gotValues) {
			t.Errorf("got array = %#v; want %#v", gotValues, values)
		}
	})
	t.Run("dropped count", func(t *testing.T) {
		values := []string{"value1", "value2", "value3", "value1", "value4", "value1", "value3", "value1", "value4"}
		wantValues := []string{"value3", "value1", "value4"}
		capacity := 3
		wantDroppedCount := len(values) - capacity

		q := newEvictedQueue(capacity)

		for _, value := range values {
			q.add(value)
		}

		gotValues := make([]string, len(wantValues))
		for i := 0; i < len(gotValues); i++ {
			gotValues[i] = q.readNext().(string)
		}

		if !reflect.DeepEqual(wantValues, gotValues) {
			t.Errorf("got array = %#v; want %#v", gotValues, wantValues)
		}

		if wantDroppedCount != q.droppedCount {
			t.Errorf("got dropped count %d want %d", q.droppedCount, wantDroppedCount)
		}
	})
}
