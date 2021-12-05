package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Pragmatic-Kernel/EveGonline/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotModified = errors.New("error not modified")

func getImage(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	imageType := strings.Split(r.URL.Path, "/")[2]
	fmt.Println(imageType)
	imageIdstr := strings.Split(r.URL.Path, "/")[3]
	imageId64, err := strconv.ParseUint(imageIdstr, 10, 64)
	imageId := uint(imageId64)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Cannot parse image URL"))
	}
	fmt.Println(imageId)
	size, err := getSizeFromUrl(*r.URL)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	payload, err := getImageFromCache(imageType, imageId, size)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Cannot get image from cache"))
	}
	if payload != nil {
		w.Header().Add("Cache-Control", "max-age=7200")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
		return
	}
	url, err := buildImageURL(imageType, imageId, size)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Cannot build image URL"))
		return
	}
	asset := common.Asset{}
	db.Where("id = ? AND size = ?", imageId, size).First(&asset)
	payload, etag, err := getImageFromEsi(url, asset.Etag)
	if err != nil {
		if err == ErrNotModified {
			err := common.TouchFile(url, imageType)
			if err != nil {
				fmt.Println(err)
				w.Write([]byte("Cannot update cache"))
				return
			}
			payload, err = common.GetCache(url, imageType, 0)
			if err != nil {
				fmt.Println(err)
				w.Write([]byte("Cannot get cache post update"))
				return
			}
		} else {
			fmt.Println(err)
			w.Write([]byte("Cannot get image from ESI"))
			return
		}
	} else {
		payload, err = common.SetCache(url, imageType, payload)
		asset.Etag = etag
		asset.ID = imageId
		asset.Size = size
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "size"}},
			DoUpdates: clause.AssignmentColumns([]string{"etag"}),
		}).Create(&asset)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte("Cannot set cache"))
			return
		}
	}
	w.Header().Add("Cache-Control", "max-age=7200")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func getImageFromCache(imageType string, imageId uint, size uint) ([]byte, error) {
	url, err := buildImageURL(imageType, imageId, size)
	if err != nil {
		// FIXME
		return nil, err
	}
	expiry := getExpiryFromType(imageType)
	payload, err := common.GetCache(url, imageType, expiry)
	if err != nil {
		// FIXME
		return nil, err
	}
	return payload, nil
}

func getImageFromEsi(url string, etag string) ([]byte, string, error) {
	url = common.EveImagesUrl + url
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}
	if err != nil {
		return nil, "", fmt.Errorf("cannot create request for image: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	etag = resp.Header.Get("etag")
	if resp.StatusCode == http.StatusNotModified {
		return nil, etag, ErrNotModified
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, etag, errors.New("not found")
	}
	body, _ := io.ReadAll(resp.Body)
	return body, etag, nil
}

func buildImageURL(imageType string, imageId uint, size uint) (string, error) {
	switch imageType {
	case "corporations":
		return imageType + "/" + fmt.Sprintf("%d", imageId) + "/logo?size=" + fmt.Sprintf("%d", size), nil
	case "characters":
		return imageType + "/" + fmt.Sprintf("%d", imageId) + "/portrait?size=" + fmt.Sprintf("%d", size), nil
	case "types":
		return fmt.Sprintf("%d", imageId) + "_" + fmt.Sprintf("%d", size) + ".png", nil
	case "renders":
		return fmt.Sprintf("%d", imageId) + ".png", nil
	}
	return "", fmt.Errorf("unable to build URL for type: %s and id: %d", imageType, imageId)
}

func getExpiryFromType(imageType string) int {
	switch imageType {
	case "corporations":
		return 86400 * 3
	case "chars":
		return 86400 * 3
	case "types":
		return 0
	case "renders":
		return 0
	}
	return 0
}

func getSizeFromUrl(url url.URL) (uint, error) {
	query := url.Query()
	sizes, ok := query["size"]
	if !ok {
		return 0, fmt.Errorf("invalid size parameter, cannot find size")
	}
	if len(sizes) != 1 {
		return 0, fmt.Errorf("invalid size parameter, too many sizes")
	}
	size, err := strconv.ParseUint(sizes[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size parameter, invalid uint")
	}
	if size != 32 && size != 64 && size != 128 && size != 256 && size != 512 {
		return 0, fmt.Errorf("invalid size parameter, invalid size: %d", size)
	}
	return uint(size), nil
}
