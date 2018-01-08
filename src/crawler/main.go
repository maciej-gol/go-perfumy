package main

import (
    "encoding/csv"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "io/ioutil"
    "net/http"
    "os"
    "strconv"
    "time"
)

func fetch_and_process_url(url string) ([]*ItemPage, []string) {
    resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("Error downloading url %s: %s.\n", url, err)
        return nil, nil
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response %s.\n", url)
    }

    return readItems(string(body))
}

func fetch_parphumes(workersChan chan int, num, step int) {
    max_page := 10000
    page := num
    directory_name := time.Now().Format("2006-01-02")
    os.MkdirAll(directory_name, os.ModePerm)
    filename := fmt.Sprintf("%s/%d.csv", directory_name, num)
    out, err := os.Create(filename)
    if err != nil {
        workersChan <- num
        return
    }
    defer out.Close()

    fmt.Printf("[%d] Writing to %q.\n", num, filename)
    csv_writer := csv.NewWriter(out)
    csv_writer.Write([]string{"brand", "name", "variant", "price"})

    for page <= max_page {
        url := fmt.Sprintf("https://www.iperfumy.pl/perfumy-new/?f=%d-2-6362", page)
        doc, err := goquery.NewDocument(url)
        if err != nil {
            fmt.Printf("Failed to fetch %q: %s.\n", url, err)
            return
        }

        max_page, err = strconv.Atoi(doc.Find("span.pages > a").Eq(-2).Text())
        if err != nil {
            fmt.Printf("Failed to parse last page: %s.\n", err)
            max_page = -1
        }

        doc.Find(".product-list li.item a:first-child").Each(func(n int, node *goquery.Selection) {
            href, exists := node.Attr("href")
            if !exists {
                return
            }

            items, _ := fetch_and_process_url(href)
            for _, item := range items {
                price := fmt.Sprintf("%.02f", item.price)
                csv_writer.Write([]string{item.brand, item.name, item.variant, price})
            }
        })

        page += step
        if ((page-num)/step)%20 == 0 {
            fmt.Printf("[%d] Fetching page %d of %d.\n", num, page, max_page)
        }
    }
    csv_writer.Flush()
    workersChan <- num
}

func main() {
    workersChan := make(chan int)
    num_workers := 5
    start := time.Now()

    for i := 1; i <= num_workers; i++ {
        go fetch_parphumes(workersChan, i, num_workers)
    }

    for i := 1; i <= num_workers; i++ {
        <-workersChan
    }

    fmt.Printf("Finished in %d seconds.\n", int64(time.Now().Sub(start)/time.Second))
}
