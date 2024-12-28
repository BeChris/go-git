package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A tool to migrate gotest test sourcecode to use testify
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage : %s <dir>\n", os.Args[0])
		os.Exit(1)
	}

	entries, err := os.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(os.Args[1], entry.Name())

		if !strings.HasSuffix(fullPath, "_test.go") {
			continue
		}

		fmt.Println("Read", fullPath)
		fileContent, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatal(err)
		}

		lines := strings.Split(string(fileContent), "\n")

		newFileContent := modifyFile(lines)

		fmt.Println("Write", fullPath)
		err = os.WriteFile(fullPath, []byte(strings.Join(newFileContent, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func modifyFile(content []string) []string {
	simpleReplacements := map[string]string{
		"(c *C)":                                  "()",
		"c.Assert(err, IsNil)":                    "s.NoError(err)",
		"c.Assert(err, Not(IsNil))":               "s.Error(err)",
		"c.Assert(err, Equals, nil)":              "s.NoError(err)",
		"func Test(t *testing.T) { TestingT(t) }": "",
		"SetUpSuite(":                             "SetupSuite(",
		"SetUpTest(":                              "SetupTest(",
	}

	regexpReplacements := []map[string]string{
		map[string]string{
			`Commentf\((.+)\)`:                     `fmt.Sprintf($1)`,
			`c\.Assert\(err, ErrorMatches, (.+)\)`: `s.ErrorContains(err, $1)`,
			`c\.Log\((.+)\)`:                       `s.T().Log($1)`,
		},
		map[string]string{
			`c\.Assert\(err, IsNil, (.+)\)`:             `s.NoError(err, $1)`,
			`c\.Assert\(len\(([^)]+)\), Equals, (.+)\)`: `s.Len($1, $2)`,
			`c\.Assert\((.+), Equals, (.+), (.+)\)`:     `s.Equal($2, $1, $3)`,
		},
		map[string]string{
			`c\.Assert\(err, Equals, (.+)\)`:        `s.Equal(err, $1)`,
			`c\.Assert\(err, DeepEquals, (.+)\)`:    `s.Equal(err, $1)`,
			`c\.Assert\((.+), Equals, true\)`:       `s.True($1)`,
			`c\.Assert\((.+), Equals, false\)`:      `s.False($1)`,
			`c\.Assert\((.+), IsNil\)`:              `s.Nil($1)`,
			`c\.Assert\((.+), Not\(IsNil\)\)`:       `s.NotNil($1)`,
			`c\.Assert\((.+), Not\(IsNil\), (.+)\)`: `s.NotNil($1, $2)`,
			`c\.Assert\((.+), NotNil\)`:             `s.NotNil($1)`,
		},
		map[string]string{
			`c\.Assert\((.+), Equals, (.+)\)`:            `s.Equal($2, $1)`,
			`c\.Assert\((.+), Not\(Equals\), (.+)\)`:     `s.NotEqual($2, $1)`,
			`c\.Assert\((.+), DeepEquals, (.+)\)`:        `s.Equal($2, $1)`,
			`c\.Assert\((.+), Not\(DeepEquals\), (.+)\)`: `s.NotEqual($2, $1)`,
			`c\.Assert\((.+), HasLen, (.+)\)`:            `s.Len($1, $2)`,
			`c\.Assert\((.+), FitsTypeOf, (.+)\)`:        `s.IsType($2, $1)`,
		},
	}

	result := []string{}

	for _, line := range content {
		for originalString, replacement := range simpleReplacements {
			line = strings.ReplaceAll(line, originalString, replacement)
		}

		for _, regexpReplacementMap := range regexpReplacements {
			for regExpression, replacement := range regexpReplacementMap {
				rx := regexp.MustCompile(regExpression)

				line = rx.ReplaceAllString(line, replacement)
			}
		}

		emptySuiteTypeRe := regexp.MustCompile(`type (.+Suite) struct\{\}`)
		emptySuiteTypeReplacedLine := emptySuiteTypeRe.ReplaceAllString(line, "type $1 struct {")

		suiteTypeRe := regexp.MustCompile(`type (.+Suite) struct \{`)
		suiteTypeReplacedLine := suiteTypeRe.ReplaceAllString(line, "type $1 struct {")

		suiteVarRe := regexp.MustCompile(`var _ = Suite\(&(.+Suite)\{\}\)`)
		suiteVarReplacedLine := suiteVarRe.ReplaceAllString(line, "func Test$1(t *testing.T) {")
		suiteRunReplacedLine := suiteVarRe.ReplaceAllString(line, "\tsuite.Run(t, new($1))")

		if emptySuiteTypeRe.FindStringSubmatch(line) != nil {
			result = append(result, emptySuiteTypeReplacedLine)
			result = append(result, "\tsuite.Suite")
			result = append(result, "}")
		} else if suiteTypeRe.FindStringSubmatch(line) != nil {
			result = append(result, suiteTypeReplacedLine)
			result = append(result, "\tsuite.Suite")
		} else if suiteVarRe.FindStringSubmatch(line) != nil {
			result = append(result, suiteVarReplacedLine)
			result = append(result, suiteRunReplacedLine)
			result = append(result, "}")
		} else {
			result = append(result, line)
		}
	}

	return result
}
