package finder

import (
	"errors"
	"io"
)

var Found = errors.New("Found it!")

type matcher struct {
	in  io.ByteReader
	mid *io.PipeWriter
	out *io.PipeReader
	f   *Finder
}

func (m *matcher) wb(b byte) { m.mid.Write([]byte{b}) }

func (m *matcher) run() {
	for {
		b, err := m.in.ReadByte()
		if err != nil {
			m.mid.CloseWithError(err)
			return
		}
		if b != m.f.s[0] {
			m.wb(b)
		} else {
			spos := 0
			for {
				b, err = m.in.ReadByte()
				if err != nil {
					m.mid.CloseWithError(err)
					return
				}
				next, ok := m.f.next[spos][b]
				if !ok {
					m.mid.Write(m.f.s[:spos+1])
					m.wb(b)
					break
				}
				if next == len(m.f.s)-1 {
					m.mid.CloseWithError(Found)
					return
				}
				m.mid.Write(m.f.s[:spos+1-next]) // nop if spos + 1 == next
				spos = next
			}
		}
	}
}

// NewReader returns an io.Reader that will read from r until it either
// encounters an error (which will be passed on) or finds the sequence known by
// f, in which case it returns Found.
func NewReader(f *Finder, r io.ByteReader) io.Reader {
	m := &matcher{in: r, f: f}
	m.out, m.mid = io.Pipe()
	go m.run()
	return m.out
}

// NewReaderBytes is like NewReader but Compiles b for you. If you will likely
// search for the same []byte again, use Compile directly.
func NewReaderBytes(b []byte, r io.ByteReader) io.Reader {
	return NewReader(Compile(b), r)
}

// A Finder contains the information necessary to find a []byte in linear time.
type Finder struct {
	next []map[byte]int
	s    []byte
}

// Compile returns a persistent *Finder that can be passed to NewReader
// multiple times. Use this if you will look for the same string more than once.
func Compile(s []byte) *Finder {
	inter := make([][]int, len(s)-1)
	// inter holds, for each char in s, where else you might be
	inter[0] = []int{-1, 0}
	for i := 1; i != len(inter); i++ {
		inter[i] = []int{-1}
		for _, pos := range inter[i-1] {
			if s[pos+1] == s[i] {
				inter[i] = append(inter[i], pos+1)
			}
		}
	}

	next := make([]map[byte]int, len(s)-1)
	for i, poss := range inter {
		next[i] = make(map[byte]int)
		for _, pos := range poss {
			next[i][s[pos+1]] = pos + 1
		}
	}
	return &Finder{next, s}
}
