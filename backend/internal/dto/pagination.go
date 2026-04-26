package dto

// ApplyPaginationDefaults sets missing pagination fields to defaults (page=1, pageSize=20)
// and computes the offset. Returns (offset, page, pageSize).
func ApplyPaginationDefaults(p, ps int) (offset, page, pageSize int) {
	return ApplyPaginationWithDefault(p, ps, 20)
}

// ApplyPaginationWithDefault sets missing pagination fields to defaults (page=1, pageSize=defaultPageSize)
// and computes the offset. Returns (offset, page, pageSize).
func ApplyPaginationWithDefault(p, ps, defaultPageSize int) (offset, page, pageSize int) {
	page = p
	pageSize = ps
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	offset = (page - 1) * pageSize
	return
}
