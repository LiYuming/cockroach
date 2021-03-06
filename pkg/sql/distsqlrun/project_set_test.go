// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package distsqlrun

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

func TestProjectSet(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := [10]sqlbase.EncDatum{}
	for i := range v {
		v[i] = intEncDatum(i)
	}
	null := nullEncDatum()

	testCases := []struct {
		description string
		spec        ProjectSetSpec
		input       sqlbase.EncDatumRows
		inputTypes  []sqlbase.ColumnType
		expected    sqlbase.EncDatumRows
	}{
		{
			description: "scalar function",
			spec: ProjectSetSpec{
				Exprs: []Expression{
					{Expr: "@1 + 1"},
				},
				GeneratedColumns: oneIntCol,
				NumColsPerGen:    []uint32{1},
			},
			input: sqlbase.EncDatumRows{
				{v[2]},
			},
			inputTypes: oneIntCol,
			expected: sqlbase.EncDatumRows{
				{v[2], v[3]},
			},
		},
		{
			description: "set-returning function",
			spec: ProjectSetSpec{
				Exprs: []Expression{
					{Expr: "generate_series(@1, 2)"},
				},
				GeneratedColumns: oneIntCol,
				NumColsPerGen:    []uint32{1},
			},
			input: sqlbase.EncDatumRows{
				{v[0]},
				{v[1]},
			},
			inputTypes: oneIntCol,
			expected: sqlbase.EncDatumRows{
				{v[0], v[0]},
				{v[0], v[1]},
				{v[0], v[2]},
				{v[1], v[1]},
				{v[1], v[2]},
			},
		},
		{
			description: "multiple exprs with different lengths",
			spec: ProjectSetSpec{
				Exprs: []Expression{
					{Expr: "0"},
					{Expr: "generate_series(0, 0)"},
					{Expr: "generate_series(0, 1)"},
					{Expr: "generate_series(0, 2)"},
				},
				GeneratedColumns: intCols(4),
				NumColsPerGen:    []uint32{1, 1, 1, 1},
			},
			input: sqlbase.EncDatumRows{
				{v[0]},
			},
			inputTypes: oneIntCol,
			expected: sqlbase.EncDatumRows{
				{v[0], v[0], v[0], v[0], v[0]},
				{v[0], null, null, v[1], v[1]},
				{v[0], null, null, null, v[2]},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.description, func(t *testing.T) {
			runProcessorTest(
				t,
				ProcessorCoreUnion{ProjectSet: &c.spec},
				PostProcessSpec{},
				c.inputTypes,
				c.input,
				append(c.inputTypes, c.spec.GeneratedColumns...), /* outputTypes */
				c.expected,
			)
		})
	}
}
