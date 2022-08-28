package main

import (
	"bufio"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"sort"
)

type colorCount struct {
	color string
	count int
}

type colorCountArray []colorCount

func (c colorCountArray) Len() int {
	return len(c)
}

func (c colorCountArray) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c colorCountArray) Less(i, j int) bool {
	return c[i].count > c[j].count
}

func readFile(path string, queue chan<- string) {
	file, _ := os.Open(path)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stringToRead := scanner.Text()
		queue <- stringToRead
	}
	close(queue)
}

func writeFile(path string, queue chan string, doneWriting chan bool) {
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	writer := bufio.NewWriter(file)
	for line := range queue {
		writer.WriteString(line)
	}
	writer.Flush()
	doneWriting <- true
}

func runThread(urls <-chan string, output chan<- string, done chan<- bool) {
	for url := range urls {
		line := downloadImage(url)
		output <- line
	}
	done <- true
}

func downloadImage(url string) string {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer response.Body.Close()
	image, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	bounds := image.Bounds()
	counts := make(map[string]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := image.At(x, y).RGBA()
			color := fmt.Sprintf("%02X%02X%02X", r>>8, g>>8, b>>8)
			counts[color] += 1
		}
	}
	vals := make(colorCountArray, 0, len(counts))
	for k, v := range counts {
		vals = append(vals, colorCount{color: k, count: v})
	}
	sort.Sort(colorCountArray(vals))
	outputString := fmt.Sprintf("%s,#%s,#%s,#%s\n", url, vals[0].color, vals[1].color, vals[2].color)
	return outputString
}

var numThreads = 10

func main() {
	input := make(chan string)
	output := make(chan string)
	done := make(chan bool)
	doneWriting := make(chan bool)
	go readFile("./input.txt", input)
	go writeFile("./output.txt", output, doneWriting)
	for i := 0; i < numThreads; i++ {
		go runThread(input, output, done)
	}

	for i := 0; i < numThreads; i++ {
		<-done
	}
	close(output)
	<-doneWriting
}
