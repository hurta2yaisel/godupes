package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func GetPath() string {
	current_path, err1 := os.Getwd()
	if err1 != nil {
		current_path = ""
	}

	var path string
	flag.StringVar(&path, "path", current_path, "Directory to scan")
	flag.Parse()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Fatalf("Folder '%s' does not exist.", path)
	}

	return path
}

func WalkPath(path string) map[int64][]string {
	file_sizes := make(map[int64][]string)
	err := filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if !info.IsDir() {
			file_size := info.Size()
			if file_size > 0 {
				if file_sizes[file_size] == nil {
					file_sizes[file_size] = make([]string, 0)
				}
				file_sizes[file_size] = append(file_sizes[file_size], filename)
			}
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	return file_sizes
}

func HashFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func BuildHashMap(file_sizes map[int64][]string) map[string][]string {
	hash_map := make(map[string][]string)
	for file_size, files := range file_sizes {
		if len(files) > 1 {
			for _, filename := range files {
				hash := HashFile(filename)
				if hash != "" {
					if hash_map[hash] == nil {
						hash_map[hash] = make([]string, 0)
					}
					filename = fmt.Sprintf("(%d bytes) %s", file_size, filename)
					hash_map[hash] = append(hash_map[hash], filename)
				}
			}
		}
	}
	return hash_map
}

func FindDupes(path string) {
	log.Printf("Scanning %s\n\n", path)
	file_sizes := WalkPath(path)
	hash_map := BuildHashMap(file_sizes)

	var duplicates int
	for _, files := range hash_map {
		len_files := len(files)
		if len_files > 1 {
			duplicates += 1
			count := 1
			for _, file := range files {
				if count == len_files {
					log.Printf("%s\n\n", file)
				} else {
					log.Printf("%s\n", file)
				}
				count += 1
			}
		}
	}
	log.Printf("Finished: %d groups of duplicates were found", duplicates)
}

func main() {
	start := time.Now()
	path := GetPath()
	FindDupes(path)
	elapsed := time.Since(start)
	log.Printf("godupes scan took %s", elapsed)
}
