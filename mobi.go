package main

import (
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
	SectionCount     uint16   //76
}

func main() {
	pd, err := GetPDFormat("file.mobi")
	fmt.Printf("%#v\n", pd, err)
}

func GetPDFormat(filename string) (pd PDFormat, err error) {
	file, err := os.Open(filename)
    if err != nil {
        return
    }
	defer file.Close()
	rdr := flate.NewReader(file)
	defer rdr.Close()
	err = binary.Read(file, binary.BigEndian, &pd)
    return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
