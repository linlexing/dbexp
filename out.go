package main

import "encoding/gob"
import "os"

type outEncode interface {
	Open(string) error
	WriteTitle([]string) error
	WriteLine([]interface{}) error
	Close() error
}
type outGob struct {
	f *os.File
	w *gob.Encoder
}

func (o *outGob) WriteTitle(data []string) error {
	return o.w.Encode(data)
}

func (o *outGob) Open(fileName string) (err error) {
	o.f, err = os.Create(fileName)
	if err != nil {
		return
	}
	o.w = gob.NewEncoder(o.f)
	return
}

func (o *outGob) WriteLine(data []interface{}) error {
	return o.w.Encode(data)
}

func (o *outGob) Close() error {
	return o.f.Close()
}
