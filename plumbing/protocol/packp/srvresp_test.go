package packp

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/suite"
)

type ServerResponseSuite struct {
	suite.Suite
}

func TestServerResponseSuite(t *testing.T) {
	suite.Run(t, new(ServerResponseSuite))
}

func (s *ServerResponseSuite) TestDecodeNAK() {
	raw := "0008NAK\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NoError(err)

	s.Len(sr.ACKs, 0)
}

func (s *ServerResponseSuite) TestDecodeNewLine() {
	raw := "\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NotNil(err)
	s.Equal("invalid pkt-len found", err.Error())
}

func (s *ServerResponseSuite) TestDecodeEmpty() {
	raw := ""

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NoError(err)
}

func (s *ServerResponseSuite) TestDecodePartial() {
	raw := "000600\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NotNil(err)
	s.Equal(fmt.Sprintf("unexpected content %q", "00"), err.Error())
}

func (s *ServerResponseSuite) TestDecodeACK() {
	raw := "0031ACK 6ecf0ef2c2dffb796033e5a02219af86ec6584e5\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NoError(err)

	s.Len(sr.ACKs, 1)
	s.Equal(plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"), sr.ACKs[0])
}

func (s *ServerResponseSuite) TestDecodeMultipleACK() {
	raw := "" +
		"0031ACK 1111111111111111111111111111111111111111\n" +
		"0031ACK 6ecf0ef2c2dffb796033e5a02219af86ec6584e5\n" +
		"00080PACK\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NoError(err)

	s.Len(sr.ACKs, 2)
	s.Equal(plumbing.NewHash("1111111111111111111111111111111111111111"), sr.ACKs[0])
	s.Equal(plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"), sr.ACKs[1])
}

func (s *ServerResponseSuite) TestDecodeMultipleACKWithSideband() {
	raw := "" +
		"0031ACK 1111111111111111111111111111111111111111\n" +
		"0031ACK 6ecf0ef2c2dffb796033e5a02219af86ec6584e5\n" +
		"00080aaaa\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NoError(err)

	s.Len(sr.ACKs, 2)
	s.Equal(plumbing.NewHash("1111111111111111111111111111111111111111"), sr.ACKs[0])
	s.Equal(plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"), sr.ACKs[1])
}

func (s *ServerResponseSuite) TestDecodeMalformed() {
	raw := "0029ACK 6ecf0ef2c2dffb796033e5a02219af86ec6584e\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), false)
	s.NotNil(err)
}

// multi_ack isn't fully implemented, this ensures that Decode ignores that fact,
// as in some circumstances that's OK to assume so.
//
// TODO: Review as part of multi_ack implementation.
func (s *ServerResponseSuite) TestDecodeMultiACK() {
	raw := "" +
		"0031ACK 1111111111111111111111111111111111111111\n" +
		"0031ACK 6ecf0ef2c2dffb796033e5a02219af86ec6584e5\n" +
		"00080PACK\n"

	sr := &ServerResponse{}
	err := sr.Decode(bufio.NewReader(bytes.NewBufferString(raw)), true)
	s.NoError(err)

	s.Len(sr.ACKs, 2)
	s.Equal(plumbing.NewHash("1111111111111111111111111111111111111111"), sr.ACKs[0])
	s.Equal(plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"), sr.ACKs[1])
}
