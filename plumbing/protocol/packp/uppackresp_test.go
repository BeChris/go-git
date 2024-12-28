package packp

import (
	"bytes"
	"io"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/stretchr/testify/suite"
)

type UploadPackResponseSuite struct {
	suite.Suite
}

func TestUploadPackResponseSuite(t *testing.T) {
	suite.Run(t, new(UploadPackResponseSuite))
}

func (s *UploadPackResponseSuite) TestDecodeNAK() {
	raw := "0008NAK\nPACK"

	req := NewUploadPackRequest()
	res := NewUploadPackResponse(req)
	defer res.Close()

	err := res.Decode(io.NopCloser(bytes.NewBufferString(raw)))
	s.NoError(err)

	pack, err := io.ReadAll(res)
	s.NoError(err)
	s.Equal([]byte("PACK"), pack)
}

func (s *UploadPackResponseSuite) TestDecodeDepth() {
	raw := "00000008NAK\nPACK"

	req := NewUploadPackRequest()
	req.Depth = DepthCommits(1)

	res := NewUploadPackResponse(req)
	defer res.Close()

	err := res.Decode(io.NopCloser(bytes.NewBufferString(raw)))
	s.NoError(err)

	pack, err := io.ReadAll(res)
	s.NoError(err)
	s.Equal([]byte("PACK"), pack)
}

func (s *UploadPackResponseSuite) TestDecodeMalformed() {
	raw := "00000008ACK\nPACK"

	req := NewUploadPackRequest()
	req.Depth = DepthCommits(1)

	res := NewUploadPackResponse(req)
	defer res.Close()

	err := res.Decode(io.NopCloser(bytes.NewBufferString(raw)))
	s.NotNil(err)
}

// multi_ack isn't fully implemented, this ensures that Decode ignores that fact,
// as in some circumstances that's OK to assume so.
//
// TODO: Review as part of multi_ack implementation.
func (s *UploadPackResponseSuite) TestDecodeMultiACK() {
	req := NewUploadPackRequest()
	req.Capabilities.Set(capability.MultiACK)

	res := NewUploadPackResponse(req)
	defer res.Close()

	err := res.Decode(io.NopCloser(bytes.NewBuffer(nil)))
	s.NoError(err)
}

func (s *UploadPackResponseSuite) TestReadNoDecode() {
	req := NewUploadPackRequest()
	req.Capabilities.Set(capability.MultiACK)

	res := NewUploadPackResponse(req)
	defer res.Close()

	n, err := res.Read(nil)
	s.Equal(err, ErrUploadPackResponseNotDecoded)
	s.Equal(0, n)
}

func (s *UploadPackResponseSuite) TestEncodeNAK() {
	pf := io.NopCloser(bytes.NewBuffer([]byte("[PACK]")))
	req := NewUploadPackRequest()
	res := NewUploadPackResponseWithPackfile(req, pf)
	defer func() { s.Nil(res.Close()) }()

	b := bytes.NewBuffer(nil)
	s.Nil(res.Encode(b))

	expected := "0008NAK\n[PACK]"
	s.Equal(expected, b.String())
}

func (s *UploadPackResponseSuite) TestEncodeDepth() {
	pf := io.NopCloser(bytes.NewBuffer([]byte("PACK")))
	req := NewUploadPackRequest()
	req.Depth = DepthCommits(1)

	res := NewUploadPackResponseWithPackfile(req, pf)
	defer func() { s.Nil(res.Close()) }()

	b := bytes.NewBuffer(nil)
	s.Nil(res.Encode(b))

	expected := "00000008NAK\nPACK"
	s.Equal(expected, b.String())
}

func (s *UploadPackResponseSuite) TestEncodeMultiACK() {
	pf := io.NopCloser(bytes.NewBuffer([]byte("[PACK]")))
	req := NewUploadPackRequest()

	res := NewUploadPackResponseWithPackfile(req, pf)
	defer func() { s.Nil(res.Close()) }()
	res.ACKs = []plumbing.Hash{
		plumbing.NewHash("5dc01c595e6c6ec9ccda4f6f69c131c0dd945f81"),
		plumbing.NewHash("5dc01c595e6c6ec9ccda4f6f69c131c0dd945f82"),
	}

	b := bytes.NewBuffer(nil)
	s.NotNil(res.Encode(b))
}

func FuzzDecoder(f *testing.F) {
	f.Add([]byte("0045ACK 5dc01c595e6c6ec9ccda4f6f69c131c0dd945f81\n"))
	f.Add([]byte("003aACK5dc01c595e6c6ec9ccda4f6f69c131c0dd945f82 \n0008NAK\n0"))

	f.Fuzz(func(t *testing.T, input []byte) {
		req := NewUploadPackRequest()
		res := NewUploadPackResponse(req)
		defer res.Close()

		res.Decode(io.NopCloser(bytes.NewReader(input)))
	})
}
