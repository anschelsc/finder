package finder

import (
	"errors"
	"io"
)

const bufsize = 100

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
				next = spos
			}
		}
	}
}

func NewReader(f *Finder, r io.ByteReader) io.Reader {
	m := &matcher{in: r, f: f}
	m.out, m.mid = io.Pipe()
	go m.run()
	return m.out
}

type Finder struct {
	next []map[byte]int
	s    []byte
}

func Compile(s []byte) *Finder {
	inter := make([][]int, len(s)-1)
	// inter holds, for each char in s, where else you might be
	inter[0] = []int{-1, 0}
	for i := 1; i != len(s); i++ {
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
