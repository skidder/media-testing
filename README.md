# Media Testing Tool

This tool is designed to automate the testing of various media files using the Lilliput library. It processes both graphical and non-graphical files, performing different checks and transformations based on the file type.

## Installation

1. Ensure you have Go installed on your system (version 1.17 or later).
2. Clone this repository:
   ```
   git clone https://github.com/skidder/media-testing.git
   cd media-testing
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```

## Usage

Run the tool using the following command:

```
go run main.go <path_to_test_files> <path_to_output_dir> [expected_failures_pattern]
```

- `<path_to_test_files>`: Directory containing the media files to be tested.
- `<path_to_output_dir>`: Directory where transformed files will be saved.
- `[expected_failures_pattern]`: (Optional) Wildcard pattern for files expected to fail the tests.

Example:
```
go run main.go ./testdata ./output '*corrupt*'
```

## Checks Performed

### Graphical Files (mp4, webm, jpg, jpeg, png, gif, webp, tiff, bmp)

1. File reading and decoding
2. Header information extraction
3. Dimension verification
4. Animation detection
5. Resizing to 50% of original dimensions
6. Conversion to WebP format
7. For animated files, generation of both animated and non-animated WebP outputs
8. Verification of resized image dimensions

### Non-Graphical Files (mp3, ogg, aac, flac, wav)

1. File reading and decoding
2. Header information extraction
3. Format identification
4. Duration check

### Specific Checks

- Video files (mp4, webm): Checks for zero dimensions and non-positive duration
- Audio files (mp3, ogg, flac, wav): Checks for non-positive duration
- AAC files: Checks for non-positive duration (noted separately due to potential negative durations)
- WebP files: Checks for zero dimensions

## Output

The tool will print information about each file processed, including any warnings or errors encountered. Transformed graphical files will be saved in the specified output directory.

If any tests fail (and are not marked as expected failures), the program will exit with a non-zero status code.

## Expected Failures

You can specify a wildcard pattern for files that are expected to fail. These failures will be noted but won't cause the overall test run to fail. This is useful for testing the tool's behavior with known problematic files.

## Dependencies

This tool uses the [Lilliput](https://github.com/discord/lilliput) library for image and video processing.
