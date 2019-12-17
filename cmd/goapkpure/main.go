package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/adamyordan/goapkpure"
    "log"
    "strings"
)

func printJson(obj interface{}) {
    prettyJSON, err := json.MarshalIndent(obj, "", "    ")
    if err != nil {
        log.Fatal("Failed to generate json", err)
    }
    fmt.Printf("%s\n", string(prettyJSON))
}

func printVersions(packageName string, inJson bool) {
    verItems, err := goapkpure.GetVersions(packageName)
    if err != nil {
        log.Fatalln(err)
    }
    if inJson {
        printJson(verItems)
    } else {
        for i, verItem := range verItems {
            tags := strings.Join(verItem.Tags, ",")
            arch := strings.Join(verItem.Architectures, ",")
            fmt.Printf("%3d | version=%-16s | updateOn=%s | tags=%-10s | size=%-8s | arch=%-16s | downloadUrl=%s\n",
                i, verItem.Version, verItem.UpdateOn, tags, verItem.Size, arch, verItem.DownloadUrl)
        }
    }
}

func getVersionIndex(verItems []goapkpure.VerItem, versionId int, versionName string) (int, error){
    verIndex := versionId
    if versionName != "" {
        found := false
        for i, verItem := range verItems {
            if verItem.Version == versionName {
                verIndex = i
                found = true
                break
            }
        }
        if !found {
            return 0, fmt.Errorf("unable to find specified version name")
        }
    }
    if versionId > len(verItems) {
        return 0, fmt.Errorf("version id exceeds version length")
    }
    return verIndex, nil
}

func main() {
    packagePtr := flag.String("package", "", "Package name")
    versionsPtr := flag.Bool("versions", false, "List all available versions")
    versionPtr := flag.Int("version", 0, "Select a version by index")
    versionNamePtr := flag.String("versionName", "", "Select a version by version name. " +
        "Get the first one if duplicate. if set, -version flag is ignored.")
    downloadPtr := flag.Bool("download", false, "Download APK from selected version. " +
        "By default download latest version. Use -version or -versionName to specify version")
    outputPtr := flag.String("output", "", "Download file output. (default: <packagename>.apk)")
    jsonPtr := flag.Bool("json", false, "Print information in JSON format.")

    flag.Parse()

    if *packagePtr == "" {
        log.Fatalln("Please specify package name with -package flag")
    }

    if *versionsPtr {
        printVersions(*packagePtr, *jsonPtr)
    } else if *jsonPtr {
        verItems, err := goapkpure.GetVersions(*packagePtr)
        if err != nil {
            log.Fatalln(err)
        }
        versionIndex, err := getVersionIndex(verItems, *versionPtr, *versionNamePtr)
        if err != nil {
            log.Fatalln(err)
        }
        printJson(verItems[versionIndex])
    } else {
        var dlLink string
        var err error
        if *versionPtr >= 0 {
            verItems, err := goapkpure.GetVersions(*packagePtr)
            if err != nil {
                log.Fatalln(err)
            }
            versionIndex, err := getVersionIndex(verItems, *versionPtr, *versionNamePtr)
            if err != nil {
                log.Fatalln(err)
            }
            dlLink, err = goapkpure.GetDownloadDirectLink(verItems[versionIndex].DownloadUrl)
            if err != nil {
                log.Fatalln(err)
            }
        } else {
            dlLink, err = goapkpure.GetLatestDownloadLink(*packagePtr)
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
            goapkpure.DownloadFile(dlLink, outputFile)
        }
    }
}
