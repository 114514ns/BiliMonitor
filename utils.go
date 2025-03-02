package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
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
