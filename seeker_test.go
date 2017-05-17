package seeker

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestSeeker(t *testing.T) {
	var (
		f   *os.File
		err error
	)
	if f, err = ioutil.TempFile("", "seeker"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		name := f.Name()
		f.Close()
		os.Remove(name)
	}()

	f.WriteString("Line 0\n")
	f.WriteString("Line 1\n")
	f.WriteString("Line 2\n")

	s := NewSeeker(f)

	if err = s.SeekToStart(); err != nil {
		t.Fatal(err)
	}

	var str string
	if err = s.ReadLine(func(buf *bytes.Buffer) {
		str = buf.String()
	}); err != nil {
		t.Fatal(err)
	}

	if str != "Line 0" {
		t.Fatalf("invalid string, expected %s and received %s", "Line 0", str)
	}

	if err = s.ReadLine(func(buf *bytes.Buffer) {
		str = buf.String()
	}); err != nil {
		t.Fatal(err)
	}

	if str != "Line 1" {
		t.Fatalf("invalid string, expected %s and received %s", "Line 1", str)
	}

	if err = s.ReadLine(func(buf *bytes.Buffer) {
		str = buf.String()
	}); err != nil {
		t.Fatal(err)
	}

	if str != "Line 2" {
		t.Fatalf("invalid string, expected %s and received %s", "Line 2", str)
	}

	if err = s.ReadLine(func(buf *bytes.Buffer) {
		str = buf.String()
	}); err != io.EOF {
		if err == nil {
			t.Fatal("no error encountered when io.EOF was expected")
		} else {
			t.Fatal(err)
		}
	}

	if err = s.PrevLine(); err != nil {
		t.Fatal(err)
	}

	if err = s.PrevLine(); err != nil {
		t.Fatal(err)
	}

	if err = s.ReadLine(func(buf *bytes.Buffer) {
		str = buf.String()
	}); err != nil {
		t.Fatal(err)
	}

	if str != "Line 1" {
		t.Fatalf("invalid string, expected %s and received %s", "Line 1", str)
	}
}
