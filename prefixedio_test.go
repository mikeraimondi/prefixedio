package prefixedio

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
)

func TestPrefixedWrite(t *testing.T) {
	t.Parallel()

	in := []byte("foo")
	buf := &bytes.Buffer{}
	n, err := PrefixedWrite(buf, in)
	if err != nil {
		t.Fatalf("Error returned: %v\n", err)
	}
	if n != len(in) {
		t.Fatalf("Wrong number of bytes returned. Expected: %v. Actual: %v\n", n, len(in))
	}

	sizeBuf := make([]byte, 8)
	if _, err = buf.Read(sizeBuf); err != nil {
		t.Fatalf("Error reading size: %v\n", err)
	}
	size := binary.BigEndian.Uint64(sizeBuf)
	if int(size) != len(in) {
		t.Fatalf("Mismatched size. Expected: %v. Actual: %v\n", n, len(in))
	}
	out := make([]byte, size)
	if _, err := buf.Read(out); err != nil {
		t.Fatalf("Error reading message: %v\n", err)
	}
	for i, c := range out {
		if in[i] != c {
			t.Fatalf("Written value doesn't match input. Expected: %v. Actual: %v\n", in[i], c)
		}
	}
}

func BenchmarkPrefixedWrite(b *testing.B) {
	testBytes := make([]byte, 1000)
	for i := range testBytes {
		testBytes[i] = byte(rand.Int())
	}
	testBuf := &bytes.Buffer{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PrefixedWrite(testBuf, testBytes)
		testBuf.Reset()
	}
}

func TestReadFromWithValidInput(t *testing.T) {
	t.Parallel()

	var pb PrefixedBuf // Reuse the same buffer between runs
	buf := &bytes.Buffer{}
	var maxLenBuf bytes.Buffer
	for i := 0; i < maxLen; i++ {
		maxLenBuf.WriteString("Q")
	}
	ins := []string{
		"foo",              // 1 < length < 8
		"foobarba",         // length == 8
		"foobarbaz",        // length > 8
		maxLenBuf.String(), // length == maxLen
		"f",                // length == 1, test last to verify reusing buffer works when content shrinks
	}
	for _, in := range ins {
		binary.Write(buf, binary.BigEndian, uint64(len(in)))
		buf.WriteString(in)

		n, err := pb.ReadFrom(buf)
		if err != nil {
			t.Fatal("Error on read: ", err)
		}
		if n != int64(len(in)) {
			t.Fatalf("Number of read bytes doesn't match input. Expected: %v. Actual: %v\n", len(in), n)
		}
		byteStr := string(pb.Bytes())
		if len(byteStr) <= 0 {
			t.Fatalf("Read bytes are empty. Expected: non-empty string. Actual: %v\n", byteStr)
		}
		if byteStr != in {
			t.Fatalf("Read bytes doesn't match input. Expected: %v. Actual: %v\n", in, byteStr)
		}
	}
}

func TestReadFromWithInvalidInput(t *testing.T) {
	t.Parallel()

	length := maxLen + 1
	buf := bytes.NewBuffer(make([]byte, 0, length+8))
	if err := binary.Write(buf, binary.BigEndian, uint64(length)); err != nil {
		t.Fatal("Error on write: ", err)
	}
	for i := 0; i < length; i++ {
		buf.WriteString("a")
	}
	var pb PrefixedBuf
	if _, err := pb.ReadFrom(buf); err == nil {
		t.Fatal("No error raised when message too long")
	}
}
