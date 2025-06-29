package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3" //nolint
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.OpenFile(fromPath, os.O_RDONLY, 0o444)
	if err != nil {
		fmt.Print(err)
		return err
	}
	defer src.Close()
	oldFileStat, err := src.Stat()
	if err != nil {
		fmt.Print(err)
		return ErrUnsupportedFile
	}
	// check if offset is correct and return immediately if it is not
	if oldFileStat.Size() < offset {
		return ErrOffsetExceedsFileSize
	}
	newFile, err := os.Create(toPath)
	if err != nil {
		fmt.Print(err)
		return err
	}
	defer newFile.Close()
	if offset > 0 {
		_, err = src.Seek(offset, io.SeekStart)
		if err != nil {
			fmt.Print(err)
			return err
		}
	}
	oldReader := io.Reader(src)
	if offset+limit > oldFileStat.Size() {
		// for correct barReader work
		limit = oldFileStat.Size() - offset
	}
	if limit > 0 {
		oldReader = io.LimitReader(src, limit)
	}
	bar := pb.Full.Start64(limit)
	barReader := bar.NewProxyReader(oldReader)
	if _, err := io.Copy(newFile, barReader); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy failed: %w", err)
	}
	bar.Finish()
	return nil
}
