package lexer

// readChar lee el siguiente carácter y avanza la posición en el input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // 0 = EOF (fin de archivo)
	} else {
		l.ch = l.input[l.readPosition]
	}

	// Track line and column position
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}

	l.position = l.readPosition
	l.readPosition++
}

// peekChar devuelve el siguiente carácter sin avanzar la posición
// Útil para operadores de dos caracteres como == o +=
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0 // EOF
	} else {
		return l.input[l.readPosition]
	}
}

// isLetter verifica si un carácter es una letra válida para identificadores
// Incluye letras minúsculas, mayúsculas y guión bajo
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_'
}

// isDigit verifica si un carácter es un dígito numérico
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
