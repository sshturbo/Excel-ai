package excel

// ExcelClient define a interface para manipulação de arquivos Excel via Excelize
type ExcelClient interface {
	// ==================== SHEETS ====================
	ListSheets() []string
	SheetExists(name string) (bool, error)
	CreateSheet(name string) error
	DeleteSheet(name string) error
	RenameSheet(oldName, newName string) error
	HideSheet(sheet string) error
	ShowSheet(sheet string) error

	// ==================== DATA ====================
	GetCellValue(sheet, cell string) (string, error)
	SetCellValue(sheet, cell string, value interface{}) error
	GetRangeValues(sheet, rng string) ([][]string, error)
	WriteRange(sheet, startCell string, data [][]interface{}) error
	ClearRange(sheet, rng string) error

	// ==================== FORMATTING ====================
	FormatRange(sheet, rng string, format Format) error
	SetColumnWidth(sheet, col string, width float64) error
	SetRowHeight(sheet, row string, height float64) error
	MergeCells(sheet, rng string) error
	UnmergeCells(sheet, rng string) error
	SetBorders(sheet, rng, style string) error
	AutoFitColumns(sheet, rng string) error

	// Conditional Formatting
	AddSimpleConditionalFormat(sheet, rng, criteria, value, bgColor string) error

	// ==================== STRUCTURE ====================
	InsertRows(sheet string, row, count int) error
	DeleteRows(sheet string, row, count int) error
	FreezePane(sheet, cell string, freezeRows, freezeCols int) error
	UnfreezePane(sheet string) error
	GroupRows(sheet string, startRow, endRow int) error
	GroupColumns(sheet, startCol, endCol string) error

	// ==================== OBJECTS ====================
	CreateChart(sheet, rng, chartType, title string) error
	DeleteChart(sheet, name string) error
	ListCharts(sheet string) ([]string, error)
	CreateTable(sheet, rng, name, style string) error
	DeleteTable(sheet, name string) error
	ListTables(sheet string) ([]string, error)
	CreatePivotTable(srcSheet, srcRange, destSheet, destCell, name string) error
	ListPivotTables(sheet string) ([]string, error)
	DeletePivotTable(sheet, name string) error

	// ==================== FILTERS & SORT ====================
	ApplyFilter(sheet, rng string) error
	ClearFilters(sheet string) error
	HasFilter(sheet string) (bool, error)
	SortRange(sheet, rng string, col int, ascending bool) error

	// ==================== VALIDATION ====================
	AddDataValidation(sheet, rng, validationType string, options []string) error
	AddDropdownList(sheet, rng string, options []string) error

	// ==================== COMMENTS ====================
	AddCellComment(sheet, cell, author, text string) error
	GetCellComment(sheet, cell string) (string, error)
	DeleteCellComment(sheet, cell string) error

	// ==================== HYPERLINKS ====================
	AddHyperlink(sheet, cell, url, display string) error
	GetHyperlink(sheet, cell string) (string, error)

	// ==================== PROTECTION ====================
	ProtectSheet(sheet, password string) error
	UnprotectSheet(sheet, password string) error
	SetCellLocked(sheet, cell string, locked bool) error
	GetSheetProtection(sheet string) (bool, error)

	// ==================== FORMULAS ====================
	SetCellFormula(sheet, cell, formula string) error
	GetCellFormula(sheet, cell string) (string, error)
	CalculateFormulas() error

	// ==================== QUERY ====================
	GetUsedRange(sheet string) (string, error)
	GetRowCount(sheet string) (int, error)
	GetColumnCount(sheet string) (int, error)
	GetHeaders(sheet, rng string) ([]string, error)

	// ==================== PRINT ====================
	SetPrintArea(sheet, rng string) error

	// ==================== LIFECYCLE ====================
	Close()
	Write() ([]byte, error)
	SaveAs(path string) error
}
