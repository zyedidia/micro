package encoding

import (
	"bytes"
	"io/ioutil"
	"testing"
)

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Write(p []byte) (n int, err error) {
	return b.Buffer.Write(p)
}

func (b *buffer) Close() error {
	return nil
}

func TestEncoding(t *testing.T) {
	test := func(name string) {
		output, settings := &buffer{}, map[string]interface{}{"password": "abc123", "size": int64(0)}
		out, err := Encoder(output, name, settings)
		if err != nil {
			t.Fatal(err)
		}
		_, err = out.Write([]byte("hello world"))
		if err != nil {
			t.Fatal(err)
		}
		err = out.Close()
		if err != nil {
			t.Fatal(err)
		}
		settings["size"] = int64(output.Len())
		in, err := Decoder(output, name, settings)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(in)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello world" {
			t.Fatalf("should be 'hello world', but is %s", string(data))
		}
	}
	test("test.asc")
	test("test.gpg")
	test("test.asc.gz")
	test("test.gpg.gz")
}
