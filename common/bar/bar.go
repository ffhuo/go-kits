package bar

import (
	"fmt"
	"io"
)

type Bar struct {
	prefix string
	io.Reader
	percent int64  //百分比
	cur     int64  //当前进度位置
	total   int64  //总进度
	rate    string //进度条
	graph   string //显示符号
}

func New(start, total int64, prefix string) *Bar {
	bar := &Bar{}
	bar.cur = start
	bar.total = total
	bar.prefix = prefix
	if bar.graph == "" {
		bar.graph = "#"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
	return bar
}

func NewWithReader(reader io.Reader, total int64, prefix string) *Bar {
	bar := &Bar{}
	bar.cur = 0
	bar.total = total
	bar.prefix = prefix
	if bar.graph == "" {
		bar.graph = "#"
	}
	bar.Reader = reader
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
	return bar
}

func (bar *Bar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%%  %8d/%d  %s  ", bar.rate, bar.percent, bar.cur, bar.total, bar.prefix)
}

func (bar *Bar) Read(p []byte) (int, error) {
	n, err := bar.Reader.Read(p)
	bar.cur += int64(n)
	bar.Play(bar.cur)
	return n, err
}

func (bar *Bar) Finish() {
	bar.Play(bar.total)
	fmt.Println()
}
