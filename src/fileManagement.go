package main
import (
    "path/filepath"
    "log"
	"github.com/xuri/excelize/v2"
    "fmt"
    "os"
    "io"
)

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

func moveFiles(imageList []string, imageDir string, destinationPath string) error {
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
    // Phase 1

    dirsToCopy := make(map[string]bool)
    foundFiles := make(map[string]bool)

    fmt.Println("Scanning source for files")
    walkErr := filepath.Walk(absImageDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        baseName := filepath.Base(path)
        for _,imgName := range  imageList{
            if baseName == imgName {
                parentDir := filepath.Dir(path)
                dirsToCopy[parentDir] = true
                foundFiles[imgName] = true
                break
            }
        }
        return nil
    })

    if walkErr != nil{
        return fmt.Errorf("error while scanning directories: %w", walkErr)
    }
    // Phase 2: Copy directories

    if len(dirsToCopy) == 0{
        fmt.Println("No directories containting matching images were found")
    }else {
        fmt.Printf("Found &d directories to copy. Starting copy process...\n", len(dirsToCopy))
        for srcDir := range dirsToCopy{

            relativeDir, err := filepath.Rel(absImageDir, srcDir)

            if err!= nil{
                log.Printf("Could not determine relative path for %s:%v. Skipping.", srcDir, err)
                continue
            }
    		destDir := filepath.Join(absDestination, relativeDir)
			
			fmt.Printf("Copying directory: %s  ->  %s\n", srcDir, destDir)
			err = copyDir(srcDir, destDir)
			if err != nil {
				log.Printf("Error copying directory %s: %v", srcDir, err)
				// Decide if you want to stop on error or continue
				// return fmt.Errorf("failed to copy directory %s: %w", srcDir, err)
			}
		}
	}
	
	// --- Phase 3: Report any files that were not found at all ---
	var notFoundFiles []string
	for _, imgName := range imageList {
		if !foundFiles[imgName] {
			notFoundFiles = append(notFoundFiles, imgName)
		}
	}
	if len(notFoundFiles) > 0 {
		fmt.Println("\nFiles from the list that were not found anywhere:")
		for _, file := range notFoundFiles {
			fmt.Println(file)
		}
	}

	fmt.Println("----------------------")
	fmt.Println("Operation completed.")
	return nil // Success
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
