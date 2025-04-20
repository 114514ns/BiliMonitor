package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bytedance/sonic"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func UploadFile(path string, alistPath string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("打开文件失败: %w", err)
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
	}
	defer resp.Body.Close()

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
