package packp

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/stretchr/testify/suite"
)

type UlReqDecodeSuite struct {
	suite.Suite
}

func TestUlReqDecodeSuite(t *testing.T) {
	suite.Run(t, new(UlReqDecodeSuite))
}

func (s *UlReqDecodeSuite) TestEmpty() {
	ur := NewUploadRequest()
	var buf bytes.Buffer
	d := newUlReqDecoder(&buf)

	err := d.Decode(ur)
	s.ErrorContains(err, "pkt-line 1: EOF")
}

func (s *UlReqDecodeSuite) TestNoWant() {
	payloads := []string{
		"foobar",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "missing 'want '")
}

func (s *UlReqDecodeSuite) testDecoderErrorMatches(input io.Reader, pattern string) {
	ur := NewUploadRequest()
	d := newUlReqDecoder(input)

	err := d.Decode(ur)
	s.ErrorContains(err, pattern)
}

func (s *UlReqDecodeSuite) TestInvalidFirstHash() {
	payloads := []string{
		"want 6ecf0ef2c2dffb796alberto2219af86ec6584e5\n",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "invalid hash")
}

func (s *UlReqDecodeSuite) TestWantOK() {
	payloads := []string{
		"want 1111111111111111111111111111111111111111",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	s.Equal([]plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
	}, ur.Wants)
}

func (s *UlReqDecodeSuite) testDecodeOK(payloads []string) *UploadRequest {
	var buf bytes.Buffer
	e := pktline.NewEncoder(&buf)
	err := e.EncodeString(payloads...)
	s.NoError(err)

	ur := NewUploadRequest()
	d := newUlReqDecoder(&buf)

	err = d.Decode(ur)
	s.NoError(err)

	return ur
}

func (s *UlReqDecodeSuite) TestWantWithCapabilities() {
	payloads := []string{
		"want 1111111111111111111111111111111111111111 ofs-delta multi_ack",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)
	s.Equal([]plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111")}, ur.Wants)

	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))
}

func (s *UlReqDecodeSuite) TestManyWantsNoCapabilities() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333",
		"want 4444444444444444444444444444444444444444",
		"want 1111111111111111111111111111111111111111",
		"want 2222222222222222222222222222222222222222",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expected := []plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
		plumbing.NewHash("2222222222222222222222222222222222222222"),
		plumbing.NewHash("3333333333333333333333333333333333333333"),
		plumbing.NewHash("4444444444444444444444444444444444444444"),
	}

	sort.Sort(byHash(ur.Wants))
	sort.Sort(byHash(expected))
	s.Equal(expected, ur.Wants)
}

type byHash []plumbing.Hash

func (a byHash) Len() int      { return len(a) }
func (a byHash) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byHash) Less(i, j int) bool {
	ii := [hash.Size]byte(a[i])
	jj := [hash.Size]byte(a[j])
	return bytes.Compare(ii[:], jj[:]) < 0
}

func (s *UlReqDecodeSuite) TestManyWantsBadWant() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333",
		"want 4444444444444444444444444444444444444444",
		"foo",
		"want 2222222222222222222222222222222222222222",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestManyWantsInvalidHash() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333",
		"want 4444444444444444444444444444444444444444",
		"want 1234567890abcdef",
		"want 2222222222222222222222222222222222222222",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "malformed hash")
}

func (s *UlReqDecodeSuite) TestManyWantsWithCapabilities() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"want 4444444444444444444444444444444444444444",
		"want 1111111111111111111111111111111111111111",
		"want 2222222222222222222222222222222222222222",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expected := []plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
		plumbing.NewHash("2222222222222222222222222222222222222222"),
		plumbing.NewHash("3333333333333333333333333333333333333333"),
		plumbing.NewHash("4444444444444444444444444444444444444444"),
	}

	sort.Sort(byHash(ur.Wants))
	sort.Sort(byHash(expected))
	s.Equal(expected, ur.Wants)

	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))
}

func (s *UlReqDecodeSuite) TestSingleShallowSingleWant() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expectedWants := []plumbing.Hash{
		plumbing.NewHash("3333333333333333333333333333333333333333"),
	}

	expectedShallows := []plumbing.Hash{
		plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	}

	s.Equal(expectedWants, ur.Wants)
	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))

	s.Equal(expectedShallows, ur.Shallows)
}

func (s *UlReqDecodeSuite) TestSingleShallowManyWants() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"want 4444444444444444444444444444444444444444",
		"want 1111111111111111111111111111111111111111",
		"want 2222222222222222222222222222222222222222",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expectedWants := []plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
		plumbing.NewHash("2222222222222222222222222222222222222222"),
		plumbing.NewHash("3333333333333333333333333333333333333333"),
		plumbing.NewHash("4444444444444444444444444444444444444444"),
	}
	sort.Sort(byHash(expectedWants))

	expectedShallows := []plumbing.Hash{
		plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	}

	sort.Sort(byHash(ur.Wants))
	s.Equal(expectedWants, ur.Wants)
	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))

	s.Equal(expectedShallows, ur.Shallows)
}

func (s *UlReqDecodeSuite) TestManyShallowSingleWant() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"shallow cccccccccccccccccccccccccccccccccccccccc",
		"shallow dddddddddddddddddddddddddddddddddddddddd",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expectedWants := []plumbing.Hash{
		plumbing.NewHash("3333333333333333333333333333333333333333"),
	}

	expectedShallows := []plumbing.Hash{
		plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		plumbing.NewHash("cccccccccccccccccccccccccccccccccccccccc"),
		plumbing.NewHash("dddddddddddddddddddddddddddddddddddddddd"),
	}
	sort.Sort(byHash(expectedShallows))

	s.Equal(expectedWants, ur.Wants)
	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))

	sort.Sort(byHash(ur.Shallows))
	s.Equal(expectedShallows, ur.Shallows)
}

func (s *UlReqDecodeSuite) TestManyShallowManyWants() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"want 4444444444444444444444444444444444444444",
		"want 1111111111111111111111111111111111111111",
		"want 2222222222222222222222222222222222222222",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"shallow cccccccccccccccccccccccccccccccccccccccc",
		"shallow dddddddddddddddddddddddddddddddddddddddd",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expectedWants := []plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
		plumbing.NewHash("2222222222222222222222222222222222222222"),
		plumbing.NewHash("3333333333333333333333333333333333333333"),
		plumbing.NewHash("4444444444444444444444444444444444444444"),
	}
	sort.Sort(byHash(expectedWants))

	expectedShallows := []plumbing.Hash{
		plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		plumbing.NewHash("cccccccccccccccccccccccccccccccccccccccc"),
		plumbing.NewHash("dddddddddddddddddddddddddddddddddddddddd"),
	}
	sort.Sort(byHash(expectedShallows))

	sort.Sort(byHash(ur.Wants))
	s.Equal(expectedWants, ur.Wants)
	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))

	sort.Sort(byHash(ur.Shallows))
	s.Equal(expectedShallows, ur.Shallows)
}

func (s *UlReqDecodeSuite) TestMalformedShallow() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shalow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestMalformedShallowHash() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "malformed hash")
}

func (s *UlReqDecodeSuite) TestMalformedShallowManyShallows() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"shalow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"shallow cccccccccccccccccccccccccccccccccccccccc",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestMalformedDeepenSpec() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen-foo 34",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected deepen")
}

func (s *UlReqDecodeSuite) TestMalformedDeepenSingleWant() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"depth 32",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestMalformedDeepenMultiWant() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"want 2222222222222222222222222222222222222222",
		"depth 32",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestMalformedDeepenWithSingleShallow() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow 2222222222222222222222222222222222222222",
		"depth 32",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestMalformedDeepenWithMultiShallow() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"shallow 2222222222222222222222222222222222222222",
		"shallow 5555555555555555555555555555555555555555",
		"depth 32",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}

func (s *UlReqDecodeSuite) TestDeepenCommits() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen 1234",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	s.IsType(DepthCommits(0), ur.Depth)
	commits, ok := ur.Depth.(DepthCommits)
	s.True(ok)
	s.Equal(1234, int(commits))
}

func (s *UlReqDecodeSuite) TestDeepenCommitsInfiniteImplicit() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen 0",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	s.IsType(DepthCommits(0), ur.Depth)
	commits, ok := ur.Depth.(DepthCommits)
	s.True(ok)
	s.Equal(0, int(commits))
}

func (s *UlReqDecodeSuite) TestDeepenCommitsInfiniteExplicit() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	s.IsType(DepthCommits(0), ur.Depth)
	commits, ok := ur.Depth.(DepthCommits)
	s.True(ok)
	s.Equal(0, int(commits))
}

func (s *UlReqDecodeSuite) TestMalformedDeepenCommits() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen -32",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "negative depth")
}

func (s *UlReqDecodeSuite) TestDeepenCommitsEmpty() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen ",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "invalid syntax")
}

func (s *UlReqDecodeSuite) TestDeepenSince() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen-since 1420167845", // 2015-01-02T03:04:05+00:00
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expected := time.Date(2015, time.January, 2, 3, 4, 5, 0, time.UTC)

	s.IsType(DepthSince(time.Now()), ur.Depth)
	since, ok := ur.Depth.(DepthSince)
	s.True(ok)
	s.True(time.Time(since).Equal(expected),
		fmt.Sprintf("obtained=%s\nexpected=%s", time.Time(since), expected))
}

func (s *UlReqDecodeSuite) TestDeepenReference() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen-not refs/heads/master",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expected := "refs/heads/master"

	s.IsType(DepthReference(""), ur.Depth)
	reference, ok := ur.Depth.(DepthReference)
	s.True(ok)
	s.Equal(expected, string(reference))
}

func (s *UlReqDecodeSuite) TestAll() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"want 4444444444444444444444444444444444444444",
		"want 1111111111111111111111111111111111111111",
		"want 2222222222222222222222222222222222222222",
		"shallow aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"shallow bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"shallow cccccccccccccccccccccccccccccccccccccccc",
		"shallow dddddddddddddddddddddddddddddddddddddddd",
		"deepen 1234",
		pktline.FlushString,
	}
	ur := s.testDecodeOK(payloads)

	expectedWants := []plumbing.Hash{
		plumbing.NewHash("1111111111111111111111111111111111111111"),
		plumbing.NewHash("2222222222222222222222222222222222222222"),
		plumbing.NewHash("3333333333333333333333333333333333333333"),
		plumbing.NewHash("4444444444444444444444444444444444444444"),
	}
	sort.Sort(byHash(expectedWants))
	sort.Sort(byHash(ur.Wants))
	s.Equal(expectedWants, ur.Wants)
	s.True(ur.Capabilities.Supports(capability.OFSDelta))
	s.True(ur.Capabilities.Supports(capability.MultiACK))

	expectedShallows := []plumbing.Hash{
		plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		plumbing.NewHash("cccccccccccccccccccccccccccccccccccccccc"),
		plumbing.NewHash("dddddddddddddddddddddddddddddddddddddddd"),
	}
	sort.Sort(byHash(expectedShallows))
	sort.Sort(byHash(ur.Shallows))
	s.Equal(expectedShallows, ur.Shallows)

	s.IsType(DepthCommits(0), ur.Depth)
	commits, ok := ur.Depth.(DepthCommits)
	s.True(ok)
	s.Equal(1234, int(commits))
}

func (s *UlReqDecodeSuite) TestExtraData() {
	payloads := []string{
		"want 3333333333333333333333333333333333333333 ofs-delta multi_ack",
		"deepen 32",
		"foo",
		pktline.FlushString,
	}
	r := toPktLines(s.T(), payloads)
	s.testDecoderErrorMatches(r, "unexpected payload")
}
