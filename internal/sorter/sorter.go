package sorter

import (
	"fmt"
	"io" // Import io for copying
	"os"
	"path/filepath" // Import runtime for OS check
	"strconv"
	"strings"

	"github.com/dhowden/tag" // Import golnk for shortcut creation
)

type Sorter struct {
	FolderPath        string
	ScriptPath        string
	NewLibraryPath    string
	FileOperationMode string   // Add file operation mode field
	Logs              []string // Add Logs field
}

func NewSorter(folderPath, scriptPath, newLibraryPath, fileOperationMode string) *Sorter {
	// Default to "preview" if mode is empty or invalid
	if fileOperationMode == "" || (fileOperationMode != "move" && fileOperationMode != "copy" && fileOperationMode != "preview") {
		fileOperationMode = "preview" // Default to preview
	}
	return &Sorter{
		FolderPath:        folderPath,
		ScriptPath:        scriptPath,
		NewLibraryPath:    newLibraryPath,
		FileOperationMode: fileOperationMode, // Store the mode
		Logs:              make([]string, 0), // Initialize Logs slice
	}
}

// ExecuteSort now returns logs and an error
func (s *Sorter) ExecuteSort() ([]string, error) {
	s.Logs = append(s.Logs, fmt.Sprintf("Starting sort process. Source: %s, Destination: %s, Mode: %s", s.FolderPath, s.NewLibraryPath, s.FileOperationMode))
	script, err := os.ReadFile(s.ScriptPath)
	if err != nil {
		s.Logs = append(s.Logs, fmt.Sprintf("Error reading script file %s: %v", s.ScriptPath, err))
		return s.Logs, err
	}
	s.Logs = append(s.Logs, fmt.Sprintf("Successfully read script file: %s", s.ScriptPath))

	cleanedScript := cleanScript(string(script))
	instructions := strings.Split(cleanedScript, "\n")

	audioFiles, err := s.getAllAudioFiles(s.FolderPath) // Call method on s
	if err != nil {
		s.Logs = append(s.Logs, fmt.Sprintf("Error getting audio files from %s: %v", s.FolderPath, err))
		return s.Logs, err
	}
	s.Logs = append(s.Logs, fmt.Sprintf("Found %d audio files in %s", len(audioFiles), s.FolderPath))

	newLibraryMap := make(map[string]string)
	processedCount := 0
	completedCount := 0
	errorCount := 0

	for _, audioFile := range audioFiles {
		fullPath := filepath.Join(s.FolderPath, audioFile)
		file, err := os.Open(fullPath)
		if err != nil {
			s.Logs = append(s.Logs, fmt.Sprintf("Error opening file %s: %v", audioFile, err))
			errorCount++
			continue
		}

		m, err := tag.ReadFrom(file)
		file.Close()
		if err != nil {
			s.Logs = append(s.Logs, fmt.Sprintf("Error reading metadata from %s: %v", audioFile, err))
			errorCount++
			continue
		}

		path := getPath(instructions, m)
		if path == "" {
			s.Logs = append(s.Logs, fmt.Sprintf("Skipping file %s based on script rules (STOP or no path generated).", audioFile))
			continue // Skip if path is empty
		}

		// Determine destination path (without extension for now)
		baseDestinationPath := filepath.Join(s.NewLibraryPath, filepath.FromSlash(path))
		fileExt := filepath.Ext(audioFile)
		destinationPath := baseDestinationPath + fileExt // Final destination path for move/copy

		s.Logs = append(s.Logs, fmt.Sprintf("Processing: %s -> %s", audioFile, destinationPath))
		processedCount++

		// Create the new directory if it doesn't exist (only if not preview)
		if s.FileOperationMode != "preview" {
			newDir := filepath.Dir(destinationPath)
			if err := os.MkdirAll(newDir, os.ModePerm); err != nil {
				s.Logs = append(s.Logs, fmt.Sprintf("Error creating directory %s: %v", newDir, err))
				errorCount++
				continue
			}
		}

		// Perform file operation based on mode
		operationLog := ""
		var operationErr error // Use a separate variable for operation error

		switch s.FileOperationMode {
		case "move":
			operationErr = os.Rename(fullPath, destinationPath)
			if operationErr == nil {
				operationLog = fmt.Sprintf("Moved: %s -> %s", audioFile, destinationPath)
				completedCount++
			} else {
				operationLog = fmt.Sprintf("Error moving file %s to %s: %v", audioFile, destinationPath, operationErr)
				errorCount++
			}
		case "copy":
			operationErr = copyFile(fullPath, destinationPath)
			if operationErr == nil {
				operationLog = fmt.Sprintf("Copied: %s -> %s", audioFile, destinationPath)
				completedCount++
			} else {
				operationLog = fmt.Sprintf("Error copying file %s to %s: %v", audioFile, destinationPath, operationErr)
				errorCount++
			}
		case "preview":
			fallthrough // Fallthrough to default for logging, but no action
		default:
			// Log the intended action without performing it
			operationLog = fmt.Sprintf("Preview: Would move/copy %s -> %s", audioFile, destinationPath)
			// In preview mode, we count it as "completed" for the summary if no errors occurred *before* this step
			completedCount++
			operationErr = nil
		}
		s.Logs = append(s.Logs, operationLog)
		if operationErr != nil {
			continue // Skip to next file on actual operation error
		}

		// Store mapping (optional, might be useful later)
		newLibraryMap[audioFile] = destinationPath // Or shortcutPath if mode is shortcut
	}

	summaryMode := s.FileOperationMode
	if summaryMode == "preview" {
		summaryMode = "previewed" // Adjust summary text for preview
	}
	s.Logs = append(s.Logs, fmt.Sprintf("Sort process finished. Processed: %d, Completed (%s): %d, Errors: %d", processedCount, summaryMode, completedCount, errorCount))
	return s.Logs, nil // Return collected logs and nil error if process finished
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func cleanScript(script string) string {
	// remove the comments on each line (everything after the #)
	// remove the empty lines and the lines that are only spaces
	// replace the 4 space with "\t"
	// replace the ' with "
	// remove the trailing spaces (conserve leading spaces)

	var result = ""
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0] // Remove comments
		}
		line = strings.TrimRight(line, " ")
		line = strings.TrimRight(line, "\t")

		if strings.TrimSpace(line) == "" {
			continue
		}

		result += line + "\n"
	}
	result = strings.TrimRight(result, "\n")
	result = strings.ReplaceAll(result, "    ", "\t") // Replace "    " with "\n"
	result = strings.ReplaceAll(result, "'", "\"")    // Replace ' with "
	return result
}

func getPath(instructions []string, meta tag.Metadata) string {
	var level = 0
	var currentResult = ""
	for _, line := range instructions {
		var lineLevel = getLevel(line)
		if lineLevel <= level {
			// we can "climb down" a level any time
			level = lineLevel
		} else {
			// but we can't "climb up" a level.
			// It is used by conditionnal :
			// 		we can only go up a level and execute theses statements if
			// 		the IF condition was met
			continue
		}
		// we can now trim the level from the line
		line = strings.TrimLeft(line, "\t")

		if strings.HasPrefix(line, "IF") {
			// IF support
			condition := strings.TrimPrefix(line, "IF")
			condition = strings.TrimSpace(condition)
			if len(condition) > 2 && condition[0] == '(' && condition[len(condition)-1] == ')' {
				condition = condition[1 : len(condition)-1] // Remove parentheses
				if evaluateCondition(condition, meta) {
					level += 1
				}
			} else {
				// Fallback to original behavior or log error for malformed IF
				// Original behavior: check if tag exists
				if getTag(condition, meta) != "" {
					level += 1
				}
			}

		} else if strings.HasPrefix(line, "ADD FOLDER") {
			// ADD FOLDER support
			currentResult += string(filepath.Separator)
			level += 1

		} else if strings.HasPrefix(line, "\"") {
			// Input string "" support
			var toAdd = strings.Split(line, "\"")[1]
			toAdd = strings.Split(toAdd, "\"")[0]
			currentResult += toAdd

		} else if strings.HasPrefix(line, "STOP") {
			// STOP support
			return ""

		} else {
			// tag support
			currentResult += getTag(line, meta)
		}
	}
	return currentResult
}

func evaluateCondition(condition string, meta tag.Metadata) bool {
	// Check for "is number"
	if strings.Contains(condition, " is number") {
		parts := strings.SplitN(condition, " is number", 2)
		tagName := strings.TrimSpace(parts[0])
		tagValue := getTag(tagName, meta)
		if tagValue == "" {
			return false // Tag doesn't exist
		}
		_, err := strconv.Atoi(tagValue)
		return err == nil // True if conversion to int is successful
	}

	// Check for "!="
	if strings.Contains(condition, "!=") {
		parts := strings.SplitN(condition, "!=", 2)
		tagName := strings.TrimSpace(parts[0])
		expectedValue := strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(expectedValue) >= 2 && expectedValue[0] == '"' && expectedValue[len(expectedValue)-1] == '"' {
			expectedValue = expectedValue[1 : len(expectedValue)-1]
		}
		actualValue := getTag(tagName, meta)
		return actualValue != expectedValue
	}

	// Check for "=="
	if strings.Contains(condition, "==") {
		parts := strings.SplitN(condition, "==", 2)
		tagName := strings.TrimSpace(parts[0])
		expectedValue := strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(expectedValue) >= 2 && expectedValue[0] == '"' && expectedValue[len(expectedValue)-1] == '"' {
			expectedValue = expectedValue[1 : len(expectedValue)-1]
		}
		actualValue := getTag(tagName, meta)
		return actualValue == expectedValue
	}

	// Default: Check if tag exists (original behavior for simple IF(TAG))
	tagName := strings.TrimSpace(condition)
	return getTag(tagName, meta) != ""
}

func getLevel(instruction string) int {
	// count the number of \t trailing
	var result = 0
	for strings.HasPrefix(instruction, "\t") {
		instruction = instruction[1:]
		result += 1
	}
	return result
}

// getAllAudioFiles should be a method of Sorter to access s.Logs
func (s *Sorter) getAllAudioFiles(directory string) ([]string, error) {
	var audioFiles []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.Logs = append(s.Logs, fmt.Sprintf("Error accessing path %s: %v", path, err))
			return err // Propagate the error
		}
		if !info.IsDir() && isAudioFile(info.Name()) {
			// Store relative path from base directory
			relPath, errRel := filepath.Rel(directory, path)
			if errRel != nil {
				// Log error but continue if possible
				s.Logs = append(s.Logs, fmt.Sprintf("Error getting relative path for %s: %v", path, errRel))
				return nil // Decide if this error is critical
			}
			audioFiles = append(audioFiles, relPath)
		}
		return nil
	})

	if err != nil {
		s.Logs = append(s.Logs, fmt.Sprintf("Error walking through directory %s: %v", directory, err))
		return nil, err
	}

	return audioFiles, nil
}

func isAudioFile(fileName string) bool {
	parts := strings.Split(fileName, ".")
	if len(parts) < 2 {
		return false
	}
	ext := strings.ToLower(parts[len(parts)-1])
	switch ext {
	case "mp3", "ogg", "flac", "wav", "opus":
		return true
	default:
		return false
	}
}

func getTag(tagName string, m tag.Metadata) string {
	if strings.HasPrefix(tagName, "CUSTOM:") {
		// Custom tag support
		tagName = strings.TrimPrefix(tagName, "CUSTOM:")
		tagName = strings.TrimSpace(tagName)
		return extractCustomTag(m, tagName)
	}
	if strings.HasPrefix(tagName, "TXXX:") {
		// TXXX tag support
		tagName = strings.TrimPrefix(tagName, "TXXX:")
		tagName = strings.TrimSpace(tagName)
		return extractCustomTag(m, tagName)
	}
	tagName = strings.ToUpper(tagName)
	tagName = strings.TrimSpace(tagName)
	switch tagName {
	case "ARTIST":
		return m.Artist()
	case "ALBUM":
		return m.Album()
	case "TITLE":
		return m.Title()
	case "ALBUMARTIST":
		return m.AlbumArtist()
	case "COMPOSER":
		return m.Composer()
	case "YEAR":
		return strconv.Itoa(m.Year())
	case "GENRE":
		return m.Genre()
	case "TRACK":
		track, _ := m.Track()
		return strconv.Itoa(track)
	case "DISC":
		disc, _ := m.Disc()
		return strconv.Itoa(disc)
	case "PICTURE":
		if m.Picture() != nil {
			return m.Picture().String()
		}
		return ""
	case "LYRICS":
		return m.Lyrics()
	case "COMMENT":
		return m.Comment()
	default:
		println("Unknown tag:", tagName)
		return "" // Return an empty string if the tag does not exist
	}
}

func extractCustomTag(m tag.Metadata, tagName string) string {
	// we try 3 times : one with the passed tag name, one with the upper case and one with the lower case
	// because every format store tags differently and may or may not save capitalization.
	// for example the mp3 with id3v2.3 stores the custom tag as "GROUP"when it should simply be "group"
	if value := _extractCustomTag(m, tagName); value != "" {
		return value
	}
	if value := _extractCustomTag(m, strings.ToUpper(tagName)); value != "" {
		return value
	}
	if value := _extractCustomTag(m, strings.ToLower(tagName)); value != "" {
		return value
	}
	return ""
}

func _extractCustomTag(m tag.Metadata, tagName string) string {
	// Check if the custom tag exists in the metadata
	if value, ok := m.Raw()[tagName]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	// for id3v2.4 tags, the custom tag is stored in a different format
	for _, v := range m.Raw() {
		// Check if the value is of type *tag.Comm
		if t, ok := v.(*tag.Comm); ok {
			// and the description is the same as the custom tag name
			if t.Description == tagName {
				return t.Text
			}
		}
	}
	return "" // Return an empty string if the custom tag does not exist
}
