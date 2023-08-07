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

func TestExpand(t *testing.T) {
	keysData := struct {
		Semi  string
		Dot   string
		Comma string
	}{";", ".", ","}
	_ = keysData

	tests := []struct {
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
		// {
		// 	template: "{keys}",
		// 	data: map[string]any{
		// 		"var":   "value",
		// 		"hello": "Hello World!",
		// 		"path":  "/foo/bar",
		// 		"list":  []string{"red", "green", "blue"},
		// 		"keys":  keysData,
		// 	},
		// 	want: "semi,%3B,dot,.,comma,%2C",
		// },
		// {
		// 	template: "{keys*}",
		// 	data: map[string]any{
		// 		"var":   "value",
		// 		"hello": "Hello World!",
		// 		"path":  "/foo/bar",
		// 		"list":  []string{"red", "green", "blue"},
		// 		"keys":  keysData,
		// 	},
		// 	want: "semi=%3B,dot=.,comma=%2C",
		// },
	}
	for _, test := range tests {
		got, err := Expand(test.template, test.data)
		if got != test.want || err != nil {
			t.Errorf("Expand(%q, %#v) = %q, %v; want %q, <nil>",
				test.template, test.data, got, err, test.want)
		}
	}
}
