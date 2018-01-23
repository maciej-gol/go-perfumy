package main

import (
    "encoding/csv"
    "flag"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
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

    return readItems(string(body), url)
}

func fetch_parphumes(workersChan chan int, num, step int, output_dir string) {
    max_page := 10000
    page := num
    directory_name := time.Now().Format("2006-01-02")
    filename := fmt.Sprintf("%d.csv", num)
    file_path := filepath.Join(output_dir, directory_name, filename)
    os.MkdirAll(filepath.Dir(file_path), os.ModePerm)
    out, err := os.Create(file_path)
    if err != nil {
        fmt.Printf("[%d] Failed to create file %q: %s.\n", num, file_path, err)
        workersChan <- num
        return
    }
    defer out.Close()

    fmt.Printf("[%d] Writing to %q.\n", num, file_path)
    csv_writer := csv.NewWriter(out)
    csv_writer.Write([]string{"brand", "name", "variant", "price", "discountInfo", "url", "html"})

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
                csv_writer.Write([]string{item.brand, item.name, item.variant, price, item.discountInfo, item.url, item.html})
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

func start_crawl(num_workers int, output_dir string) {
    workersChan := make(chan int)
    output_dir, err := filepath.Abs(output_dir)
    if err != nil {
        fmt.Printf("Failed to get path.")
        return
    }

    for i := 1; i <= num_workers; i++ {
        go fetch_parphumes(workersChan, i, num_workers, output_dir)
    }

    for i := 1; i <= num_workers; i++ {
        <-workersChan
    }
}

func main() {
    start := time.Now()

    output_dir_ptr := flag.String("output-dir", "", "Directory to output the crawls.")
    flag.Parse()

    start_crawl(5, *output_dir_ptr)

    fmt.Printf("Finished in %d seconds.\n", int64(time.Now().Sub(start)/time.Second))
}
