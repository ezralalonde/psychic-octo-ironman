package main

import (
	"bytes"
	"compress/flate"
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

type Header struct {
	Format   PDFormat
	Sections []PDRecordInfoSection
}

func main() {
	hd, err := GetHeader("file.mobi")
	check(err)
	fmt.Printf("%#v\n", hd.Format)
	fmt.Printf("%v %v\n", hd.Sections[0], hd.Sections[181])
}

//GetPDRecordInfoSection reads the Record Info Section of `file` into
//`section`, starting data at offset `index`.
//Returns the number of bytes read and any error.
func GetPDRecordInfoSection(file *os.File, section *PDRecordInfoSection, index int) (rd int, err error) {
	b := make([]byte, 8)
	rd, err = file.ReadAt(b, int64(index))
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	err = binary.Read(buf, binary.BigEndian, section)
	return
}

//GetPDRecordInfoSectionList reads `count` items from `file`,
//starting at byte `offset` and placing the result in in `ris`.
//Returns the number of records read, and any error.
func GetPDRecordInfoSectionList(file *os.File, ris *[]PDRecordInfoSection, count int, start int) (ii int, err error) {
	for ii = 0; ii < count; ii++ {
		var section PDRecordInfoSection
		_, err = GetPDRecordInfoSection(file, &section, start+ii*8)
		*ris = append(*ris, section)
	}
	return
}

//GetPDFormat reads the PDFormat from the first 78 bytes of `file`.
//Returns the number of bytes read, and any error.
func GetPDFormat(file *os.File, pd *PDFormat) (rd int, err error) {
	b := make([]byte, 78)
	rd, err = file.ReadAt(b, int64(rd))
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	err = binary.Read(buf, binary.BigEndian, pd)
	return
}

//GetHeader reads the header information from the file at path `path`.
//Returns the Header as read, and any error.
func GetHeader(path string) (hd Header, err error) {
	file, err := os.Open(path)
	defer file.Close()
	start := 0
	if err != nil {
		return
	}

	start, err = GetPDFormat(file, &hd.Format)
	if err != nil {
		return
	}

	rdr := flate.NewReader(file)
	defer rdr.Close()

	_, err = GetPDRecordInfoSectionList(file, &hd.Sections, int(hd.Format.SectionCount), start)
	fmt.Println(start)
	return
}

//check helps panic when there's an error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
