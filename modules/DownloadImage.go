package modules

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// isImage checks for correct image type and returns true
// ifit matches the Image related content types.
func isImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg":
		return true
	case "image/png":
		return true
	case "image/bmp":
		return true
	default:
		return false
	}
}

// unsupported file response
var errUnsportedFile = errors.New("The file cannot be set as wallpaper. It is not an image file")
var errFileExists = errors.New("File already exists")

// downloadImage function is used to download image
func DownloadImage(imgURL, nameOfImage, dirname string) (string, error) {

	filename := filepath.Join(dirname, nameOfImage+".jpg")

	fmt.Println("The filename is: ", filename)

	if _, err := os.Stat(filename); err == nil {
		fmt.Println("File already exists")
		return filename, errFileExists
	}

	// send http get request to fetch the resource (Image).
	resp, err := http.Get(imgURL)
	if err != nil {
		return "", err
	}
	// get the url returned from response's request
	// so that if there was a shortned url it will be the final url of image file.
	respURL := resp.Request.URL.String()
	fmt.Println("URL: ", respURL)

	// get the content type of response
	contentType := resp.Header.Get("Content-Type")
	isImage := isImageType(contentType)
	if !isImage { // if the response if not an image the return an error!
		return "", errUnsportedFile
	}

	defer resp.Body.Close()

	// Create the image file at the given path
	file, err := os.Create(filename)

	if err != nil {
		fmt.Println("Error while creating file")
		return "", err
	}
	// close the file when function ends
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error while copying response contents to file")
		return "", err
	}
	fmt.Println("File downloaded successfully!")

	return filename, nil
}
