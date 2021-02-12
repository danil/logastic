package log64_test

import (
	"bytes"
	"encoding"
	"fmt"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/danil/log64"
	"github.com/kinbiko/jsonassert"
)

var WriteTestCases = []struct {
	name      string
	line      int
	log       log64.Logger
	input     []byte
	kv        []log64.KV
	expected  string
	benchmark bool
}{
	{
		name: "nil",
		line: line(),
		log:  dummy,
		expected: `{
	    "message":null,
			"excerpt":"_EMPTY_"
		}`,
	},
	{
		name: `"string" key with "foo" value and "string" key with "bar" value`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			KV:    []log64.KV{log64.String("string", "foo")},
			Keys:  [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: []byte("Hello, World!"),
		kv:    []log64.KV{log64.String("string", "bar")},
		expected: `{
			"message":"Hello, World!",
		  "string": "bar"
		}`,
		benchmark: true,
	},
	{
		name:  "kv is nil",
		line:  line(),
		log:   dummy,
		input: []byte("Hello, World!"),
		kv:    nil,
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name: `bytes appends to the "message" key with "string value"`,
		line: line(),
		log: &log64.Log{
			KV:      []log64.KV{log64.String("message", "string value")},
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("Hello,\nWorld!"),
		expected: `{
			"message":"string value",
			"excerpt":"Hello, World!",
			"trail":"Hello,\nWorld!"
		}`,
	},
	{
		name:  `bytes appends to the "message" key with "string value"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello,\nWorld!"),
		kv:    []log64.KV{log64.String("message", "string value")},
		expected: `{
			"message":"string value",
			"excerpt":"Hello, World!",
			"trail":"Hello,\nWorld!"
		}`,
	},
	{
		name: `bytes is nil and "message" key with "string value"`,
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.String("message", "string value")},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message")},
		},
		expected: `{
			"message":"string value"
		}`,
	},
	{
		name: `input is nil and "message" key with "string value"`,
		line: line(),
		log:  dummy,
		kv:   []log64.KV{log64.String("message", "string value")},
		expected: `{
			"message":"string value"
		}`,
	},
	{
		name:  `bytes appends to the integer key "message"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello, World!\n"),
		kv:    []log64.KV{log64.StringInt("message", 1)},
		expected: `{
			"message":1,
			"excerpt":"Hello, World!",
			"trail":"Hello, World!\n"
		}`,
	},
	{
		name:  `bytes appends to the float 32 bit key "message"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello,\nWorld!"),
		kv:    []log64.KV{log64.StringFloat32("message", 4.2)},
		expected: `{
			"message":4.2,
			"excerpt":"Hello, World!",
			"trail":"Hello,\nWorld!"
		}`,
	},
	{
		name:  `bytes appends to the float 64 bit key "message"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello,\nWorld!"),
		kv:    []log64.KV{log64.StringFloat64("message", 4.2)},
		expected: `{
			"message":4.2,
			"excerpt":"Hello, World!",
			"trail":"Hello,\nWorld!"
		}`,
	},
	{
		name:  `bytes appends to the boolean key "message"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello,\nWorld!"),
		kv:    []log64.KV{log64.StringBool("message", true)},
		expected: `{
			"message":true,
			"excerpt":"Hello, World!",
			"trail":"Hello,\nWorld!"
		}`,
	},
	{
		name:  `bytes will appends to the nil key "message"`,
		line:  line(),
		log:   dummy,
		input: []byte("Hello, World!"),
		kv:    []log64.KV{log64.StringReflect("message", nil)},
		expected: `{
			"message":null,
			"trail":"Hello, World!"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "message" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message")},
			Key:   log64.Original,
		},
		kv: []log64.KV{log64.String("message", "foo")},
		expected: `{
			"message":"foo"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "message" key is present and with replace`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("message", "foo\n")},
		expected: `{
			"message":"foo\n"
		}`,
	},
	{
		name: `default key is original and bytes is present and "message" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("message", "bar")},
		expected: `{
			"message":"bar",
			"trail":"foo"
		}`,
	},
	{
		name: `default key is original and bytes is present and "message" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:   log64.Original,
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar")},
		expected: `{
			"message":"bar",
			"excerpt":"foo",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is original and bytes is present and "message" key is present and with replace input bytes and key`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar\n")},
		expected: `{
			"message":"bar\n",
			"excerpt":"foo",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "excerpt" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Original,
		},
		kv: []log64.KV{log64.String("excerpt", "foo")},
		expected: `{
			"excerpt":"foo"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "excerpt" key is present and with replace`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("excerpt", "foo\n")},
		expected: `{
			"excerpt":"foo\n"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Original,
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" key is present and with replace input bytes and rey`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar\n")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar\n"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "excerpt" and "message" keys is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Original,
		},
		kv: []log64.KV{log64.String("message", "foo"), log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is original and bytes is nil and "excerpt" and "message" keys is present and replace keys`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("message", "foo\n"), log64.String("excerpt", "bar\n")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar\n"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" and "message" keys is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:   log64.Original,
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("message", "bar"), log64.String("excerpt", "xyz")},
		expected: `{
			"message":"bar",
			"excerpt":"xyz",
			"trail":"foo"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" and "message" keys is present and replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv: []log64.KV{
			log64.String("message", "bar"),
			log64.String("excerpt", "xyz"),
		},
		expected: `{
			"message":"bar",
			"excerpt":"xyz",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is original and bytes is present and "excerpt" and "message" keys is present and replace input bytes and keys`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Original,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar\n"), log64.String("excerpt", "xyz\n")},
		expected: `{
			"message":"bar\n",
			"excerpt":"xyz\n",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "message" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message")},
			Key:   log64.Excerpt,
		},
		kv: []log64.KV{log64.String("message", "foo")},
		expected: `{
			"message":"foo"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "message" key is present and with replace`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("message", "foo\n")},
		expected: `{
			"message":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "message" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("message", "bar")},
		expected: `{
			"message":"bar",
			"excerpt":"foo"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "message" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:   log64.Excerpt,
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar")},
		expected: `{
			"message":"bar",
			"excerpt":"foo",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "message" key is present and with replace input bytes and key`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar\n")},
		expected: `{
			"message":"bar\n",
			"excerpt":"foo",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "excerpt" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Excerpt,
		},
		kv: []log64.KV{log64.String("excerpt", "foo")},
		expected: `{
			"excerpt":"foo"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "excerpt" key is present and with replace`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("excerpt", "foo\n")},
		expected: `{
			"excerpt":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" key is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Excerpt,
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" key is present and with replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" key is present and with replace input bytes and rey`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("excerpt", "bar\n")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "excerpt" and "message" keys is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:   log64.Excerpt,
		},
		kv: []log64.KV{log64.String("message", "foo"), log64.String("excerpt", "bar")},
		expected: `{
			"message":"foo",
			"excerpt":"bar"
		}`,
	},
	{
		name: `default key is excerpt and bytes is nil and "excerpt" and "message" keys is present and replace keys`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		kv: []log64.KV{log64.String("message", "foo\n"), log64.String("excerpt", "bar\n")},
		expected: `{
			"message":"foo\n",
			"excerpt":"bar\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" and "message" keys is present`,
		line: line(),
		log: &log64.Log{
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:   log64.Excerpt,
		},
		input: []byte("foo"),
		kv:    []log64.KV{log64.String("message", "bar"), log64.String("excerpt", "xyz")},
		expected: `{
			"message":"bar",
			"excerpt":"xyz",
			"trail":"foo"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" and "message" keys is present and replace input bytes`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar"), log64.String("excerpt", "xyz")},
		expected: `{
			"message":"bar",
			"excerpt":"xyz",
			"trail":"foo\n"
		}`,
	},
	{
		name: `default key is excerpt and bytes is present and "excerpt" and "message" keys is present and replace input bytes and keys`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail")},
			Key:     log64.Excerpt,
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: []byte("foo\n"),
		kv:    []log64.KV{log64.String("message", "bar\n"), log64.String("excerpt", "xyz\n")},
		expected: `{
			"message":"bar\n",
			"excerpt":"xyz\n",
			"trail":"foo\n"
		}`,
	},
	{
		name: `bytes is nil and bytes "message" key with json`,
		line: line(),
		log:  dummy,
		kv:   []log64.KV{log64.StringBytes("message", []byte(`{"foo":"bar"}`))},
		expected: `{
			"message":"{\"foo\":\"bar\"}"
		}`,
	},
	{
		name: `bytes is nil and raw "message" key with json`,
		line: line(),
		log:  dummy,
		kv:   []log64.KV{log64.StringRaw("message", []byte(`{"foo":"bar"}`))},
		expected: `{
			"message":{"foo":"bar"}
		}`,
	},
	{
		name: "bytes is nil and flag is long file",
		line: line(),
		log: &log64.Log{
			Flag: log.Llongfile,
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		kv: []log64.KV{log64.String("foo", "bar")},
		expected: `{
			"foo":"bar"
		}`,
	},
	{
		name: "bytes is one char and flag is long file",
		line: line(),
		log: &log64.Log{
			Flag: log.Llongfile,
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: []byte("a"),
		expected: `{
			"message":"a"
		}`,
	},
	{
		name: "bytes is two chars and flag is long file",
		line: line(),
		log: &log64.Log{
			Flag: log.Llongfile,
			Keys: [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
		},
		input: []byte("ab"),
		expected: `{
			"message":"ab",
			"file":"a"
		}`,
	},
	{
		name: "bytes is three chars and flag is long file",
		line: line(),
		log: &log64.Log{
			Flag: log.Llongfile,
			Keys: [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
		},
		input: []byte("abc"),
		expected: `{
			"message":"abc",
			"file":"ab"
		}`,
	},
	{
		name: "permanent kv overwritten by the additional kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{log64.String("foo", "bar")},
		},
		kv: []log64.KV{log64.String("foo", "baz")},
		expected: `{
			"foo":"baz"
		}`,
	},
	{
		name: "permanent kv and first additional kv overwritten by the second additional kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{log64.String("foo", "bar")},
		},
		kv: []log64.KV{
			log64.String("foo", "baz"),
			log64.String("foo", "xyz"),
		},
		expected: `{
			"foo":"xyz"
		}`,
	},
}

func TestWrite(t *testing.T) {
	_, testFile, _, _ := runtime.Caller(0)
	for _, tc := range WriteTestCases {
		tc := tc
		t.Run(fmt.Sprintf("io writer %s %d", tc.name, tc.line), func(t *testing.T) {
			t.Parallel()
			linkToExample := fmt.Sprintf("%s:%d", testFile, tc.line)

			var buf bytes.Buffer

			l, ok := tc.log.(*log64.Log)
			if !ok {
				t.Fatal("logger type is not appropriate")
			}
			l.Output = &buf

			_, err := tc.log.With(tc.kv...).Write(tc.input)
			if err != nil {
				t.Fatalf("write error: %s", err)
			}

			ja := jsonassert.New(testprinter{t: t, link: linkToExample})
			ja.Assertf(buf.String(), tc.expected)
		})
	}
}

var FprintWriteTestCases = []struct {
	name      string
	line      int
	log       log64.Logger
	input     interface{}
	expected  string
	benchmark bool
}{
	{
		name: "readme example 1",
		log: &log64.Log{
			Trunc:   12,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Marks:   [3][]byte{[]byte("…")},
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		line:  line(),
		input: "Hello,\nWorld!",
		expected: `{
			"message":"Hello,\nWorld!",
			"excerpt":"Hello, World…"
		}`,
	},
	{
		name: "readme example 2",
		line: line(),
		log: func() *log64.Log {
			lg := log64.GELF()
			lg.Func = []func() log64.KV{
				func() log64.KV {
					return log64.StringInt64("timestamp", time.Date(2020, time.October, 15, 18, 9, 0, 0, time.UTC).Unix())
				},
			}
			lg.KV = []log64.KV{log64.String("version", "1.1")}
			return lg
		}(),
		input: "Hello,\nGELF!",
		expected: `{
			"version":"1.1",
			"short_message":"Hello, GELF!",
			"full_message":"Hello,\nGELF!",
			"timestamp":1602785340
		}`,
	},
	{
		name: "readme example 3.1",
		log: &log64.Log{
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		line:  line(),
		input: 3.21,
		expected: `{
			"message":"3.21"
		}`,
	},
	{
		name: "readme example 3.2",
		log: &log64.Log{
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		line:  line(),
		input: 123,
		expected: `{
			"message":"123"
		}`,
	},
	{
		name:  "string",
		line:  line(),
		log:   dummy,
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name:  "integer type appears in the messages excerpt as a string",
		line:  line(),
		log:   dummy,
		input: 123,
		expected: `{
			"message":"123"
		}`,
	},
	{
		name:  "float type appears in the messages excerpt as a string",
		line:  line(),
		log:   dummy,
		input: 3.21,
		expected: `{
			"message":"3.21"
		}`,
	},
	{
		name:  "empty message",
		line:  line(),
		log:   dummy,
		input: "",
		expected: `{
	    "message":"",
			"excerpt":"_EMPTY_"
		}`,
	},
	{
		name:  "blank message",
		line:  line(),
		log:   dummy,
		input: " ",
		expected: `{
	    "message":" ",
			"excerpt":"_BLANK_"
		}`,
	},
	{
		name:  "single quotes",
		line:  line(),
		log:   dummy,
		input: "foo 'bar'",
		expected: `{
			"message":"foo 'bar'"
		}`,
	},
	{
		name:  "double quotes",
		line:  line(),
		log:   dummy,
		input: `foo "bar"`,
		expected: `{
			"message":"foo \"bar\""
		}`,
	},
	{
		name:  `leading/trailing "spaces"`,
		line:  line(),
		log:   dummy,
		input: " \n\tHello, World! \t\n",
		expected: `{
			"message":" \n\tHello, World! \t\n",
			"excerpt":"Hello, World!"
		}`,
	},
	{
		name:  "JSON string",
		line:  line(),
		log:   dummy,
		input: `{"foo":"bar"}`,
		expected: `{
			"message":"{\"foo\":\"bar\"}"
		}`,
	},
	{
		name: `"string" key with "foo" value`,
		line: line(),
		log: &log64.Log{
			KV:   []log64.KV{log64.String("string", "foo")},
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!",
		  "string": "foo"
		}`,
	},
	{
		name: `"integer" key with 123 value`,
		line: line(),
		log: &log64.Log{
			KV:   []log64.KV{log64.StringInt("integer", 123)},
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!",
		  "integer": 123
		}`,
	},
	{
		name: `"float" key with 3.21 value`,
		line: line(),
		log: &log64.Log{
			KV:   []log64.KV{log64.StringFloat32("float", 3.21)},
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!",
		  "float": 3.21
		}`,
	},
	{
		name:  "fmt.Fprint prints nil as <nil>",
		line:  line(),
		log:   dummy,
		input: nil,
		expected: `{
			"message":"<nil>"
		}`,
	},
	{
		name:  "multiline string",
		line:  line(),
		log:   dummy,
		input: "Hello,\nWorld\n!",
		expected: `{
			"message":"Hello,\nWorld\n!",
			"excerpt":"Hello, World !"
		}`,
	},
	{
		name:  "long string",
		line:  line(),
		log:   dummy,
		input: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		expected: `{
			"message":"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			"excerpt":"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliq…"
		}`,
	},
	{
		name:  "multiline long string with leading spaces",
		line:  line(),
		log:   dummy,
		input: " \n \tLorem ipsum dolor sit amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		expected: `{
			"message":" \n \tLorem ipsum dolor sit amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			"excerpt":"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliq…"
		}`,
	},
	{
		name:  "multiline long string with leading spaces and multibyte character",
		line:  line(),
		log:   dummy,
		input: " \n \tLorem ipsum dolor sit amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor incididunt ut labore et dolore magna Ää.",
		expected: `{
			"message":" \n \tLorem ipsum dolor sit amet,\nconsectetur adipiscing elit,\nsed do eiusmod tempor incididunt ut labore et dolore magna Ää.",
			"excerpt":"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna Ää…"
		}`,
		benchmark: true,
	},
	{
		name: "zero maximum length",
		log: &log64.Log{
			Keys:  [4]encoding.TextMarshaler{log64.String("message")},
			Trunc: 0,
		},
		line:  line(),
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name: "without message key names",
		log: &log64.Log{
			Keys: [4]encoding.TextMarshaler{},
		},
		line:  line(),
		input: "Hello, World!",
		expected: `{
			"":"Hello, World!"
		}`,
	},
	{
		name: "only original message key name",
		log: &log64.Log{
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		line:  line(),
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name: "explicit byte slice as message excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.StringBytes("excerpt", []byte("Explicit byte slice"))},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":"Explicit byte slice"
		}`,
	},
	{
		name: "explicit string as message excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.String("excerpt", "Explicit string")},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":"Explicit string"
		}`,
	},
	{
		name: "explicit integer as message excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.StringInt("excerpt", 42)},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":42
		}`,
	},
	{
		name: "explicit float as message excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.StringFloat32("excerpt", 4.2)},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":4.2
		}`,
	},
	{
		name: "explicit boolean as message excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.StringBool("excerpt", true)},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":true
		}`,
	},
	{
		name: "explicit rune slice as messages excerpt key",
		line: line(),
		log: &log64.Log{
			KV:    []log64.KV{log64.StringRunes("excerpt", []rune("Explicit rune slice"))},
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
		},
		input: "Hello, World!",
		expected: `{
		  "message": "Hello, World!",
			"excerpt":"Explicit rune slice"
		}`,
	},
	{
		name: `dynamic "time" key`,
		line: line(),
		log: &log64.Log{
			Func: []func() log64.KV{
				func() log64.KV {
					return log64.String("time", time.Date(2020, time.October, 15, 18, 9, 0, 0, time.UTC).String())
				},
			},
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!",
			"time":"2020-10-15 18:09:00 +0000 UTC"
		}`,
	},
	{
		name: `"standard flag" do not respects file path`,
		line: line(),
		log: &log64.Log{
			Flag: log.LstdFlags,
			Keys: [4]encoding.TextMarshaler{log64.String("message")},
		},
		input: "path/to/file1:23: Hello, World!",
		expected: `{
			"message":"path/to/file1:23: Hello, World!"
		}`,
	},
	{
		name: `"long file" flag respects file path`,
		line: line(),
		log: &log64.Log{
			Flag:  log.Llongfile,
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
		},
		input: "path/to/file1:23: Hello, World!",
		expected: `{
			"message":"path/to/file1:23: Hello, World!",
			"excerpt":"Hello, World!",
			"file":"path/to/file1:23"
		}`,
	},
	{
		name: "replace newline character by whitespace character",
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		input: "Hello,\nWorld!",
		expected: `{
			"message":"Hello,\nWorld!",
			"excerpt":"Hello, World!"
		}`,
	},
	{
		name: "remove exclamation marks",
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Replace: [][2][]byte{[2][]byte{[]byte("!")}},
		},
		input: "Hello, World!!!",
		expected: `{
			"message":"Hello, World!!!",
			"excerpt":"Hello, World"
		}`,
	},
	{
		name: `replace word "World" by world "Work"`,
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt")},
			Replace: [][2][]byte{[2][]byte{[]byte("World"), []byte("Work")}},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!",
			"excerpt":"Hello, Work!"
		}`,
	},
	{
		name: "ignore pointless replace",
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message")},
			Replace: [][2][]byte{[2][]byte{[]byte("!"), []byte("!")}},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name: "ignore empty replace",
		line: line(),
		log: &log64.Log{
			Trunc:   120,
			Keys:    [4]encoding.TextMarshaler{log64.String("message")},
			Replace: [][2][]byte{[2][]byte{}},
		},
		input: "Hello, World!",
		expected: `{
			"message":"Hello, World!"
		}`,
	},
	{
		name: "file path with empty message",
		line: line(),
		log: &log64.Log{
			Flag:  log.Llongfile,
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
			Marks: [3][]byte{[]byte("…"), []byte("_EMPTY_")},
		},
		input: "path/to/file1:23:",
		expected: `{
			"message":"path/to/file1:23:",
			"excerpt":"_EMPTY_",
			"file":"path/to/file1:23"
		}`,
	},
	{
		name: "file path with blank message",
		line: line(),
		log: &log64.Log{
			Flag:  log.Llongfile,
			Trunc: 120,
			Keys:  [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
			Marks: [3][]byte{[]byte("…"), []byte("_EMPTY_"), []byte("_BLANK_")},
		},
		input: "path/to/file4:56:  ",
		expected: `{
			"message":"path/to/file4:56:  ",
			"excerpt":"_BLANK_",
			"file":"path/to/file4:56"
		}`,
	},
	{
		name: "GELF",
		line: line(),
		log: func() *log64.Log {
			lg := log64.GELF()
			lg.Func = []func() log64.KV{
				func() log64.KV {
					return log64.StringInt64("timestamp", time.Date(2020, time.October, 15, 18, 9, 0, 0, time.UTC).Unix())
				},
			}
			lg.KV = []log64.KV{log64.String("version", "1.1"), log64.String("host", "example.tld")}
			return lg
		}(),
		input: "Hello, GELF!",
		expected: `{
			"version":"1.1",
			"short_message":"Hello, GELF!",
			"host":"example.tld",
			"timestamp":1602785340
		}`,
	},
	{
		name: "GELF with file path",
		line: line(),
		log: func() *log64.Log {
			lg := log64.GELF()
			lg.Flag = log.Llongfile
			lg.Func = []func() log64.KV{
				func() log64.KV {
					return log64.StringInt64("timestamp", time.Date(2020, time.October, 15, 18, 9, 0, 0, time.UTC).Unix())
				},
			}
			lg.KV = []log64.KV{log64.String("version", "1.1"), log64.String("host", "example.tld")}
			return lg
		}(),
		input: "path/to/file7:89: Hello, GELF!",
		expected: `{
			"version":"1.1",
			"short_message":"Hello, GELF!",
			"full_message":"path/to/file7:89: Hello, GELF!",
			"host":"example.tld",
			"timestamp":1602785340,
			"_file":"path/to/file7:89"
		}`,
	},
}

func TestFprintWrite(t *testing.T) {
	_, testFile, _, _ := runtime.Caller(0)
	for _, tc := range FprintWriteTestCases {
		tc := tc
		t.Run(fmt.Sprintf("io writer via fmt fprint %s %d", tc.name, tc.line), func(t *testing.T) {
			t.Parallel()
			linkToExample := fmt.Sprintf("%s:%d", testFile, tc.line)

			var buf bytes.Buffer

			l, ok := tc.log.(*log64.Log)
			if !ok {
				t.Fatal("logger type is not appropriate")
			}

			l.Output = &buf

			_, err := fmt.Fprint(tc.log, tc.input)
			if err != nil {
				t.Fatalf("write error: %s", err)
			}

			ja := jsonassert.New(testprinter{t: t, link: linkToExample})
			ja.Assertf(buf.String(), tc.expected)
		})
	}
}

func BenchmarkLog64(b *testing.B) {
	for _, tc := range WriteTestCases {
		if !tc.benchmark {
			continue
		}
		b.Run(fmt.Sprintf("io.Writer %d", tc.line), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer

				l, ok := tc.log.(*log64.Log)
				if !ok {
					b.Fatal("logger type is not appropriate")
				}

				l.Output = &buf

				lg := tc.log
				for _, kv := range tc.kv {
					lg = lg.With(kv)
				}
				_, err := lg.Write(tc.input)
				if err != nil {
					fmt.Println(err)
				}
			}
		})
	}

	for _, tc := range FprintWriteTestCases {
		if !tc.benchmark {
			continue
		}
		b.Run(fmt.Sprintf("fmt.Fprint io.Writer %d", tc.line), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer

				l, ok := tc.log.(*log64.Log)
				if !ok {
					b.Fatal("logger type is not appropriate")
				}

				l.Output = &buf

				_, err := fmt.Fprint(tc.log, tc.input)
				if err != nil {
					fmt.Println(err)
				}
			}
		})
	}
}

var dummy = &log64.Log{
	Trunc:   120,
	Keys:    [4]encoding.TextMarshaler{log64.String("message"), log64.String("excerpt"), log64.String("trail"), log64.String("file")},
	Key:     log64.Original,
	Marks:   [3][]byte{[]byte("…"), []byte("_EMPTY_"), []byte("_BLANK_")},
	Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
}

func TestLogWriteTrailingNewLine(t *testing.T) {
	var buf bytes.Buffer

	lg := &log64.Log{Output: &buf}

	_, err := lg.Write([]byte("Hello, Wrold!"))
	if err != nil {
		t.Fatalf("write error: %s", err)
	}

	if buf.Bytes()[len(buf.Bytes())-1] != '\n' {
		t.Errorf("trailing new line expected but not present: %q", buf.String())
	}
}

var TruncateTestCases = []struct {
	name      string
	line      int
	log       log64.Logger
	input     []byte
	expected  []byte
	benchmark bool
}{
	{
		name:     "do nothing",
		log:      &log64.Log{},
		line:     line(),
		input:    []byte("Hello,\nWorld!"),
		expected: []byte("Hello,\nWorld!"),
	},
	{
		name: "truncate last character",
		log: &log64.Log{
			Trunc: 12,
		},
		line:     line(),
		input:    []byte("Hello, World!"),
		expected: []byte("Hello, World"),
	},
	{
		name: "truncate last character and places ellipsis instead",
		log: &log64.Log{
			Trunc: 12,
			Marks: [3][]byte{[]byte("…")},
		},
		line:     line(),
		input:    []byte("Hello, World!"),
		expected: []byte("Hello, World…"),
	},
	{
		name: "replace new lines by spaces",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte(" ")}},
		},
		line:     line(),
		input:    []byte("Hello\n,\nWorld\n!"),
		expected: []byte("Hello , World !"),
	},
	{
		name: "replace new lines by empty string",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("\n"), []byte("")}},
		},
		line:     line(),
		input:    []byte("Hello\n,\nWorld\n!"),
		expected: []byte("Hello,World!"),
	},
	{
		name: "remove new lines",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("\n")}},
		},
		line:     line(),
		input:    []byte("Hello\n,\nWorld\n!"),
		expected: []byte("Hello,World!"),
	},
	{
		name: "replace three characters by one",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("foo"), []byte("f")}, [2][]byte{[]byte("bar"), []byte("b")}},
		},
		line:     line(),
		input:    []byte("foobar"),
		expected: []byte("fb"),
	},
	{
		name: "replace one characters by three",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("f"), []byte("foo")}, [2][]byte{[]byte("b"), []byte("bar")}},
		},
		line:     line(),
		input:    []byte("fb"),
		expected: []byte("foobar"),
	},
	{
		name: "remove three characters",
		log: &log64.Log{
			Replace: [][2][]byte{[2][]byte{[]byte("foo")}, [2][]byte{[]byte("bar")}},
		},
		line:     line(),
		input:    []byte("foobar foobar"),
		expected: []byte(" "),
	},
}

func TestTruncate(t *testing.T) {
	_, testFile, _, _ := runtime.Caller(0)
	for _, tc := range TruncateTestCases {
		tc := tc
		t.Run(fmt.Sprintf("truncate %s %d", tc.name, tc.line), func(t *testing.T) {
			t.Parallel()
			linkToExample := fmt.Sprintf("%s:%d", testFile, tc.line)

			l, ok := tc.log.(*log64.Log)
			if !ok {
				t.Fatal("logger type is not appropriate")
			}

			n := len(tc.input) + 10*10
			for _, m := range l.Marks {
				if n < len(m) {
					n = len(m)
				}
			}

			excerpt := make([]byte, n)

			n, err := l.Truncate(excerpt, tc.input)
			if err != nil {
				t.Fatalf("write error: %s", err)
			}

			excerpt = excerpt[:n]

			if !bytes.Equal(excerpt, tc.expected) {
				t.Errorf("unexpected excerpt, expected: %q, received %q %s", tc.expected, excerpt, linkToExample)
			}
		})
	}
}

var WithTestCases = []struct {
	name      string
	line      int
	log       log64.Logger
	kv        []log64.KV
	expected  string
	benchmark bool
}{
	{
		name: "one kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{
				log64.String("foo", "bar"),
			},
		},
		expected: `{
			"foo":"bar"
		}`,
	},
	{
		name: "two kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{
				log64.String("foo", "bar"),
				log64.String("baz", "xyz"),
			},
		},
		expected: `{
			"foo":"bar",
			"baz":"xyz"
		}`,
	},
	{
		name: "one additional kv",
		line: line(),
		log:  &log64.Log{},
		kv: []log64.KV{
			log64.String("baz", "xyz"),
		},
		expected: `{
			"baz":"xyz"
		}`,
	},
	{
		name: "two additional kv",
		line: line(),
		log:  &log64.Log{},
		kv: []log64.KV{
			log64.String("foo", "bar"),
			log64.String("baz", "xyz"),
		},
		expected: `{
			"foo":"bar",
			"baz":"xyz"
		}`,
	},
	{
		name: "one kv with additional one kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{
				log64.String("foo", "bar"),
			},
		},
		kv: []log64.KV{
			log64.String("baz", "xyz"),
		},
		expected: `{
			"foo":"bar",
			"baz":"xyz"
		}`,
	},
	{
		name: "two kv with two additional kv",
		line: line(),
		log: &log64.Log{
			KV: []log64.KV{
				log64.String("foo", "bar"),
				log64.String("abc", "dfg"),
			},
		},
		kv: []log64.KV{
			log64.String("baz", "xyz"),
			log64.String("hjk", "lmn"),
		},
		expected: `{
			"foo":"bar",
			"abc":"dfg",
			"baz":"xyz",
			"hjk":"lmn"
		}`,
	},
}

func TestWith(t *testing.T) {
	_, testFile, _, _ := runtime.Caller(0)
	for _, tc := range WithTestCases {
		tc := tc
		t.Run(fmt.Sprintf("with %s %d", tc.name, tc.line), func(t *testing.T) {
			t.Parallel()
			linkToExample := fmt.Sprintf("%s:%d", testFile, tc.line)

			var buf bytes.Buffer

			l, ok := tc.log.(*log64.Log)
			if !ok {
				t.Fatal("logger type is not appropriate")
			}

			l.Output = &buf

			_, err := tc.log.With(tc.kv...).Write(nil)
			if err != nil {
				t.Fatalf("write error: %s", err)
			}

			ja := jsonassert.New(testprinter{t: t, link: linkToExample})
			ja.Assertf(buf.String(), tc.expected)
		})
	}
}
