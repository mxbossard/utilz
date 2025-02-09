package filez

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"mby.fr/utils/errorz"
)

func InitSeq(pathes ...string) (err error) {
	seqFilepath := filepath.Join(pathes...)
	err = os.WriteFile(seqFilepath, []byte("0"), 0600)
	if err != nil {
		err = fmt.Errorf("cannot initialize seq file (%s): %w", seqFilepath, err)
	}
	return
}

func IncrementSeq(pathes ...string) (seq uint32) {
	// return an increment for test indexing
	seqFilepath := filepath.Join(pathes...)

	file, err := os.OpenFile(seqFilepath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		errorz.Fatalf("cannot open seq file (%s) to increment: %s", seqFilepath, err)
	}
	defer file.Close()
	var strSeq string
	_, err = fmt.Fscanln(file, &strSeq)
	if err != nil && err != io.EOF {
		errorz.Fatalf("cannot read seq file (%s) to increment: %s", seqFilepath, err)
	}
	if strSeq == "" {
		seq = 0
	} else {
		var i int
		i, err = strconv.Atoi(strSeq)
		if err != nil {
			errorz.Fatalf("cannot convert seq file (%s) to an integer to increment: %s", seqFilepath, err)
		}
		seq = uint32(i)
	}

	newSec := seq + 1
	_, err = file.WriteAt([]byte(fmt.Sprint(newSec)), 0)
	if err != nil {
		errorz.Fatalf("cannot write seq file (%s) to increment: %s", seqFilepath, err)
	}

	//fmt.Printf("Incremented seq(%s %s %s): %d => %d\n", testSuite, token, filename, seq, newSec)
	return newSec
}

func ReadSeq(pathes ...string) (c uint32) {
	// return the count of run test
	seqFilepath := filepath.Join(pathes...)

	file, err := os.OpenFile(seqFilepath, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		errorz.Fatalf("cannot open seq file (%s) to read: %s", seqFilepath, err)
	}
	defer file.Close()
	var strSeq string
	_, err = fmt.Fscanln(file, &strSeq)
	if err != nil {
		if err == io.EOF {
			return 0
		}
		errorz.Fatalf("cannot read seq file (%s) to read: %s", seqFilepath, err)
	}
	var i int
	i, err = strconv.Atoi(strSeq)
	if err != nil {
		errorz.Fatalf("cannot convert seq file (%s) as an integer to read: %s", seqFilepath, err)
	}
	c = uint32(i)
	return
}
