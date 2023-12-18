package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

type ImageInfo struct {
	Digest     string
	Repository string
}

type OCIIndex struct {
	Manifests []OCIManifest `json:"manifests"`
}
type OCIManifest struct {
	MediaType string `json:"mediaType"`
	Size      uint   `json:"size"`
	Digest    string `json:"digest"`
}

const repoSearchOn = "readonly REPOSITORY_FILE=\""

func getImageInfo(manifest string) ImageInfo {
	file, err := os.ReadFile(manifest)
	if err != nil {
		log.Fatalf("could not open %s %v", manifest, err)
	}

	paths := make([]string, 0)

	err = json.Unmarshal(file, &paths)
	if err != nil {
		log.Fatalf("could not read manifest %s %v", manifest, err)
	}

	imageInfo := ImageInfo{}

	for _, path := range paths {
		pathFile, err := os.Open(path)
		if err != nil {
			log.Fatalf("could not open %s %v", path, err)
		}
		stat, _ := pathFile.Stat()

		// this our push file
		if strings.HasSuffix(path, ".sh") {
			scanner := bufio.NewScanner(pathFile)
			for scanner.Scan() {
				text := scanner.Text()
				if strings.HasPrefix(text, repoSearchOn) {
					image, found := strings.CutPrefix(text, repoSearchOn)
					if found {
						repoPath, _ := strings.CutSuffix(image, "\"")
						repoFile, err := os.ReadFile(repoPath)
						if err != nil {
							log.Fatalf("could not read repo file %s %v", repoPath, err)
						}
						imageInfo.Repository = strings.ReplaceAll(string(repoFile), "\n", "")
					}
				}
			}
		}

		// Our image
		if stat.IsDir() {
			indexFile, err := os.ReadFile(fmt.Sprintf("%s/index.json", path))
			if err != nil {
				log.Fatalf("could not open %s/index.json %v", path, err)
			}

			ociIndex := OCIIndex{}

			_ = json.Unmarshal(indexFile, &ociIndex)

			imageInfo.Digest = ociIndex.Manifests[0].Digest
		}

	}

	return imageInfo
}

func writeTemplate(path, tpl string, data ImageInfo) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Cannot create %s file", path)
	}

	parts := strings.Split(tpl, "/")
	templateName := parts[len(parts)-1]

	t := template.Must(template.New(templateName).ParseFiles(tpl))
	err = t.Execute(f, data)
	if err != nil {
		log.Fatalf("Cannot generate %s template, %v", tpl, err)
	}
}

func main() {
	imageTemplate := flag.String("template", "", "Template file to populate")
	outPath := flag.String("out", "", "Output path")
	manifest := flag.String("manifest", "", "Image manifest")
	flag.Parse()

	imageInfo := getImageInfo(*manifest)
	if imageInfo.Digest == "" || imageInfo.Repository == "" {
		log.Fatalf("invalid image info: %v", imageInfo)
	}

	writeTemplate(*outPath, *imageTemplate, imageInfo)
}
