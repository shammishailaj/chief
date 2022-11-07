package compress

import (
	"github.com/mholt/archiver"
	"log"
)

// CompressTBZ2 - a function to compress one or more file(s) into a compressed TBZ2 (.tzr.bz2) archive
func CompressTBZ2(sources []string, destination string) error {
	a := archiver.NewTarBz2()
	a.CompressionLevel = 1 // This is the best compression-level for BZ2 format
	err := a.Archive(sources, destination)
	if err != nil {
		log.Printf("Error archiving file(s) %#v into %s. Details: %s", sources, destination, err.Error())
	} else {
		log.Printf("Successfully archived %#v into file %s", sources, destination)
	}
	return err
}

// ExtractTBZ2 - a function to compress one or more file(s) into a compressed TBZ2 (.tzr.bz2) archive
func ExtractTBZ2(source string, destination string) error {
	a := archiver.NewTarBz2()
	err := a.Unarchive(source, destination)
	if err != nil {
		log.Printf("Error Decompressing file: %s into %s. Reason: %s", source, destination, err.Error())
	}
	return err
}
