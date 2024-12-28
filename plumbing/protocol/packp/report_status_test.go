package packp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/stretchr/testify/suite"
)

type ReportStatusSuite struct {
	suite.Suite
}

func TestReportStatusSuite(t *testing.T) {
	suite.Run(t, new(ReportStatusSuite))
}

func (s *ReportStatusSuite) TestError() {
	rs := NewReportStatus()
	rs.UnpackStatus = "ok"
	s.Nil(rs.Error())
	rs.UnpackStatus = "OK"
	s.ErrorContains(rs.Error(), "unpack error: OK")
	rs.UnpackStatus = ""
	s.ErrorContains(rs.Error(), "unpack error: ")

	cs := &CommandStatus{ReferenceName: plumbing.ReferenceName("ref")}
	rs.UnpackStatus = "ok"
	rs.CommandStatuses = append(rs.CommandStatuses, cs)

	cs.Status = "ok"
	s.Nil(rs.Error())
	cs.Status = "OK"
	s.ErrorContains(rs.Error(), "command error on ref: OK")
	cs.Status = ""
	s.ErrorContains(rs.Error(), "command error on ref: ")
}

func (s *ReportStatusSuite) testEncodeDecodeOk(rs *ReportStatus, lines ...string) {
	s.testDecodeOk(rs, lines...)
	s.testEncodeOk(rs, lines...)
}

func (s *ReportStatusSuite) testDecodeOk(expected *ReportStatus, lines ...string) {
	r := toPktLines(s.T(), lines)
	rs := NewReportStatus()
	s.Nil(rs.Decode(r))
	s.Equal(expected, rs)
}

func (s *ReportStatusSuite) testDecodeError(errorMatch string, lines ...string) {
	r := toPktLines(s.T(), lines)
	rs := NewReportStatus()
	s.ErrorContains(rs.Decode(r), errorMatch)
}

func (s *ReportStatusSuite) testEncodeOk(input *ReportStatus, lines ...string) {
	expected := pktlines(s.T(), lines...)
	var buf bytes.Buffer
	s.Nil(input.Encode(&buf))
	obtained := buf.Bytes()

	comment := fmt.Sprintf("\nobtained = %s\nexpected = %s\n", string(obtained), string(expected))

	s.Equal(expected, obtained, comment)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkOneReference() {
	rs := NewReportStatus()
	rs.UnpackStatus = "ok"
	rs.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testEncodeDecodeOk(rs,
		"unpack ok\n",
		"ok refs/heads/master\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkOneReferenceFailed() {
	rs := NewReportStatus()
	rs.UnpackStatus = "my error"
	rs.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "command error",
	}}

	s.testEncodeDecodeOk(rs,
		"unpack my error\n",
		"ng refs/heads/master command error\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkMoreReferences() {
	rs := NewReportStatus()
	rs.UnpackStatus = "ok"
	rs.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}, {
		ReferenceName: plumbing.ReferenceName("refs/heads/a"),
		Status:        "ok",
	}, {
		ReferenceName: plumbing.ReferenceName("refs/heads/b"),
		Status:        "ok",
	}}

	s.testEncodeDecodeOk(rs,
		"unpack ok\n",
		"ok refs/heads/master\n",
		"ok refs/heads/a\n",
		"ok refs/heads/b\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkMoreReferencesFailed() {
	rs := NewReportStatus()
	rs.UnpackStatus = "my error"
	rs.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}, {
		ReferenceName: plumbing.ReferenceName("refs/heads/a"),
		Status:        "command error",
	}, {
		ReferenceName: plumbing.ReferenceName("refs/heads/b"),
		Status:        "ok",
	}}

	s.testEncodeDecodeOk(rs,
		"unpack my error\n",
		"ok refs/heads/master\n",
		"ng refs/heads/a command error\n",
		"ok refs/heads/b\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkNoReferences() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"

	s.testEncodeDecodeOk(expected,
		"unpack ok\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestEncodeDecodeOkNoReferencesFailed() {
	rs := NewReportStatus()
	rs.UnpackStatus = "my error"

	s.testEncodeDecodeOk(rs,
		"unpack my error\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestDecodeErrorOneReferenceNoFlush() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("missing flush",
		"unpack ok\n",
		"ok refs/heads/master\n",
	)
}

func (s *ReportStatusSuite) TestDecodeErrorEmpty() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("unexpected EOF")
}

func (s *ReportStatusSuite) TestDecodeErrorMalformed() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("malformed unpack status: unpackok",
		"unpackok\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestDecodeErrorMalformed2() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("malformed unpack status: UNPACK OK",
		"UNPACK OK\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestDecodeErrorMalformedCommandStatus() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("malformed command status: ko refs/heads/master",
		"unpack ok\n",
		"ko refs/heads/master\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestDecodeErrorMalformedCommandStatus2() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("malformed command status: ng refs/heads/master",
		"unpack ok\n",
		"ng refs/heads/master\n",
		pktline.FlushString,
	)
}

func (s *ReportStatusSuite) TestDecodeErrorPrematureFlush() {
	expected := NewReportStatus()
	expected.UnpackStatus = "ok"
	expected.CommandStatuses = []*CommandStatus{{
		ReferenceName: plumbing.ReferenceName("refs/heads/master"),
		Status:        "ok",
	}}

	s.testDecodeError("premature flush",
		pktline.FlushString,
	)
}
