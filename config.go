package main

type Config struct {
	Driver   string
	DBUrl    string
	Table    string
	Query    string
	Limit    int64
	Filename string
	Outfmt   string
}
