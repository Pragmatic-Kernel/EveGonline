package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func setCache(url string, body []byte) ([]byte, error) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	file, err := os.Create(filepath.Join("cache", url))
	if err != nil {
		return nil, fmt.Errorf("unable to create cache file: %w", err)
	}
	file.Write(body)
	if err != nil {
		return nil, fmt.Errorf("unable to write cache file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close cache file: %w", err)
	}
	fmt.Printf("Ok cache set %s\n", url)
	return body, nil
}

func getCache(url string, maxage int) ([]byte, error) {
	// if maxage = 0; no expiry
	url = strings.TrimPrefix(url, "https://")
	url = strings.ReplaceAll(url, "/", "_")
	body := bytes.Buffer{}
	file, err := os.Open(filepath.Join("cache", url))
	if err != nil {
		return nil, nil
	}
	if maxage != 0 {
		fileAge, err := checkFileAge(file)
		if err != nil {
			return nil, fmt.Errorf("error while checking age for cache file: %w", err)
		}
		if fileAge > maxage {
			fmt.Printf("File cache too old, moving.\n")
			currentTimefmt := time.Now().Format("_2006-01-02_15_04")
			e := os.Rename(file.Name(), file.Name()+currentTimefmt)
			if e != nil {
				return nil, fmt.Errorf("error while renaming cache file: %w", err)
			}
			err = file.Close()
			if err != nil {
				return nil, fmt.Errorf("unable to close cache file: %w", err)
			}
			return nil, nil
		}
	}
	_, err = io.Copy(&body, file)
	if err != nil {
		return nil, fmt.Errorf("unable to write cache file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close cache file: %w", err)
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
