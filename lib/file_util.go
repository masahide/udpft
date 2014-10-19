package lib

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const MaxDepth = 20

const Bufsize = 1 * 1024 * 1024

type DataBlock struct {
	id   int
	size int
	data []byte
}

func FileRead(done <-chan struct{}, path string) (<-chan DataBlock, <-chan error) {
	datas := make(chan DataBlock, 2)
	errc := make(chan error, 1)
	go func() {
		defer close(datas)
		f, err := os.Open(path)
		if err != nil {
			errc <- err
			return
		}
		for id := 0; ; id++ {
			buf := make([]byte, Bufsize)
			size, readError := f.Read(buf)
			data := DataBlock{id: id, size: size, data: buf[0:size]}
			if readError != nil && readError == io.EOF {
				datas <- data
				err = nil
				break
			} else if readError != nil {
				errc <- readError
				return
			}
			datas <- data
		}
		errc <- err
	}()
	return datas, errc
}

func Validate(file string) (string, error) {
	_, err := FileSize(file)
	if err != nil {
		return "", err
	}
	return file, err
}

func FileSize(file string) (int64, error) {
	fileinfo, err := os.Stat(file)
	if !os.IsNotExist(err) {
		return fileinfo.Size(), nil
	}
	return 0, err

}

func Md5sum(file *os.File) (string, error) {
	h := md5.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return "", err
	}
	file.Seek(0, os.SEEK_SET)
	hex := fmt.Sprintf("%x", h.Sum(nil))
	return hex, nil
}

func ChooseNonEmpty(a string, b string) string {
	if a == "" {
		return b
	}
	return a
}

func ListFiles(searchPath string, walkFn filepath.WalkFunc, symlinkDepth int) []error {
	errors := []error{}
	fi, err := os.Lstat(searchPath)
	if err != nil {
		return []error{walkFn(searchPath, fi, err)}
	}
	if IsSymlink(fi.Mode()) {
		fi, err = os.Stat(searchPath)
		if err != nil {
			return []error{walkFn(searchPath, fi, err)}
		}
		symlinkDepth++
		if symlinkDepth > MaxDepth {
			return []error{&ListFilesError{searchPath}}
		}
	}
	if fi.IsDir() {
		fis, err := ioutil.ReadDir(searchPath)
		if err != nil {
			return []error{walkFn(searchPath, fi, err)}
		}
		for _, fi := range fis {
			fullPath := filepath.Join(searchPath, fi.Name())
			errs := ListFiles(fullPath, walkFn, symlinkDepth)
			t := make([]error, len(errors)+len(errs))
			copy(t, errors)
			copy(t[len(errors):], errs)
			errors = t
		}
	} else if fi.Mode().IsRegular() {
		err := walkFn(searchPath, fi, err)
		if err != nil {
			return []error{err}
		}
	}
	return errors

}

func IsSymlink(fileMode os.FileMode) bool {
	return fileMode&os.ModeSymlink == os.ModeSymlink
}

//Error

type ListFilesError struct {
	Path string
}

func (e *ListFilesError) Error() string {
	return fmt.Sprintf("%s :symlinkDepth > %d", e.Path, MaxDepth)
}
