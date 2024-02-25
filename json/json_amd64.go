package json

import (
	"bytes"
	runtimejson "encoding/json"
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
)

type Decoder struct {
	*decoder.Decoder
}

type Encoder struct {
	*encoder.StreamEncoder
}

type Number string

func newDecoder(r io.Reader) *Decoder {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return &Decoder{decoder.NewDecoder(buf.String())}
}

func newEncoder(w io.Writer) *Encoder {
	return &Encoder{encoder.NewStreamEncoder(w)}
}

var (
	NewDecoder    = newDecoder
	NewEncoder    = newEncoder
	Marshal       = sonic.Marshal
	Unmarshal     = sonic.Unmarshal
	MarshalIndent = runtimejson.MarshalIndent
)
