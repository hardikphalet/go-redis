package resp

import (
	"bufio"
	"fmt"
)

type Writer struct {
	writer *bufio.Writer
}

// NewWriter creates a new RESP Writer
func NewWriter(writer *bufio.Writer) *Writer {
	return &Writer{writer: writer}
}

// WriteString writes a RESP Simple String ("+OK\r\n")
func (w *Writer) WriteString(s string) error {
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", s)
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// WriteError writes a RESP Error ("-Error message\r\n")
func (w *Writer) WriteError(err error) error {
	_, err2 := fmt.Fprintf(w.writer, "-ERR %s\r\n", err.Error())
	if err2 != nil {
		return err2
	}
	return w.writer.Flush()
}

// WriteInteger writes a RESP Integer (":1000\r\n")
func (w *Writer) WriteInteger(i int64) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", i)
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// WriteBulkString writes a RESP Bulk String ("$5\r\nhello\r\n")
func (w *Writer) WriteBulkString(s string) error {
	if s == "" {
		// Empty string is encoded as "$0\r\n\r\n"
		_, err := fmt.Fprintf(w.writer, "$0\r\n\r\n")
		if err != nil {
			return err
		}
		return w.writer.Flush()
	}

	// Write the length prefix
	_, err := fmt.Fprintf(w.writer, "$%d\r\n%s\r\n", len(s), s)
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// WriteNull writes a RESP Null value ("$-1\r\n")
func (w *Writer) WriteNull() error {
	_, err := fmt.Fprintf(w.writer, "$-1\r\n")
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

// WriteArray writes a RESP Array ("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
func (w *Writer) WriteArray(arr []string) error {
	if arr == nil {
		// Null array is encoded as "*-1\r\n"
		_, err := fmt.Fprintf(w.writer, "*-1\r\n")
		if err != nil {
			return err
		}
		return w.writer.Flush()
	}

	// Write array length
	_, err := fmt.Fprintf(w.writer, "*%d\r\n", len(arr))
	if err != nil {
		return err
	}

	// Write each element as a bulk string
	for _, s := range arr {
		err := w.WriteBulkString(s)
		if err != nil {
			return err
		}
	}

	return w.writer.Flush()
}

// WriteInterface writes any interface{} value in the appropriate RESP format
func (w *Writer) WriteInterface(v interface{}) error {
	if v == nil {
		return w.WriteNull()
	}

	switch val := v.(type) {
	case string:
		return w.WriteBulkString(val)
	case []string:
		return w.WriteArray(val)
	case int:
		return w.WriteInteger(int64(val))
	case int64:
		return w.WriteInteger(val)
	case error:
		return w.WriteError(val)
	case []interface{}:
		// Convert []interface{} to []string
		strArr := make([]string, len(val))
		for i, v := range val {
			strArr[i] = fmt.Sprintf("%v", v)
		}
		return w.WriteArray(strArr)
	default:
		// Convert anything else to string
		return w.WriteBulkString(fmt.Sprintf("%v", v))
	}
}
