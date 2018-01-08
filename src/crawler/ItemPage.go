package main

import (
    "errors"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "html"
    "regexp"
    "strconv"
    "strings"
)

var (
    ITEM_RE    = regexp.MustCompile("<li class=\"item\"[\\s\\S]+?</li>")
    VARIANT_RE = regexp.MustCompile("<span class=\"name\">([^<]+)")
    PRICE_RE   = regexp.MustCompile("<span.+?itemprop=\"price\" content=\"([^\"]+?)\">")
    STRIP_RE   = regexp.MustCompile("\\s{2,}")
    BRAND_RE   = regexp.MustCompile("<span class=\"brand\">([^<]+?)</span> <strong>([^<]+?)</strong>")
)

type ItemPage struct {
    content string
    name    string
    price   float64
    brand   string
    variant string
}

func create(body, brand, name string) (*ItemPage, error) {
    item := new(ItemPage)
    item.content = body
    item.brand = brand
    item.name = name

    variant, err := item.getItemVariant()
    if err != nil {
        return nil, err
    }
    item.variant = variant
    price, err := item.getItemPrice()
    if err != nil {
        return nil, err
    }
    item.price = price
    return item, nil
}

func createFromNode(node *goquery.Selection, brand, name string) (*ItemPage, error) {
    item := new(ItemPage)
    var err error = nil
    item.brand = brand
    item.name = name

    item.variant = strings.Replace(node.Find("span.name").Text(), "Wyprzedaż", "", -1)
    price_str, exists := node.Find("span[itemprop=\"price\"]").Attr("content")
    if !exists {
        item.price = 0
    } else {
        item.price, err = strconv.ParseFloat(price_str, 64)
    }
    return item, err
}

func (ip *ItemPage) getItemVariant() (string, error) {
    match := VARIANT_RE.FindStringSubmatch(ip.content)
    if len(match) == 0 {
        return "", nil
    }
    variant := html.UnescapeString(match[1])
    variant = strings.Replace(variant, "\u00a0", "", -1)
    variant = strings.Replace(variant, "\xa0", "", -1)
    variant = STRIP_RE.ReplaceAllLiteralString(variant, "")
    return variant, nil
}

func (ip *ItemPage) getItemPrice() (float64, error) {
    match := PRICE_RE.FindStringSubmatch(ip.content)
    if len(match) == 0 {
        return 0, nil
    }
    val, err := strconv.ParseFloat(match[1], 64)
    if err != nil {
        return 0, errors.New("Failed to convert price.")
    }
    return val, nil
}

func readItems(body string) ([]*ItemPage, []string) {
    var items []*ItemPage
    var others []string

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
    if err != nil {
        fmt.Println("Failed to parse results.")
        return items, others
    }
    base_info := doc.Find(".product-base-info h1")
    brand := base_info.Find(".brand").Text()
    name := base_info.Find("strong").Text()

    variants := doc.Find("#variants li.item")

    variants.Each(func(n int, node *goquery.Selection) {
        item, err := createFromNode(node, brand, name)
        if err != nil {
            fmt.Printf("err %s\n", err)
            return
        }
        items = append(items, item)
    })

    doc.Find("#other-products li.item").Each(func(n int, node *goquery.Selection) {
        attr, exists := node.Find("a").First().Attr("href")
        if exists {
            others = append(others, attr)
        }
    })
    return items, others
}
