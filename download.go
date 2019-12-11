package main

import (
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

func printDownloadPercent(done chan int64, dest string, total int64) {
    stop := false
    for {
        select {
        case <-done:
            stop = true
        default:
            file, err := os.Open(dest)
            if err != nil {
                log.Fatal(err)
            }
            fi, err := file.Stat()
            if err != nil {
                log.Fatal(err)
            }
            size := fi.Size()
            percent := float64(size) / float64(total) * 100
            log.Printf("Downloaded %d/%d (%.2f%%)\n", size, total, percent)
        }
        if stop {
            break
        }
        time.Sleep(5 * time.Second)
    }
}

func DownloadFile(url string, dest string) {
    log.Printf("Downloading from %s\n", url)
    start := time.Now()

    out, err := os.Create(dest)
    if err != nil {
        log.Fatalln(err)
    }
    defer out.Close()

    headerReq, _ := http.NewRequest("GET", url, nil)
    headerReq.Header.Set("Range", "bytes=0-0")
    headerRes, err := http.DefaultClient.Do(headerReq)
    if err != nil {
        log.Fatalln(err)
    }
    defer headerRes.Body.Close()

    sizeStr := strings.Split(headerRes.Header.Get("Content-Range"), "/")[1]
    size, err := strconv.Atoi(sizeStr)
    if err != nil {
        log.Fatalln(err)
    }

    doneChan := make(chan int64)
    go printDownloadPercent(doneChan, dest, int64(size))

    res, err := http.Get(url)
    if err != nil {
        log.Fatalln(err)
    }
    defer res.Body.Close()

    n, err := io.Copy(out, res.Body)
    if err != nil {
        log.Fatalln(err)
    }
    doneChan <- n
    elapsed := time.Since(start)
    log.Printf("Download completed in %s", elapsed)
}
