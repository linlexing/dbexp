package main

import (
	"bufio"
	"fmt"
	"os"

	"strings"

	log "github.com/Sirupsen/logrus"
)

//flat是一种特殊的格式，要求传入的所有记录全部是字符串格式，然后每行拼接所有字段
//数据里的换行回车被替换成空格
type outFlat struct {
	f *os.File
	w *bufio.Writer
}

func (o *outFlat) Open(fileName string) (err error) {
	o.f, err = os.Create(fileName)
	if err != nil {
		return
	}
	o.w = bufio.NewWriter(o.f)
	return
}

func (o *outFlat) WriteTitle(data []string) error {
	return nil
}

func (o *outFlat) WriteLine(data []interface{}) error {
	for i, one := range data {
		var str string
		switch tv := one.(type) {
		case string:
			str = tv
		case []byte:
			str = string(tv)
		default:
			err := fmt.Errorf("not string")
			log.WithFields(log.Fields{
				"data": one,
				"col":  i,
			}).Error(err.Error())
			return err
		}
		str = strings.Replace(
			strings.Replace(str, "\r", " ", -1),
			"\n", " ", -1)
		rstr := []rune(str)
		if len(rstr) < cfg.FieldSize[i] {
			str = str + strings.Repeat(" ", cfg.FieldSize[i]-len(rstr))
		} else if len(rstr) > cfg.FieldSize[i] {
			str = string(rstr[:cfg.FieldSize[i]])
		}
		if _, err := o.w.WriteString(str); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(o.w); err != nil {
		return err
	}
	return nil
}

func (o *outFlat) Close() (err error) {
	if err = o.w.Flush(); err != nil {
		return
	}
	return o.f.Close()
}
