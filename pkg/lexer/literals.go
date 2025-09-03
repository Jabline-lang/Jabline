package lexer

// readIdentifier lee un identificador (variables, funciones, etc.)
// Los identificadores pueden contener letras, dígitos y guiones bajos
func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// readNumber lee un número (entero o decimal)
// Soporta enteros (123) y decimales (3.14)
func (l *Lexer) readNumber() string {
	start := l.position

	// Lee la parte entera
	for isDigit(l.ch) {
		l.readChar()
	}

	// Si encuentra un punto decimal, lee la parte decimal
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume el punto
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[start:l.position]
}

// readString lee una cadena de texto entre comillas dobles
// Maneja secuencias de escape como \n, \t, \", \\
func (l *Lexer) readString() string {
	var result []byte

	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}

		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				// Si no es una secuencia de escape reconocida, incluir ambos caracteres
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
	}

	return string(result)
}

// readTemplateLiteral lee un template literal entre backticks
// Maneja interpolación con ${expression} dentro del template
func (l *Lexer) readTemplateLiteral() string {
	var result []byte

	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			break
		}

		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '`':
				result = append(result, '`')
			default:
				// Si no es una secuencia de escape reconocida, incluir ambos caracteres
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
	}

	return string(result)
}
