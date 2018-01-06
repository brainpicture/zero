package zero

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// J joins whatever passed to an string
func J(a ...interface{}) string {
	out := ""
	for _, arg := range a {
		switch t := arg.(type) {
		case int:
			out += strconv.Itoa(t)
		case int64:
			out += fmt.Sprintf("%d", t)
		case uint64:
			out += fmt.Sprintf("%d", t)
		case float64:
			out += fmt.Sprintf("%f", t)
		case string:
			out += t
		case bool:
			if t {
				out += "true"
			} else {
				out += "false"
			}
		case []byte:
			out += string(out)
		}
	}
	return out
}

// I converts string to int without errors
func I(a string) int {
	res, _ := strconv.Atoi(a)
	//res, _ := strconv.ParseInt(a, 10, 64)
	return res
}

// I64 converts string to int64 without errors
func I64(a string) int64 {
	res, _ := strconv.ParseInt(a, 10, 64)
	return res
}

// Now returns unixtime
func Now() int {
	return int(time.Now().Unix())
}

// Split divide string into slice by an of chars from splits string
func Split(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

// IsInt checks if string is integer and returns bool
func IsInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// SplitToInts splits any string to slice of int using seporator
func SplitToInts(str string, sep string) (res []int) {
	chunks := strings.Split(str, sep)

	for _, chunk := range chunks {
		num, err := strconv.Atoi(chunk)
		if err != nil {
			continue
		}
		res = append(res, num)
	}
	return res
}

// Log prints any objects
func Log(objs ...interface{}) {
	format := ""
	for i := 0; i < len(objs); i++ {
		format += "%+v "
	}

	fmt.Printf(format+"\n", objs...)
}

// ParsePath returns path out of an url
func ParsePath(path string) string {
	path = strings.Replace(path, "//", "/", -1)
	pathParts := strings.Split(path, "?")
	path = pathParts[0]
	return path
}

// Base64UrlEncode encode string to base64
func Base64UrlEncode(data string) string {
	str := base64.StdEncoding.EncodeToString([]byte(data))
	str = strings.Replace(str, "+", "-", -1)
	str = strings.Replace(str, "/", "_", -1)
	str = strings.Replace(str, "=", "", -1)
	return str
}

// Base64UrlDecode decode string from base64
func Base64UrlDecode(str string) (string, error) {
	str = strings.Replace(str, "-", "+", -1)
	str = strings.Replace(str, "_", "/", -1)
	for len(str)%4 != 0 {
		str += "="
	}
	bytes, err := base64.StdEncoding.DecodeString(str)
	return string(bytes), err
}

// MD5 return string representation of md5
func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Reverse returns any slice in reverse order
func Reverse(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// SortByValueInt sorts any map[string]int by value
func SortByValueInt(m map[string]int) {
	n := map[int][]string{}
	var a []int
	for k, v := range m {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(a)))

	for _, k := range a {
		for _, s := range n[k] {

			fmt.Printf("%s, %d\n", s, k)
		}
	}
}
