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

func replicateBlueprintFromSource(blueprintDir, sourceDir, destinationDir string) error {
	// --- Phase 1: Index all files in the source directory for fast lookup ---
	fmt.Println("--- Phase 1: Indexing all files in the source directory... ---")
	sourceFileIndex := make(map[string]string) // Map of filename -> full path

	walkSourceErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We only care about files, not directories
		if !info.IsDir() {
			// Add the file to our index. If a filename is duplicated in the source,
			// the last one found will be used.
			sourceFileIndex[info.Name()] = path
		}
		return nil
	})

	if walkSourceErr != nil {
		return fmt.Errorf("error indexing source directory '%s': %w", sourceDir, walkSourceErr)
	}
	fmt.Printf("Found and indexed %d unique filenames in the source.\n", len(sourceFileIndex))

	// --- Phase 2: Walk the blueprint, find matches in our index, and copy ---
	fmt.Println("\n--- Phase 2: Processing blueprint and copying files... ---")
	var copiedCount int
	var notFoundInSource []string

	walkBlueprintErr := filepath.Walk(blueprintDir, func(blueprintPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We only use files from the blueprint as our guide
		if info.IsDir() {
			return nil
		}

		fileName := info.Name()

		// Look for this filename in our source index
		sourcePath, found := sourceFileIndex[fileName]

		if !found {
			// The file from the blueprint doesn't exist anywhere in the source directory
			log.Printf("SKIPPED: File '%s' from blueprint not found in source index.\n", fileName)
			notFoundInSource = append(notFoundInSource, fileName)
			return nil
		}

		// --- We found a match, now let's construct the destination path ---
		// Get the relative path of the file from the root of the blueprint directory
        relativePath, err := filepath.Rel(blueprintDir, blueprintPath)
        if err != nil {
            return fmt.Errorf("could not determine relative path for %s:%w", blueprintPath, err)
        }

        destinationPath := filepath.Join(destinationDir,relativePath)

        if err:= os.MkdirAll(filepath.Dir(destinationPath), os.ModePerm); err!= nil{
            return fmt.Errorf("Could not create sub directory %s: %w", filepath.Dir(destinationPath), err) 
        }


		// --- Copy the file ---
		fmt.Printf("COPYING: %s  ->  %s\n", sourcePath, destinationPath)
		if err := copyFile(sourcePath, destinationPath); err != nil {
			log.Printf("ERROR: Failed to copy file %s: %v", fileName, err)
		} else {
			copiedCount++
		}

		return nil
	})

	if walkBlueprintErr != nil {
		return fmt.Errorf("error processing blueprint directory '%s': %w", blueprintDir, walkBlueprintErr)
	}

	// --- Final Summary ---
	fmt.Println("\n----------------------")
	fmt.Println("Operation Completed.")
	fmt.Printf("Successfully copied %d files.\n", copiedCount)

	if len(notFoundInSource) > 0 {
		fmt.Println("\nThe following files from the blueprint were NOT found in the source directory:")
		for _, file := range notFoundInSource {
			fmt.Printf("- %s\n", file)
		}
	}

	return nil
}
