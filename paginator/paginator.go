package paginator

import (
	"fmt"
	"reflect"
)

var (
	DefaultPerPage int64 = 10
	DefaultPage    int64 = 1
)

// 分页对象
type Paginator struct {
	perPage int64 // 每页多少条记录
	page    int64 // 当前是第多少页
}

func (p *Paginator) init() {
	if p.perPage < 0 {
		p.perPage = DefaultPerPage
	}
	if p.page <= 0 {
		p.page = DefaultPage
	}
}

func (p *Paginator) GetPageSize() int64 {
	return p.perPage
}

func (p *Paginator) GetPage() int64 {
	return p.page
}

func (p *Paginator) GetLimit() int {
	return int(p.perPage)
}

func (p *Paginator) GetOffset() int {
	return int(p.perPage*p.page - p.perPage)
}

func (p *Paginator) GetLimitAndOffset() (limit, offset int) {
	return p.GetLimit(), p.GetOffset()
}

func (p *Paginator) GetPageData(data interface{}) ([]interface{}, error) {
	page := p.page
	pageSize := p.perPage

	result := make([]interface{}, 0)
	s := reflect.Indirect(reflect.ValueOf(data))
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		return result, fmt.Errorf("not slice")
	}

	index := 0
	if page > 0 {
		index = int((page - 1) * pageSize)
	}
	if index < 0 {
		index = 0
	}

	for i := index; i < s.Len(); i++ {
		interf := s.Index(i).Interface()
		result = append(result, interf)
		if len(result) == int(pageSize) {
			break
		}
	}
	return result, nil
}

func NewPaginator(page, pageSize int64) *Paginator {
	p := &Paginator{}
	if pageSize < 0 {
		p.perPage = DefaultPerPage
	} else {
		p.perPage = pageSize
	}

	if page == 0 {
		p.page = DefaultPage
	} else {
		p.page = page
	}

	p.init()
	return p
}
