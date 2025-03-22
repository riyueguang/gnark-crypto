// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package polynomial

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: Property based tests?
func TestFoldBilinear(t *testing.T) {

	for i := 0; i < 100; i++ {

		// f = c₀ + c₁ X₁ + c₂ X₂ + c₃ X₁ X₂
		var coefficients [4]fr.Element
		fr.Vector(coefficients[:]).MustSetRandom()

		var r fr.Element
		r.MustSetRandom()

		// interpolate at {0,1}²:
		m := make(MultiLin, 4)
		m[0] = coefficients[0]
		m[1].Add(&coefficients[0], &coefficients[2])
		m[2].Add(&coefficients[0], &coefficients[1])
		m[3].
			Add(&m[1], &coefficients[1]).
			Add(&m[3], &coefficients[3])

		m.Fold(r)

		// interpolate at {r}×{0,1}:
		var expected0, expected1 fr.Element
		expected0.
			Mul(&r, &coefficients[1]).
			Add(&expected0, &coefficients[0])

		expected1.
			Mul(&r, &coefficients[3]).
			Add(&expected1, &coefficients[2]).
			Add(&expected0, &expected1)

		if !m[0].Equal(&expected0) || !m[1].Equal(&expected1) {
			t.Fail()
		}
	}
}

// TODO: Benchmark folding? Algorithms is pretty straightforward; unless we want to measure how well memory management is working

func TestFoldedEqTable(t *testing.T) {
	q := make([]fr.Element, 2)
	q[0].SetInt64(2)
	q[1].SetInt64(3)

	m := make(MultiLin, 4)
	m[0].SetOne()
	m.Eq(q)

	eq := make([]fr.Element, 4)
	p := make([]fr.Element, 2)

	var one fr.Element
	one.SetOne()

	for p0 := 0; p0 < 2; p0++ {
		p[1].SetZero()
		for p1 := 0; p1 < 2; p1++ {
			eq[p0*2+p1] = EvalEq(q, p)
			p[1].Add(&p[1], &one)
		}
		p[0].Add(&p[0], &one)
	}

	for i := 0; i < 4; i++ {
		assert.Equal(t, eq[i], m[i], "folded table disagrees with EqEval", i)
	}

}
