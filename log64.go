package log64

import (
	"bytes"
	"encoding"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/danil/log64/marshal"
	jsoniter "github.com/json-iterator/go"
)

type Logger interface {
	io.Writer
	With(...KV) Logger
}

type KV interface {
	encoding.TextMarshaler
	json.Marshaler
}

const (
	Original = iota
	Excerpt
	Trail
	File
)

const (
	Trunc = iota
	Empty
	Blank
)

// Log is a JSON logger/writer.
type Log struct {
	Output  io.Writer                 // Output is a destination for output.
	Flag    int                       // Flag is a log properties.
	KV      []KV                      // Key-values.
	KVF     []func() KV               // KVF ia a dynamically calculated key-values. Existing kv will not overwritten by the dynamically calculated key-values.
	Keys    [4]encoding.TextMarshaler // Keys: 0 = original message; 1 = message excerpt; 2 = message trail; 3 = file path.
	Key     uint8                     // Key is a default/sticky message key: all except 1 = original message; 1 = message excerpt.
	Trunc   int                       // Maximum length of the message excerpt after which the message excerpt is truncated.
	Marks   [3][]byte                 // Marks: 0 = truncate; 1 = empty; 2 = blank.
	Replace [][2][]byte               // Replace ia a pairs of byte slices to replace in the message excerpt.
}

func (l Log) Write(src []byte) (int, error) {
	j, err := l.json(src)
	if err != nil {
		return 0, err
	}
	return l.Output.Write(j)
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

var (
	mapPool     = sync.Pool{New: func() interface{} { return make(map[string]json.Marshaler) }}
	excerptPool = sync.Pool{New: func() interface{} { return new([]byte) }}
)

func (l Log) json(src []byte) ([]byte, error) {
	tmpKV := mapPool.Get().(map[string]json.Marshaler)
	for k := range tmpKV {
		delete(tmpKV, k)
	}
	defer mapPool.Put(tmpKV)

	for _, kv := range l.KV {
		p, err := kv.MarshalText()
		if err != nil {
			return nil, err
		}
		tmpKV[string(p)] = kv
	}

	for _, fn := range l.KVF {
		kv := fn()
		p, err := kv.MarshalText()
		if err != nil {
			return nil, err
		}
		s := string(p)
		if _, ok := tmpKV[s]; ok {
			continue
		}
		tmpKV[s] = kv
	}

	var tail, file int

	if len(src) != 0 {
		switch l.Flag {
		case log.Lshortfile, log.Llongfile:
			i := bytes.Index(src, []byte(": "))
			if i == -1 {
				file = len(src) - 1
				tail = file + 1
			} else {
				file = i
				tail = i + 2
			}
		}
	}

	var originalKey string

	if l.Keys[Original] == nil {
		originalKey = ""
	} else {
		p, err := l.Keys[Original].MarshalText()
		if err != nil {
			return nil, err
		}
		originalKey = string(p)
	}

	var excerptKey string

	if l.Keys[Excerpt] == nil {
		excerptKey = ""
	} else {
		p, err := l.Keys[Excerpt].MarshalText()
		if err != nil {
			return nil, err
		}
		excerptKey = string(p)
	}

	excerpt := *excerptPool.Get().(*[]byte)
	excerpt = excerpt[:0]
	defer excerptPool.Put(&excerpt)

	if tmpKV[excerptKey] == nil {
		if tail == len(src) && tmpKV[originalKey] == nil {
			excerpt = append(excerpt[:0], l.Marks[Empty]...)

		} else if tail != len(src) {
			n := len(src) + len(l.Marks[Trunc])
			for _, m := range l.Marks {
				if n < len(m) {
					n = len(m)
				}
			}

			excerpt = append(excerpt[:0], make([]byte, n)...)
			n, err := l.Truncate(excerpt, src[tail:])
			if err != nil {
				return nil, err
			}

			excerpt = excerpt[:n]
		}
	}

	var trailKey string

	if l.Keys[Trail] == nil {
		trailKey = ""
	} else {
		p, err := l.Keys[Trail].MarshalText()
		if err != nil {
			return nil, err
		}
		trailKey = string(p)
	}

	if bytes.Equal(src, excerpt) && src != nil {
		if l.Key == Excerpt {
			tmpKV[excerptKey] = marshal.Bytes(src)

		} else {
			if tmpKV[originalKey] == nil {
				tmpKV[originalKey] = marshal.Bytes(src)
			} else if len(src) != 0 {
				tmpKV[trailKey] = marshal.Bytes(src)
			}
		}

	} else if !bytes.Equal(src, excerpt) {
		if tmpKV[originalKey] == nil {
			tmpKV[originalKey] = marshal.Bytes(src)
		} else if tmpKV[originalKey] != nil && len(src) != 0 {
			tmpKV[trailKey] = marshal.Bytes(src)
		}

		if tmpKV[excerptKey] == nil && len(excerpt) != 0 {
			tmpKV[excerptKey] = marshal.Bytes(excerpt)
		}
	}

	var fileKey string

	if l.Keys[File] == nil {
		fileKey = ""
	} else {
		p, err := l.Keys[File].MarshalText()
		if err != nil {
			return nil, err
		}
		fileKey = string(p)
	}

	if file != 0 {
		tmpKV[fileKey] = marshal.Bytes(src[:file])
	}

	p, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(tmpKV)
	if err != nil {
		return nil, err
	}

	return append(p, '\n'), nil
}

// lastIndexFunc is the same as bytes.LastIndexFunc except that if
// truth==false, the sense of the predicate function is
// inverted.
// lastIndexFunc copied from the bytes package.
func lastIndexFunc(s []byte, f func(r rune) bool, truth bool) int {
	for i := len(s); i > 0; {
		r, size := rune(s[i-1]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeLastRune(s[0:i])
		}
		i -= size
		if f(r) == truth {
			return i
		}
	}
	return -1
}

// Truncate writes excerpt of the src to the dst and returns number of the written bytes
// and error if occurre.
func (l Log) Truncate(dst, src []byte) (int, error) {
	var start, end int
	begin := true

	for {
		r, n := utf8.DecodeRune(src[end:])
		if n == 0 {
			break
		}

		// Rids of off all leading space, as defined by Unicode.
		if begin {
			c := src[end]

			// Fast path for ASCII: look for the first ASCII non-space byte or
			// if we run into a non-ASCII byte, fall back
			// to the slower unicode-aware method
			if c < utf8.RuneSelf && asciiSpace[c] == 1 {
				start++
				end++

				continue
			} else if unicode.IsSpace(r) {
				start += n
				end += n

				continue
			} else {
				begin = false
			}
		}

		if end-start >= len(src) || (l.Trunc > 0 && end-start >= l.Trunc) {
			break
		}

		end += n
	}

	truncate := end-start < len(src[start:])

	// Rids of off all trailing white space,
	// as defined by Unicode.
	// Look for the first ASCII non-space byte from the end.
	for ; end > start; end-- {
		c := src[end-1]
		if c >= utf8.RuneSelf {
			end = lastIndexFunc(src[:end], unicode.IsSpace, false)
			if end >= 0 && src[end] >= utf8.RuneSelf {
				_, wid := utf8.DecodeRune(src[end:])
				end += wid
			} else {
				end++
			}
			break
		}
		if asciiSpace[c] == 0 {
			break
		}
	}

	n := copy(dst, src[start:end])

replace:
	for _, r := range l.Replace {
		for offset := 0; offset < n; {
			if len(r[0]) == 0 || bytes.Equal(r[0], r[1]) {
				continue replace
			}

			idx := bytes.Index(dst[offset:n], r[0])
			if idx == -1 {
				continue replace
			}

			offset += idx

			copy(dst, append(dst[:offset], append(r[1], dst[offset+len(r[0]):]...)...))

			offset += len(r[1])
			n += len(r[1]) - len(r[0])
		}
	}

	if end-start == 0 {
		n += copy(dst[n:], l.Marks[Blank])
	}

	if end-start != 0 && truncate {
		n += copy(dst[n:], l.Marks[Trunc])
	}

	return n, nil
}

// With returns copy of the logger with additional key-values.
// Copy of the original key-values overwritten by the additional key-values.
func (l *Log) With(kv ...KV) Logger {
	l2 := *l
	l2.KV = append(kv[:0], append(l.KV, kv...)...)
	return &l2
}

// GELF returns a GELF formater <https://docs.graylog.org/en/latest/pages/gelf.html>.
func GELF() *Log {
	return &Log{
		// GELF spec version – "1.1"; Must be set by client library.
		// <https://docs.graylog.org/en/latest/pages/gelf.html#gelf-payload-specification>,
		// <https://github.com/graylog-labs/gelf-rb/issues/41#issuecomment-198266505>.
		KV: []KV{String("version", "1.1")},
		KVF: []func() KV{
			func() KV { return StringInt64("timestamp", time.Now().Unix()) },
		},
		Trunc: 120,
		Keys: [4]encoding.TextMarshaler{
			String("full_message"),
			String("short_message"),
			String("_trail"),
			String("_file"),
		},
		Key:     Excerpt,
		Marks:   [3][]byte{[]byte("…"), []byte("_EMPTY_"), []byte("_BLANK_")},
		Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
	}
}
