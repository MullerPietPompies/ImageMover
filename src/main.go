package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	fmt.Println("Welcome to the Image Mover Util!")
	fmt.Println("--------------------------------")

	fmt.Println("Enter The Excel File Path: ")

	var path string
	fmt.Scanln(&path)
	imageList := getImageList(path)

	var imageDir string
	fmt.Println("Select image directory")
	fmt.Scanln(&imageDir)

	var destinationPath string
	fmt.Println("Choose Destination Folder: ")
	fmt.Scanln(&destinationPath)

	fmt.Printf("Moving Images! \n")
	moveFiles(imageList, imageDir, destinationPath)

	fmt.Println("----------------------")
	fmt.Println("Thank you for using this utility")
}

func getImageList(path string) []string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Invalid file path: %v", err)
	}

	var sheet string
	fmt.Println("Enter Sheet Name")
	fmt.Scanln(&sheet)

	fmt.Println("Reading File")
	file, err := excelize.OpenFile(absPath)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := file.GetRows(sheet)
	if err != nil {
		log.Fatal(err)
	}

	var imageList []string

	fmt.Println("Reading Sheet")
	fmt.Println("Files That will be Copied!")
	fmt.Println("---------------------------------")
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) > 0 {
			imageList = append(imageList, rows[i][0])
			fmt.Print(rows[i][0], "\n")
		}
	}

	return imageList
}

func moveFiles(imageList []string, imageDir string, destinationPath string) {
	absImageDir, err := filepath.Abs(imageDir)
	if err != nil {
		log.Fatalf("Error getting absolute Image path: %v", err)
	}

	absDestination, err := filepath.Abs(destinationPath)
	if err != nil {
		log.Fatalf("Error getting absolute Destination path: %v", err)
	}

	if _, err := os.Stat(absDestination); os.IsNotExist(err) {
		err = os.MkdirAll(absDestination, os.ModePerm)
		if err != nil {
			log.Fatalf("Error Creating destination directory: %v", err)
		}
	}

	foundFiles := make(map[string]bool)
	notFoundFiles := make([]string, 0)

	err = filepath.Walk(absImageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, imgName := range imageList {
			if filepath.Base(path) == imgName && !foundFiles[imgName] {
				destPath := filepath.Join(absDestination, imgName)

				//Copy File

				sourceFile, err := os.Open(path)
				if err != nil {
					fmt.Printf("Error opening source file %s: %v\n", imgName, err)
					continue
				}
				defer sourceFile.Close()

				destFile, err := os.Create(destPath)
				if err != nil {
					fmt.Printf("Error creating destination file %s: %v\n", imgName, err)
					continue
				}
				defer destFile.Close()

				_, err = io.Copy(destFile, sourceFile)
				if err != nil {
					fmt.Printf("Error Copying file %s: %v\n", imgName, err)
					continue
				}
				fmt.Printf("Copied: %s (from %s)\n", imgName, path)
				foundFiles[imgName] = true
				break
			}
		}
		return nil
	})
	for _, imgName := range imageList {
		if !foundFiles[imgName] {
			notFoundFiles = append(notFoundFiles, imgName)
		}
	}
	if len(notFoundFiles) > 0 {
		fmt.Println("\n Files not found")
		for _, file := range notFoundFiles {
			fmt.Println(file)
		}
	}

	if err != nil {
		log.Fatalf("Error While walking through directories: %v", err)
	}
}
