package packp

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/stretchr/testify/assert"
)

// returns a byte slice with the pkt-lines for the given payloads.
func pktlines(t *testing.T, payloads ...string) []byte {
	var buf bytes.Buffer
	e := pktline.NewEncoder(&buf)

	err := e.EncodeString(payloads...)
	assert.NoError(t, err, fmt.Sprintf("building pktlines for %v\n", payloads))

	return buf.Bytes()
}

func toPktLines(t *testing.T, payloads []string) io.Reader {
	var buf bytes.Buffer
	e := pktline.NewEncoder(&buf)
	err := e.EncodeString(payloads...)
	assert.NoError(t, err)

	return &buf
}
