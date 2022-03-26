package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

func Marshal(bufSize uint32, parts ...interface{}) (result []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	for _, part := range parts {
		err = binary.Write(buf, binary.BigEndian, part)
		if err != nil {
			return
		}
	}

	result = buf.Bytes()
	return
}

func Unmarshal(reader io.Reader, data any) error {
	return binary.Read(reader, binary.BigEndian, data)
}
