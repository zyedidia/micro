package encoding

import (
	"errors"
	"io"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

func init() {
	entry := Entry{
		Extensions: []string{"asc"},
		Settings:   []string{"password", "size"},
		Encoding:   &armorgpg{},
	}
	Add(entry)
}

type armorgpg struct {
}

type armorgpgWriter struct {
	out       io.Closer
	armor     io.Closer
	plaintext io.WriteCloser
}

func (w *armorgpgWriter) Write(p []byte) (n int, err error) {
	return w.plaintext.Write(p)
}

func (w *armorgpgWriter) Close() error {
	err := w.plaintext.Close()
	if err != nil {
		return err
	}
	err = w.armor.Close()
	if err != nil {
		return err
	}
	return w.out.Close()
}

func (a *armorgpg) Encode(writer io.WriteCloser, settings map[string]interface{}) (io.WriteCloser, error) {
	password := settings["password"].(string)
	if password == "" {
		return writer, nil
	}

	arm, err := armor.Encode(writer, "PGP SIGNATURE", nil)
	if err != nil {
		return arm, err
	}

	plaintext, err := openpgp.SymmetricallyEncrypt(arm, []byte(password), nil, nil)
	if err != nil {
		return plaintext, err
	}

	plaintext = &armorgpgWriter{
		out:       writer,
		armor:     arm,
		plaintext: plaintext,
	}

	return plaintext, nil
}

func (a *armorgpg) Decode(reader io.Reader, settings map[string]interface{}) (io.Reader, error) {
	password := settings["password"].(string)
	if settings["size"].(int64) == 0 || password == "" {
		return reader, nil
	}

	unarmored, err := armor.Decode(reader)
	if err != nil {
		return reader, err
	}
	reader = unarmored.Body

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
