package lexer

func isSpace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r'
}

func isSymbol(c byte) bool {
	switch c {
	case '*', ',', '=', ';', '(', ')':
		return true
	}
	return false
}
