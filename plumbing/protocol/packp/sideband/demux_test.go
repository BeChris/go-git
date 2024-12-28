package sideband

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/stretchr/testify/suite"
)

type SidebandSuite struct {
	suite.Suite
}

func TestSidebandSuite(t *testing.T) {
	suite.Run(t, new(SidebandSuite))
}

func (s *SidebandSuite) TestDecode() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode(PackData.WithPayload(expected[0:8]))
	e.Encode(ProgressMessage.WithPayload([]byte{'F', 'O', 'O', '\n'}))
	e.Encode(PackData.WithPayload(expected[8:16]))
	e.Encode(PackData.WithPayload(expected[16:26]))

	content := make([]byte, 26)
	d := NewDemuxer(Sideband64k, buf)
	n, err := io.ReadFull(d, content)
	s.NoError(err)
	s.Equal(26, n)
	s.Equal(expected, content)
}

func (s *SidebandSuite) TestDecodeMoreThanContain() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode(PackData.WithPayload(expected))

	content := make([]byte, 42)
	d := NewDemuxer(Sideband64k, buf)
	n, err := io.ReadFull(d, content)
	s.Equal(err, io.ErrUnexpectedEOF)
	s.Equal(26, n)
	s.Equal(expected, content[0:26])
}

func (s *SidebandSuite) TestDecodeWithError() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode(PackData.WithPayload(expected[0:8]))
	e.Encode(ErrorMessage.WithPayload([]byte{'F', 'O', 'O', '\n'}))
	e.Encode(PackData.WithPayload(expected[8:16]))
	e.Encode(PackData.WithPayload(expected[16:26]))

	content := make([]byte, 26)
	d := NewDemuxer(Sideband64k, buf)
	n, err := io.ReadFull(d, content)
	s.ErrorContains(err, "unexpected error: FOO\n")
	s.Equal(8, n)
	s.Equal(expected[0:8], content[0:8])
}

type mockReader struct{}

func (r *mockReader) Read([]byte) (int, error) { return 0, errors.New("foo") }

func (s *SidebandSuite) TestDecodeFromFailingReader() {
	content := make([]byte, 26)
	d := NewDemuxer(Sideband64k, &mockReader{})
	n, err := io.ReadFull(d, content)
	s.ErrorContains(err, "foo")
	s.Equal(0, n)
}

func (s *SidebandSuite) TestDecodeWithProgress() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	input := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(input)
	e.Encode(PackData.WithPayload(expected[0:8]))
	e.Encode(ProgressMessage.WithPayload([]byte{'F', 'O', 'O', '\n'}))
	e.Encode(PackData.WithPayload(expected[8:16]))
	e.Encode(PackData.WithPayload(expected[16:26]))

	output := bytes.NewBuffer(nil)
	content := make([]byte, 26)
	d := NewDemuxer(Sideband64k, input)
	d.Progress = output

	n, err := io.ReadFull(d, content)
	s.NoError(err)
	s.Equal(26, n)
	s.Equal(expected, content)

	progress, err := io.ReadAll(output)
	s.NoError(err)
	s.Equal([]byte{'F', 'O', 'O', '\n'}, progress)
}

func (s *SidebandSuite) TestDecodeFlushEOF() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	input := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(input)
	e.Encode(PackData.WithPayload(expected[0:8]))
	e.Encode(ProgressMessage.WithPayload([]byte{'F', 'O', 'O', '\n'}))
	e.Encode(PackData.WithPayload(expected[8:16]))
	e.Encode(PackData.WithPayload(expected[16:26]))
	e.Flush()
	e.Encode(PackData.WithPayload([]byte("bar\n")))

	output := bytes.NewBuffer(nil)
	content := bytes.NewBuffer(nil)
	d := NewDemuxer(Sideband64k, input)
	d.Progress = output

	n, err := content.ReadFrom(d)
	s.NoError(err)
	s.Equal(int64(26), n)
	s.Equal(expected, content.Bytes())

	progress, err := io.ReadAll(output)
	s.NoError(err)
	s.Equal([]byte{'F', 'O', 'O', '\n'}, progress)
}

func (s *SidebandSuite) TestDecodeWithUnknownChannel() {
	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode([]byte{'4', 'F', 'O', 'O', '\n'})

	content := make([]byte, 26)
	d := NewDemuxer(Sideband64k, buf)
	n, err := io.ReadFull(d, content)
	s.ErrorContains(err, "unknown channel 4FOO\n")
	s.Equal(0, n)
}

func (s *SidebandSuite) TestDecodeWithPending() {
	expected := []byte("abcdefghijklmnopqrstuvwxyz")

	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode(PackData.WithPayload(expected[0:8]))
	e.Encode(PackData.WithPayload(expected[8:16]))
	e.Encode(PackData.WithPayload(expected[16:26]))

	content := make([]byte, 13)
	d := NewDemuxer(Sideband64k, buf)
	n, err := io.ReadFull(d, content)
	s.NoError(err)
	s.Equal(13, n)
	s.Equal(expected[0:13], content)

	n, err = d.Read(content)
	s.NoError(err)
	s.Equal(13, n)
	s.Equal(expected[13:26], content)
}

func (s *SidebandSuite) TestDecodeErrMaxPacked() {
	buf := bytes.NewBuffer(nil)
	e := pktline.NewEncoder(buf)
	e.Encode(PackData.WithPayload(bytes.Repeat([]byte{'0'}, MaxPackedSize+1)))

	content := make([]byte, 13)
	d := NewDemuxer(Sideband, buf)
	n, err := io.ReadFull(d, content)
	s.Equal(err, ErrMaxPackedExceeded)
	s.Equal(0, n)
}
