package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
)

//problems with endianess
//work in big endian but input and output should be in little endian
//e.g. when reading hte hash read f first
//maybe consider using a []bytes instead of string for hash
//https://golang.org/src/encoding/binary/binary.go
//https://en.wikipedia.org/wiki/Endianness - use little endian
//https://stackoverflow.com/questions/34701187/go-byte-to-little-big-endian-signed-integer-or-float

const key = "abcdef"

func hashToByteArr(s string) []byte {
	//store string as little endian byte slice
	var bytes []byte
	for _, c := range s {
		bytes = append(bytes, byte(c))
	}
	fmt.Printf("bytes: %v\n", bytes)
	//returns correct value for blocks
	return bytes
}

func appendOgLen2bin(originalLength uint64, bytes []byte) []byte {
	//https://stackoverflow.com/questions/35371385/how-can-i-convert-an-int64-into-a-byte-array-in-go
	//should be little endian
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, originalLength)
	fmt.Printf("oglength: %d\nlength as buf: %v\n", originalLength, buf)
	bytes = append(bytes, buf...)
	return bytes
}

func padByteArr(bytes []byte, s string) []byte {
	originalLength := uint64(8 * len(s))
	bytes = append(bytes, 128)
	for ((8 * len(bytes)) % 512) != 448 {
		bytes = append(bytes, 0)
	}
	//appends the original length as a uint64 in little endian
	bytes = appendOgLen2bin(originalLength, bytes)

	return bytes
}

func splitByteArr(bytes []byte) (hashArr [][]uint32) {
	//floor function of no. of bits in bytes divided by 512 + 1
	no512Blocks := int((8 * len(bytes)) / 512)
	var bytesSlice []uint32
	for i := 0; i < no512Blocks; i++ {
		for j := 0; j < 16; j++ {
			pos := (i * 512) + (j * 4)
			bytesSlice = append(bytesSlice, binary.LittleEndian.Uint32(bytes[pos:pos+4]))
			//fmt.Printf("j: %d\nbytes slice: %v\ndecmial: %d\nbinary: %032b\n", j, bytes[pos:pos+4], binary.LittleEndian.Uint32(bytes[pos:pos+4]), binary.LittleEndian.Uint32(bytes[pos:pos+4]))
		}
		fmt.Printf("%d: %v\n", i, bytesSlice)
		hashArr = append(hashArr, bytesSlice)
	}
	fmt.Printf("hashArr: %v\n", hashArr)
	return hashArr
}

func initialiseTables() (s [64]uint32, k [64]uint32, g [64]uint32) {
	//k[i] = abs(sin(i + 1)) * 2^32
	var i uint32
	for i = 0; i < 64; i++ {
		k[i] = uint32(math.Abs(math.Sin(float64(i+1))) * math.Pow(2.0, 32.0))
		//create S
		if i < 16 {
			g[i] = i
		} else if i < 32 {
			g[i] = (5*i + 1) % 16
		} else if i < 48 {
			g[i] = (3*i + 5) % 16
		} else {
			g[i] = (7 * i) % 16
		}
	}
	for i := 0; i < 16; i++ {
		switch i % 4 {
		case 0:
			s[i] = 7
			s[i+16] = 5
			s[i+32] = 4
			s[i+48] = 6
		case 1:
			s[i] = 12
			s[i+16] = 9
			s[i+32] = 11
			s[i+48] = 10
		case 2:
			s[i] = 17
			s[i+16] = 14
			s[i+32] = 16
			s[i+48] = 15
		case 3:
			s[i] = 22
			s[i+16] = 20
			s[i+32] = 23
			s[i+48] = 21
		}

	}
	//fmt.Printf("%v\n", s)
	return s, k, g
}

func NOT(B uint32) uint32 { return 0xFFFFFFFF ^ B }
func logicFunction(i int, B uint32, C uint32, D uint32) uint32 {
	i = int(i / 16)
	switch i {
	case 0:
		return ((B & C) | ((NOT(B)) & D))
	case 1:
		return ((B & D) | C&NOT(D))
	case 2:
		return (B ^ C ^ D)
	case 3:
		return (C ^ (B | NOT(D)))
	default:
		fmt.Printf("Error in logic functions. Switch broken\n")
	}
	return 0
}

func leftRotate(A uint32, f uint32, k uint32, m uint32, c uint32) uint32 {
	//fmt.Printf("A: %d, f: %d, k: %d, m: %d\n", A, f, k, m)
	//	fmt.Printf("A: %d\nf: %d\nk[32]: %d\nm[g[32]]: %d\ns[i]: %d\n", A, f, k, m, c)
	x := A + f + k + m
	var Rint uint32 = ((x << c) | (x >> (32 - c)))
	//	fmt.Printf("x: %d\n", x)

	//	fmt.Printf("x: %032b\nx shifted: %032b\n\n", x, ((x << c) | (x >> (32 - c))))
	//	fmt.Printf("leftRotate: %d\n", Rint)

	return Rint
}

func mainHash(m []uint32, A uint32, B uint32, C uint32, D uint32, k [64]uint32, s [64]uint32, g [64]uint32) (uint32, uint32, uint32, uint32) {
	//doesn't line up. intialised correctly but breaks down here
	for i := 0; i < 64; i++ {
		//		fmt.Printf("i: %d\n", i)
		f := logicFunction(i, B, C, D)
		if f == 0 {
			fmt.Printf("An error has occured in logic function.\n")
			os.Exit(1)
		}

		lRotate := leftRotate(A, f, k[i], m[g[i]], s[i])

		A = D
		D = C
		C = B
		B = B + lRotate
		//x := []uint32{A, B, C, D}

		//		fmt.Printf("ai: %d\nA: %d\nB: %d\nC: %d\nD: %d\n\n", i, A, B, C, D)
		//prints in order that https://cse.unl.edu/~ssamal/crypto/genhash.php shows
		//fmt.Printf("A: %d\nB: %d\nC: %d\nD: %d\n\n", x[(i+1)%4], x[(i+2)%4], x[(i+3)%4], x[(i+4)%4])
		//32 breaks

	}
	return A, B, C, D
}

func MD5Hash(i int) []byte {
	hashInput := key + strconv.FormatInt(int64(i), 10)
	fmt.Printf("hash: %s\n", hashInput)
	hashBytes := hashToByteArr(hashInput)
	//pads our hash to 448 % 512 (512-64) characters
	hashBytes = padByteArr(hashBytes, hashInput)
	fmt.Printf("hashBytes: %v\n", hashBytes)
	hashTable := splitByteArr(hashBytes)
	a0, b0, c0, d0 := uint32(0x67452301), uint32(0xEFCDAB89), uint32(0x98BADCFE), uint32(0x10325476)
	s, k, g := initialiseTables()
	//fmt.Printf("a0: %d\nb0: %d\nc0: %d\nd0: %d\n\n", a0, b0, c0, d0)

	for _, hash := range hashTable {
		A := a0
		B := b0
		C := c0
		D := d0
		// don't really understand this. Makes the buffers all 64 bits long
		A, B, C, D = mainHash(hash, A, B, C, D, k, s, g)
		//A,B,C, and D are 64 bit when should be 32 bit
		a0 += A
		b0 += B
		c0 += C
		d0 += D
		//fmt.Printf("A: %d\nB: %d\nC: %d\nD: %d\n", A, B, C, D)
		fmt.Printf("a0: %d\nb0: %d\nc0: %d\nd0: %d\n", a0, b0, c0, d0)
	}
	//output in little endian
	//output := fmt.Sprintf("%032b%032b%032b%032b", d0, c0, b0, a0)

	//doesn't work. Might need to put into little endian
	outputSlice := []uint32{a0, b0, c0, d0}

	var output []byte
	for _, n := range outputSlice {
		var buf []byte
		//error here. Index out of range
		binary.LittleEndian.PutUint32(buf, n)
		output = append(output, buf...)
	}

	return output
}

func main() {
	fmt.Printf("\n")
	testHash := MD5Hash(1)
	encodedStr := hex.EncodeToString(testHash)
	fmt.Printf("hash: %v\n\n", encodedStr)

}
