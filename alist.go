package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"
)

func UploadFile(path string, alistPath string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("打开文件失败: %v", err)
	}
	defer file.Close()

	bodyReader, bodyWriter := io.Pipe() // 创建 Pipe
	writer := multipart.NewWriter(bodyWriter)

	go func() {
		defer bodyWriter.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			bodyWriter.CloseWithError(fmt.Errorf("创建表单文件失败: %w", err))
			return
		}

		_, err = io.Copy(part, file)
		if err != nil {
			bodyWriter.CloseWithError(fmt.Errorf("复制文件数据失败: %w", err))
			return
		}

		writer.Close()
	}()

	req, err := http.NewRequest("PUT", config.AlistServer+"api/fs/form", bodyReader)
	if err != nil {
		log.Println("创建请求失败:", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", GetAlistToken())
	req.Header.Set("File-Path", alistPath)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("上传请求失败: %w", err.Error())
	} else {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应失败: %w", err)
	}

	log.Printf("[%s] %d %s\n", alistPath, resp.StatusCode, string(body))
	return nil
}
func UploadBytes(bytes0 []byte, alistPath string) error {

	req, err := http.NewRequest("PUT", config.AlistServer+"api/fs/put", bytes.NewReader(bytes0))
	if err != nil {
		log.Println("创建请求失败:", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", GetAlistToken())
	req.Header.Set("File-Path", alistPath)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("上传请求失败: %w", err.Error())
	} else {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应失败: %w", err)
	}

	log.Printf("[%s] %d %s\n", alistPath, resp.StatusCode, string(body))
	return nil
}

func GetAlistToken() string {
	type LoginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var sum = sha256.Sum256([]byte(config.AlistPass + "-https://github.com/alist-org/alist"))
	var req = LoginRequest{Username: config.AlistUser, Password: hex.EncodeToString(sum[:])}
	alist, err := client.R().SetBody(req).Post(config.AlistServer + "api/auth/login/hash")

	if err != nil {
		log.Println(err)
	}
	var res = LoginResponse{}
	sonic.Unmarshal(alist.Body(), &res)
	return res.Data.Token
}

func GetFile(path string) string {
	type Request struct {
		Path string `json:"path"`
	}
	var req = Request{Path: path}
	res, _ := client.R().SetBody(req).Post(config.AlistServer + "api/fs/get/")

	var obj map[string]interface{}
	sonic.Unmarshal(res.Body(), &obj)

	return (getString(obj, "data.raw_url"))
}

type File struct {
	FileName  string
	CreatedAt string
	Sign      string
	Size      int64
	Link      string
}

func ListFile(path string) []File {
	type Request struct {
		Path    string `json:"path"`
		Refresh bool   `json:"refresh"`
	}
	var obj map[string]interface{}
	res, _ := queryClient.R().SetBody(Request{Refresh: false, Path: path}).Post(config.AlistServer + "api/fs/list/")

	sonic.Unmarshal(res.Body(), &obj)

	var results []File

	if getInt(obj, "code") == 200 {
		for _, item := range getArray(obj, "data.content") {
			var f File
			f.FileName = getString(item, "name")
			f.CreatedAt = getString(item, "created")
			f.Sign = getString(item, "sign")
			f.Link = config.AlistServer + "d/Microsoft365" + strings.Replace((getString(item, "path")), "Live/", "", 1)

			results = append(results, f)
		}
	}

	return results

}
