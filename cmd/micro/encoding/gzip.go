// +build phat

package encoding

import (
	"compress/gzip"
	"io"
)

func init() {
	entry := Entry{
		Extensions: []string{"gz"},
		Settings:   []string{"size"},
		Encoding:   &gzipEncoding{},
	}
	Add(entry)
}

type gzipEncoding struct {
}

type gzipWriter struct {
	out       io.Closer
	plaintext io.WriteCloser
}

func (w *gzipWriter) Write(p []byte) (n int, err error) {
	return w.plaintext.Write(p)
}

func (w *gzipWriter) Close() error {
	err := w.plaintext.Close()
	if err != nil {
		return err
	}
	return w.out.Close()
}

func (g *gzipEncoding) Encode(writer io.WriteCloser, settings map[string]interface{}) (io.WriteCloser, error) {
	plaintext, err := gzip.NewWriterLevel(writer, gzip.BestCompression)
	if err != nil {
		return plaintext, err
	}

	w := &gzipWriter{
		out:       writer,
		plaintext: plaintext,
	}

	return w, nil
}

func (g *gzipEncoding) Decode(reader io.Reader, settings map[string]interface{}) (io.Reader, error) {
	if settings["size"].(int64) == 0 {
		return reader, nil
	}
	return gzip.NewReader(reader)
}
