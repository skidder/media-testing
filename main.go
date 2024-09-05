package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/discord/lilliput"
)

var graphicalFormats = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".tiff": true, ".bmp": true,
}

var testsFailed bool
var expectedFailures string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <path_to_test_files> <path_to_output_dir> [expected_failures_pattern]")
		os.Exit(1)
	}

	testDir := os.Args[1]
	outDir := os.Args[2]
	if len(os.Args) > 3 {
		expectedFailures = os.Args[3]
	}

	processTestFiles(testDir, outDir)

	if testsFailed {
		fmt.Println("Some tests or transformations failed.")
		os.Exit(1)
	} else {
		fmt.Println("All tests passed successfully.")
	}
}

func processTestFiles(dir string, outDir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	fileGroups := make(map[string][]string)

	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			fileGroups[ext] = append(fileGroups[ext], filepath.Join(dir, file.Name()))
		}
	}

	for ext, files := range fileGroups {
		fmt.Printf("Processing %s files:\n", ext)
		for _, file := range files {
			if graphicalFormats[ext] {
				testGraphicalFile(file, outDir)
			} else {
				testNonGraphicalFile(file, outDir)
			}
		}
		fmt.Println()
	}
}

func isExpectedFailure(filePath string) bool {
	if expectedFailures == "" {
		return false
	}
	match, err := filepath.Match(expectedFailures, filepath.Base(filePath))
	if err != nil {
		log.Printf("Error matching pattern: %v\n", err)
		return false
	}
	return match
}

func testGraphicalFile(filePath string, outDir string) {
	fmt.Printf("Testing graphical file: %s\n", filePath)
	expectedToFail := isExpectedFailure(filePath)
	failed := false

	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		failed = true
	} else {
		decoder, err := lilliput.NewDecoder(buffer)
		if err != nil {
			log.Printf("Error creating decoder for %s: %v\n", filePath, err)
			failed = true
		} else {
			defer decoder.Close()

			header, err := decoder.Header()
			if err != nil {
				log.Printf("Error reading header for %s: %v\n", filePath, err)
				failed = true
			} else {
				fmt.Printf("  Format: %s\n", decoder.Description())
				fmt.Printf("  Dimensions: %d x %d\n", header.Width(), header.Height())
				fmt.Printf("  Animated: %v\n", header.IsAnimated())

				ops := lilliput.NewImageOps(8192)
				defer ops.Close()

				resizeWidth := header.Width() / 2
				resizeHeight := header.Height() / 2

				// Define common options
				commonOptions := &lilliput.ImageOptions{
					FileType:              ".webp",
					Width:                 resizeWidth,
					Height:                resizeHeight,
					ResizeMethod:          lilliput.ImageOpsFit,
					NormalizeOrientation:  true,
					EncodeOptions:         map[int]int{lilliput.WebpQuality: 90},
					EncodeTimeout:         time.Second * 300,
					DisableAnimatedOutput: true,
				}

				// Perform non-animated transformation
				if err := performTransform(decoder, ops, commonOptions, filePath, outDir, "_resized.webp"); err != nil {
					failed = true
				}

				// For animated formats, also generate an animated output
				if header.IsAnimated() {
					// reinitialize the decoder
					decoder, err = lilliput.NewDecoder(buffer)
					if err != nil {
						log.Printf("Error creating decoder for animated output %s: %v\n", filePath, err)
						failed = true
					} else {
						defer decoder.Close()
						animatedOptions := *commonOptions
						animatedOptions.DisableAnimatedOutput = false
						if err := performTransform(decoder, ops, &animatedOptions, filePath, outDir, "_resized_animated.webp"); err != nil {
							failed = true
						}
					}
				}
			}
		}
	}

	if failed {
		if !expectedToFail {
			testsFailed = true
		} else {
			fmt.Println("  Note: This failure was expected")
		}
	} else {
		fmt.Println("  Test completed successfully")
	}
}

func performTransform(decoder lilliput.Decoder, ops *lilliput.ImageOps, options *lilliput.ImageOptions, filePath, outDir, outputSuffix string) error {
	resizedBuffer := make([]byte, 50*1024*1024) // 50MB buffer
	resizedBuffer, err := ops.Transform(decoder, options, resizedBuffer)
	if err != nil {
		log.Printf("Error resizing %s: %v\n", filePath, err)
		return err
	}

	// Verify dimensions of resized image
	resizedDecoder, err := lilliput.NewDecoder(resizedBuffer)
	if err != nil {
		log.Printf("Error creating decoder for resized image: %v\n", err)
		return err
	}
	defer resizedDecoder.Close()

	resizedHeader, err := resizedDecoder.Header()
	if err != nil {
		log.Printf("Error reading header of resized image: %v\n", err)
		return err
	}

	if resizedHeader.Width() != options.Width || resizedHeader.Height() != options.Height {
		log.Printf("Warning: Resized dimensions (%dx%d) do not match specified dimensions (%dx%d)\n",
			resizedHeader.Width(), resizedHeader.Height(), options.Width, options.Height)
		return fmt.Errorf("resized dimensions do not match specified dimensions")
	}

	fmt.Printf("  Resized dimensions match: %dx%d\n", resizedHeader.Width(), resizedHeader.Height())

	// write to outDir using the original filename
	outputPath := filepath.Join(outDir, filepath.Base(filePath)+outputSuffix)
	if err := ioutil.WriteFile(outputPath, resizedBuffer, 0644); err != nil {
		log.Printf("Error writing output file %s: %v\n", outputPath, err)
		return err
	}

	fmt.Printf("  Resized and saved to: %s\n", outputPath)
	return nil
}

func testNonGraphicalFile(filePath string, outDir string) {
	fmt.Printf("Inspecting file: %s\n", filePath)
	expectedToFail := isExpectedFailure(filePath)
	failed := false

	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		failed = true
	} else {
		decoder, err := lilliput.NewDecoder(buffer)
		if err != nil {
			log.Printf("Error creating decoder for %s: %v\n", filePath, err)
			fmt.Println("  Unable to decode file with Lilliput")
			failed = true
		} else {
			defer decoder.Close()

			header, err := decoder.Header()
			if err != nil {
				log.Printf("Error reading header for %s: %v\n", filePath, err)
				fmt.Println("  Unable to read file header with Lilliput")
				failed = true
			} else {
				fmt.Printf("  Format: %s\n", decoder.Description())
				fmt.Printf("  Dimensions: %d x %d\n", header.Width(), header.Height())
				fmt.Printf("  Duration: %v\n", decoder.Duration())
				fmt.Printf("  Animated: %v\n", header.IsAnimated())

				if decoder.Duration() < 0 {
					fmt.Println("  Note: Negative duration (typical for images)")
				}

				// Additional checks based on file type
				ext := strings.ToLower(filepath.Ext(filePath))
				switch ext {
				case ".mp4", ".webm":
					if header.Width() == 0 || header.Height() == 0 {
						fmt.Println("  Warning: Video file has zero dimensions")
						failed = true
					}
					if decoder.Duration() <= 0 {
						fmt.Println("  Warning: Video file has non-positive duration")
						failed = true
					}
				case ".mp3", ".ogg", ".flac", ".wav":
					if decoder.Duration() <= 0 {
						fmt.Println("  Warning: Audio file has non-positive duration")
						failed = true
					}
				case ".aac":
					if decoder.Duration() <= 0 {
						fmt.Println("  Warning: AAC audio file has non-positive duration")
					}
				case ".webp":
					if header.Width() == 0 || header.Height() == 0 {
						fmt.Println("  Warning: WebP file has zero dimensions")
						failed = true
					}
				}
			}
		}
	}

	if failed {
		if !expectedToFail {
			testsFailed = true
		} else {
			fmt.Println("  Note: This failure was expected")
		}
	}

	fmt.Println("  Inspection completed")
}
