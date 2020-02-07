package app

type luaReportTable struct {
	head []string
	rows []*luaReportTableRow
}

func (x *luaReportTable) AddRow(s string) *luaReportTableRow {
	row := &luaReportTableRow{head: s}
	x.rows = append(x.rows, row)
	return row
}

type luaReportTableRow struct {
	head  string
	cells []*luaReportTableCell
}

func (x *luaReportTableRow) AddCell(s string) {
	x.cells = append(x.cells, &luaReportTableCell{text: s})
}

func (x *luaReportTableRow) AddCellOk(s string) {
	x.cells = append(x.cells, &luaReportTableCell{text: s, ok: ptrBool(true)})
}

func (x *luaReportTableRow) AddCellErr(s string) {
	x.cells = append(x.cells, &luaReportTableCell{text: s, ok: ptrBool(true)})
}

type luaReportTableCell struct {
	text string
	ok   *bool
}

func ptrBool(x bool) *bool {
	return &x
}
