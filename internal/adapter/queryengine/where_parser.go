package queryengine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var allowedColumns = map[string]struct{}{
	"title":      {},
	"author":     {},
	"rating":     {},
	"read_date":  {},
	"genre":      {},
	"publisher":  {},
	"edition":    {},
	"imprint":    {},
	"series":     {},
	"translator": {},
}

var rejectedKeywords = map[string]struct{}{
	"select": {}, "insert": {}, "update": {}, "delete": {},
	"create": {}, "alter": {}, "drop": {},
	"join": {}, "union": {}, "order": {}, "by": {},
}

func parseWhere(input string) (expr, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	parser := whereParser{tokens: tokens}
	parsed, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	if parser.peek().kind != tokenEOF {
		return nil, fmt.Errorf("unsupported where syntax near %q", parser.peek().text)
	}
	return parsed, nil
}

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenIdent
	tokenNumber
	tokenString
	tokenOperator
	tokenLParen
	tokenRParen
)

type token struct {
	kind tokenKind
	text string
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		r := rune(input[i])
		if unicode.IsSpace(r) {
			i++
			continue
		}
		if input[i] == '-' && i+1 < len(input) && input[i+1] == '-' {
			return nil, fmt.Errorf("comments are not supported in --where")
		}
		if input[i] == ';' {
			return nil, fmt.Errorf("multiple statements are not supported in --where")
		}
		if input[i] == '(' {
			tokens = append(tokens, token{kind: tokenLParen, text: "("})
			i++
			continue
		}
		if input[i] == ')' {
			tokens = append(tokens, token{kind: tokenRParen, text: ")"})
			i++
			continue
		}
		if strings.ContainsRune("=!<>", rune(input[i])) {
			op := string(input[i])
			if i+1 < len(input) && input[i+1] == '=' {
				op += "="
				i++
			}
			if op == "!" {
				return nil, fmt.Errorf("unsupported operator: !")
			}
			tokens = append(tokens, token{kind: tokenOperator, text: op})
			i++
			continue
		}
		if input[i] == '"' || input[i] == '\'' {
			text, next, err := readString(input, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{kind: tokenString, text: text})
			i = next
			continue
		}
		if unicode.IsDigit(r) {
			start := i
			for i < len(input) && unicode.IsDigit(rune(input[i])) {
				i++
			}
			tokens = append(tokens, token{kind: tokenNumber, text: input[start:i]})
			continue
		}
		if isIdentStart(r) {
			start := i
			for i < len(input) && isIdentPart(rune(input[i])) {
				i++
			}
			text := strings.ToLower(input[start:i])
			if _, rejected := rejectedKeywords[text]; rejected {
				return nil, fmt.Errorf("unsupported keyword in --where: %s", text)
			}
			tokens = append(tokens, token{kind: tokenIdent, text: text})
			continue
		}
		return nil, fmt.Errorf("unsupported character in --where: %q", input[i])
	}
	tokens = append(tokens, token{kind: tokenEOF})
	return tokens, nil
}

func readString(input string, start int) (string, int, error) {
	quote := input[start]
	var builder strings.Builder
	for i := start + 1; i < len(input); i++ {
		if input[i] == quote {
			return builder.String(), i + 1, nil
		}
		if input[i] == '\\' && i+1 < len(input) {
			i++
			builder.WriteByte(input[i])
			continue
		}
		builder.WriteByte(input[i])
	}
	return "", 0, fmt.Errorf("unterminated string in --where")
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

type whereParser struct {
	tokens []token
	pos    int
}

func (parser *whereParser) parseOr() (expr, error) {
	left, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}
	for parser.matchIdent("or") {
		right, err := parser.parseAnd()
		if err != nil {
			return nil, err
		}
		left = logical{op: "or", left: left, right: right}
	}
	return left, nil
}

func (parser *whereParser) parseAnd() (expr, error) {
	left, err := parser.parseNot()
	if err != nil {
		return nil, err
	}
	for parser.matchIdent("and") {
		right, err := parser.parseNot()
		if err != nil {
			return nil, err
		}
		left = logical{op: "and", left: left, right: right}
	}
	return left, nil
}

func (parser *whereParser) parseNot() (expr, error) {
	if parser.matchIdent("not") {
		inner, err := parser.parsePrimary()
		if err != nil {
			return nil, err
		}
		return negation{inner: inner}, nil
	}
	return parser.parsePrimary()
}

func (parser *whereParser) parsePrimary() (expr, error) {
	if parser.match(tokenLParen) {
		inner, err := parser.parseOr()
		if err != nil {
			return nil, err
		}
		if !parser.match(tokenRParen) {
			return nil, fmt.Errorf("missing closing parenthesis in --where")
		}
		if parser.peek().kind == tokenLParen {
			return nil, fmt.Errorf("function calls are not supported in --where")
		}
		return inner, nil
	}
	return parser.parseComparison()
}

func (parser *whereParser) parseComparison() (expr, error) {
	column := parser.advance()
	if column.kind != tokenIdent {
		return nil, fmt.Errorf("expected column name in --where")
	}
	if _, ok := allowedColumns[column.text]; !ok {
		return nil, fmt.Errorf("unsupported column in --where: %s", column.text)
	}
	if parser.matchIdent("is") {
		if parser.matchIdent("not") {
			if !parser.matchIdent("null") {
				return nil, fmt.Errorf("expected null after is not")
			}
			return comparison{column: column.text, op: "is not null"}, nil
		}
		if !parser.matchIdent("null") {
			return nil, fmt.Errorf("expected null after is")
		}
		return comparison{column: column.text, op: "is null"}, nil
	}

	op := parser.advance()
	if op.kind == tokenIdent && op.text == "like" {
		if !isTextColumn(column.text) {
			return nil, fmt.Errorf("like is only supported for text columns")
		}
		return parser.parseBinary(column.text, "like")
	}
	if op.kind != tokenOperator {
		return nil, fmt.Errorf("expected operator after %s", column.text)
	}
	return parser.parseBinary(column.text, op.text)
}

func (parser *whereParser) parseBinary(column, op string) (expr, error) {
	value := parser.advance()
	switch column {
	case "rating":
		if value.kind != tokenNumber {
			return nil, fmt.Errorf("rating comparison requires a number")
		}
		if _, err := strconv.Atoi(value.text); err != nil {
			return nil, fmt.Errorf("invalid number in --where: %s", value.text)
		}
		return comparison{column: column, op: op, value: value.text}, nil
	case "read_date":
		if value.kind != tokenString {
			return nil, fmt.Errorf("read_date comparison requires a quoted YYYY-MM-DD date")
		}
		if _, err := time.Parse("2006-01-02", value.text); err != nil {
			return nil, fmt.Errorf("read_date comparison requires a quoted YYYY-MM-DD date")
		}
		return comparison{column: column, op: op, value: value.text}, nil
	}
	if value.kind != tokenString {
		return nil, fmt.Errorf("%s comparison requires a quoted string", column)
	}
	return comparison{column: column, op: op, value: value.text}, nil
}

func isTextColumn(column string) bool {
	return column != "rating" && column != "read_date"
}

func (parser *whereParser) peek() token {
	return parser.tokens[parser.pos]
}

func (parser *whereParser) advance() token {
	current := parser.peek()
	if current.kind != tokenEOF {
		parser.pos++
	}
	return current
}

func (parser *whereParser) match(kind tokenKind) bool {
	if parser.peek().kind != kind {
		return false
	}
	parser.advance()
	return true
}

func (parser *whereParser) matchIdent(text string) bool {
	if parser.peek().kind != tokenIdent || parser.peek().text != text {
		return false
	}
	parser.advance()
	return true
}
