package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	src := flag.String("src", "", "путь к файлу с URL-адресами")
	dst := flag.String("dst", "", "путь к директории, где будут храниться .html страницы")

	flag.Parse()

	if *src == "" || *dst == "" {
		fmt.Println("Необходимо указать путь к файлу с URL-адресами и путь к директории для сохранения.")
		return
	}

	timeNow := time.Now()
	fmt.Println("Program has been started on", timeNow)

	file, err := os.Open(*src)
	if err != nil {
		fmt.Println("Error file didn't open", err)
		return
	}
	defer file.Close()

	var urls []string
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		url := scan.Text() // Сохраняем текст в переменной url

		// Проверяем  URL с http:// или https://
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url 
		}
		urls = append(urls, url) 
	}

	if err := scan.Err(); err != nil {
		fmt.Println("Error of read file:", err)
		return
	}

	err = os.MkdirAll(*dst, 0777)
	if err != nil {
		fmt.Println("Error of Directory", err)
		return
	}

	var wg sync.WaitGroup // Создаем WaitGroup для ожидания завершения горутин

	for _, urlStr := range urls {
		wg.Add(1) // Увеличиваем счетчик WaitGroup на 1

		go func(urlStr string) {
			defer wg.Done() // Уменьшаем счетчик WaitGroup на 1 при завершении горутины

			start := time.Now()

			response, err := http.Get(urlStr)
			if err != nil {
				fmt.Println("Error of http get ", err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error of Read data", urlStr, err)
				return
			}

			// Извлекаем имя файла из URL 
			parsedUrl, err := url.Parse(urlStr)
			if err != nil {
				fmt.Println("Error parsing URL:", urlStr, err)
				return
			}
			
			fileName := parsedUrl.Host + ".html"

			filename := filepath.Join(*dst, fileName)

			err = os.WriteFile(filename, body, 0644)
			if err != nil {
				fmt.Println("Error of storage data in the file ", err)
				return
			}

			requestTime := time.Since(start)

			fmt.Printf("Страница %s сохранена в %s (время выполнения: %v)\n", urlStr, filename, requestTime)
		}(urlStr) // Передаем urlStr в горутину
	}

	

	wg.Wait() // Ожидаем завершения всех горутин

	totalTime := time.Since(timeNow)
	fmt.Println("Программа завершена. Общее время выполнения:", totalTime)
}


