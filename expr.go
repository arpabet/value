/*
 *
 * Copyright 2020-present Arpabet, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package value

import (
	"strings"
)

const (
	exprSep = "."
)

type expression struct {
	path []string
}

func Expression(str string) *expression {
	return &expression{strings.Split(str, exprSep)}
}

func (e expression) Empty() bool {
	return len(e.path) == 0
}

func (e expression) Size() int {
	return len(e.path)
}

func (e expression) GetAt(index int) string {
	if index < 0 || index >= len(e.path) {
		return ""
	}
	return e.path[index]
}

func (e expression) GetPath() []string {
	return e.path
}

func (e expression) String() string {
	return strings.Join(e.path, exprSep)
}
