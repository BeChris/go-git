package pktline_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/stretchr/testify/suite"
)

type SuiteEncoder struct {
	suite.Suite
}

func TestSuiteEncoder(t *testing.T) {
	suite.Run(t, new(SuiteEncoder))
}

func (s *SuiteEncoder) TestFlush() {
	var buf bytes.Buffer
	e := pktline.NewEncoder(&buf)

	err := e.Flush()
	s.NoError(err)

	obtained := buf.Bytes()
	s.Equal(pktline.FlushPkt, obtained)
}

func (s *SuiteEncoder) TestEncode() {
	for i, test := range [...]struct {
		input    [][]byte
		expected []byte
	}{
		{
			input: [][]byte{
				[]byte("hello\n"),
			},
			expected: []byte("000ahello\n"),
		}, {
			input: [][]byte{
				[]byte("hello\n"),
				pktline.Flush,
			},
			expected: []byte("000ahello\n0000"),
		}, {
			input: [][]byte{
				[]byte("hello\n"),
				[]byte("world!\n"),
				[]byte("foo"),
			},
			expected: []byte("000ahello\n000bworld!\n0007foo"),
		}, {
			input: [][]byte{
				[]byte("hello\n"),
				pktline.Flush,
				[]byte("world!\n"),
				[]byte("foo"),
				pktline.Flush,
			},
			expected: []byte("000ahello\n0000000bworld!\n0007foo0000"),
		}, {
			input: [][]byte{
				[]byte(strings.Repeat("a", pktline.MaxPayloadSize)),
			},
			expected: []byte(
				"fff0" + strings.Repeat("a", pktline.MaxPayloadSize)),
		}, {
			input: [][]byte{
				[]byte(strings.Repeat("a", pktline.MaxPayloadSize)),
				[]byte(strings.Repeat("b", pktline.MaxPayloadSize)),
			},
			expected: []byte(
				"fff0" + strings.Repeat("a", pktline.MaxPayloadSize) +
					"fff0" + strings.Repeat("b", pktline.MaxPayloadSize)),
		},
	} {
		comment := fmt.Sprintf("input %d = %v\n", i, test.input)

		var buf bytes.Buffer
		e := pktline.NewEncoder(&buf)

		err := e.Encode(test.input...)
		s.NoError(err, comment)

		s.Equal(test.expected, buf.Bytes(), comment)
	}
}

func (s *SuiteEncoder) TestEncodeErrPayloadTooLong() {
	for i, input := range [...][][]byte{
		{
			[]byte(strings.Repeat("a", pktline.MaxPayloadSize+1)),
		},
		{
			[]byte("hello world!"),
			[]byte(strings.Repeat("a", pktline.MaxPayloadSize+1)),
		},
		{
			[]byte("hello world!"),
			[]byte(strings.Repeat("a", pktline.MaxPayloadSize+1)),
			[]byte("foo"),
		},
	} {
		comment := fmt.Sprintf("input %d = %v\n", i, input)

		var buf bytes.Buffer
		e := pktline.NewEncoder(&buf)

		err := e.Encode(input...)
		s.Equal(pktline.ErrPayloadTooLong, err, comment)
	}
}

func (s *SuiteEncoder) TestEncodeStrings() {
	for i, test := range [...]struct {
		input    []string
		expected []byte
	}{
		{
			input: []string{
				"hello\n",
			},
			expected: []byte("000ahello\n"),
		}, {
			input: []string{
				"hello\n",
				pktline.FlushString,
			},
			expected: []byte("000ahello\n0000"),
		}, {
			input: []string{
				"hello\n",
				"world!\n",
				"foo",
			},
			expected: []byte("000ahello\n000bworld!\n0007foo"),
		}, {
			input: []string{
				"hello\n",
				pktline.FlushString,
				"world!\n",
				"foo",
				pktline.FlushString,
			},
			expected: []byte("000ahello\n0000000bworld!\n0007foo0000"),
		}, {
			input: []string{
				strings.Repeat("a", pktline.MaxPayloadSize),
			},
			expected: []byte(
				"fff0" + strings.Repeat("a", pktline.MaxPayloadSize)),
		}, {
			input: []string{
				strings.Repeat("a", pktline.MaxPayloadSize),
				strings.Repeat("b", pktline.MaxPayloadSize),
			},
			expected: []byte(
				"fff0" + strings.Repeat("a", pktline.MaxPayloadSize) +
					"fff0" + strings.Repeat("b", pktline.MaxPayloadSize)),
		},
	} {
		comment := fmt.Sprintf("input %d = %v\n", i, test.input)

		var buf bytes.Buffer
		e := pktline.NewEncoder(&buf)

		err := e.EncodeString(test.input...)
		s.NoError(err, comment)
		s.Equal(test.expected, buf.Bytes(), comment)
	}
}

func (s *SuiteEncoder) TestEncodeStringErrPayloadTooLong() {
	for i, input := range [...][]string{
		{
			strings.Repeat("a", pktline.MaxPayloadSize+1),
		},
		{
			"hello world!",
			strings.Repeat("a", pktline.MaxPayloadSize+1),
		},
		{
			"hello world!",
			strings.Repeat("a", pktline.MaxPayloadSize+1),
			"foo",
		},
	} {
		comment := fmt.Sprintf("input %d = %v\n", i, input)

		var buf bytes.Buffer
		e := pktline.NewEncoder(&buf)

		err := e.EncodeString(input...)
		s.Equal(pktline.ErrPayloadTooLong, err, comment)
	}
}

func (s *SuiteEncoder) TestEncodef() {
	format := " %s %d\n"
	str := "foo"
	d := 42

	var buf bytes.Buffer
	e := pktline.NewEncoder(&buf)

	err := e.Encodef(format, str, d)
	s.NoError(err)

	expected := []byte("000c foo 42\n")
	s.Equal(expected, buf.Bytes())
}
