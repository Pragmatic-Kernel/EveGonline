package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrCacheExpired = errors.New("cache too old")

func TouchFile(url string, directory string) error {
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, "=", "_")
	url = strings.ReplaceAll(url, "?", "_")
	currentTime := time.Now()
	filePath := filepath.Join("cache", directory, url)
	err := os.Chtimes(filePath, currentTime, currentTime)
	if err != nil {
		return fmt.Errorf("unable to touch cache file: %w", err)
	}
	return nil
}

func SetCache(url string, directory string, body []byte) ([]byte, error) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, "=", "_")
	url = strings.ReplaceAll(url, "?", "_")
	file, err := os.Create(filepath.Join("cache", directory, url))
	if err != nil {
		return nil, fmt.Errorf("unable to create cache file: %w", err)
	}
	defer file.Close()
	file.Write(body)
	if err != nil {
		return nil, fmt.Errorf("unable to write cache file: %w", err)
	}
	fmt.Printf("Ok cache set %s\n", url)
	return body, nil
}

func MoveCacheFile(url string, directory string) error {
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, "=", "_")
	url = strings.ReplaceAll(url, "?", "_")
	file, err := os.Open(filepath.Join("cache", directory, url))
	if err != nil {
		return err
	}
	defer file.Close()
	currentTimefmt := time.Now().Format("_2006-01-02_15_04")
	e := os.Rename(file.Name(), file.Name()+currentTimefmt)
	if e != nil {
		return fmt.Errorf("error while renaming cache file: %w", err)
	}
	return nil
}

func GetCache(url string, directory string, maxage int) ([]byte, error) {
	// if maxage = 0; no expiry
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, "=", "_")
	url = strings.ReplaceAll(url, "?", "_")
	body := bytes.Buffer{}
	file, err := os.Open(filepath.Join("cache", directory, url))
	if err != nil {
		return nil, nil
	}
	defer file.Close()
	if maxage != 0 {
		fileAge, err := checkFileAge(file)
		if err != nil {
			return nil, fmt.Errorf("error while checking age for cache file: %w", err)
		}
		if fileAge > maxage {
			return nil, ErrCacheExpired
		}
	}
	_, err = io.Copy(&body, file)
	if err != nil {
		return nil, fmt.Errorf("unable to write cache file: %w", err)
	}
	fmt.Printf("Ok cache get %s: %d bytes\n", url, len(body.Bytes()))
	return body.Bytes(), nil
}

func checkFileAge(file *os.File) (int, error) {
	currentTime := time.Now()
	fi, err := os.Stat(file.Name())
	if err != nil {
		return -1, fmt.Errorf("unable to get stat for cache file: %w", err)
	}
	diff := currentTime.Sub(fi.ModTime())
	return int(diff.Seconds()), nil
}
