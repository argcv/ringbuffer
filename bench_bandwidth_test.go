package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	argcv "github.com/argcv/ringbuffer"
	smallnest "github.com/smallnest/ringbuffer"
)

// Identical benchmark structure, only library differs

func benchmarkCopy(lib string, bufSize int, b *testing.B) {
	data := []byte(strings.Repeat("a", 1024*1024))
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src := bytes.NewReader(data)
		dst := &bytes.Buffer{}
		if lib == "smallnest" {
			rb := smallnest.New(bufSize)
			_, _ = rb.Copy(dst, src)
		} else {
			rb := argcv.New(bufSize)
			_, _ = rb.Copy(dst, src)
		}
	}
}

func BenchmarkCopySmallnest_4K(b *testing.B)  { benchmarkCopy("smallnest", 4096, b) }
func BenchmarkCopySmallnest_64K(b *testing.B) { benchmarkCopy("smallnest", 64*1024, b) }
func BenchmarkCopyArgcv_4K(b *testing.B)      { benchmarkCopy("argcv", 4096, b) }
func BenchmarkCopyArgcv_64K(b *testing.B)     { benchmarkCopy("argcv", 64*1024, b) }

// Also test io.Copy baseline with same pattern
func BenchmarkIoCopy_1M(b *testing.B) {
	data := []byte(strings.Repeat("a", 1024*1024))
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src := bytes.NewReader(data)
		dst := &bytes.Buffer{}
		_, _ = io.Copy(dst, src)
	}
}
