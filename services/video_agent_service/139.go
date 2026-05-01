package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type CMDrive struct {
	Authorization string
	resty         *resty.Client
	ChunkSize     int64
	Phone         string
}

type DriveItem struct {
	FileID   string
	FileName string
	Size     int64
}

func NewDrive(auth string) *CMDrive {
	var r = resty.New()
	r.OnBeforeRequest(func(client *resty.Client, request *resty.Request) error {
		if strings.Contains(request.URL, "personal-kd-njs.yun.139.com") {
			request.Header.Set("Authorization", auth)
			request.Header.Set("x-yun-module-type", "100")
			request.Header.Set("x-yun-client-info", "||9|7.17.3|chrome|148.0.0.0|||windows 10||en|||dWmVk||")
			request.Header.Set("x-yun-app-channel", "10000034")
			request.Header.Set("x-yun-api-version", "v1")
		}
		return nil
	})
	return &CMDrive{
		Authorization: auth,
		resty:         r,
		ChunkSize:     1024 * 1024 * 1024,
	}
}

func (drive *CMDrive) ListFiles(parent string, offset ...string) []DriveItem {
	var items []DriveItem
	var off = ""
	if len(offset) >= 1 {
		off = offset[0]
	}
	type PageInfo struct {
		PageSize   int    `json:"pageSize"`
		PageOffset string `json:"pageCursor"`
	}
	res, _ := drive.resty.R().SetBody(struct {
		Parent   string   `json:"parentFileId"`
		PageInfo PageInfo `json:"pageInfo"`
	}{
		Parent: parent,
		PageInfo: PageInfo{
			PageSize:   100,
			PageOffset: off,
		},
	}).Post("https://personal-kd-njs.yun.139.com/hcy/file/list")
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)

	if getBool(obj, "success") == true {
		for _, i := range getArray(obj, "data.items") {
			items = append(items, DriveItem{
				FileID:   getString(i, "fileId"),
				FileName: getString(i, "name"),
				Size:     getInt64(i, "size"),
			})
		}

	}

	return items

}

func (drive *CMDrive) CreateFile(location, parent, name string) ([]string, func()) {

	file, err := os.Open(location)
	if err != nil {
		return nil, nil
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, nil
	}

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, nil
	}
	contentHash := hex.EncodeToString(h.Sum(nil))

	var partInfos []map[string]interface{}
	size := info.Size()
	chunkSize := drive.ChunkSize
	numParts := (size + chunkSize - 1) / chunkSize
	if numParts == 0 {
		numParts = 1
	}

	for i := int64(0); i < numParts; i++ {
		offset := i * chunkSize
		partSize := chunkSize
		if offset+partSize > size {
			partSize = size - offset
		}
		partInfos = append(partInfos, map[string]interface{}{
			"parallelHashCtx": map[string]interface{}{
				"partOffset": offset,
			},
			"partNumber": i + 1,
			"partSize":   partSize,
		})
	}

	body := map[string]interface{}{
		"fileRenameMode":       "auto_rename",
		"contentType":          "application/oct-stream",
		"type":                 "file",
		"name":                 name,
		"size":                 size,
		"contentHashAlgorithm": "SHA256",
		"contentHash":          contentHash,
		"partInfos":            partInfos,
		"parentFileId":         parent,
		"commonAccountInfo": map[string]interface{}{
			"account":     drive.Phone,
			"accountType": 1,
		},
	}

	res, _ := drive.resty.R().SetBody(body).Post("https://personal-kd-njs.yun.139.com/hcy/file/create")
	var obj map[string]interface{}
	json.Unmarshal(res.Body(), &obj)
	log.Println(res.String())
	var uploadUrls []string
	for _, i := range getArray(obj, "data.partInfos") {
		uploadUrls = append(uploadUrls, getString(i, "uploadUrl"))
	}
	return uploadUrls, func() {
		//complete handler
		// https://personal-kd-njs.yun.139.com/hcy/file/complete
		/*
			{
			    "fileId": "FqZvJSoY3kHMFHjU875hM2rw3_UmJN4Xz",
			    "uploadId": "2~DWk7p0auu249M1zWoFMm0q7R-QtFplS",
			    "contentHash": "1d5ff169f13a2830ac63f7225aab21f1bfbb4814fc00984d9115f1b3b5b77fe2",
			    "contentHashAlgorithm": "SHA256"
			}
		*/
		drive.resty.R().SetBody(map[string]interface{}{
			"fileId":               getString(obj, "data.fileId"),
			"uploadId":             getString(obj, "data.uploadId"),
			"contentHash":          contentHash,
			"contentHashAlgorithm": "SHA256",
		}).Post("https://personal-kd-njs.yun.139.com/hcy/file/complete")
	}
}

func (drive *CMDrive) MkDir(parent, name string) {
	drive.resty.R().SetBody(fmt.Sprintf(`
{"parentFileId":"%s","name":"%s","description":"","type":"folder","fileRenameMode":"force_rename"}
`, parent, name)).Post("https://personal-kd-njs.yun.139.com/hcy/file/create")
}

func (drive *CMDrive) UploadFile(location, parent, name string) error {
	urls, complete := drive.CreateFile(location, parent, name)

	if urls == nil || complete == nil {
		return fmt.Errorf("failed to init file upload")
	}

	file, err := os.Open(location)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	size := info.Size()

	for i, url := range urls {
		offset := int64(i) * drive.ChunkSize
		partSize := drive.ChunkSize
		if offset+partSize > size {
			partSize = size - offset
		}

		section := io.NewSectionReader(file, offset, partSize)

		req, err := http.NewRequest(http.MethodPut, url, section)
		if err != nil {
			return err
		}

		// 显式设置 ContentLength，禁用 chunked 编码
		req.ContentLength = partSize
		req.Header.Set("Content-Type", "application/octet-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		status := resp.StatusCode
		resp.Body.Close()

		if status != http.StatusOK {
			return fmt.Errorf("part %d upload failed with status %d", i+1, status)
		}
	}

	complete()
	return nil
}
