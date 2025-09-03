package lexer

// skipWhitespace omite espacios en blanco, tabs, saltos de línea y retornos de carro
// Continúa leyendo caracteres hasta encontrar uno que no sea whitespace
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment omite comentarios de línea que comienzan con //
// Continúa leyendo hasta encontrar un salto de línea o el final del archivo
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipMultiLineComment omite comentarios multilínea que comienzan con /* y terminan con */
// Maneja comentarios anidados y continúa hasta encontrar el cierre correspondiente
func (l *Lexer) skipMultiLineComment() {
	depth := 1
	l.readChar() // skip the '*' after '/'

	for depth > 0 && l.ch != 0 {
		if l.ch == '/' && l.peekChar() == '*' {
			// Nested comment start
			depth++
			l.readChar() // skip '/'
			l.readChar() // skip '*'
		} else if l.ch == '*' && l.peekChar() == '/' {
			// Comment end
			depth--
			l.readChar() // skip '*'
			l.readChar() // skip '/'
		} else {
			l.readChar()
		}
	}
}
