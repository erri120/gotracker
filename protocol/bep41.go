package protocol

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
)

func ExtractExtensionData(reader *bytes.Reader) ([]byte, error) {
	var optionType uint8
	if err := Unmarshal(reader, &optionType); err != nil {
		return nil, err
	}

	option := BEP41OptionType(optionType)

	switch option {
	case BEP41OptionTypeEndOfOptions:
		return nil, nil
	case BEP41OptionTypeNOP:
		return nil, nil
	case BEP41OptionTypeURLData:
		var length uint8
		if err := Unmarshal(reader, &length); err != nil {
			return nil, err
		}

		if length == 0 {
			return nil, nil
		}

		if length == math.MaxUint8 {
			// TODO: BEP 41 says you can go around this limit by appending multiple URLData fields to the request
			return nil, nil
		}

		urlData := make([]byte, length)
		if _, err := reader.Read(urlData); err != nil {
			return nil, err
		}

		return urlData, nil
	default:
		return nil, fmt.Errorf("Unknown BEP 41 option type %d", optionType)
	}
}

func ConvertUrlDataToUrl(urlData []byte) (*url.URL, error) {
	if len(urlData) == 0 {
		return &url.URL{}, fmt.Errorf("URL data is empty")
	}

	urlString := string(urlData)

	if urlString[0] != '/' {
		return &url.URL{}, fmt.Errorf("URL data must start with /")
	}

	return url.Parse(urlString)
}
