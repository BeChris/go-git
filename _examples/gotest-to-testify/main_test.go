package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	for _, testData := range []struct {
		lines    []string
		expected []string
	}{
		{
			[]string{"func (s *SubmoduleSuite) TestGitSubmodulesSymlink(c *C) {"},
			[]string{"func (s *SubmoduleSuite) TestGitSubmodulesSymlink() {"},
		},
		{
			[]string{"c.Assert(err, IsNil)"},
			[]string{"s.NoError(err)"},
		},
		// c.Assert(err, Not(IsNil))
		{
			[]string{"c.Assert(err, Not(IsNil))"},
			[]string{"s.Error(err)"},
		},
		{
			[]string{"c.Assert(err, IsNil, comment)"},
			[]string{"s.NoError(err, comment)"},
		},
		{
			[]string{"c.Assert(err, Equals, nil)"},
			[]string{"s.NoError(err)"},
		},
		{
			[]string{"c.Assert(err, Equals, ErrGitModulesSymlink)"},
			[]string{"s.Equal(err, ErrGitModulesSymlink)"},
		},
		{
			[]string{"c.Assert(str, Equals, expected[i].s)"},
			[]string{"s.Equal(expected[i].s, str)"},
		},
		//
		{
			[]string{"c.Assert(idx2, Equals, uint32(idx))"},
			[]string{"s.Equal(uint32(idx), idx2)"},
		},
		//
		{
			[]string{"c.Assert(ru.Entries[1].Stages[OurMode], Not(Equals), plumbing.ZeroHash)"},
			[]string{"s.NotEqual(plumbing.ZeroHash, ru.Entries[1].Stages[OurMode])"},
		},
		// c.Assert(b.Hash(), Not(DeepEquals), bb.Hash())
		{
			[]string{"c.Assert(b.Hash(), Not(DeepEquals), bb.Hash())"},
			[]string{"s.NotEqual(bb.Hash(), b.Hash())"},
		},
		{
			[]string{`c.Assert(hash.String()+":"+parent.String(), Equals, hash.String()+":"+commitData.ParentHashes[i].String())`},
			[]string{`s.Equal(hash.String()+":"+commitData.ParentHashes[i].String(), hash.String()+":"+parent.String())`},
		},
		{
			[]string{"c.Assert(obtained, Equals, test.expected, comment)"},
			[]string{"s.Equal(test.expected, obtained, comment)"},
		},
		//
		{
			[]string{"c.Assert(result, DeepEquals, expected)"},
			[]string{"s.Equal(expected, result)"},
		},
		{
			[]string{"c.Assert(err, DeepEquals, e)"},
			[]string{"s.Equal(err, e)"},
		},
		{
			[]string{"c.Assert(ok, Equals, true)"},
			[]string{"s.True(ok)"},
		},
		// c.Assert((&Option{Key: "key"}).IsKey("key"), Equals, true)
		{
			[]string{`c.Assert((&Option{Key: "key"}).IsKey("key"), Equals, true)`},
			[]string{`s.True((&Option{Key: "key"}).IsKey("key"))`},
		},
		{
			[]string{"c.Assert(checked, Equals, false)"},
			[]string{"s.False(checked)"},
		},
		{
			[]string{"c.Assert(obj, IsNil)"},
			[]string{"s.Nil(obj)"},
		},
		// c.Assert(obj, Not(IsNil))
		{
			[]string{"c.Assert(obj, Not(IsNil))"},
			[]string{"s.NotNil(obj)"},
		},
		// c.Assert(err, Not(IsNil), comment)
		{
			[]string{"c.Assert(obj, Not(IsNil), comment)"},
			[]string{"s.NotNil(obj, comment)"},
		},
		{
			[]string{"c.Assert(obj, NotNil)"},
			[]string{"s.NotNil(obj)"},
		},
		{
			[]string{"c.Assert(len(commitData.ParentIndexes), Equals, 0)"},
			[]string{"s.Len(commitData.ParentIndexes, 0)"},
		},
		{
			[]string{"c.Assert(ps, HasLen, 2)"},
			[]string{"s.Len(ps, 2)"},
		},
		{
			[]string{`c.Assert(fmt.Sprintf("%x", idx.PackfileChecksum), Equals, f.PackfileHash)`},
			[]string{`s.Equal(f.PackfileHash, fmt.Sprintf("%x", idx.PackfileChecksum))`},
		},
		{
			[]string{"c.Assert(p, FitsTypeOf, &Parser{})"},
			[]string{"s.IsType(&Parser{}, p)"},
		},
		{
			[]string{"type ParserSuite struct{}"},
			[]string{
				"type ParserSuite struct {",
				"\tsuite.Suite",
				"}",
			},
		},
		{
			[]string{"type ObjectSuite struct {"},
			[]string{
				"type ObjectSuite struct {",
				"\tsuite.Suite",
			},
		},
		{
			[]string{"var _ = Suite(&ParserSuite{})"},
			[]string{
				"func TestParserSuite(t *testing.T) {",
				"\tsuite.Run(t, new(ParserSuite))",
				"}",
			},
		},
		{
			[]string{"func Test(t *testing.T) { TestingT(t) }"},
			[]string{""},
		},
		{
			[]string{
				"SetUpSuite(",
				"SetUpTest(",
			},
			[]string{
				"SetupSuite(",
				"SetupTest(",
			},
		},
		{
			[]string{`comment := Commentf("input %d = %v\n", i, test.input)`},
			[]string{`comment := fmt.Sprintf("input %d = %v\n", i, test.input)`},
		},
		{
			[]string{`c.Assert(err, ErrorMatches, "malformed change.*")`},
			[]string{`s.ErrorContains(err, "malformed change.*")`},
		},
		{
			[]string{`c.Log("Executing test cases:", tc.description)`},
			[]string{`s.T().Log("Executing test cases:", tc.description)`},
		},
	} {
		modifiedContent := modifyFile(testData.lines)

		assert.Equal(t, testData.expected, modifiedContent, testData.lines)
	}
}
