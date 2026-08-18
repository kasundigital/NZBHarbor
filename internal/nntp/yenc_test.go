package nntp

import "testing"

func TestDecodeYEnc(t *testing.T) {
	// "ABC" encoded by adding 42 to every byte.
	article := []byte("=ybegin line=128 size=3 name=x.bin\nklm\n=yend size=3\n")
	got, err := DecodeYEnc(article)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ABC" {
		t.Fatalf("got %q", got)
	}
}
