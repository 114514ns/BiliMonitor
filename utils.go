package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	mathRand "math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}
func formatTime(input string) string {
	if input == "0000-00-00 00:00:00" {
		return "Invalid Date"
	}

	// Define layout compatible with the input
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, input)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return "Parsing Error"
	}
	return t.Format(layout)
}
func Last(dir string) (fileName string, modTime time.Time, err error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return "", time.Time{}, err
	}

	var onlyFlvFiles []os.DirEntry
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".flv") {
			onlyFlvFiles = append(onlyFlvFiles, entry)
		}
	}
	if len(onlyFlvFiles) == 0 {
		return "", time.Time{}, fmt.Errorf("no .flv files found in the directory: %s", dir)
	}

	sort.Slice(onlyFlvFiles, func(i, j int) bool {
		infoI, _ := onlyFlvFiles[i].Info()
		infoJ, _ := onlyFlvFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})
	latestFile := onlyFlvFiles[0]
	info, err := latestFile.Info()
	if err != nil {
		return "", time.Time{}, err
	}
	return latestFile.Name(), info.ModTime(), nil
}
func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	secs := (duration % time.Minute) / time.Second

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
func abs(a int) int {
	if a < 0 {
		return -a
	} else {
		return a
	}
}
func toInt64(s string) int64 {
	i64, _ := strconv.ParseInt(s, 10, 64)
	return i64
}
func toInt(s string) int {
	i64, _ := strconv.ParseInt(s, 10, 64)
	return int(i64)
}
func AppendElement[T any](queue []T, maxSize int, element T) []T {
	if len(queue) >= maxSize {
		queue = queue[1:]
	}
	return append(queue, element)
}
func DeepCopy[T any](src T) (T, error) {
	var dst T
	data, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}
	err = json.Unmarshal(data, &dst)
	return dst, err
}
func getCorrespondPath(ts int64) string {
	const publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDLgd2OAkcGVtoE3ThUREbio0Eg
Uc/prcajMKXvkCKFCWhJYJcLkcM2DKKcSeFpD/j6Boy538YXnR6VhcuUJOhH2x71
nzPjfdTcqMz7djHum0qSZA0AyCBDABUqCrfNgCiJ00Ra7GmRj+YCK1NJEuewlb40
JNrRuoEUXpabUzGB8QIDAQAB
-----END PUBLIC KEY-----
`
	pubKeyBlock, _ := pem.Decode([]byte(publicKeyPEM))
	hash := sha256.New()
	random := rand.Reader
	msg := []byte(fmt.Sprintf("refresh_%d", ts))
	var pub *rsa.PublicKey
	pubInterface, parseErr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	if parseErr != nil {
		return ""
	}
	pub = pubInterface.(*rsa.PublicKey)
	encryptedData, encryptErr := rsa.EncryptOAEP(hash, random, pub, msg, nil)
	if encryptErr != nil {
		return ""
	}
	return hex.EncodeToString(encryptedData)
}

func Index(s string, index int) string {
	runes := bytes.Runes([]byte(s))
	for i, rune := range runes {
		if i == int(index) {
			return string(rune)
		}
	}
	return ""
}
func extractTextFromHTML(htmlStr string) string {
	doc, _ := html.Parse(strings.NewReader(htmlStr))

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text := f(c)
			if text != "" {
				return text
			}
		}
		return ""
	}

	return f(doc)
}
func Has[T comparable](a []T, b T) bool {
	for _, s := range a {
		if b == s {
			return true
		}
	}
	return false
}
func remove[T comparable](slice []T, value T) []T {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
func chunkSlice[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
func checkIP(c *resty.Client) string {
	res, _ := c.R().SetHeader("Connection", "close").Get("https://api.bilibili.com/x/web-interface/zone")
	return res.String()
}
func RandomPick[T any](arr []T) T {
	if len(arr) == 0 {
		var zero T
		return zero // 如果数组为空，返回零值
	}
	mathRand.Seed(time.Now().UnixNano()) // 初始化随机种子
	index := mathRand.Intn(len(arr))
	return arr[index]
}
