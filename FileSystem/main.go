package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

type Config struct {
	Host string
	Port string
}

type FileInfo struct {
	Path  string 
	Size  string  
	IsDir bool   
}

type SizeUnit int

const (
	B SizeUnit = iota // 0
	KB                 // 1
	MB                 // 2
	GB                 // 3
)

// Функция для получения файлов и его размеров
func getFilesAndSizes(root string) ([]FileInfo, error) {
	var fileInfos []FileInfo
	var mu sync.Mutex
	var wg sync.WaitGroup

	fmt.Printf("Scanning directory: %s\n", root)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return nil 
		}

		if filepath.Dir(path) == root {
			wg.Add(1)
			go func(path string, info os.FileInfo) {
				defer wg.Done()
				mu.Lock() 
				defer mu.Unlock()

				var size float64
				if info.IsDir() {
			
					size = getDirSize(path)
				} else {
					size = float64(info.Size())
				}

				fileInfos = append(fileInfos, FileInfo{
					Path:  path,
					Size:  convertSize(size), 
					IsDir: info.IsDir(),
				})
			}(path, info)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the directory: %v", err)
	}

	wg.Wait()

	return fileInfos, nil
}

// Функция для подсчета размера директорий и файлов 
func getDirSize(path string) float64 {
	var size float64

	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return nil 
		}
		if info.IsDir() {
			
			if info.Name() != filepath.Base(path) {
				size += float64(info.Size())
			}
		} else {
			size += float64(info.Size())
		}
		return nil
	})
	if err != nil {
		log.Printf("ошибка при вычислении размера директории %s: %v", path, err)
		return 0
	}

	return size
}

// Функция для сортировки по размеру (возрастание\убывание)
func sortFiles(fileInfos []FileInfo, order string) {
	sort.Slice(fileInfos, func(i, j int) bool {
		if order == "asc" {
			return fileInfos[i].Size < fileInfos[j].Size
		} else {
			return fileInfos[i].Size > fileInfos[j].Size
		}
	})
}

// Функция для форматирования единиц размерности
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

// Обработчик для вывода JSON
func jsonHandler(root string, order string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileInfos, err := getFilesAndSizes(root)
		if err != nil {
			http.Error(w, "Ошибка при получении файлов: "+err.Error(), http.StatusInternalServerError)
			return
		}

		sortFiles(fileInfos, order)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(fileInfos); err != nil {
			http.Error(w, "Ошибка при выводе JSON: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки .env файла: %v", err)
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	if host == "" || port == "" {
		return nil, fmt.Errorf("HOST или PORT не заданы в .env файле")
	}

	return &Config{
		Host: host,
		Port: port,
	}, nil
}

// Основная функция
func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := flag.String("root", "/home/danil", "choose a directory")
	sortOrder := flag.String("sort", "asc", "choose sorting of directory (asc/desc)")
	flag.Parse()

	if *root == "" {
		fmt.Println("Please specify a directory using the -root flag.")
		return
	}

	if _, err := os.Stat(*root); os.IsNotExist(err) {
		fmt.Println("Directory does not exist.")
		return
	}

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler: http.HandlerFunc(jsonHandler(*root, *sortOrder)),
	}

	
	go func() {
		fmt.Printf("Сервер запущен на http://%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}
	}()

	
	<-signalChan
	fmt.Println("Получен сигнал завершения, начинаем завершение работы...")

	
	cancel()

	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}

	fmt.Println("Сервер завершил работу.")
}