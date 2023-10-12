package iptc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

type TagFormat int

const (
	TagFormatByte TagFormat = iota
	TagFormatShort
	TagFormatLong
	TagFormatString
	TagFormatBinary
	TagFormatDate
	TagFormatTime
)

type TagKey struct {
	RecordID byte
	TagID    byte
}

type TagInfo struct {
	Name       string
	Title      string
	Format     TagFormat
	Required   bool
	Repeatable bool
	MinSize    int
	MaxSize    int
}

var tagInfoMap = map[TagKey]TagInfo{
	{1, 0}: {
		"ModelVersion",
		"Model Version",
		TagFormatShort, true, false, 2, 2,
	},
	{1, 5}: {
		"Destination",
		"Destination",
		TagFormatString, false, true, 0, 1024,
	},
	{1, 20}: {
		"FileFormat",
		"File Format",
		TagFormatShort, true, false, 2, 2,
	},
	{1, 22}: {
		"FileVersion",
		"File Version",
		TagFormatShort, true, false, 2, 2,
	},
	{1, 30}: {
		"ServiceID",
		"Service Identifier",
		TagFormatString, true, false, 0, 10,
	},
	{1, 40}: {
		"EnvelopeNum",
		"Envelope Number",
		TagFormatString, true, false, 8, 8,
	},
	{1, 50}: {
		"ProductID",
		"Product I.D.",
		TagFormatString, false, true, 0, 32,
	},
	{1, 60}: {
		"EnvelopePriority",
		"Envelope Priority",
		TagFormatString, false, false, 1, 1,
	},
	{1, 70}: {
		"DateSent",
		"Date Sent",
		TagFormatDate, true, false, 8, 8,
	},
	{1, 80}: {
		"TimeSent",
		"Time Sent",
		TagFormatTime, false, false, 11, 11,
	},
	{1, 90}: {
		"CharacterSet",
		"Coded Character Set",
		TagFormatBinary, false, false, 0, 32,
	},
	{1, 100}: {
		"UNO",
		"Unique Name of Object",
		TagFormatString, false, false, 14, 80,
	},
	{1, 120}: {
		"ARMID",
		"ARM Identifier",
		TagFormatShort, false, false, 2, 2,
	},
	{1, 122}: {
		"ARMVersion",
		"ARM Version",
		TagFormatShort, false, false, 2, 2,
	},
	{2, 0}: {
		"RecordVersion",
		"Record Version",
		TagFormatShort, true, false, 2, 2,
	},
	{2, 3}: {
		"ObjectType",
		"Object Type Reference",
		TagFormatString, false, false, 3, 67,
	},
	{2, 4}: {
		"ObjectAttribute",
		"Object Attribute Reference",
		TagFormatString, false, true, 4, 68,
	},
	{2, 5}: {
		"ObjectName",
		"Object Name",
		TagFormatString, false, false, 0, 64,
	},
	{2, 7}: {
		"EditStatus",
		"Edit Status",
		TagFormatString, false, false, 0, 64,
	},
	{2, 8}: {
		"EditorialUpdate",
		"Editorial Update",
		TagFormatString, false, false, 2, 2,
	},
	{2, 10}: {
		"Urgency",
		"Urgency",
		TagFormatString, false, false, 1, 1,
	},
	{2, 12}: {
		"SubjectRef",
		"Subject Reference",
		TagFormatString, false, true, 13, 236,
	},
	{2, 15}: {
		"Category",
		"Category",
		TagFormatString, false, false, 0, 3,
	},
	{2, 20}: {
		"SupplCategory",
		"Supplemental Category",
		TagFormatString, false, true, 0, 32,
	},
	{2, 22}: {
		"FixtureID",
		"Fixture Identifier",
		TagFormatString, false, false, 0, 32,
	},
	{2, 25}: {
		"Keywords",
		"Keywords",
		TagFormatString, false, true, 0, 64,
	},
	{2, 26}: {
		"ContentLocCode",
		"Content Location Code",
		TagFormatString, false, true, 3, 3,
	},
	{2, 27}: {
		"ContentLocName",
		"Content Location Name",
		TagFormatString, false, true, 0, 64,
	},
	{2, 30}: {
		"ReleaseDate",
		"Release Date",
		TagFormatDate, false, false, 8, 8,
	},
	{2, 35}: {
		"ReleaseTime",
		"Release Time",
		TagFormatTime, false, false, 11, 11,
	},
	{2, 37}: {
		"ExpirationDate",
		"Expiration Date",
		TagFormatDate, false, false, 8, 8,
	},
	{2, 38}: {
		"ExpirationTime",
		"Expiration Time",
		TagFormatTime, false, false, 11, 11,
	},
	{2, 40}: {
		"SpecialInstructions",
		"Special Instructions",
		TagFormatString, false, false, 0, 256,
	},
	{2, 42}: {
		"ActionAdvised",
		"Action Advised",
		TagFormatString, false, false, 2, 2,
	},
	{2, 45}: {
		"RefService",
		"Reference Service",
		TagFormatString, false, true, 0, 10,
	},
	{2, 47}: {
		"RefDate",
		"Reference Date",
		TagFormatDate, false, true, 8, 8,
	},
	{2, 50}: {
		"RefNumber",
		"Reference Number",
		TagFormatString, false, true, 8, 8,
	},
	{2, 55}: {
		"DateCreated",
		"Date Created",
		TagFormatDate, false, false, 8, 19,
	},
	{2, 60}: {
		"TimeCreated",
		"Time Created",
		TagFormatTime, false, false, 6, 11,
	},
	{2, 62}: {
		"DigitalCreationDate",
		"Digital Creation Date",
		TagFormatDate, false, false, 8, 8,
	},
	{2, 63}: {
		"DigitalCreationTime",
		"Digital Creation Time",
		TagFormatTime, false, false, 11, 11,
	},
	{2, 65}: {
		"OriginatingProgram",
		"Originating Program",
		TagFormatString, false, false, 0, 32,
	},
	{2, 70}: {
		"ProgramVersion",
		"Program Version",
		TagFormatString, false, false, 0, 10,
	},
	{2, 75}: {
		"ObjectCycle",
		"Object Cycle",
		TagFormatString, false, false, 1, 1,
	},
	{2, 80}: {
		"Byline",
		"By-line",
		TagFormatString, false, true, 0, 32,
	},
	{2, 85}: {
		"BylineTitle",
		"By-line Title",
		TagFormatString, false, true, 0, 32,
	},
	{2, 90}: {
		"City",
		"City",
		TagFormatString, false, false, 0, 32,
	},
	{2, 92}: {
		"Sublocation",
		"Sub-location",
		TagFormatString, false, false, 0, 32,
	},
	{2, 95}: {
		"State",
		"Province/State",
		TagFormatString, false, false, 0, 32,
	},
	{2, 100}: {
		"CountryCode",
		"Country Code",
		TagFormatString, false, false, 3, 3,
	},
	{2, 101}: {
		"CountryName",
		"Country Name",
		TagFormatString, false, false, 0, 64,
	},
	{2, 103}: {
		"OrigTransRef",
		"Original Transmission Reference",
		TagFormatString, false, false, 0, 32,
	},
	{2, 105}: {
		"Headline",
		"Headline",
		TagFormatString, false, false, 0, 256,
	},
	{2, 110}: {
		"Credit",
		"Credit",
		TagFormatString, false, false, 0, 32,
	},
	{2, 115}: {
		"Source",
		"Source",
		TagFormatString, false, false, 0, 32,
	},
	{2, 116}: {
		"CopyrightNotice",
		"Copyright Notice",
		TagFormatString, false, false, 0, 128,
	},
	{2, 118}: {
		"Contact",
		"Contact",
		TagFormatString, false, true, 0, 128,
	},
	{2, 120}: {
		"Caption",
		"Caption/Abstract",
		TagFormatString, false, false, 0, 2000,
	},
	{2, 122}: {
		"WriterEditor",
		"Writer/Editor",
		TagFormatString, false, true, 0, 32,
	},
	{2, 125}: {
		"RasterizedCaption",
		"Rasterized Caption",
		TagFormatBinary, false, false, 7360, 7360,
	},
	{2, 130}: {
		"ImageType",
		"Image Type",
		TagFormatString, false, false, 2, 2,
	},
	{2, 131}: {
		"ImageOrientation",
		"Image Orientation",
		TagFormatString, false, false, 1, 1,
	},
	{2, 135}: {
		"LanguageID",
		"Language Identifier",
		TagFormatString, false, false, 2, 3,
	},
	{2, 150}: {
		"AudioType",
		"Audio Type",
		TagFormatString, false, false, 2, 2,
	},
	{2, 151}: {
		"AudioSamplingRate",
		"Audio Sampling Rate",
		TagFormatString, false, false, 6, 6,
	},
	{2, 152}: {
		"AudioSamplingRes",
		"Audio Sampling Resolution",
		TagFormatString, false, false, 2, 2,
	},
	{2, 153}: {
		"AudioDuration",
		"Audio Duration",
		TagFormatString, false, false, 6, 6,
	},
	{2, 154}: {
		"AudioOutcue",
		"Audio Outcue",
		TagFormatString, false, false, 0, 64,
	},
	{2, 200}: {
		"PreviewFileFormat",
		"Preview File Format",
		TagFormatShort, false, false, 2, 2,
	},
	{2, 201}: {
		"PreviewFileFormatVer",
		"Preview File Format Version",
		TagFormatShort, false, false, 2, 2,
	},
	{2, 202}: {
		"PreviewData",
		"Preview Data",
		TagFormatBinary, false, false, 0, 256000,
	},
	{7, 10}: {
		"SizeMode",
		"Size Mode",
		TagFormatByte, true, false, 1, 1,
	},
	{7, 20}: {
		"MaxSubfileSize",
		"Max Subfile Size",
		TagFormatLong, true, false, 3, 4,
	},
	{7, 90}: {
		"ObjectSizeAnnounced",
		"Object Size Announced",
		TagFormatLong, false, false, 3, 4,
	},
	{7, 95}: {
		"MaxObjectSize",
		"Maximum Object Size",
		TagFormatLong, false, false, 3, 4,
	},
	{8, 10}: {
		"Subfile",
		"Subfile",
		TagFormatBinary, true, true, 0, math.MaxUint32,
	},
	{9, 10}: {
		"ConfirmedDataSize",
		"Confirmed Data Size",
		TagFormatLong, true, false, 3, 4,
	},
}

func GetTagInfo(key TagKey) (TagInfo, error) {
	info, infoFound := tagInfoMap[key]
	if !infoFound {
		return TagInfo{}, fmt.Errorf("unknown tag %d:%d", key.RecordID, key.TagID)
	}
	return info, nil
}

type TagValue struct {
	Format TagFormat
	Raw    []byte
}

func (v TagValue) Typecast() interface{} {
	switch v.Format {
	case TagFormatByte, TagFormatShort, TagFormatLong:
		return v.Int()
	case TagFormatBinary:
		return v.Raw
	default:
		return string(v.Raw)
	}
}

func (v TagValue) Int() int {
	switch len(v.Raw) {
	// Check zero data just in case
	case 0:
		return 0
	case 1:
		return int(v.Raw[0])
	case 2:
		return int(binary.BigEndian.Uint16(v.Raw))
	case 3:
		return int(binary.BigEndian.Uint16(v.Raw[:2]))<<8 + int(v.Raw[2])
	default:
		return int(binary.BigEndian.Uint32(v.Raw[:4]))
	}
}

func (v TagValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Typecast())
}
