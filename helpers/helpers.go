package helpers

import (
	"fmt"
	"strings"
	"os"
	"io"
	"hash/fnv"
	"bufio"
)

func CopyFile(dst, src string) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("error opening src file %s: %s", dst, err)
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("error opening dst file %s: %s", dst, err)
	}
	defer dstFile.Close()
	n, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return 0, fmt.Errorf("error in copy from %s to %s: %s", src, dst, err)
	}
	return n, nil
}

func GetDomain(url string) (string) {
	arr := strings.Split(url, "/")
	if len(arr) < 3 { // didnt start with http://
		return url
	}
	arr = strings.Split(arr[2], ".") // split host by .
	if len(arr) < 3 {
		return strings.Join(arr, ".")
	}
	subArr := make([]string, 0)
	if arr[len(arr) - 2] == "co" || arr[len(arr) - 2] == "com" {
		subArr = append(subArr, arr[len(arr) - 3])
	}
	subArr = append(subArr, arr[len(arr) - 2], arr[len(arr) - 1])
	return strings.Join(subArr, ".")
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func ReadInput(messages chan string) {
	fmt.Println("Type q to quit")
	bio := bufio.NewReader(os.Stdin)
	for {
		input, _, err := bio.ReadLine()
		if err != nil { panic(err) }
		messages <- string(input)
	}
}
