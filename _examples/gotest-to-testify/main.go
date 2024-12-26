package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Basic example of how to clone a repository using clone options.
func main() {
	filePath := os.Args[1]

	fmt.Println("Read", filePath)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(fileContent), "\n")

	newFileContent := modifyFile(lines)

	fmt.Println("Write", filePath)
	err = os.WriteFile(filePath, []byte(strings.Join(newFileContent, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func modifyFile(content []string) []string {
	simpleReplacements := map[string]string{
		"(c *C)":                     "()",
		"c.Assert(err, IsNil)":       "s.NoError(err)",
		"c.Assert(err, Not(IsNil))":  "s.Error(err)",
		"c.Assert(err, Equals, nil)": "s.NoError(err)",
	}

	regexpReplacements := []map[string]string{
		map[string]string{
			`c\.Assert\(err, IsNil, (.+)\)`:             `s.NoError(err, $1)`,
			`c\.Assert\(len\(([^)]+)\), Equals, (.+)\)`: `s.Len($1, $2)`,
		},
		map[string]string{
			`c\.Assert\(err, Equals, (.+)\)`:        `s.ErrorIs(err, $1)`,
			`c\.Assert\(err, DeepEquals, (.+)\)`:    `s.ErrorIs(err, $1)`,
			`c\.Assert\((.+), Equals, true\)`:       `s.True($1)`,
			`c\.Assert\((.+), Equals, false\)`:      `s.False($1)`,
			`c\.Assert\((.+), IsNil\)`:              `s.Nil($1)`,
			`c\.Assert\((.+), Not\(IsNil\)\)`:       `s.NotNil($1)`,
			`c\.Assert\((.+), Not\(IsNil\), (.+)\)`: `s.NotNil($1, $2)`,
			`c\.Assert\((.+), NotNil\)`:             `s.NotNil($1)`,
		},
		map[string]string{
			`c\.Assert\((.+), Equals, (.+), (.+)\)`:      `s.Equal($2, $1, $3)`,
			`c\.Assert\((.+), Equals, (.+)\)`:            `s.Equal($2, $1)`,
			`c\.Assert\((.+), Not\(Equals\), (.+)\)`:     `s.NotEqual($2, $1)`,
			`c\.Assert\((.+), DeepEquals, (.+)\)`:        `s.Equal($2, $1)`,
			`c\.Assert\((.+), Not\(DeepEquals\), (.+)\)`: `s.NotEqual($2, $1)`,
			`c\.Assert\((.+), HasLen, (.+)\)`:            `s.Len($1, $2)`,
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

		result = append(result, line)
	}

	return result
}
