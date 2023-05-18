// DSUL - Disturb State USB Light : IPC module tests.
package ipc

import (
	"bytes"
	"testing"
)

func TestEncodeToBytes(t *testing.T) {
	in_value := Message{"", "", "", ""}
	out_value := encodeToBytes(in_value)
	want := []byte{59, 255, 129, 3, 1, 1, 7, 77, 101, 115, 115, 97, 103, 101, 1, 255, 130, 0, 1, 4, 1, 4, 84, 121, 112, 101, 1, 12, 0, 1, 3, 75, 101, 121, 1, 12, 0, 1, 5, 86, 97, 108, 117, 101, 1, 12, 0, 1, 6, 83, 101, 99, 114, 101, 116, 1, 12, 0, 0, 0, 3, 255, 130, 0}

	if !bytes.Equal(out_value, want) {
		t.Errorf("Wrong(%q) == %q, want %q", in_value, out_value, want)
	}
}

func TestDecodeToMessage(t *testing.T) {
	in_value := []byte{59, 255, 129, 3, 1, 1, 7, 77, 101, 115, 115, 97, 103, 101, 1, 255, 130, 0, 1, 4, 1, 4, 84, 121, 112, 101, 1, 12, 0, 1, 3, 75, 101, 121, 1, 12, 0, 1, 5, 86, 97, 108, 117, 101, 1, 12, 0, 1, 6, 83, 101, 99, 114, 101, 116, 1, 12, 0, 0, 0, 3, 255, 130, 0}
	out_value := decodeToMessage(in_value)
	want := Message{"", "", "", ""}

	if out_value != want {
		t.Errorf("Wrong(%q) == %q, want %q", in_value, out_value, want)
	}
}

func TestDecodeToString(t *testing.T) {
	in_value := []byte{10, 12, 0, 7, 101, 120, 97, 109, 112, 108, 101}
	out_value := decodeToString(in_value)
	want := "example"

	if out_value != want {
		t.Errorf("Wrong(%q) == %q, want %q", in_value, out_value, want)
	}
}
