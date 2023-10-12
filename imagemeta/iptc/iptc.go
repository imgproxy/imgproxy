package iptc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

var (
	iptcTagHeader = byte(0x1c)

	errInvalidDataSize = errors.New("invalid IPTC data size")
)

type IptcMap map[TagKey][]TagValue

func (m IptcMap) AddTag(key TagKey, data []byte) error {
	info, infoFound := tagInfoMap[key]
	if !infoFound {
		return fmt.Errorf("unknown tag %d:%d", key.RecordID, key.TagID)
	}

	dataSize := len(data)
	if dataSize < info.MinSize || dataSize > info.MaxSize {
		return fmt.Errorf("invalid tag data size. Min: %d, Max: %d, Has: %d", info.MinSize, info.MaxSize, dataSize)
	}

	value := TagValue{info.Format, data}

	if info.Repeatable {
		m[key] = append(m[key], value)
	} else {
		m[key] = []TagValue{value}
	}

	return nil
}

func (m IptcMap) MarshalJSON() ([]byte, error) {
	mm := make(map[string]interface{}, len(m))
	for key, values := range m {
		info, infoFound := tagInfoMap[key]
		if !infoFound {
			continue
		}

		if info.Repeatable {
			mm[info.Title] = values
		} else {
			mm[info.Title] = values[0]
		}

		// Add some additional fields for backward compatibility
		if key.RecordID == 2 {
			if key.TagID == 5 {
				mm["Name"] = values[0]
			} else if key.TagID == 120 {
				mm["Caption"] = values[0]
			}
		}
	}
	return json.Marshal(mm)
}

func Parse(data []byte, m IptcMap) error {
	buf := bytes.NewBuffer(data)

	// Min tag size is 5 (2 tagHeader)
	for buf.Len() >= 5 {
		if buf.Next(1)[0] != iptcTagHeader {
			continue
		}

		recordID, _ := buf.ReadByte()
		tagID, _ := buf.ReadByte()
		dataSize16 := binary.BigEndian.Uint16(buf.Next(2))

		var dataSize int

		if dataSize16 < 32768 {
			dataSize = int(dataSize16)
		} else {
			dataSizeSize := dataSize16 & 32767

			switch dataSizeSize {
			case 4:
				dataSize32 := uint32(0)
				if err := binary.Read(buf, binary.BigEndian, &dataSize32); err != nil {
					return fmt.Errorf("%s: %s", errInvalidDataSize, err)
				}
				dataSize = int(dataSize32)
			case 8:
				dataSize64 := uint64(0)
				if err := binary.Read(buf, binary.BigEndian, &dataSize64); err != nil {
					return fmt.Errorf("%s: %s", errInvalidDataSize, err)
				}
				dataSize = int(dataSize64)
			default:
				return errInvalidDataSize
			}
		}

		// Ignore errors here. If tag is invalid, just don't add it
		m.AddTag(TagKey{recordID, tagID}, buf.Next(dataSize))
	}

	return nil
}

func (m IptcMap) Dump() []byte {
	buf := new(bytes.Buffer)

	for key, values := range m {
		for _, value := range values {
			dataSize := len(value.Raw)
			// Skip tags with too big data size
			if dataSize > math.MaxUint32 {
				continue
			}

			buf.WriteByte(iptcTagHeader)
			buf.WriteByte(key.RecordID)
			buf.WriteByte(key.TagID)

			if dataSize < (1 << 15) {
				binary.Write(buf, binary.BigEndian, uint16(dataSize))
			} else {
				binary.Write(buf, binary.BigEndian, uint16(4+(1<<15)))
				binary.Write(buf, binary.BigEndian, uint32(dataSize))
			}

			buf.Write(value.Raw)
		}
	}

	return buf.Bytes()
}
