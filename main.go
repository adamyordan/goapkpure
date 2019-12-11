package main

import (
    "flag"
    "fmt"
    "log"
    "strings"
)

func main() {
    packagePtr := flag.String("package", "", "Package name")
    versionsPtr := flag.Bool("versions", false, "List all available versions")
    versionPtr := flag.Int("version", 0, "Get direct link for a version index")
    downloadPtr := flag.Bool("download", false, "Download APK from direct link. " +
        "By default download latest version. Use -version to specify version ID")
    outputPtr := flag.String("output", "", "Download file output. (default: <packagename>.apk)")

    flag.Parse()

    if *packagePtr == "" {
        log.Fatalln("Please specify package name with -package flag")
    }

    if *versionsPtr {
        verItems, err := GetVersions(*packagePtr)
        if err != nil {
            log.Fatalln(err)
        }
        for i, verItem := range verItems {
            tags := strings.Join(verItem.Tags, ",")
            arch := strings.Join(verItem.Architectures, ",")
            fmt.Printf("%3d | version=%-16s | updateOn=%s | tags=%-10s | size=%-8s | arch=%-16s | downloadUrl=%s\n",
                i, verItem.Version, verItem.UpdateOn, tags, verItem.Size, arch, verItem.DownloadUrl)
        }
    } else {
        var dlLink string
        var err error
        if *versionPtr >= 0 {
            verItems, err := GetVersions(*packagePtr)
            if err != nil {
                log.Fatalln(err)
            }
            if *versionPtr >= len(verItems) {
                log.Fatalln("version ID invalid")
            }
            dlLink, err = GetDownloadDirectLink(verItems[*versionPtr].DownloadUrl)
            if err != nil {
                log.Fatalln(err)
            }
        } else {
            dlLink, err = GetLatestDownloadLink(*packagePtr)
            if err != nil {
                log.Fatalln(err)
            }
        }

        if !*downloadPtr {
            fmt.Println(dlLink)
        } else {
            outputFile := *outputPtr
            if outputFile == "" {
                outputFile = *packagePtr + ".apk"
            }
            DownloadFile(dlLink, outputFile)
        }
    }
}
