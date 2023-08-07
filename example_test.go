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

package uritemplate_test

import (
	"fmt"

	"zombiezen.com/go/uritemplate"
)

func ExampleExpand() {
	expanded, err := uritemplate.Expand("/foo{?var}", map[string]any{
		"var": "value",
	})
	if err != nil {
		// handle error
	}
	fmt.Println(expanded)
	// Output:
	// /foo?var=value
}

func ExampleExpand_struct() {
	var data struct {
		Color []string
	}
	data.Color = []string{"r", "g", "b"}

	expanded, err := uritemplate.Expand("/foo{?color*}", data)
	if err != nil {
		// handle error
	}
	fmt.Println(expanded)
	// Output:
	// /foo?color=r&color=g&color=b
}
