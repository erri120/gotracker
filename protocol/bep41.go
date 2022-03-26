package protocol

import (
	"bytes"
	"fmt"
	"math"
)

func ExtractExtensionData(reader *bytes.Reader) (string, error) {
	var optionType uint8
	if err := Unmarshal(reader, &optionType); err != nil {
		return "", err
	}

	option := BEP41OptionType(optionType)

	switch option {
	case BEP41OptionTypeEndOfOptions:
		return "", nil
	case BEP41OptionTypeNOP:
		return "", nil
	case BEP41OptionTypeURLData:
		var length uint8
		if err := Unmarshal(reader, &length); err != nil {
			return "", err
		}

		if length == 0 {
			return "", nil
		}

		if length == math.MaxUint8 {
			// TODO: BEP 41 says you can go around this limit by appending multiple URLData fields to the request
			return "", nil
		}

		urlData := make([]byte, length)
		if _, err := reader.Read(urlData); err != nil {
			return "", err
		}

		urlPart := string(urlData)
		return urlPart, nil
	default:
		return "", fmt.Errorf("Unknown BEP 41 option type %d", optionType)
	}
}
