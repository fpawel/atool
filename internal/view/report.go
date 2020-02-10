package view

type Report struct {
	headers []string
	tables  []*Table
}

func (x *Report) AddHeader(header string) {
	x.headers = append(x.headers, header)
}

func (x *Report) AddTable(head string) *Table {
	tab := &Table{head: head}
	x.tables = append(x.tables, tab)
	return tab
}

type Table struct {
	head string
	rows []*TableRow
}

func (x *Table) AddRow(s string) *TableRow {
	row := &TableRow{head: s}
	x.rows = append(x.rows, row)
	return row
}

type TableRow struct {
	head  string
	cells []*TableCell
}

func (x *TableRow) AddCell(s string) {
	x.cells = append(x.cells, &TableCell{text: s})
}

func (x *TableRow) AddCellOk(s string) {
	x.cells = append(x.cells, &TableCell{text: s, ok: ptrBool(true)})
}

func (x *TableRow) AddCellErr(s string) {
	x.cells = append(x.cells, &TableCell{text: s, ok: ptrBool(false)})
}

type TableCell struct {
	text string
	ok   *bool
}

func ptrBool(x bool) *bool {
	return &x
}
