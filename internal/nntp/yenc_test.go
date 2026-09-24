package nntp

import "testing"

func TestDecodeYEnc(t *testing.T) {
	// "ABC" encoded by adding 42 to every byte.
	article := []byte("=ybegin line=128 size=3 name=x.bin\nklm\n=yend size=3 pcrc32=a3830348\n")
	got, err := DecodeYEnc(article)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ABC" {
		t.Fatalf("got %q", got)
	}
}

func TestDecodeYEncRejectsBadCRC(t *testing.T) {
	article := []byte("=ybegin line=128 size=3 name=x.bin\nklm\n=yend size=3 pcrc32=00000000\n")
	if _, err := DecodeYEnc(article); err == nil {
		t.Fatal("expected CRC mismatch")
	}
}

func TestDecodeYEncRejectsBadSize(t *testing.T) {
	article := []byte("=ybegin line=128 size=3 name=x.bin\nklm\n=yend size=4\n")
	if _, err := DecodeYEnc(article); err == nil {
		t.Fatal("expected size mismatch")
	}
}
