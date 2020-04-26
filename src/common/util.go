package common

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	chars = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
		"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
		"Y", "Z", "~", "!", "@", "#", "$", "%", "^", "&",
		"*", "(", ")", "-", "_", "=", "+", "[", "]", "{",
		"}", "|", "<", ">", "?", "/", ".", ",", ";", ":"}

	numberChars = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
)

func Shuffle2(a []string) []string {
	i := len(a) - 1
	for i > 0 {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		i--
	}
	return a
}

func GetToken(n int) string {
	if n < 1 {
		return ""
	}
	var tokens []string
	for i := 0; i < n; i++ {
		tokens = append(tokens, chars[rand.Intn(90)]) // 90 是 Chars 的长度
	}
	return strings.Join(tokens, "")
}

// id 的第一位从 1 开始
func GetID(n int) int {
	if n < 1 {
		return -1
	}
	min := math.Pow10(n - 1)
	id := int(min) + rand.Intn(int(math.Pow10(n)-min))
	return id
}

func HttpPost(url string, data string) ([]byte, error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		return body, nil
	}
	return nil, err
}
func Atoi(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

var todayCode = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C",
	"D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
	"Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

func GetTodayCode(n int) string {
	newWords := ""
	for i := 0; i < n; i++ {
		newWords += todayCode[rand.Intn(len(todayCode))]
	}
	return newWords
}

func OneDay0ClockTimestamp(t time.Time) int64 {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}