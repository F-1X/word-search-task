package searcher

import (
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSearcher_Search(t *testing.T) {
	type fields struct {
		FS fs.FS
	}
	type args struct {
		word string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "Ok",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"file3.txt": {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "World"},
			wantFiles: []string{"file1", "file3"},
			wantErr:   false,
		},
		{
			name: "err word not found",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": &fstest.MapFile{},
				},
			},
			args:      args{word: "any"},
			wantFiles: nil,
			wantErr:   true,
		},
		{
			name: "err no files found",
			fields: fields{
				FS: fstest.MapFS{},
			},
			args:      args{word: "any"},
			wantFiles: nil,
			wantErr:   true,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				FS: tt.fields.FS,
			}
			gotFiles, err := s.Search(tt.args.word)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}


func TestSearcher_Search_VerboseError(t *testing.T) {
	type fields struct {
		FS fs.FS
	}
	type args struct {
		word string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   string
	}{
		{
			name: "Ok",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"file3.txt": {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "World"},
			wantFiles: []string{"file1", "file3"},
			wantErr:   "",
		},
		{
			name: "err word not found",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": &fstest.MapFile{},
				},
			},
			args:      args{word: "any"},
			wantFiles: nil,
			wantErr:   "err word not found",
		},
		{
			name: "err no files found",
			fields: fields{
				FS: fstest.MapFS{},
			},
			args:      args{word: "any"},
			wantFiles: nil,
			wantErr:   "err no files found",
		},


	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				FS: tt.fields.FS,
			}
			gotFiles, err := s.Search(tt.args.word)
			if (err != nil) && err.Error() != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}
