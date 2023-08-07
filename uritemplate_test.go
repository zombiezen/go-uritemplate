// Copyright 2023 Ross Light
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		 https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package uritemplate

import "testing"

var keysData = struct {
	Semi  string
	Dot   string
	Comma string
}{";", ".", ","}

var expansionSectionData = map[string]any{
	"count":     []string{"one", "two", "three"},
	"dom":       []string{"example", "com"},
	"dub":       "me/too",
	"hello":     "Hello World!",
	"half":      "50%",
	"var":       "value",
	"who":       "fred",
	"base":      "http://example.com/home/",
	"path":      "/foo/bar",
	"list":      []string{"red", "green", "blue"},
	"keys":      keysData,
	"v":         "6",
	"x":         "1024",
	"y":         "768",
	"empty":     "",
	"emptyKeys": map[string]any{},
	"undef":     nil,
}

var tests = []struct {
	template string
	data     any
	want     string
}{
	// Introduction examples.
	{
		template: "http://www.example.com/foo{?query,number}",
		data: map[string]any{
			"query":  "mycelium",
			"number": 100,
		},
		want: "http://www.example.com/foo?query=mycelium&number=100",
	},
	{
		template: "http://www.example.com/foo{?query,number}",
		data: map[string]any{
			"number": 100,
		},
		want: "http://www.example.com/foo?number=100",
	},
	{
		template: "http://www.example.com/foo{?query,number}",
		data:     nil,
		want:     "http://www.example.com/foo",
	},

	// Simple string expansion.
	{
		template: "{var}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
		},
		want: "value",
	},
	{
		template: "{hello}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
		},
		want: "Hello%20World%21",
	},
	{
		template: "{half}",
		data:     expansionSectionData,
		want:     "50%25",
	},
	{
		template: "O{empty}X",
		data:     expansionSectionData,
		want:     "OX",
	},
	{
		template: "O{undef}X",
		data:     expansionSectionData,
		want:     "OX",
	},
	{
		template: "{x,y}",
		data:     expansionSectionData,
		want:     "1024,768",
	},
	{
		template: "{x,y}",
		data:     struct{ X, Y int }{1024, 768},
		want:     "1024,768",
	},
	{
		template: "{x,y}",
		data: struct {
			Bar int `uritemplate:"y"`
			Foo int `uritemplate:"x"`
		}{768, 1024},
		want: "1024,768",
	},
	{
		template: "{x,hello,y}",
		data:     expansionSectionData,
		want:     "1024,Hello%20World%21,768",
	},
	{
		template: "?{x,empty}",
		data:     expansionSectionData,
		want:     "?1024,",
	},
	{
		template: "?{x,undef}",
		data:     expansionSectionData,
		want:     "?1024",
	},
	{
		template: "?{undef,y}",
		data:     expansionSectionData,
		want:     "?768",
	},

	// Reserved string expansion.
	{
		template: "{+var}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "value",
	},
	{
		template: "{+hello}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "Hello%20World!",
	},
	{
		template: "{+path}/here",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "/foo/bar/here",
	},
	{
		template: "here?ref={+path}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "here?ref=/foo/bar",
	},
	{
		template: "{base}index",
		data:     expansionSectionData,
		want:     "http%3A%2F%2Fexample.com%2Fhome%2Findex",
	},
	{
		template: "{+base}index",
		data:     expansionSectionData,
		want:     "http://example.com/home/index",
	},
	{
		template: "O{+empty}X",
		data:     expansionSectionData,
		want:     "OX",
	},
	{
		template: "O{+undef}X",
		data:     expansionSectionData,
		want:     "OX",
	},

	// Fragment expansion, crosshatch-prefixed.
	{
		template: "X{#var}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "X#value",
	},
	{
		template: "X{#hello}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
		},
		want: "X#Hello%20World!",
	},

	// String expansion with multiple variables.
	{
		template: "map?{x,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "map?1024,768",
	},
	{
		template: "{x,hello,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "1024,Hello%20World%21,768",
	},

	// Reserved expansion with multiple variables.
	{
		template: "{+x,hello,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "1024,Hello%20World!,768",
	},
	{
		template: "{+path,x}/here",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "/foo/bar,1024/here",
	},

	// Fragment expansion with multiple variables.
	{
		template: "{#x,hello,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "#1024,Hello%20World!,768",
	},
	{
		template: "{#path,x}/here",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "#/foo/bar,1024/here",
	},

	// Label expansion, dot-prefixed.
	{
		template: "X{.var}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "X.value",
	},
	{
		template: "X{.x,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "X.1024.768",
	},

	// Path segments, slash-prefixed.
	{
		template: "{/var}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "/value",
	},
	{
		template: "{/var,x}/here",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "/value/1024/here",
	},

	// Path-style parameters, semicolon-prefixed.
	{
		template: "{;x,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: ";x=1024;y=768",
	},
	{
		template: "{;x,y,empty}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: ";x=1024;y=768;empty",
	},

	// Form-style query, ampersand-separated.
	{
		template: "{?x,y}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "?x=1024&y=768",
	},
	{
		template: "{?x,y,empty}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "?x=1024&y=768&empty=",
	},

	// Form-style query continuation.
	{
		template: "?fixed=yes{&x}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "?fixed=yes&x=1024",
	},
	{
		template: "{&x,y,empty}",
		data: map[string]string{
			"var":   "value",
			"hello": "Hello World!",
			"empty": "",
			"path":  "/foo/bar",
			"x":     "1024",
			"y":     "768",
		},
		want: "&x=1024&y=768&empty=",
	},

	// String expansion with value modifiers.
	{
		template: "{var:3}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "val",
	},
	{
		template: "{var:30}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "value",
	},
	{
		template: "{semi}",
		data: map[string]any{
			"var":  "value",
			"semi": ";",
		},
		want: "%3B",
	},
	{
		template: "{semi:2}",
		data: map[string]any{
			"var":  "value",
			"semi": ";",
		},
		want: "%3B",
	},
	{
		template: "{list}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "red,green,blue",
	},
	{
		template: "{list*}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "red,green,blue",
	},
	{
		template: "{keys}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "semi,%3B,dot,.,comma,%2C",
	},
	{
		template: "{keys*}",
		data: map[string]any{
			"var":   "value",
			"hello": "Hello World!",
			"path":  "/foo/bar",
			"list":  []string{"red", "green", "blue"},
			"keys":  keysData,
		},
		want: "semi=%3B,dot=.,comma=%2C",
	},
	{
		template: "find{?year*}",
		data: map[string]any{
			"year": []string{"1965", "2000", "2012"},
			"dom":  []string{"example", "com"},
		},
		want: "find?year=1965&year=2000&year=2012",
	},
	{
		template: "www{.dom*}",
		data: map[string]any{
			"year": []string{"1965", "2000", "2012"},
			"dom":  []string{"example", "com"},
		},
		want: "www.example.com",
	},

	// Variable expansion.
	{
		template: "{count}",
		data:     expansionSectionData,
		want:     "one,two,three",
	},
	{
		template: "{count*}",
		data:     expansionSectionData,
		want:     "one,two,three",
	},
	{
		template: "{/count}",
		data:     expansionSectionData,
		want:     "/one,two,three",
	},
	{
		template: "{/count*}",
		data:     expansionSectionData,
		want:     "/one/two/three",
	},
	{
		template: "{;count}",
		data:     expansionSectionData,
		want:     ";count=one,two,three",
	},
	{
		template: "{;count*}",
		data:     expansionSectionData,
		want:     ";count=one;count=two;count=three",
	},
	{
		template: "{?count}",
		data:     expansionSectionData,
		want:     "?count=one,two,three",
	},
	{
		template: "{?count*}",
		data:     expansionSectionData,
		want:     "?count=one&count=two&count=three",
	},
	{
		template: "{&count*}",
		data:     expansionSectionData,
		want:     "&count=one&count=two&count=three",
	},
}

func TestExpand(t *testing.T) {
	for _, test := range tests {
		got, err := Expand(test.template, test.data)
		if got != test.want || err != nil {
			t.Errorf("Expand(%q, %#v) = %q, %v; want %q, <nil>",
				test.template, test.data, got, err, test.want)
		}
	}
}

func BenchmarkExpand(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Expand("{var}", expansionSectionData)
		}
	})

	b.Run("Complex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Expand("{.dom*}/{keys}{?list}", expansionSectionData)
		}
	})
}

func FuzzExpand(f *testing.F) {
	for _, test := range tests {
		f.Add(test.template)
	}

	f.Fuzz(func(t *testing.T, template string) {
		Expand(template, expansionSectionData)
	})
}
