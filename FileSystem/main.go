package main

import (
	"flag"
	"fmt"
	"math"
	"sync"
	"time"

	// "math"

	// "io/fs"
	"os"
	"path/filepath"
	"sort"
)

func getFilesAndSizes(root string) ([]string, []float64, error) {
	var files []string
	var sizes []float64
	var mu sync.Mutex
	var wg sync.WaitGroup


	fmt.Printf("Scanning directory: %s\n", root)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем корневую директорию
		if path == root {
			return nil
		}

		// Проверяем, является ли файл или директория на первом уровне
		if filepath.Dir(path) == root {
			wg.Add(1)
			go func(path string, info os.FileInfo) {
				defer wg.Done() // Уменьшаем счетчик при завершении горутины
				mu.Lock()       // Блокируем мьютекс для безопасного доступа к общим данным
				defer mu.Unlock()

				files = append(files, path)
				if info.IsDir() {
					// Получаем размер директории, включая все вложенные файлы и поддиректории
					size := getDirSize(path)
					sizes = append(sizes, float64(size))
				} else {
					sizes = append(sizes, float64(info.Size()))
				}
			}(path, info)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	wg.Wait() // Ожидаем завершения всех горутин

	return files, sizes, err
}

func getDirSize(path string) float64 {
	var size float64

	// Рекурсивно обходим все файлы и поддиректории.
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Для файлов добавляем их размер.
			if info.Name() != filepath.Base(path){
				size += float64(info.Size())
			}
		} else{
			size += float64(info.Size())
		}
		return nil
	})
	if err != nil {
		fmt.Println("ошибка при вычислении размера директории:", err)
		return 0
	}
	fmt.Println("path", path, size)

	return size
}

func sortFiles(files []string, sizes []float64, order string) {
	type fileSize struct {
		path string
		size float64
	}
	var fileSizes []fileSize
	for i := range files {
		fileSizes = append(fileSizes, fileSize{files[i], sizes[i]})
	}

	sort.Slice(fileSizes, func(i, j int) bool {
		if order == "asc" {
			return fileSizes[i].size < fileSizes[j].size
		} else {
			return fileSizes[i].size > fileSizes[j].size
		}
	})

	for i := 0; i < len(fileSizes); i++ {
		files[i] = fileSizes[i].path
		sizes[i] = fileSizes[i].size
	}
}


func printFiles(files []string, sizes []float64) error {
	if len(files) == 0 {
		fmt.Println("No files or directories found.")
		return nil
	}

	for i, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Error getting info for %s: %s\n", file, err)
			return err
		}

		name := filepath.Base(file)
		sizeFormatted := convertSize(sizes[i])

		if info.IsDir() {
			fmt.Printf("[DIR]  %s (%s)\n", name, sizeFormatted)
		} else {
			fmt.Printf("[FILE] %s (%s)\n", name, sizeFormatted)
		}
	}
	return nil
}

func convertSize(size float64) string {
	floatSize := float64(size)
	counter := 0
	var value string

	for {
		if floatSize >= 1000 {
			floatSize = floatSize / 1000
			counter += 1
		} else {
			break
		}
	}
	roundedSize := math.Round(floatSize*10) / 10
	switch counter {
	case 0:
		value = "B"
	case 1:
		value = "KB"
	case 2:
		value = "MB"
	case 3:
		value = "GB"
	}
	return fmt.Sprintf("%v %s", roundedSize, value) 
}

func main() {
	root := flag.String("root", "", "choose a directory")
	sortOrder := flag.String("sort", "asc", "choose sorting of directory (asc/desc)")
	flag.Parse()

	if *root == "" {
		fmt.Println("Please warinng a directory using the -root flag.")
		return
	}

	if _, err := os.Stat(*root); os.IsNotExist(err) {
		fmt.Println("Directory does not exist.")
		return
	}
	startTime := time.Now()

	files, sizes, err := getFilesAndSizes(*root)
	if err != nil {
		fmt.Printf("Error walking the directory: %s\n", err)
		return
	}

	sortFiles(files, sizes, *sortOrder)

	printFiles(files, sizes)

	TimeEnd := time.Since(startTime)
	fmt.Printf("time end: %s\n", TimeEnd)
}
