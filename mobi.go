package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type PDFormat struct {
	Name             [32]byte //start at byte 0
	Attributes       uint16   //32
	Version          uint16   //34
	CreationDate     uint32   //36
	ModifyDate       uint32   //40
	BackupDate       uint32   //44
	ModifyNumber     uint32   //48
	AppInfoID        uint32   //52
	SortInfoID       uint32   //56
	Type             [4]byte  //60
	Creator          [4]byte  //64
	UniqueIDSeed     uint32   //68
	NextRecordListID uint32   //72
	SectionCount     uint16   //76-78
}

type PDRecordInfoSection struct {
	DataOffset uint32  //starts at byte 0
	Attributes byte    //4
	UniqueID   [3]byte //5-8
}

type PDHeader struct {
	CompressionType uint16 //starts at byte 0
	_               uint16 //2 (always zero?)
	TextLength      uint32 //4
	RecordCount     uint16 //8
	RecordSize      uint16 //10
	CurrentPosition uint32 //12-16
}

type Mobi8Header struct {
	//note that since the values are being read
	//by bytes.Read, we can't use "_" as a name,
	//or have any unexported fields.
	//This is stupid, but whatever.
	//We use the name "SkipX" instead.
	CompressionType     uint16   //start at byte 0
	Skip1               uint16   //2 (always zero?)
	TextLength          uint32   //4
	RecordCount         uint16   //8
	RecordSize          uint16   //10
	CryptoType          uint16   //12
	Skip2               uint16   //14 filler
	Identifier          [4]byte  //16
	HeaderLength        uint32   //20
	Type                uint32   //24
	TextEncoding        uint32   //28
	UniqueID            uint32   //32
	Version             uint32   //36
	OrtographicIndex    uint32   //40
	IncflectionIndex    uint32   //44
	IndexNames          uint32   //48
	IndexKeys           uint32   //52
	Extra               [24]byte //56
	FirstNontext        uint32   //80
	TitleOffset         uint32   //84
	TitleLength         uint32   //88
	Locale              uint32   //92
	InputLanguage       uint32   //96
	OutputLanguage      uint32   //100
	MinVersion          uint32   //104
	FirstImageOffset    uint32   //108
	HuffmanRecordOffset uint32   //112
	HuffmanRecordCount  uint32   //116
	HuffmanTableOffset  uint32   //120
	HuffTableLength     uint32   //124
	ExthFlags           uint32   //128
	Skip3               [32]byte //132
	Unknown0            uint32   //164
	DrmOffset           uint32   //168
	DrmCount            uint32   //172
	DrmSize             uint32   //176
	DrmFlags            uint32   //180
	Skip4               [8]byte  //184
	FirstContentNumber  uint16   //192
	LastContentNumber   uint16   //194
	FdstFlowCount       uint32   //196
	FcisOffset          uint32   //200
	FcisCount           uint32   //204
	FlisOffset          uint32   //208
	FlisCount           uint32   //212
	Skip5               [8]byte  //216
	SrcsOffset          uint32   //224
	SrcsCount           uint32   //228
	Skip6               [8]byte  //232
	TrailDataFlags      uint32   //240
	NcxIndex            uint32   //244-248
}

func GetStruct(file *os.File, hd interface{}, length int, offset int64) (rd int, err error) {
	b := make([]byte, length)
	rd, err = file.ReadAt(b, offset)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	err = binary.Read(buf, binary.BigEndian, hd)
	return
}

type ExthHeader struct {
	Identifier  [4]byte //starts at byte 0
	Length      uint32  //4
	RecordCount uint32  //8-12
}

type ExthData struct {
	Header  ExthHeader
	Records []ExthRecord
}

type ExthRecordInfo struct {
	Type   uint32 //starts at 0
	Length uint32 //4-8
}

type ExthRecordData []byte

type ExthRecord struct {
	Info ExthRecordInfo
	Data ExthRecordData
}

type FileHeader struct {
	Format            PDFormat
	Sections          []PDRecordInfoSection
	MobiHeader        Mobi8Header
	Fcis              FcisRecord
	Flis              FlisRecord
	Eof               EofRecord
	Exth              ExthData
	RawContents       []ContentRecord
	DecryptedContents []ContentRecord
}

type ContentRecord []byte

func (cr *ContentRecord) Decode(in []byte) (rd int, err error) {
	for p := 0; p < len(in); {
		c := in[p]
		p += 1
		if c >= 1 && c <= 8 {
			for ii := 0; ii < int(c); ii++ {
				if p < len(in) {
					*cr = append(*cr, in[p])
					p += 1
				}
			}
		} else if c < 128 {
			*cr = append(*cr, c)
		} else if c >= 192 {
			*cr = append(*cr, byte(' '))
			*cr = append(*cr, c^128)
		} else {
			if p < len(in) {
				d := uint16(c)<<8 + uint16(in[p])
				p += 1
				m := uint16(d>>3) & 0x7ff
				n := uint16(d&7) + 3
				length := len(*cr)
				for ii := uint16(0); ii < n; ii++ {
					if length-int(m-ii) >= 0 && length-int(m-ii) < len(*cr) {
						one := []byte(*cr)[length-int(m-ii)]
						*cr = append(*cr, one)
					}
				}
			}
		}
	}
	return
}

type FlisRecord struct {
	Identifier [4]byte //starts at 0, "FLIS"
	Skip1      uint32  //4            8
	Skip2      uint16  //8            65
	Skip3      uint16  //10           0
	Skip4      uint32  //12           0
	Skip5      uint32  //16          -1
	Skip6      uint16  //20           1
	Skip7      uint16  //22           3
	Skip8      uint32  //24           3
	Skip9      uint32  //28           1
	Skip10     uint32  //32-36       -1
}

type FcisRecord struct {
	Identifier [4]byte //starts at 0, to 4 "FCIS"
	Skip1      uint32  //4     20
	Skip2      uint32  //8     16
	Skip3      uint32  //12    1
	Skip4      uint32  //16    0
	TextLength uint32  //20    Same as PDHeader
	Skip5      uint32  //24    0
	Skip6      uint32  //28    32
	Skip7      uint32  //32    8
	Skip8      uint16  //36    1
	Skip9      uint16  //38    1
	Skip10     uint32  //40-44 0
}

type EofRecord struct {
	Skip1 uint8 //starts at 0, to 1  233
	Skip2 uint8 //1                  142
	Skip3 uint8 //2                   13
	Skip4 uint8 //3-4                 10
}

func main() {
	hd, err := GetFileHeader("file.mobi")
	check(err)
	//	fmt.Printf("%#v\n", hd.Format)
	//	fmt.Printf("%#v %#v\n", hd.Sections[0], hd.Sections[181])
	//	fmt.Printf("%#v\n", hd.MobiHeader)
	//	fmt.Printf("%#v\n", hd.Fcis)
	//	fmt.Printf("%#v\n", hd.Flis)
	//	fmt.Printf("%#v\n", hd.Eof)
	//	fmt.Printf("%#v\n", hd.Exth.Header)
	//	fmt.Printf("%#v\n", hd.Exth.Header.Identifier)
	//	fmt.Printf("%#v %#v\n", string(hd.Exth.Records[0].Data), string(hd.Exth.Records[16].Data))
	//	fmt.Printf("%#v\n", string(hd.Exth.Records[len(hd.Exth.Records)-1].Data))
	//	fmt.Printf("%#v\n", hd.MobiHeader.FirstContentNumber)
	//	fmt.Printf("%#v\n", hd.MobiHeader.LastContentNumber)
	//	fmt.Printf("%#v\n", len(hd.RawContents))
	//	fmt.Printf("%#v\n", len(hd.RawContents[0]))
	//	fmt.Printf("%#v\n", len(hd.RawContents[len(hd.RawContents)-1]))
	fmt.Printf("%v\n", string(hd.DecryptedContents[125]))
	fmt.Printf("%#v\n", len(hd.DecryptedContents[125]))
	WriteContentsToFile("./testout.html", hd.DecryptedContents)
}

func WriteContentsToFile(path string, content []ContentRecord) (err error) {
	file, err := os.Create(path)
	check(err)
	for ii := 0; ii < len(content); ii++ {
		file.Write(content[ii])
	}
    return
}

//GetPDRecordInfoSectionList reads `count` items from `file`,
//starting at byte `offset` and placing the result in in `ris`.
//Returns the number of records read, and any error.
func GetPDRecordInfoSectionList(file *os.File, ris *[]PDRecordInfoSection, count int, start int) (ii int, err error) {
	for ii = 0; ii < count; ii++ {
		var section PDRecordInfoSection
		_, err = GetStruct(file, &section, 8, int64(start+ii*8))
		if err != nil {
			return
		}
		*ris = append(*ris, section)
	}
	return
}

//GetFileHeader reads the header information from the file at path `path`.
//Returns the FileHeader as read, and any error.
func GetFileHeader(path string) (hd FileHeader, err error) {
	file, err := os.Open(path)
	defer file.Close()
	start := 0
	if err != nil {
		return
	}

	start, err = GetStruct(file, &hd.Format, 78, 0)
	fmt.Println(start)
	if err != nil {
		return
	}

	bytesRead, err := GetPDRecordInfoSectionList(file, &hd.Sections, int(hd.Format.SectionCount), start)
	fmt.Println(bytesRead)
	if err != nil {
		return
	}
	offset := int64(hd.Sections[0].DataOffset)
	bytesRead, err = GetStruct(file, &hd.MobiHeader, 248, offset)
	offset += int64(bytesRead)

	_, err = GetContents(file, &hd.RawContents, hd.Sections, hd.MobiHeader.FirstContentNumber, hd.MobiHeader.LastContentNumber)
	if err != nil {
		return
	}

	if length := len(hd.RawContents); length > 0 {
		for ii := uint16(0); ii < hd.MobiHeader.RecordCount; ii++ {
			var out ContentRecord
			out.Decode(hd.RawContents[ii])
			hd.DecryptedContents = append(hd.DecryptedContents, out)
		}
	}

	if hd.MobiHeader.ExthFlags&64 == 64 {
		_, err = GetExthData(file, &hd.Exth, offset)
	}

	if hd.MobiHeader.FcisCount > 0 {
		offset := int64(hd.Sections[hd.MobiHeader.FcisOffset].DataOffset)
		bytesRead, err = GetStruct(file, &hd.Fcis, 44, offset)
		fmt.Println("FCIS", bytesRead, err)
	}

	if hd.MobiHeader.FlisCount > 0 {
		offset := int64(hd.Sections[hd.MobiHeader.FlisOffset].DataOffset)
		bytesRead, err = GetStruct(file, &hd.Flis, 36, offset)
		fmt.Println("FLIS", bytesRead, err)
	}

	offset = int64(hd.Sections[len(hd.Sections)-1].DataOffset)
	bytesRead, err = GetStruct(file, &hd.Eof, 4, offset)
	fmt.Println("EOF", bytesRead, err)

	return
}

//GetContents reads contents of `file` into `crs`, with record bounds being
//the difference off the DataOffsets in `sections`[first : last + 1]
//Returns the number of content records read, and the error.
func GetContents(file *os.File, crs *[]ContentRecord, sections []PDRecordInfoSection, first, last uint16) (count int, err error) {
	count = 0
	for ii := first; ii <= last; ii++ {
		offset := int64(sections[ii].DataOffset)
		length := int(sections[ii+1].DataOffset - uint32(offset))
		cr := make(ContentRecord, length)
		bytesRead, err := GetStruct(file, &cr, length, offset)
		if bytesRead != length {
			// do something?
		}
		if err != nil {
			// do something?
		}
		*crs = append(*crs, cr)
		count++
	}
	return
}

//GetExthData reads exth data at index `offset` in `file`, and copies
//the contents into into `ex`. Offset is relative to the start of the file,
//not the start of the record.
//Return the total number of bytes read and the error.
func GetExthData(file *os.File, ex *ExthData, offset int64) (total int, err error) {
	read, err := GetStruct(file, &ex.Header, 12, offset)
	total += read
	for ii := uint32(0); ii < ex.Header.RecordCount; ii++ {
		var record ExthRecord
		read, err = GetStruct(file, &record.Info, 8, offset+int64(total))
		total += read
		record.Data = make([]byte, int(record.Info.Length-8))
		read, err = GetStruct(file, &record.Data, len(record.Data), offset+int64(total))
		total += read
		ex.Records = append(ex.Records, record)
	}
	return
}

//check panics when there's an error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
