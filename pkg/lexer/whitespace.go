package lexer

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipMultiLineComment() {
	depth := 1
	l.readChar()

	for depth > 0 && l.ch != 0 {
		if l.ch == '/' && l.peekChar() == '*' {

			depth++
			l.readChar()
			l.readChar()
		} else if l.ch == '*' && l.peekChar() == '/' {

			depth--
			l.readChar()
			l.readChar()
		} else {
			l.readChar()
		}
	}
}
