package client

import (
	"net/http"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) TestNewClientSSH() {
	e, err := transport.NewEndpoint("ssh://github.com/src-d/go-git")
	s.NoError(err)

	output, err := NewClient(e)
	s.NoError(err)
	s.NotNil(output)
}

func (s *ClientSuite) TestNewClientUnknown() {
	e, err := transport.NewEndpoint("unknown://github.com/src-d/go-git")
	s.NoError(err)

	_, err = NewClient(e)
	s.NotNil(err)
}

func (s *ClientSuite) TestNewClientNil() {
	Protocols["newscheme"] = nil
	e, err := transport.NewEndpoint("newscheme://github.com/src-d/go-git")
	s.NoError(err)

	_, err = NewClient(e)
	s.NotNil(err)
}

func (s *ClientSuite) TestInstallProtocol() {
	InstallProtocol("newscheme", &dummyClient{})
	s.NotNil(Protocols["newscheme"])
}

func (s *ClientSuite) TestInstallProtocolNilValue() {
	InstallProtocol("newscheme", &dummyClient{})
	InstallProtocol("newscheme", nil)

	_, ok := Protocols["newscheme"]
	s.False(ok)
}

type dummyClient struct {
	*http.Client
}

func (*dummyClient) NewUploadPackSession(*transport.Endpoint, transport.AuthMethod) (
	transport.UploadPackSession, error) {
	return nil, nil
}

func (*dummyClient) NewReceivePackSession(*transport.Endpoint, transport.AuthMethod) (
	transport.ReceivePackSession, error) {
	return nil, nil
}
