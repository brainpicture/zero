package zero

import (
	"errors"
	"image"
	_ "image/jpeg" // need to fetch supported formats
	_ "image/png"  // png as well
	"io"
	"mime/multipart"
	"os"
)

// File is a basic type to operate files
type File struct {
	Handler *multipart.FileHeader
}

// Save saves file on the disk
func (f *File) Save(path string) error {
	file, err := f.Handler.Open()
	defer file.Close()
	if err != nil {
		return errors.New("File open error")
	}
	dst, err := os.Create(path)
	defer dst.Close()
	if err != nil {
		return errors.New("Saved file create error")
	}
	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return errors.New("Copy file create error")
	}
	return nil
}

// GetImageDimension return image dimensions
func (f *File) GetImageDimension() (int, int, error) {
	file, err := f.Handler.Open()
	defer file.Close()
	if err != nil {
		return 0, 0, errors.New("File open error")
	}

	image, _, err := image.DecodeConfig(file)

	if err != nil {
		return 0, 0, err
	}
	return image.Width, image.Height, nil
}
