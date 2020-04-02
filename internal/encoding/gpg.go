package encoding

import (
	"errors"
	"io"

	"golang.org/x/crypto/openpgp"
)

func init() {
	entry := Entry{
		Extensions: []string{"gpg"},
		Settings:   []string{"password", "size"},
		Encoding:   &gpg{},
	}
	Add(entry)
}

type gpg struct {
}

type gpgWriter struct {
	out       io.Closer
	plaintext io.WriteCloser
}

func (w *gpgWriter) Write(p []byte) (n int, err error) {
	return w.plaintext.Write(p)
}

func (w *gpgWriter) Close() error {
	err := w.plaintext.Close()
	if err != nil {
		return err
	}
	return w.out.Close()
}

func (a *gpg) Encode(writer io.WriteCloser, settings map[string]interface{}) (io.WriteCloser, error) {
	password := settings["password"].(string)
	if password == "" {
		return writer, nil
	}

	plaintext, err := openpgp.SymmetricallyEncrypt(writer, []byte(password), nil, nil)
	if err != nil {
		return plaintext, err
	}

	plaintext = &gpgWriter{
		out:       writer,
		plaintext: plaintext,
	}

	return plaintext, nil
}

func (a *gpg) Decode(reader io.Reader, settings map[string]interface{}) (io.Reader, error) {
	password := settings["password"].(string)
	if settings["size"].(int64) == 0 || password == "" {
		return reader, nil
	}

	attempts := 0
	md, err := openpgp.ReadMessage(reader, nil, func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if attempts > 0 {
			return []byte{}, errors.New("invalid password")
		}
		attempts++
		return []byte(password), nil
	}, nil)
	if err != nil {
		return reader, err
	}
	reader = md.UnverifiedBody

	return reader, nil
}
