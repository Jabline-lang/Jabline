package code

type SourcePos struct {
	Line   int
	Column int
}

type SourceMap map[int]SourcePos
