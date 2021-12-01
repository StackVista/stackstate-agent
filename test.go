package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"strings"
)

func stringSeedToUInt64(inputValue string) (uint64, error) {
	h := md5.New()
	if _, err := io.WriteString(h, inputValue); err != nil {
		return 0, err
	}

	var seed = binary.BigEndian.Uint64(h.Sum(nil))
	rand.Seed(int64(seed))
	return rand.Uint64(), nil
}

func convertStringToUint64(input string) (*uint64, error) {
	// https://www.geeksforgeeks.org/rune-in-golang/
	// https://yourbasic.org/golang/convert-string-to-rune-slice/
	// https://yourbasic.org/golang/build-append-concatenate-strings-efficiently/

	// Int Ascii values representing the string values
	runes := []rune(input)

	// We use string builder for performance
	// The actual integer values are written to a string to create a massive big itn
	var sb strings.Builder
	var stringBuilderError error

	for _, r := range runes {
		runeIntToString := fmt.Sprintf("%v", r)
		_, err := sb.WriteString(runeIntToString)
		if err != nil {
			stringBuilderError = err
			break
		}
	}

	if stringBuilderError != nil {
		return nil, stringBuilderError
	}

	runeString := sb.String()

	// Convert a massive number to an uint64 representation
	bigInt := big.NewInt(0)
	bigInt.SetString(runeString, 10)
	var uint64Representation = bigInt.Uint64()

	return &uint64Representation, nil
}

func main() {
	resultA, _ := convertStringToUint64("!\"#$%&\\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\\\]^_`abcdefghijklmnopqrstuvwxyz{|}~")
	fmt.Println(*resultA == 337105592495713734)

	resultB, _ := convertStringToUint64("YZ0T8B2Ll8IIzMv3EfFIqQ==")
	fmt.Println(*resultB == 9702618976578468961)

	resultC, _ := convertStringToUint64("yjXK+2eLD+s=")
	fmt.Println(*resultC == 5327617794819146761)

	resultD, _ := convertStringToUint64("Y3OrG+/srMM=")
	fmt.Println(*resultD == 864179905714572449)
}















