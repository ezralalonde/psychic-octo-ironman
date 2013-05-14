package main

import (
  "compress/flate"
	"encoding/binary"
	"fmt"
	"os"
)

type Head struct {
	Name           [32]byte //0
	DbAttributes   uint16   //32
	FileVersion    uint16   //34
	CreationDate   uint32   //36
	ModifyDate     uint32   //40
	BackupDate     uint32   //44
	ModifyNo       uint32   //48
	AppInfoOffet   uint32   //52
	SortInfoOffset uint32   //56
	Type           [4]byte  //60
	Creator        [4]byte  //64
	UniqueSeed     uint32   //68
	ExpectedZero   uint32   //72
	SectionCount   uint16   //76
}

func main() {
	file, _ := os.Open("file.mobi")
	defer file.Close()
	rdr := flate.NewReader(file)
	defer rdr.Close()
	var header Head

	err := binary.Read(file, binary.BigEndian, &header)
	check(err)
	fmt.Printf("%#v\n", header)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
