package main

import (
	"encoding/binary"
	"testing"

	argcv "github.com/argcv/ringbuffer"
)

// === Generic Ring[int] vs byte-based workaround ===
// Simulating an int queue using bytes (what you'd have to do without generics)

func BenchmarkIntQueue_Generic(b *testing.B) {
	rb := argcv.NewRing[int](1024)
	data := []int{1, 2, 3, 4, 5, 6, 7, 8}
	buf := make([]int, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}

func BenchmarkIntQueue_ByteWorkaround(b *testing.B) {
	rb := argcv.New(1024 * 8) // 8 bytes per int64
	data := make([]byte, 64)  // 8 ints * 8 bytes
	for i := range 8 {
		binary.BigEndian.PutUint64(data[i*8:], uint64(i+1))
	}
	buf := make([]byte, 64)
	out := make([]int, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
		for j := range 8 {
			out[j] = int(binary.BigEndian.Uint64(buf[j*8:]))
		}
	}
}

// === Generic Ring[struct] vs byte serialization ===

type LogEntry struct {
	ID     uint64
	Level  uint32
	MsgLen uint32
}

func BenchmarkStructQueue_Generic(b *testing.B) {
	rb := argcv.NewRing[LogEntry](1024)
	data := []LogEntry{{ID: 1, Level: 2, MsgLen: 100}}
	buf := make([]LogEntry, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}

func BenchmarkStructQueue_ByteWorkaround(b *testing.B) {
	rb := argcv.New(1024 * 16) // 16 bytes per LogEntry
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[0:], 1)
	binary.BigEndian.PutUint32(data[8:], 2)
	binary.BigEndian.PutUint32(data[12:], 100)
	buf := make([]byte, 16)
	var out LogEntry

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
		out.ID = binary.BigEndian.Uint64(buf[0:])
		out.Level = binary.BigEndian.Uint32(buf[8:])
		out.MsgLen = binary.BigEndian.Uint32(buf[12:])
	}
}

// === Iterator: iter.Seq[byte] vs Bytes() copy ===

func BenchmarkIterator_All(b *testing.B) {
	rb := argcv.New(1024)
	_, _ = rb.Write([]byte("hello world this is a test message"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range rb.All() {
		}
	}
}

func BenchmarkIterator_BytesCopy(b *testing.B) {
	rb := argcv.New(1024)
	_, _ = rb.Write([]byte("hello world this is a test message"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := rb.Bytes(nil)
		for _, _ = range data {
		}
	}
}

// === Ring[byte] vs RingBuffer: confirm no overhead ===

func BenchmarkRingByte_Sync(b *testing.B) {
	rb := argcv.NewRing[byte](1024)
	data := []byte("abcdefgh")
	buf := make([]byte, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}

func BenchmarkRingBuffer_Sync(b *testing.B) {
	rb := argcv.New(1024)
	data := []byte("abcdefgh")
	buf := make([]byte, 8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rb.Write(data)
		_, _ = rb.Read(buf)
	}
}
