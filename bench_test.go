package main

import (
	"context"
	"io"
	"strings"
	"testing"

	argcv "github.com/argcv/ringbuffer"
	smallnest "github.com/smallnest/ringbuffer"
)

// === Sync Read/Write ===

func BenchmarkSmallnest_Sync(b *testing.B) {
	rb := smallnest.New(1024)
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}

func BenchmarkArgcv_Sync(b *testing.B) {
	rb := argcv.New(1024)
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}

// === Async Read Blocking ===

func BenchmarkSmallnest_AsyncReadBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := smallnest.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			_, _ = rb.Read(buf)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
	}
}

func BenchmarkArgcv_AsyncReadBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := argcv.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			_, _ = rb.Read(buf)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
	}
}

// === Async Write Blocking ===

func BenchmarkSmallnest_AsyncWriteBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := smallnest.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			_, _ = rb.Write(data)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Read(buf)
	}
}

func BenchmarkArgcv_AsyncWriteBlocking(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := argcv.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		for {
			_, _ = rb.Write(data)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Read(buf)
	}
}

// === ReadFrom ===

type repeatReader struct {
	b      []byte
	doCopy bool
}

func (r repeatReader) Read(b []byte) (n int, err error) {
	n = len(b)
	for r.doCopy && len(b) > 0 {
		n2 := copy(b, r.b)
		b = b[n2:]
	}
	return n, nil
}

func BenchmarkSmallnest_ReadFrom(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := smallnest.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		_, _ = rb.ReadFrom(repeatReader{b: data})
	}()

	b.ResetTimer()
	b.SetBytes(sz)
	for i := 0; i < b.N; i++ {
		_, _ = io.ReadFull(rb, buf)
	}
	rb.CloseWithError(context.Canceled)
}

func BenchmarkArgcv_ReadFrom(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := argcv.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))
	buf := make([]byte, sz)

	go func() {
		_, _ = rb.ReadFrom(repeatReader{b: data})
	}()

	b.ResetTimer()
	b.SetBytes(sz)
	for i := 0; i < b.N; i++ {
		_, _ = io.ReadFull(rb, buf)
	}
	rb.CloseWithError(context.Canceled)
}

// === WriteTo ===

func BenchmarkSmallnest_WriteTo(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := smallnest.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))

	go func() {
		_, _ = rb.WriteTo(io.Discard)
	}()

	b.ResetTimer()
	b.SetBytes(sz)
	for i := 0; i < b.N; i++ {
		_, err := rb.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	rb.CloseWithError(context.Canceled)
}

func BenchmarkArgcv_WriteTo(b *testing.B) {
	const sz = 512
	const buffers = 10
	rb := argcv.New(sz * buffers)
	rb.SetBlocking(true)
	data := []byte(strings.Repeat("a", sz))

	go func() {
		_, _ = rb.WriteTo(io.Discard)
	}()

	b.ResetTimer()
	b.SetBytes(sz)
	for i := 0; i < b.N; i++ {
		_, err := rb.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	rb.CloseWithError(context.Canceled)
}

// === io.Pipe baseline ===

func BenchmarkIoPipe(b *testing.B) {
	pr, pw := io.Pipe()
	data := []byte(strings.Repeat("a", 512))
	buf := make([]byte, 512)

	go func() {
		for {
			_, _ = pw.Write(data)
		}
	}()

	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, _ = pr.Read(buf)
	}
}
