// Copyright 2020 ConsenSys Software Inc.
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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package gkr

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/test_vector_utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestNoGateTwoInstances(t *testing.T) {
	// Testing a single instance is not possible because the sumcheck implementation doesn't cover the trivial 0-variate case
	testNoGate(t, []fr.Element{four, three})
}

func TestNoGate(t *testing.T) {
	testManyInstances(t, 1, testNoGate)
}

func TestSingleMulGateTwoInstances(t *testing.T) {
	testSingleMulGate(t, []fr.Element{four, three}, []fr.Element{two, three})
}

func TestSingleMulGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMulGate)
}

func TestSingleInputTwoIdentityGatesTwoInstances(t *testing.T) {

	testSingleInputTwoIdentityGates(t, []fr.Element{two, three})
}

func TestSingleInputTwoIdentityGates(t *testing.T) {

	testManyInstances(t, 2, testSingleInputTwoIdentityGates)
}

func TestSingleInputTwoIdentityGatesComposedTwoInstances(t *testing.T) {
	testSingleInputTwoIdentityGatesComposed(t, []fr.Element{two, one})
}

func TestSingleInputTwoIdentityGatesComposed(t *testing.T) {
	testManyInstances(t, 1, testSingleInputTwoIdentityGatesComposed)
}

func TestSingleMimcCipherGateTwoInstances(t *testing.T) {
	testSingleMimcCipherGate(t, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestSingleMimcCipherGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMimcCipherGate)
}

func TestATimesBSquaredTwoInstances(t *testing.T) {
	testATimesBSquared(t, 2, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestShallowMimcTwoInstances(t *testing.T) {
	testMimc(t, 2, []fr.Element{one, one}, []fr.Element{one, two})
}
func TestMimcTwoInstances(t *testing.T) {
	testMimc(t, 93, []fr.Element{one, one}, []fr.Element{one, two})
}

func TestMimc(t *testing.T) {
	testManyInstances(t, 2, generateTestMimc(93))
}

func generateTestMimc(numRounds int) func(*testing.T, ...[]fr.Element) {
	return func(t *testing.T, inputAssignments ...[]fr.Element) {
		testMimc(t, numRounds, inputAssignments...)
	}
}

func TestRecreateSumcheckErrorFromSingleInputTwoIdentityGatesGateTwoInstances(t *testing.T) {
	circuit := Circuit{{Wire{
		Gate:       nil,
		Inputs:     []*Wire{},
		NumOutputs: 2,
	}}}

	wire := &circuit[0][0]

	assignment := WireAssignment{&circuit[0][0]: []fr.Element{two, three}}

	claimsManagerGen := func() *claimsManager {
		manager := newClaimsManager(circuit, assignment)
		manager.add(wire, []fr.Element{three}, five)
		manager.add(wire, []fr.Element{four}, six)
		return &manager
	}

	transcriptGen := sumcheck.NewMessageCounterGenerator(4, 1)

	proof := sumcheck.Prove(claimsManagerGen().getClaim(wire), transcriptGen())
	sumcheck.Verify(claimsManagerGen().getLazyClaim(wire), proof, transcriptGen())
}

// complete the circuit evaluation from input values
func (a WireAssignment) complete(c Circuit) WireAssignment {
	numEvaluations := len(a[&c[len(c)-1][0]])

	for i := len(c) - 2; i >= 0; i-- { //there can only be input wires in the bottommost layer
		layer := c[i]
		for j := 0; j < len(layer); j++ {
			wire := &layer[j]

			if !wire.IsInput() {
				evals := make([]fr.Element, numEvaluations)
				ins := make([]fr.Element, len(wire.Inputs))
				for k := 0; k < numEvaluations; k++ {
					for inI, in := range wire.Inputs {
						ins[inI] = a[in][k]
					}
					evals[k] = wire.Gate.Evaluate(ins...)
				}
				a[wire] = evals
			}
		}
	}
	return a
}

var one, two, three, four, five, six fr.Element

func init() {
	one.SetOne()
	two.Double(&one)
	three.Add(&two, &one)
	four.Double(&two)
	five.Add(&three, &two)
	six.Double(&three)
}

var testManyInstancesLogMaxInstances = -1

func getLogMaxInstances(t *testing.T) int {
	if testManyInstancesLogMaxInstances == -1 {

		s := os.Getenv("GKR_LOG_INSTANCES")
		if s == "" {
			testManyInstancesLogMaxInstances = 5
		} else {
			var err error
			testManyInstancesLogMaxInstances, err = strconv.Atoi(s)
			if err != nil {
				t.Error(err)
			}
		}

	}
	return testManyInstancesLogMaxInstances
}

func testManyInstances(t *testing.T, numInput int, test func(*testing.T, ...[]fr.Element)) {
	fullAssignments := make([][]fr.Element, numInput)
	maxSize := 1 << getLogMaxInstances(t)

	t.Log("Entered test orchestrator, assigning and randomizing inputs")

	for i := range fullAssignments {
		fullAssignments[i] = polynomial.Make(maxSize)
		setRandom(fullAssignments[i])
	}

	defer polynomial.Dump(fullAssignments...)

	inputAssignments := make([][]fr.Element, numInput)
	for numEvals := maxSize; numEvals <= maxSize; numEvals *= 2 {
		for i, fullAssignment := range fullAssignments {
			inputAssignments[i] = fullAssignment[:numEvals]
		}

		t.Log("Selected inputs for test")
		test(t, inputAssignments...)
	}
}

func testNoGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := Circuit{
		{
			{
				Inputs:     []*Wire{},
				NumOutputs: 1,
				Gate:       nil,
			},
		},
	}

	assignment := WireAssignment{&c[0][0]: inputAssignments[0]}

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	// Even though a hash is called here, the proof is empty

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}
}

func testSingleMulGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{{
		Inputs:     []*Wire{&c[1][0], &c[1][1]},
		NumOutputs: 1,
		Gate:       mulGate{},
	}}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0], &c[1][1]: inputAssignments[1]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testSingleInputTwoIdentityGates(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 2,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       IdentityGate{},
		},
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       IdentityGate{},
		},
	}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testSingleMimcCipherGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{
		{
			Inputs:     []*Wire{&c[1][0], &c[1][1]},
			NumOutputs: 1,
			Gate:       mimcCipherGate{},
		},
	}
	t.Log("Evaluating all circuit wires")
	assignment := WireAssignment{&c[1][0]: inputAssignments[0], &c[1][1]: inputAssignments[1]}.complete(c)
	t.Log("Circuit evaluation complete")
	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))
	t.Log("Proof complete")
	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}
	t.Log("Successful verification complete")
	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
	t.Log("Unsuccessful verification complete")
}

func testSingleInputTwoIdentityGatesComposed(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 3)

	c[2] = CircuitLayer{{
		Gate:       nil,
		Inputs:     []*Wire{},
		NumOutputs: 1,
	}}
	c[1] = CircuitLayer{{
		Gate:       IdentityGate{},
		Inputs:     []*Wire{&c[2][0]},
		NumOutputs: 1,
	}}
	c[0] = CircuitLayer{{
		Gate:       IdentityGate{},
		Inputs:     []*Wire{&c[1][0]},
		NumOutputs: 1,
	}}

	assignment := WireAssignment{&c[2][0]: inputAssignments[0]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testMimc(t *testing.T, numRounds int, inputAssignments ...[]fr.Element) {
	//TODO: Implement mimc correctly. Currently, the computation is mimc(a,b) = cipher( cipher( ... cipher(a, b), b) ..., b)
	// @AlexandreBelling: Please explain the extra layers in https://github.com/ConsenSys/gkr-mimc/blob/81eada039ab4ed403b7726b535adb63026e8011f/examples/mimc.go#L10

	c := make(Circuit, numRounds+1)

	c[numRounds] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: numRounds,
			Gate:       nil,
		},
	}

	for i := numRounds; i > 0; i-- {
		c[i-1] = CircuitLayer{
			{
				Inputs:     []*Wire{&c[i][0], &c[numRounds][1]},
				NumOutputs: 1,
				Gate:       mimcCipherGate{}, //TODO: Put arks in there
			},
		}
	}

	t.Log("Evaluating all circuit wires")
	assignment := WireAssignment{&c[numRounds][0]: inputAssignments[0], &c[numRounds][1]: inputAssignments[1]}.complete(c)
	t.Log("Circuit evaluation complete")

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	t.Log("Proof finished")
	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	t.Log("Successful verification finished")
	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
	t.Log("Unsuccessful verification finished")
}

func testATimesBSquared(t *testing.T, numRounds int, inputAssignments ...[]fr.Element) {
	// This imitates the MiMC circuit

	c := make(Circuit, numRounds+1)

	c[numRounds] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: numRounds,
			Gate:       nil,
		},
	}

	for i := numRounds; i > 0; i-- {
		c[i-1] = CircuitLayer{
			{
				Inputs:     []*Wire{&c[i][0], &c[numRounds][1]},
				NumOutputs: 1,
				Gate:       mulGate{},
			},
		}
	}

	assignment := WireAssignment{&c[numRounds][0]: inputAssignments[0], &c[numRounds][1]: inputAssignments[1]}.complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

func setRandom(slice []fr.Element) {
	for i := range slice {
		slice[i].SetRandom()
	}
}

func generateTestProver(path string) func(t *testing.T) {
	return func(t *testing.T) {
		testCase, err := newTestCase(path)
		assert.NoError(t, err)
		testCase.Transcript.Update(0)
		proof := Prove(testCase.Circuit, testCase.FullAssignment, testCase.Transcript)
		assert.NoError(t, proofEquals(testCase.Proof, proof))
	}
}

func generateTestVerifier(path string) func(t *testing.T) {
	return func(t *testing.T) {
		testCase, err := newTestCase(path)
		assert.NoError(t, err)
		testCase.Transcript.Update(0)
		success := Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
		assert.True(t, success)

		testCase, err = newTestCase(path)
		assert.NoError(t, err)
		testCase.Transcript.Update(1)
		success = Verify(testCase.Circuit, testCase.InOutAssignment, testCase.Proof, testCase.Transcript)
		assert.False(t, success)
	}
}

func TestGkrVectors(t *testing.T) {

	testDirPath := "../../../../internal/generator/gkr/test_vectors"
	dirEntries, err := os.ReadDir(testDirPath)
	assert.NoError(t, err)
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {

			if filepath.Ext(dirEntry.Name()) == ".json" {
				path := filepath.Join(testDirPath, dirEntry.Name())
				noExt := dirEntry.Name()[:len(dirEntry.Name())-len(".json")]

				t.Run(noExt+"_prover", generateTestProver(path))
				t.Run(noExt+"_verifier", generateTestVerifier(path))

			}
		}
	}
}

// TODO: Move into test_vector_utils package
func TestTestHash(t *testing.T) {
	m, err := test_vector_utils.GetHash("../../../../internal/generator/gkr/test_vectors/resources/hash.json")
	assert.NoError(t, err)
	var one, two, negFour fr.Element
	one.SetOne()
	two.SetInt64(2)
	negFour.SetInt64(-4)

	h := m.FindPair(&one, &two)
	assert.True(t, h.Equal(&negFour), "expected -4, saw %s", h.Text(10))
}

func proofEquals(expected Proof, seen Proof) error {
	if len(expected) != len(seen) {
		return fmt.Errorf("length mismatch %d ≠ %d", len(expected), len(seen))
	}
	for i, x := range expected {
		xSeen := seen[i]
		if len(expected) != len(seen) {
			return fmt.Errorf("length mismatch %d ≠ %d", len(x), len(xSeen))
		}
		for j, y := range x {
			ySeen := xSeen[j]

			if ySeen.FinalEvalProof == nil {
				if seenFinalEval := y.FinalEvalProof.([]fr.Element); 0 != len(seenFinalEval) {
					return fmt.Errorf("length mismatch %d ≠ %d", 0, len(seenFinalEval))
				}
			} else {
				if err := test_vector_utils.SliceEquals(y.FinalEvalProof.([]fr.Element), ySeen.FinalEvalProof.([]fr.Element)); err != nil {
					return fmt.Errorf("final evaluation proof mismatch")
				}
			}
			if err := test_vector_utils.PolynomialSliceEquals(y.PartialSumPolys, ySeen.PartialSumPolys); err != nil {
				return err
			}
		}
	}
	return nil
}

type WireInfo struct {
	Gate   string  `json:"gate"`
	Inputs [][]int `json:"inputs"`
}

type CircuitInfo [][]WireInfo

var circuitCache = make(map[string]Circuit)

func getCircuit(path string) (Circuit, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if circuit, ok := circuitCache[path]; ok {
		return circuit, nil
	}
	var bytes []byte
	if bytes, err = os.ReadFile(path); err == nil {
		var circuitInfo CircuitInfo
		if err = json.Unmarshal(bytes, &circuitInfo); err == nil {
			circuit := circuitInfo.toCircuit()
			circuitCache[path] = circuit
			return circuit, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (c CircuitInfo) toCircuit() (circuit Circuit) {
	isOutput := make(map[*Wire]interface{})
	circuit = make(Circuit, len(c))
	for i := len(c) - 1; i >= 0; i-- {
		circuit[i] = make(CircuitLayer, len(c[i]))
		for j, wireInfo := range c[i] {
			circuit[i][j].Gate = gates[wireInfo.Gate]
			circuit[i][j].Inputs = make([]*Wire, len(wireInfo.Inputs))
			isOutput[&circuit[i][j]] = nil
			for k, inputCoord := range wireInfo.Inputs {
				if len(inputCoord) != 2 {
					panic("circuit wire has two coordinates")
				}
				input := &circuit[inputCoord[0]][inputCoord[1]]
				input.NumOutputs++
				circuit[i][j].Inputs[k] = input
				delete(isOutput, input)
			}
			if (i == len(c)-1) != (len(circuit[i][j].Inputs) == 0) {
				panic("wire is input if and only if in last layer")
			}
		}
	}

	for k := range isOutput {
		k.NumOutputs = 1
	}

	return
}

var gates map[string]Gate

func init() {
	gates = make(map[string]Gate)
	gates["identity"] = IdentityGate{}
	gates["mul"] = mulGate{}
	gates["mimc"] = mimcCipherGate{} //TODO: Add ark
}

type mimcCipherGate struct {
	ark fr.Element
}

func (m mimcCipherGate) Evaluate(input ...fr.Element) (res fr.Element) {
	var sum fr.Element

	sum.
		Add(&input[0], &input[1]).
		Add(&sum, &m.ark)

	res.Square(&sum)    // sum^2
	res.Mul(&res, &sum) // sum^3
	res.Square(&res)    //sum^6
	res.Mul(&res, &sum) //sum^7

	return
}

func (m mimcCipherGate) Degree() int {
	return 7
}

type PrintableProof [][]PrintableSumcheckProof

type PrintableSumcheckProof struct {
	FinalEvalProof  interface{}     `json:"finalEvalProof"`
	PartialSumPolys [][]interface{} `json:"partialSumPolys"`
}

func unmarshalProof(printable PrintableProof) (Proof, error) {
	proof := make(Proof, len(printable))
	for i := range printable {
		proof[i] = make([]sumcheck.Proof, len(printable[i]))
		for j, printableSumcheck := range printable[i] {
			finalEvalProof := []fr.Element(nil)

			if printableSumcheck.FinalEvalProof != nil {
				finalEvalSlice := reflect.ValueOf(printableSumcheck.FinalEvalProof)
				finalEvalProof = make([]fr.Element, finalEvalSlice.Len())
				for k := range finalEvalProof {
					if _, err := test_vector_utils.SetElement(&finalEvalProof[k], finalEvalSlice.Index(k).Interface()); err != nil {
						return nil, err
					}
				}
			}

			proof[i][j] = sumcheck.Proof{
				PartialSumPolys: make([]polynomial.Polynomial, len(printableSumcheck.PartialSumPolys)),
				FinalEvalProof:  finalEvalProof,
			}
			for k := range printableSumcheck.PartialSumPolys {
				var err error
				if proof[i][j].PartialSumPolys[k], err = test_vector_utils.SliceToElementSlice(printableSumcheck.PartialSumPolys[k]); err != nil {
					return nil, err
				}
			}
		}
	}
	return proof, nil
}

type TestCase struct {
	Circuit         Circuit
	Transcript      sumcheck.ArithmeticTranscript
	Proof           Proof
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
}

type TestCaseInfo struct {
	Hash    string          `json:"hash"`
	Circuit string          `json:"circuit"`
	Input   [][]interface{} `json:"input"`
	Output  [][]interface{} `json:"output"`
	Proof   PrintableProof  `json:"proof"`
}

type ParsedTestCase struct {
	FullAssignment  WireAssignment
	InOutAssignment WireAssignment
	Proof           Proof
	Hash            test_vector_utils.HashMap
	Circuit         Circuit
}

var parsedTestCases = make(map[string]*ParsedTestCase)

func newTestCase(path string) (*TestCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)

	parsedCase, ok := parsedTestCases[path]
	if !ok {
		var bytes []byte
		if bytes, err = os.ReadFile(path); err == nil {
			var info TestCaseInfo
			err = json.Unmarshal(bytes, &info)
			if err != nil {
				return nil, err
			}

			var circuit Circuit
			if circuit, err = getCircuit(filepath.Join(dir, info.Circuit)); err != nil {
				return nil, err
			}
			var hash test_vector_utils.HashMap
			if hash, err = test_vector_utils.GetHash(filepath.Join(dir, info.Hash)); err != nil {
				return nil, err
			}
			var proof Proof
			if proof, err = unmarshalProof(info.Proof); err != nil {
				return nil, err
			}

			fullAssignment := make(WireAssignment)
			inOutAssignment := make(WireAssignment)
			assignmentSize := len(info.Input[0])

			{
				i := len(circuit) - 1

				if expected, seen := len(circuit[i]), len(info.Input); expected != seen {
					return nil, fmt.Errorf("input layer length %d must match that of input vector %d", expected, seen)
				}

				for j := range circuit[i] {
					wire := &circuit[i][j]
					var wireAssignment []fr.Element
					if wireAssignment, err = test_vector_utils.SliceToElementSlice(info.Input[j]); err != nil {
						return nil, err
					}
					fullAssignment[wire] = wireAssignment
					inOutAssignment[wire] = wireAssignment
				}
			}

			for i := len(circuit) - 2; i >= 0; i-- {
				for j := range circuit[i] {
					wire := &circuit[i][j]
					assignment := make(polynomial.MultiLin, assignmentSize)
					in := make([]fr.Element, len(wire.Inputs))
					for k := range assignment {
						for l, inputWire := range circuit[i][j].Inputs {
							in[l] = fullAssignment[inputWire][k]
						}
						assignment[k] = wire.Gate.Evaluate(in...)
					}

					fullAssignment[wire] = assignment
				}
			}

			if expected, seen := len(circuit[0]), len(info.Output); expected != seen {
				return nil, fmt.Errorf("output layer length %d must match that of input vector %d", expected, seen)
			}
			for j := range circuit[0] {
				wire := &circuit[0][j]
				if inOutAssignment[wire], err = test_vector_utils.SliceToElementSlice(info.Output[j]); err != nil {
					return nil, err
				}
				if err = test_vector_utils.SliceEquals(inOutAssignment[wire], fullAssignment[wire]); err != nil {
					return nil, err
				}
			}

			parsedCase = &ParsedTestCase{
				FullAssignment:  fullAssignment,
				InOutAssignment: inOutAssignment,
				Proof:           proof,
				Hash:            hash,
				Circuit:         circuit,
			}

			parsedTestCases[path] = parsedCase
		} else {
			return nil, err
		}
	}

	return &TestCase{
		Circuit:         parsedCase.Circuit,
		Transcript:      &test_vector_utils.MapHashTranscript{HashMap: parsedCase.Hash},
		FullAssignment:  parsedCase.FullAssignment,
		InOutAssignment: parsedCase.InOutAssignment,
		Proof:           parsedCase.Proof,
	}, nil
}

type mulGate struct{}

func (m mulGate) Evaluate(element ...fr.Element) (result fr.Element) {
	result.Mul(&element[0], &element[1])
	return
}

func (m mulGate) Degree() int {
	return 2
}
