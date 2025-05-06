# Go Sort Music UI

## Overview
Go Sort Music UI is a user-friendly interface for the SortLibrary script, designed to help users easily sort their music libraries based on customizable parameters. This project integrates a web-based interface with the existing sorting logic, allowing for seamless interaction and execution of sorting operations.
![image](https://github.com/user-attachments/assets/d2a5557c-4130-4906-a0b5-43f910e798e1)

- **cmd/app/main.go**: Entry point of the application. Initializes the web server and sets up routes.
- **internal/sorter/sorter.go**: Contains the core logic for sorting the music library, utilizing the existing SortLibrary code.

## Setup Instructions
1. **Clone the Repository**
   ```
   git clone https://github.com/jere344/GoSortMusicLibrary.git
   cd GoSortMusicLibrary

   ```

1. **Install Dependencies**
   Ensure you have Go installed (https://go.dev/doc/install). Run the following command to download the necessary dependencies:
   ```
   go mod tidy
   ```

2. **Run the Application**
   Start the web server by executing:
   ```
   go run cmd/app/main.go
   ```
   The application will be accessible at `http://localhost:8080`.

## Usage
- Open your web browser and navigate to `http://localhost:8080`.
- Use the provided form to enter your sorting parameters.
- Submit the form to execute the sorting operation.
- View the execution logs displayed on the interface.
  
## Script
The script is interpreted line by line for every music file found.
Here are all the scripts available instructions:

### **ADD FOLDER**: 
Defines that we are in a folder creation context. The script always starts with this instruction, and the last one will correspond to the file name.

###  **IF (condition)**: 
Defines a conditional context. the incremented part following the condition will only be executed if the condition is true. The condition can be any of the following:
- **[TAG_NAME]** : if the tag exists and is not empty
- **[TAG_NAME] == "VALUE"** : if the tag exists and is equal to VALUE
- **[TAG_NAME] != "VALUE"** : if the tag exists and is not equal to VALUE
- **[TAG_NAME] is number** : if the tag exists and is a number

### **TAG_NAME**: 
Insert the value of the specified tag. Supported standard tags include:
- **ARTIST**: The artist name
- **ALBUM**: The album name
- **TITLE**: The track title
- **ALBUMARTIST**: The album artist name
- **COMPOSER**: The composer name
- **YEAR**: The release year
- **GENRE**: The genre
- **TRACK**: The track number
- **DISC**: The disc number
- **PICTURE**: Image data
- **LYRICS**: Lyrics text
- **COMMENT**: Comments

### **CUSTOM:TAG_NAME** or **TXXX:TAG_NAME**:
Access custom tags in your music files.

### **"TEXT"**: 
Insert literal text. For example, `"Disc "` will insert the text "Disc ".

### **STOP**:
Stops processing the current file. The file will be skipped and not included in the sorted library. Meant for conditionals, if you wish to skip a file based on a condition.

## Script Examples

### Basic Organization Example
```
ADD FOLDER
    ARTIST

ADD FOLDER
    ALBUM

ADD FOLDER
    TRACK
    ". "
    TITLE
```
This script will organize files into folders as: `Artist/Album/Track. Title`

### Advanced Example with Conditionals
```
ADD FOLDER
    GENRE

ADD FOLDER
    ARTIST

ADD FOLDER
    IF (YEAR)
        YEAR
        " - "
    ALBUM

ADD FOLDER
    IF (TRACK is number)
        TRACK
        ". "
    TITLE
```
This script will organize files into folders as: `Genre/Artist/[Year - ]Album/[Track. ]Title`


## License
This project is licensed under the MIT License. See the LICENSE file for more details.
