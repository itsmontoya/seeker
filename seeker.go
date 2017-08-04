package seeker

import (
	"bytes"
	"io"
	"os"

	"github.com/itsmontoya/async/file"
	"github.com/missionMeteora/toolkit/errors"
)

const (
	// ErrLineNotFound is returned when a line is not found while calling SeekNextLine
	ErrLineNotFound = errors.Error("line not found")
	// ErrInvalidLineNumber is returned when an invalid line number is provided
	ErrInvalidLineNumber = errors.Error("invalid line number provided")
	// ErrEndEarly is returned when line reading has ended early
	ErrEndEarly = errors.Error("ended early")
)

const (
	// Buffer size used for file seeking
	seekerBufSize = 32
)

const (
	charNewline = '\n'
)

// New will return a pointer to a new instance of Seeker
func New(f file.Interface) *Seeker {
	var s Seeker
	s.f = f
	s.lbuf = bytes.NewBuffer(nil)
	return &s
}

// Seeker is a file seeker
type Seeker struct {
	// File
	f file.Interface
	// Seek buffer, used for storing read data while seeking
	sbuf [seekerBufSize]byte
	// Line buffer, used for storing lines
	lbuf *bytes.Buffer
}

func (s *Seeker) getPosition() (pos int64) {
	pos, _ = s.f.Seek(0, os.SEEK_CUR)
	return
}

func (s *Seeker) seekBackwards(cc int64) (nc int64, err error) {
	if cc > seekerBufSize {
		cc = seekerBufSize
	}

	return s.f.Seek(-cc, os.SEEK_CUR)
}

func (s *Seeker) readChunks(fn func(int) bool) (err error) {
	var n int

	for n, err = s.f.Read(s.sbuf[:]); ; n, err = s.f.Read(s.sbuf[:]) {
		if err == io.EOF && n == 0 {
			err = nil
			break
		} else if err != nil {
			break
		}

		if fn(n) {
			break
		}
	}

	return
}

func (s *Seeker) readReverseChunks(fn func(int) bool) (err error) {
	var (
		curr = s.getPosition()

		n    int
		cc   int64
		done bool
	)

	for !done {
		if curr > seekerBufSize {
			cc = seekerBufSize
		} else {
			cc = curr
			done = true
		}

		if curr, err = s.seekBackwards(curr); err != nil {
			return
		}

		if n, err = s.f.Read(s.sbuf[:cc]); err == io.EOF && n == 0 {
			err = nil
			break
		} else if err != nil {
			break
		}

		if fn(n) {
			break
		}

		if done {
			break
		}

		if curr, err = s.seekBackwards(curr); err != nil {
			break
		}
	}

	return
}

// SetFile will set a file
func (s *Seeker) SetFile(f file.Interface) {
	s.f = f
}

// SeekToStart will seek the file to the start
func (s *Seeker) SeekToStart() (err error) {
	_, err = s.f.Seek(0, os.SEEK_SET)
	return
}

// SeekToEnd will seek the file to the end
func (s *Seeker) SeekToEnd() (err error) {
	_, err = s.f.Seek(0, os.SEEK_END)
	return
}

// SeekToLine will seek to line
func (s *Seeker) SeekToLine(n int) (err error) {
	if n < 0 {
		err = ErrInvalidLineNumber
		return
	}

	curr := -1

	if _, err = s.f.Seek(0, os.SEEK_SET); err != nil {
		goto END
	}

READ:
	if err = s.ReadLine(func(b *bytes.Buffer) error {
		curr++
		return nil
	}); err != nil {
		goto END
	}

	if curr < n {
		goto READ
	}

END:
	if curr != n {
		err = ErrLineNotFound
	} else {
		err = s.PrevLine()
	}

	return
}

// NextLine will position the file at the next line
func (s *Seeker) NextLine() (err error) {
	var (
		nlf    bool
		offset int64 = -1
	)

	pcfn := func(n int) (end bool) {
		for i, b := range s.sbuf[:n] {
			if b == charNewline {
				nlf = true
			} else if nlf {
				offset = int64(n - i)
				return true
			}
		}

		return
	}

	if err = s.readChunks(pcfn); err != nil {
		return
	}

	if offset == -1 {
		return ErrLineNotFound
	}

	_, err = s.f.Seek(-offset, os.SEEK_CUR)
	return
}

// PrevLine will position the file at the previous line
func (s *Seeker) PrevLine() (err error) {
	var (
		nlc    int
		rel    int64
		offset int64 = -1
	)

	pcfn := func(n int) (end bool) {
		s := s.sbuf[:n]
		reverseByteSlice(s)

		for i, b := range s {
			if b != charNewline {
				continue
			}

			if nlc++; nlc == 2 {
				offset = int64(i)
				return true
			}
		}

		return
	}

	// Get current index
	if rel, err = s.f.Seek(0, os.SEEK_CUR); err != nil {
		return
	}

	if rel == 0 {
		return io.EOF
	}

	if err = s.readReverseChunks(pcfn); err != nil {
		return
	}

	if offset == -1 {
		_, err = s.f.Seek(0, os.SEEK_SET)
	} else {
		_, err = s.f.Seek(-offset, os.SEEK_CUR)
	}

	return
}

// ReadLine will return a line in the form of an a bytes.Buffer
func (s *Seeker) ReadLine(fn func(*bytes.Buffer) error) (err error) {
	var (
		n   int
		b   []byte
		idx int
	)

	for err == nil {
		if n, err = s.f.Read(s.sbuf[:]); err != nil && err != io.EOF {
			break
		}

		b = s.sbuf[:n]
		if idx = getNewlineIndex(b); idx == -1 {
			s.lbuf.Write(b)
			continue
		}

		s.lbuf.Write(b[:idx])
		s.f.Seek(int64(-(n - idx - 1)), os.SEEK_CUR)
		err = fn(s.lbuf)
		break
	}

	s.lbuf.Reset()
	return
}

// ReadLines will return all lines in the form of an a bytes.Buffer
// Provided function can return true to end iteration early
func (s *Seeker) ReadLines(fn func(*bytes.Buffer) error) (err error) {
	for err == nil {
		err = s.ReadLine(fn)
	}

	if err == io.EOF || err == ErrEndEarly {
		err = nil
	}

	return
}
