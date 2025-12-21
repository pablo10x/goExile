package game

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		newPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(newPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		dstFile, err := os.Create(newPath)
		if err != nil {
			return err
		}
		defer func() { _ = dstFile.Close() }()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(newPath, info.Mode())
	})
}

// zipDir compresses a directory into a zip file, excluding specific subdirectories.
func zipDir(src string, dest string, excludes []string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	archive := zip.NewWriter(zipFile)
	defer func() { _ = archive.Close() }()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle exclusions
		for _, exclude := range excludes {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Get relative path for header
		headerPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if headerPath == "." {
			return nil
		}

		// Use forward slashes for zip compatibility
		headerPath = filepath.ToSlash(headerPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = headerPath
		header.Method = zip.Deflate // High compression

		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(writer, file)
		return err
	})
}

func unzipDir(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			func() { _ = outFile.Close() }()
			return err
		}

		_, err = io.Copy(outFile, rc)
		func() { _ = outFile.Close() }()
		_ = rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
