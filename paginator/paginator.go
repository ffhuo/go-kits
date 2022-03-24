package paginator

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
