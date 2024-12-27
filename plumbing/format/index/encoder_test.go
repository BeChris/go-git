package index

import (
	"bytes"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/google/go-cmp/cmp"
)

func (s *IndexSuite) TestEncode() {
	idx := &Index{
		Version: 2,
		Entries: []*Entry{{
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Dev:        4242,
			Inode:      424242,
			UID:        84,
			GID:        8484,
			Size:       42,
			Stage:      TheirMode,
			Hash:       plumbing.NewHash("e25b29c8946e0e192fae2edc1dabf7be71e8ecf3"),
			Name:       "foo",
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       "bar",
			Size:       82,
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       strings.Repeat(" ", 20),
			Size:       82,
		}},
	}

	buf := bytes.NewBuffer(nil)
	e := NewEncoder(buf)
	err := e.Encode(idx)
	s.NoError(err)

	output := &Index{}
	d := NewDecoder(buf)
	err = d.Decode(output)
	s.NoError(err)

	s.True(cmp.Equal(idx, output))

	s.Equal(strings.Repeat(" ", 20), output.Entries[0].Name)
	s.Equal("bar", output.Entries[1].Name)
	s.Equal("foo", output.Entries[2].Name)

}

func (s *IndexSuite) TestEncodeV4() {
	idx := &Index{
		Version: 4,
		Entries: []*Entry{{
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Dev:        4242,
			Inode:      424242,
			UID:        84,
			GID:        8484,
			Size:       42,
			Stage:      TheirMode,
			Hash:       plumbing.NewHash("e25b29c8946e0e192fae2edc1dabf7be71e8ecf3"),
			Name:       "foo",
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       "bar",
			Size:       82,
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       strings.Repeat(" ", 20),
			Size:       82,
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       "baz/bar",
			Size:       82,
		}, {
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			Name:       "baz/bar/bar",
			Size:       82,
		}},
	}

	buf := bytes.NewBuffer(nil)
	e := NewEncoder(buf)
	err := e.Encode(idx)
	s.NoError(err)

	output := &Index{}
	d := NewDecoder(buf)
	err = d.Decode(output)
	s.NoError(err)

	s.True(cmp.Equal(idx, output))

	s.Equal(strings.Repeat(" ", 20), output.Entries[0].Name)
	s.Equal("bar", output.Entries[1].Name)
	s.Equal("baz/bar", output.Entries[2].Name)
	s.Equal("baz/bar/bar", output.Entries[3].Name)
	s.Equal("foo", output.Entries[4].Name)
}

func (s *IndexSuite) TestEncodeUnsupportedVersion() {
	idx := &Index{Version: 5}

	buf := bytes.NewBuffer(nil)
	e := NewEncoder(buf)
	err := e.Encode(idx)
	s.Equal(err, ErrUnsupportedVersion)
}

func (s *IndexSuite) TestEncodeWithIntentToAddUnsupportedVersion() {
	idx := &Index{
		Version: 3,
		Entries: []*Entry{{IntentToAdd: true}},
	}

	buf := bytes.NewBuffer(nil)
	e := NewEncoder(buf)
	err := e.Encode(idx)
	s.NoError(err)

	output := &Index{}
	d := NewDecoder(buf)
	err = d.Decode(output)
	s.NoError(err)

	s.True(cmp.Equal(idx, output))
	s.True(output.Entries[0].IntentToAdd)
}

func (s *IndexSuite) TestEncodeWithSkipWorktreeUnsupportedVersion() {
	idx := &Index{
		Version: 3,
		Entries: []*Entry{{SkipWorktree: true}},
	}

	buf := bytes.NewBuffer(nil)
	e := NewEncoder(buf)
	err := e.Encode(idx)
	s.NoError(err)

	output := &Index{}
	d := NewDecoder(buf)
	err = d.Decode(output)
	s.NoError(err)

	s.True(cmp.Equal(idx, output))
	s.True(output.Entries[0].SkipWorktree)
}
