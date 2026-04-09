package fm

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FmFullDump struct {
	path string
}

func NewFmFullDump(path string) *FmFullDump {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	return &FmFullDump{
		path: path,
	}
}

func (f *FmFullDump) Flush(data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpPath := f.path + "_tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, f.path); err != nil {
		return err
	}

	return nil
}

func (f *FmFullDump) Load(data any) error {
	file, err := os.Open(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}
