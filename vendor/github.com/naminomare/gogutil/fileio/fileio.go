package fileio

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// ErrOverMaxLoop maxLoopを超えてもファイルが存在し続けた場合のエラー
	ErrOverMaxLoop = errors.New("maxLoopを超えてもファイルが存在し続けた場合のエラー")

	// ExtAll すべてを検索する場合の拡張子
	ExtAll = "*"
)

// IsExist ファイルが存在するかチェック
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileName path[len(dir):]
func FileName(path string) string {
	sl := strings.LastIndex(path, "/")
	bs := strings.LastIndex(path, "\\")
	if sl == -1 && bs == -1 {
		// /も\も無いので、それそのものがファイル名だろう
		return path
	}

	if sl >= 0 {
		return path[sl+1:]
	}
	return path[bs+1:]
}

// GetNonExistFileName pathが存在した場合に、path0, path1のようなファイル名を返す
func GetNonExistFileName(path string, maxLoop int) (string, error) {
	if !IsExist(path) {
		return path, nil
	}

	// 存在した場合
	ext := filepath.Ext(path)
	filenamebase := path[0 : len(path)-len(ext)]

	for i := 0; i < maxLoop; i++ {
		filename := filenamebase + strconv.Itoa(i) + ext
		if IsExist(filename) {
			continue
		}
		return filename, nil
	}
	return "", ErrOverMaxLoop
}

// GetFiles targetDirectoryからextの拡張子を持ったファイルを取得する
// extが"*"の場合はすべての拡張子
// topDirectoryOnlyがtrueの場合はtopのみ
func GetFiles(targetDirectory, targetExt string, topDirectoryOnly bool) []string {
	var files []string
	filepath.Walk(
		targetDirectory,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if topDirectoryOnly {
					if targetDirectory == path {
						return nil
					}
					return filepath.SkipDir
				}
			}

			if targetExt != ExtAll {
				ext := filepath.Ext(info.Name())
				if ext != targetExt {
					return nil
				}
			}

			files = append(files, path)
			return nil
		},
	)
	return files
}
