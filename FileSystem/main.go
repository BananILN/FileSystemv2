package main

import (
	"flag"
	"fmt"
	"math"
	"sync"
	"time"
	"os"
	"path/filepath"
	"sort"
)
type fileSize struct {
	path string
	size float64
}


type SizeUnit int

const (
	B SizeUnit = iota // 0
	KB                 // 1
	MB                 // 2
	GB                 // 3
)

// Функция для получения файлов и его размеров
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

	

		if filepath.Dir(path) == root {
			wg.Add(1)
			go func(path string, info os.FileInfo) {
				defer wg.Done() 
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

	wg.Wait() 

	return files, sizes, err
}

// Функция для подсчета размера директорий и файлов 
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

// Функция для сортировки по размеру (возрастание\убывание)
func sortFiles(files []string, sizes []float64, order string) {
	
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

// Функция для вывода файлов и его размеров
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
// Функция для форматирований единиц размерности
func convertSize(size float64) string {
	floatSize := float64(size)
	var unit SizeUnit

	for {
		if floatSize >= 1000 {
			floatSize = floatSize / 1000
			unit++
		} else {
			break
		}
	}
	roundedSize := math.Round(floatSize*10) / 10
	var unitString string

	switch unit {
	case B:
		unitString = "B"
	case KB:
		unitString = "KB"
	case MB:
		unitString = "MB"
	case GB:
		unitString = "GB"
	default:
		unitString = "Unknown" 
	}
	return fmt.Sprintf("%v %s", roundedSize, unitString) 
}

// Основная функция
func main() {
	root := flag.String("root", "", "choose a directory")
	sortOrder := flag.String("sort", "asc", "choose sorting of directory (asc/desc)")
	flag.Parse()

	if len(os.Args) < 2 {
		flag.PrintDefaults()
		return
	}

	if *root == "" {
		fmt.Println("Please warinng a directory using the -root flag.")
		return
	}
	if *root != *sortOrder{
		fmt.Println("You can use -sort flag to choose sorting of directory (asc/desc), default sort = asc")
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
