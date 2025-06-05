package resp

import "bufio"

type Writer struct {
	writer *bufio.Writer
}

func (w *Writer) WriteString(s string) error
func (w *Writer) WriteError(err error) error
func (w *Writer) WriteArray(arr []string) error
