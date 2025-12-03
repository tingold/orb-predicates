package predicates

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
)

// JTS XML test file format structures

// JTSTestRun represents the root element of a JTS test XML file
type JTSTestRun struct {
	XMLName xml.Name  `xml:"run"`
	Cases   []JTSCase `xml:"case"`
}

// JTSCase represents a single test case with geometries and operations
type JTSCase struct {
	Desc  string         `xml:"desc"`
	A     string         `xml:"a"`
	B     string         `xml:"b"`
	Tests []JTSTestBlock `xml:"test"`
}

// JTSTestBlock contains one or more operations to test
type JTSTestBlock struct {
	Op JTSOperation `xml:"op"`
}

// JTSOperation represents a single predicate operation test
type JTSOperation struct {
	Name     string `xml:"name,attr"`
	Arg1     string `xml:"arg1,attr"`
	Arg2     string `xml:"arg2,attr"`
	Arg3     string `xml:"arg3,attr"` // Used for relate pattern
	Expected string `xml:",chardata"`
}

// predicateFunc is a function type for spatial predicates
type predicateFunc func(a, b orb.Geometry) bool

// supportedPredicates maps JTS operation names to our predicate functions
var supportedPredicates = map[string]predicateFunc{
	"intersects": Intersects,
	"contains":   Contains,
	"within":     Within,
	"covers":     Covers,
	"coveredby":  CoveredBy, // JTS uses lowercase 'b'
	"crosses":    Crosses,
	"overlaps":   Overlaps,
	"touches":    Touches,
	"disjoint":   Disjoint,
}

// parseJTSTestFile reads and parses a JTS XML test file
func parseJTSTestFile(path string) (*JTSTestRun, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var testRun JTSTestRun
	if err := xml.Unmarshal(data, &testRun); err != nil {
		return nil, err
	}

	return &testRun, nil
}

// parseWKT parses a WKT string into an orb.Geometry
func parseWKT(wktStr string) (orb.Geometry, error) {
	// Clean up whitespace in WKT string
	wktStr = strings.TrimSpace(wktStr)
	// Normalize internal whitespace (JTS XML often has newlines in WKT)
	wktStr = strings.Join(strings.Fields(wktStr), " ")

	return wkt.Unmarshal(wktStr)
}

// parseExpected parses the expected result string to a boolean
func parseExpected(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "true"
}

// TestJTSPredicates runs all JTS XML test files against our predicate implementations
func TestJTSPredicates(t *testing.T) {
	files, err := filepath.Glob("testdata/jts/*.xml")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No JTS test files found in testdata/jts/")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			runJTSTestFile(t, file)
		})
	}
}

// runJTSTestFile executes all test cases in a single JTS XML file
func runJTSTestFile(t *testing.T, path string) {
	testRun, err := parseJTSTestFile(path)
	if err != nil {
		t.Fatalf("Failed to parse test file %s: %v", path, err)
	}

	for i, tc := range testRun.Cases {
		t.Run(tc.Desc, func(t *testing.T) {
			runJTSTestCase(t, tc, i)
		})
	}
}

// runJTSTestCase executes a single JTS test case
func runJTSTestCase(t *testing.T, tc JTSCase, caseIndex int) {
	// Parse geometry A
	geomA, err := parseWKT(tc.A)
	if err != nil {
		t.Logf("Skipping case %d (%s): failed to parse geometry A: %v", caseIndex, tc.Desc, err)
		t.SkipNow()
		return
	}

	// Parse geometry B (may be empty for some tests)
	var geomB orb.Geometry
	if strings.TrimSpace(tc.B) != "" {
		geomB, err = parseWKT(tc.B)
		if err != nil {
			t.Logf("Skipping case %d (%s): failed to parse geometry B: %v", caseIndex, tc.Desc, err)
			t.SkipNow()
			return
		}
	}

	// Run each test operation
	for _, test := range tc.Tests {
		op := test.Op
		opName := strings.ToLower(op.Name)

		// Determine argument order
		var argA, argB orb.Geometry
		if strings.ToUpper(op.Arg1) == "A" {
			argA = geomA
		} else {
			argA = geomB
		}
		if strings.ToUpper(op.Arg2) == "A" {
			argB = geomA
		} else {
			argB = geomB
		}

		// Skip if either geometry is nil
		if argA == nil || argB == nil {
			continue
		}

		// Skip operations we don't support
		predFunc, supported := supportedPredicates[opName]
		if !supported {
			continue
		}

		expected := parseExpected(op.Expected)
		actual := predFunc(argA, argB)

		if actual != expected {
			t.Errorf("%s(%s, %s) = %v, expected %v\n  A: %s\n  B: %s",
				opName, op.Arg1, op.Arg2, actual, expected,
				strings.TrimSpace(tc.A), strings.TrimSpace(tc.B))
		}
	}
}

// TestJTSSummary provides a summary of JTS test coverage
func TestJTSSummary(t *testing.T) {
	files, err := filepath.Glob("testdata/jts/*.xml")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No JTS test files found")
	}

	totalCases := 0
	totalOps := 0
	opCounts := make(map[string]int)

	for _, file := range files {
		testRun, err := parseJTSTestFile(file)
		if err != nil {
			t.Logf("Warning: Failed to parse %s: %v", file, err)
			continue
		}

		totalCases += len(testRun.Cases)
		for _, tc := range testRun.Cases {
			for _, test := range tc.Tests {
				opName := strings.ToLower(test.Op.Name)
				opCounts[opName]++
				totalOps++
			}
		}
	}

	t.Logf("JTS Test Summary:")
	t.Logf("  Files: %d", len(files))
	t.Logf("  Total cases: %d", totalCases)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Operations by type:")
	for op, count := range opCounts {
		_, supported := supportedPredicates[op]
		status := "supported"
		if !supported {
			status = "not implemented"
		}
		t.Logf("    %s: %d (%s)", op, count, status)
	}
}
