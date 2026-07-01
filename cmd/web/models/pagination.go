package models

func totalPages(totalCount int64, perPage int) int {
	if perPage <= 0 {
		return 1
	}
	if totalCount == 0 {
		return 1
	}
	return int((totalCount + int64(perPage) - 1) / int64(perPage))
}

func ClampPage(page int, totalCount int64, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || totalCount == 0 {
		return page
	}
	if page > totalPages(totalCount, perPage) {
		return totalPages(totalCount, perPage)
	}
	return page
}

func (p TablePagination) TotalPages() int {
	return totalPages(p.TotalCount, p.PerPage)
}

func (p TablePagination) StartItem() int64 {
	if p.TotalCount == 0 {
		return 0
	}
	return int64((p.Page-1)*p.PerPage + 1)
}

func (p TablePagination) EndItem() int64 {
	end := int64(p.Page * p.PerPage)
	if end > p.TotalCount {
		return p.TotalCount
	}
	return end
}

func (p TablePagination) HasPrev() bool {
	return p.Page > 1
}

func (p TablePagination) HasNext() bool {
	return p.Page < p.TotalPages()
}
