package goapkpure

import (
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "io/ioutil"
    "net/http"
    "regexp"
    "sort"
    "strings"
    "sync"
)

type VerItem struct {
    Version string
    DownloadUrl string
    Size string
    Tags []string
    Title string
    UpdateOn string
    Signature string
    Sha1 string
    AndroidVer string
    Architectures []string
    ScreenDPI string
}

type VerItemWithIndex struct {
    i    int
    j    int
    Item VerItem
}

const URL_BASE = "https://apkpure.com"

func GetPackagePageUrl(packageName string) (string, error) {
    url := URL_BASE + "/search?q=" + packageName
    res, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return "", err
    }
    r, err := regexp.Compile("/([^/]*)/" + packageName)
    if err != nil {
        return "", err
    }
    match := r.FindString(string(body))
    return URL_BASE + match, nil
}

func GetDownloadDirectLink(downloadUrl string) (string, error) {
    res, err := http.Get(downloadUrl)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        return "", err
    }
    dlNode := doc.Find("#download_link")
    dlLink, exists := dlNode.First().Attr("href")
    if !exists {
        return "", fmt.Errorf("unable to find download link")
    }
    return dlLink, nil
}

func GetLatestDownloadLink(packageName string) (string, error) {
    packageUrl, err := GetPackagePageUrl(packageName)
    if err != nil {
        return "", err
    }
    downloadUrl := packageUrl + "/download"
    return GetDownloadDirectLink(downloadUrl)
}

func parseVerItem(ch chan<-VerItemWithIndex, wg *sync.WaitGroup, i int, s *goquery.Selection) {
    defer wg.Done()
    size := s.Find(".ver-item-s").First().Text()
    tags := s.Find(".ver-item-t").Map(func(i int, s2 *goquery.Selection) string {
        return s2.Text()
    })
    title := s.Find(".ver-item-a").Children().First().Text()
    nextUrl, _ := s.Find("a").First().Attr("href")

    if s.Has(".ver-info").Length() > 0 {
        infos := make(map[string]string)
        s.Find(".ver-info-m p").Each(func(_ int, s2 *goquery.Selection) {
            y := strings.Split(s2.Text(), ":")
            if len(y) > 1 {
                infos[y[0]] = strings.TrimSpace(y[1])
            }
        })
        verItem := VerItem{}
        verItem.Size = size
        verItem.Tags = tags
        verItem.Title = title
        verItem.DownloadUrl = URL_BASE + nextUrl
        verItem.Version = s.Find(".ver-info-top").Text()[len(verItem.Title) + 1:]
        verItem.UpdateOn = infos["Update on"]
        verItem.Signature = infos["Signature"]
        verItem.Sha1 = infos["File SHA1"]
        verItem.AndroidVer = infos["Requires Android"]
        verItem.Architectures = strings.Split(infos["Architecture"], ", ")
        verItem.ScreenDPI = infos["Screen DPI"]
        ch <- VerItemWithIndex{i, 0, verItem}
    } else {
        res2, err := http.Get(URL_BASE + nextUrl)
        if err != nil {
            return
        }
        defer res2.Body.Close()
        doc2, err := goquery.NewDocumentFromReader(res2.Body)
        if err != nil {
            return
        }
        doc2.Find(".table").Children().Next().Each(func(j int, sRow *goquery.Selection) {
            infos := make(map[string]string)
            sRow.Find(".ver-info-m p").Each(func(_ int, s2 *goquery.Selection) {
                y := strings.Split(s2.Text(), ":")
                if len(y) > 1 {
                    infos[y[0]] = strings.TrimSpace(y[1])
                } else {
                    dlLink, exists := s2.Find("a").First().Attr("href")
                    if exists {
                        infos["Download"] = dlLink
                    }
                }
            })
            verItem := VerItem{}
            verItem.Size = size
            verItem.Tags = tags
            verItem.Title = title
            verItem.DownloadUrl = URL_BASE + infos["Download"]
            verItem.Version = sRow.Find(".ver-info-top").Text()[len(verItem.Title) + 1:]
            verItem.UpdateOn = infos["Update on"]
            verItem.Signature = infos["Signature"]
            verItem.Sha1 = infos["File SHA1"]
            verItem.AndroidVer = infos["Requires Android"]
            verItem.Architectures = strings.Split(infos["Architecture"], ", ")
            verItem.ScreenDPI = infos["Screen DPI"]
            ch <- VerItemWithIndex{i, j,verItem}
        })
    }
}

func GetVersions(packageName string) ([]VerItem, error) {
    packageUrl, err := GetPackagePageUrl(packageName)
    if err != nil {
        return nil, err
    }
    url := packageUrl + "/versions"
    res, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        return nil, err
    }

    versionNodes := doc.Find(".ver-wrap li")
    var wg sync.WaitGroup
    verItemChan := make(chan VerItemWithIndex, versionNodes.Length() * 10)
    versionNodes.Each(func(i int, s *goquery.Selection) {
        wg.Add(1)
        go parseVerItem(verItemChan, &wg, i, s)
    })
    wg.Wait()
    close(verItemChan)

    var itemsWithIndex []VerItemWithIndex
    for tp := range verItemChan {
        itemsWithIndex = append(itemsWithIndex, tp)
    }

    sort.Slice(itemsWithIndex, func(i, j int) bool {
        if itemsWithIndex[i].i == itemsWithIndex[j].i {
            return itemsWithIndex[i].j < itemsWithIndex[j].j
        }
        return itemsWithIndex[i].i < itemsWithIndex[j].i
    })

    var verItems []VerItem
    for _, itemWithIndex := range itemsWithIndex {
        verItems = append(verItems, itemWithIndex.Item)
    }
    return verItems, nil
}
