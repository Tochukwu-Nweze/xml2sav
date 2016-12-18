package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

type BytecodeWriter struct {
	io.Writer
	bias    float64
	command [8]byte
	index   int
	data    bytes.Buffer
}

func NewBytecodeWriter(w io.Writer, bias float64) *BytecodeWriter {
	return &BytecodeWriter{Writer: w, bias: bias}
}

func (w *BytecodeWriter) checkAndWrite() error {
	if w.index >= len(w.command) {
		if _, err := w.Write(w.command[:]); err != nil {
			return err
		}
		if _, err := w.Write(w.data.Bytes()); err != nil {
			return err
		}
		w.index = 0
		w.data.Truncate(0)
	}
	return nil
}

func (w *BytecodeWriter) WriteMissing() error {
	w.command[w.index] = 255
	w.index++
	return w.checkAndWrite()
}

func (w *BytecodeWriter) WriteNumber(number float64) error {
	for i := 1.0; i <= 251; i++ {
		if number == i-w.bias {
			w.command[w.index] = byte(i)
			w.index++
			return w.checkAndWrite()
		}
	}
	w.command[w.index] = 253
	w.index++
	binary.Write(&w.data, endian, number)
	return w.checkAndWrite()
}

func (w *BytecodeWriter) WriteString(val string, elements int) error {
	for i := 0; i < elements; i++ {
		var p string
		if len(val) > 8 {
			p = val[:8]
			val = val[8:]
		} else {
			p = val
			val = ""
		}

		if len(p) < 8 {
			p += strings.Repeat(" ", 8-len(p))
		}

		if p == "        " {
			w.command[w.index] = 254
			w.index++
		} else {
			w.command[w.index] = 253
			w.index++
			w.data.Write([]byte(p))
		}

		if err := w.checkAndWrite(); err != nil {
			return err
		}
	}

	return nil
}

func (w *BytecodeWriter) Flush() error {
	for w.index < 8 {
		w.command[w.index] = 0
		w.index++
	}
	return w.checkAndWrite()
}
