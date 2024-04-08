package searcher

import (
	"bufio"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
	"word-search-in-files/pkg/internal/dir"
)

type Searcher struct {
	FS           fs.FS
	Dir          string
	Index        map[string][]string        // инвертированный индекс слов в каждом файле
	mu           sync.Mutex                 // для доступа к индексам
	indexedMap   map[string]map[string]bool // Мап для отслеживания проиндексированных файлов . также была идея сохранять индексы файлов в бд или файл
	muIndexedMap sync.Mutex                 // мьютекс для синхронизации проверки indexedMap
	Recursive    bool                       // вложенные директории
}

func NewSearcher(fsys fs.FS) *Searcher {
	return &Searcher{
		FS:         fsys,
		indexedMap: make(map[string]map[string]bool),
		Recursive:  true,
	}
}

func (s *Searcher) IndexFiles() error {

	if s.indexedMap == nil {
		s.indexedMap = make(map[string]map[string]bool)
	}

	if s.Index == nil {
		s.Index = make(map[string][]string)
	}

	// имеет смысл вернуть ошибки связанные с чтением файлов? не реализовано в плане отправки  ошибки стороне клиента
	errorsMap := make(map[string]error)

	if s.Dir == "" {
		s.Dir = "."
	}

	files, err := dir.FilesFS(s.FS, ".")
	// fmt.Println("files", files)
	if err != nil {
		fmt.Println("err", err)
		return err
	}

	var wg sync.WaitGroup

	for _, fileName := range files {
		dir := s.Dir
		s.muIndexedMap.Lock()

		if s.indexedMap[s.Dir] == nil {
			s.indexedMap[s.Dir] = make(map[string]bool)
		}

		if s.indexedMap[dir][fileName] {
			s.muIndexedMap.Unlock()
			continue
		}
		s.muIndexedMap.Unlock()

		wg.Add(1)

		go func(fileName string, dir string) {

			defer wg.Done()
			file, err := s.FS.Open(fileName)
			if err != nil {
				s.mu.Lock()
				errorsMap[fileName] = err
				s.mu.Unlock()

				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				tokens := tokenize(line)
				for _, token := range tokens {
					s.mu.Lock()
					s.Index[token] = append(s.Index[token], fileName)
					s.mu.Unlock()

					s.muIndexedMap.Lock()
					s.indexedMap[dir][fileName] = true
					s.muIndexedMap.Unlock()
				}
			}

			if err := scanner.Err(); err != nil {
				errorsMap[fileName] = err
				return
			}

		}(fileName, dir)

	}

	wg.Wait()

	return nil
}

func (s *Searcher) Search(word string) ([]string, error) {

	// индексанция производится 1 раз при первом запуске в текущей директории "." (main.go -> os.DirFS("."))
	// но лучше получить актуальный список файлов и переиндексировать, если возникает ситуация удаления/добавления файла
	// также стоит создать структуру файл-время изменения (при модификации) и пересчета текущей метаинформации о времени изменении файла для переиндексации
	if s.Index == nil {
		if err := s.IndexFiles(); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	files, found := s.Index[word]
	s.mu.Unlock()

	if !found {
		return nil, fmt.Errorf("err word not found")
	}

	// ответ учитывает заданный параметр - dir и recursive
	answer := []string{}
	for _, fullName := range files {
		dir := filepath.Dir(fullName)
		if dir == s.Dir || s.Dir == "." {
			answer = append(answer, fullName)
			continue
		}

		if s.Recursive && strings.HasPrefix(dir, s.Dir) {
			answer = append(answer, fullName)
		}
	}

	// убираем расширение файлов для ответа
	filesBase := []string{}
	for _, fullName := range answer {
		filesBase = append(filesBase, fullName[:len(fullName)-len(filepath.Ext(fullName))])
	}

	// сортируем, тк при тесте была ошибка, из-за порядка файлов file3 file1 вместо file1 file3
	sort.Strings(filesBase)
	return filesBase, nil

}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}
