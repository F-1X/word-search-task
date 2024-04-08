package dir

import (
	"fmt"
	"io/fs"
)

func FilesFS(fsys fs.FS, dir string) ([]string, error) {
	if dir == "" {
		dir = "."
	}
	var fileNames []string
	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// возможно добавить регулярку на типы файлов, хотя принципиально нужно задавать базовую директорию под конкретную работу с индексами, без "лишних" файлов
		if !d.IsDir() {
			fileNames = append(fileNames, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(fileNames) == 0 {
		return nil, fmt.Errorf("err no files found")
	}
	return fileNames, nil
}
