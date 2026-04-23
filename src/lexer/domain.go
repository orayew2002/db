package lexer

const (
	ILLEGAL   TokenType = "ILLEGAL"
	EOF       TokenType = "EOF"
	IDENT     TokenType = "IDENT"
	NUMBER    TokenType = "NUMBER"
	STRING    TokenType = "STRING"
	SELECT    TokenType = "SELECT"
	FROM      TokenType = "FROM"
	WHERE     TokenType = "WHERE"
	AND       TokenType = "AND"
	EQ        TokenType = "EQ"
	GT        TokenType = "GT"
	COMMA     TokenType = "COMMA"
	SEMICOLON TokenType = "SEMICOLON"
)
