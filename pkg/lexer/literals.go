package lexer

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[start:l.position]
}

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
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
	}

	return string(result)
}

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
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
	}

	return string(result)
}
