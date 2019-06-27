package zero

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// H allow to describe custom json easily
type H map[string]interface{}

// S allow to describe string hashmaps easily
type S map[string]string

// V allow to pass list of elements of any type
type V []interface{}

// KV represents an key value array
type KV struct {
	Key   string
	Value string
}

var reURL *regexp.Regexp

// String return field as string
func (h H) String(field string) string {
	val, ok := h[field]
	if ok {
		return J(val)
	}
	return ""
}

// Int return field as int
func (h H) Int(field string) int {
	val, ok := h[field]
	if ok {
		res, ok := val.(int)
		if ok {
			return res
		}
	}
	return 0
}

// Int64 return field as int64
func (h H) Int64(field string) int64 {
	val, ok := h[field]
	if ok {
		res, ok := val.(int64)
		if ok {
			return res
		}
	}
	return int64(0)
}

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
			out += string(t)
		case byte:
			out += string([]byte{t})
		default:
			bytes, _ := json.Marshal(arg)
			out += string(bytes)
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

// UI64 converts string to int64 without errors
func UI64(a string) uint64 {
	res, _ := strconv.ParseUint(a, 10, 64)
	return res
}

// Now returns unixtime
func Now() int64 {
	return time.Now().Unix()
}

// NowNano returns unixnano time
func NowNano() int64 {
	return time.Now().UnixNano()
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

// SplitToInt splits any string to slice of int using separator
func SplitToInt(str string, sep string) (res []int) {
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

// SplitToInt64 splits any string to slice of int using separator
func SplitToInt64(str string, sep string) (res []int64) {
	chunks := strings.Split(str, sep)

	for _, chunk := range chunks {
		num, err := strconv.ParseInt(chunk, 10, 0)
		if err != nil {
			continue
		}
		res = append(res, num)
	}
	return res
}

// SplitDoubleInt64 return two int64 using
func SplitDoubleInt64(str string, sep string) (int64, int64) {
	n1 := int64(0)
	n2 := int64(0)
	chunks := strings.Split(str, sep)
	if len(chunks) > 0 {
		n1 = I64(chunks[0])
		if len(chunks) > 1 {
			n2 = I64(chunks[1])
		}
	}
	return n1, n2
}

// SplitDoubleString return two strings with splitting
func SplitDoubleString(str string, sep string) (n1 string, n2 string) {
	chunks := strings.SplitN(str, sep, 2)
	if len(chunks) > 0 {
		n1 = chunks[0]
		if len(chunks) > 1 {
			n2 = chunks[1]
		}
	}
	return n1, n2
}

// SplitTrippleString return three string
func SplitTrippleString(str string, sep string) (n1, n2, n3 string) {
	chunks := strings.SplitN(str, sep, 3)
	if len(chunks) > 0 {
		n1 = chunks[0]
		if len(chunks) > 1 {
			n2 = chunks[1]
		}
		if len(chunks) > 2 {
			n3 = chunks[2]
		}
	}
	return n1, n2, n3
}

// SplitTrippleInt64 return three int64 by splitting sting
func SplitTrippleInt64(str string, sep string) (n1, n2, n3 int64) {
	chunks := strings.SplitN(str, sep, 3)
	if len(chunks) > 0 {
		n1 = I64(chunks[0])
		if len(chunks) > 1 {
			n2 = I64(chunks[1])
		}
		if len(chunks) > 2 {
			n3 = I64(chunks[2])
		}
	}
	return n1, n2, n3
}

// SplitDoubleInt return two int using
func SplitDoubleInt(str string, sep string) (int, int) {
	n1 := 0
	n2 := 0
	chunks := strings.Split(str, sep)
	if len(chunks) > 0 {
		n1 = I(chunks[0])
		if len(chunks) > 1 {
			n2 = I(chunks[1])
		}
	}
	return n1, n2
}

// SplitIntString return int and string
func SplitIntString(str string, sep string) (int, string) {
	n1 := 0
	n2 := ""
	chunks := strings.SplitN(str, sep, 2)
	if len(chunks) > 0 {
		n1 = I(chunks[0])
		if len(chunks) > 1 {
			n2 = chunks[1]
		}
	}
	return n1, n2
}

// SplitInt64String return int64 and string
func SplitInt64String(str string, sep string) (int64, string) {
	n1 := int64(0)
	n2 := ""
	chunks := strings.SplitN(str, sep, 2)
	if len(chunks) > 0 {
		n1 = I64(chunks[0])
		if len(chunks) > 1 {
			n2 = chunks[1]
		}
	}
	return n1, n2
}

// Log prints any objects
func Log(objs ...interface{}) {
	format := ""
	for i := 0; i < len(objs); i++ {
		format += "%+v "
	}

	fmt.Printf(format+"\n", objs...)
}

// Ok prints any objects
func Ok(objs ...interface{}) {
	os.Stdout.WriteString("\x1b[92m")
	for i := 0; i < len(objs); i++ {
		out := fmt.Sprintf("%+v", objs[i])
		if i != 0 {
			out = " " + out
		}
		os.Stdout.WriteString(out)
	}
	os.Stdout.WriteString("\x1b[0m\n")
}

// Err prints any objects
func Err(objs ...interface{}) {
	os.Stdout.WriteString("\x1b[91m")
	for i := 0; i < len(objs); i++ {
		out := fmt.Sprintf("%+v", objs[i])
		if i != 0 {
			out = " " + out
		}
		os.Stdout.WriteString(out)
	}
	os.Stdout.WriteString("\x1b[0m\n")
}

// LogJSON will show the json representation of logged content
func LogJSON(objs ...interface{}) {
	jsons := []string{}
	for i := 0; i < len(objs); i++ {
		json, _ := json.Marshal(objs[i])
		jsons = append(jsons, string(json))
	}

	fmt.Println(strings.Join(jsons, " "))
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

// Trim checks string is not longer than provided length and shorted it with .. postfix
func Trim(str string, length int) string {
	chars := []rune(str)
	if len(chars) > length {
		return string(chars[:length-2]) + ".."
	}
	return str
}

// DeleteFromStrings removes element from items
func DeleteFromStrings(items []string, toRemove string) []string {
	for k, v := range items {
		if v == toRemove {
			return append(items[:k], items[k+1:]...)
		}
	}
	return items
}

// DeleteFromInt64 removes element from items
func DeleteFromInt64(items []int64, toRemove int64) []int64 {
	for k, v := range items {
		if v == toRemove {
			return append(items[:k], items[k+1:]...)
		}
	}
	return items
}

// OneOf check is element presented inside list
func OneOf(el string, items ...string) bool {
	for _, v := range items {
		if el == v {
			return true
		}
	}
	return false
}

// Plot2D is an basic struct for 2d plot representation
type Plot2D struct {
	Labels []string `json:"labels"`
	Points []int64  `json:"points"`
}

// AlphabetEncodingSymbols is list of symbols you can use to encode different strings
const AlphabetEncodingSymbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// I64Str converts int to short string with numbers and letters
func I64Str(val int64) string {
	return EncodeInt64(val, AlphabetEncodingSymbols)
}

// StrI64 converts int to short string with numbers and letters
func StrI64(val string) int64 {
	return DecodeInt64(val, AlphabetEncodingSymbols)
}

// EncodeInt64 encodes int using passed symbols
func EncodeInt64(val int64, symbols string) string {
	l := int64(len(symbols))
	link := ""

	for val >= 1 {
		link += string(symbols[int(val%l)])
		val /= l
	}
	return link
}

// DecodeInt64 decodes int using passed symbols
func DecodeInt64(link string, symbols string) int64 {
	val := int64(0)
	l := int64(len(symbols))
	pow := int64(1)
	for _, symbol := range link {
		i := strings.Index(symbols, string(symbol))
		val += int64(i) * pow
		pow *= l
	}
	return val
}

// RandomHex will generate random hex string
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Parallel do several tasks and return result, no more than 100 params supported
func Parallel(tasks ...func() H) (result []H) {
	num := len(tasks)
	resultChan := make(chan H, num) // no more than 100 params supported
	for _, task := range tasks {
		go func(t func() H) {
			resultChan <- t()
		}(task)
	}
	for i := 0; i < num; i++ {
		result = append(result, <-resultChan)
	}
	return
}

func init() {
	reURL = regexp.MustCompile(`((http|ftp|https):\/\/)?([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
}
