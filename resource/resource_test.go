// Copyright 2018, OpenCensus Authors
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

package resource

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	cases := []struct {
		a, b, expect *Resource
	}{
		{
			a: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1", "b": "2"},
			},
			b: &Resource{
				Type: "t2",
				Tags: map[string]string{"a": "1", "b": "3", "c": "4"},
			},
			expect: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1", "b": "2", "c": "4"},
			},
		},
		{
			a: nil,
			b: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1"},
			},
			expect: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1"},
			},
		},
		{
			a: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1"},
			},
			b: nil,
			expect: &Resource{
				Type: "t1",
				Tags: map[string]string{"a": "1"},
			},
		},
	}
	for _, c := range cases {
		res := Merge(c.a, c.b)
		if !reflect.DeepEqual(res, c.expect) {
			t.Fatalf("unexpected result: want %+v, got %+v", c.expect, res)
		}
	}
}

func TestDecodeTags(t *testing.T) {
	cases := []struct {
		s      string
		expect map[string]string
		fail   bool
	}{
		{
			s:      `example.org/test-1="test ¥ \"" ,un=quøted,  Abc=Def`,
			expect: map[string]string{"example.org/test-1": "test ¥ \"", "un": "quøted", "Abc": "Def"},
		}, {
			s:      `single=key`,
			expect: map[string]string{"single": "key"},
		},
		{s: `invalid-char-ü=test`, fail: true},
		{s: `missing="trailing-quote`, fail: true},
		{s: `missing=leading-quote"`, fail: true},
		{s: `extra=chars, a`, fail: true},
		{s: `a, extra=chars`, fail: true},
		{s: `a, extra=chars`, fail: true},
	}
	for i, c := range cases {
		t.Logf("test %d: %s", i, c.s)

		res, err := DecodeTags(c.s)
		if err != nil && !c.fail {
			t.Fatalf("unexpected error: %s", err)
		}
		if c.fail && err == nil {
			t.Fatalf("expected failure but got none, result: %v", res)
		}
		if !reflect.DeepEqual(res, c.expect) {
			t.Fatalf("expected result %v, got %v", c.expect, res)
		}
	}
}

func TestEncodeTags(t *testing.T) {
	s := EncodeTags(map[string]string{
		"example.org/test-1": "test ¥ \"",
		"un":                 "quøted",
		"Abc":                "Def",
	})
	if exp := `example.org/test-1="test ¥ \"",un="quøted",Abc="Def"`; s != exp {
		t.Fatalf("expected %q, got %q", exp, s)
	}
}
