package zero

import (
	"errors"
	"image"
	_ "image/jpeg" // need to fetch supported formats
	_ "image/png"  // png as well
	"io"
	"mime/multipart"
	"os"
	//"github.com/rwcarlsen/goexif/exif"
)

// File is a basic type to operate files
type File struct {
	//Handler *multipart.FileHeader
	reader        io.ReadCloser
	ContentLength int64
	open          func() error
	name          string
}

// SetMultipart will file open handler using multipart filehandler
func (f *File) SetMultipart(handler *multipart.FileHeader) error {
	f.open = func() error {
		file, err := handler.Open()
		if err != nil {
			return err
		}
		f.reader = file
		f.ContentLength = handler.Size
		return nil
	}
	f.name = handler.Filename
	return nil
}

// GetFileName return name of the file
func (f *File) GetFileName() string {
	return f.name
}

// SetReadCloser sets the reader, and handles the closer
func (f *File) SetReadCloser(name string, contentLength int64, closer io.ReadCloser) {
	f.reader = closer
	f.ContentLength = contentLength
}

// GetReadCloser gets the reader
func (f *File) GetReadCloser() (io.ReadCloser, error) {
	var err error
	if f.open != nil {
		err = f.open()
	}
	return f.reader, err
}

// Save saves file on the disk
func (f *File) Save(path string) error {
	//file, err := f.Handler.Open()
	//defer file.Close()
	if f.open != nil {
		openErr := f.open()
		if openErr != nil {
			return openErr
		}
	}

	defer f.reader.Close()
	//if err != nil {
	//	return errors.New("File open error")
	//}
	dst, err := os.Create(path)
	defer dst.Close()
	if err != nil {
		return errors.New("Saved file create error")
	}
	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, f.reader); err != nil {
		return errors.New("Copy file create error")
	}
	return nil
}

// GetImageDimension return image dimensions
func (f *File) GetImageDimension() (int, int, error) {
	if f.open != nil {
		openErr := f.open()
		if openErr != nil {
			return 0, 0, openErr
		}
	}
	image, _, err := image.DecodeConfig(f.reader)

	if err != nil {
		return 0, 0, err
	}
	return image.Width, image.Height, nil
}
