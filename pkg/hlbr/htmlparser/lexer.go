package htmlparser

import "strings"

type TokenType int

const (
	TextToken TokenType = iota
	StartTagToken
	EndTagToken
	CommentToken
	DoctypeToken
)

type Token struct {
	Type       TokenType
	Data       string
	Attr       []Attribute
	SelfClosed bool
}

type Attribute struct {
	Key   string
	Value string
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TextToken, Data: ""}
	}

	if l.input[l.pos] == '<' {
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '!' {
			if strings.HasPrefix(l.input[l.pos:], "<!--") {
				return l.readComment()
			}
			if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), "<!DOCTYPE") {
				return l.readDoctype()
			}
		}
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			return l.readEndTag()
		}
		return l.readStartTag()
	}

	return l.readText()
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.pos++
	}
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func (l *Lexer) readText() Token {
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '<' {
		l.pos++
	}
	return Token{Type: TextToken, Data: l.input[start:l.pos]}
}

func (l *Lexer) readStartTag() Token {
	l.pos++ // skip '<'
	tagName := l.readTagName()

	token := Token{Type: StartTagToken, Data: strings.ToLower(tagName)}

	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) || l.input[l.pos] == '>' || (l.input[l.pos] == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '>') {
			break
		}
		attr := l.readAttribute()
		if attr.Key != "" {
			token.Attr = append(token.Attr, attr)
		}
	}

	if l.pos < len(l.input) && l.input[l.pos] == '/' {
		token.SelfClosed = true
		l.pos++
	}
	if l.pos < len(l.input) && l.input[l.pos] == '>' {
		l.pos++
	}

	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	if voidElements[token.Data] {
		token.SelfClosed = true
	}

	return token
}

func (l *Lexer) readTagName() string {
	start := l.pos
	for l.pos < len(l.input) && !isWhitespace(l.input[l.pos]) && l.input[l.pos] != '>' && l.input[l.pos] != '/' {
		l.pos++
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readAttribute() Attribute {
	if l.pos >= len(l.input) {
		return Attribute{}
	}

	start := l.pos
	for l.pos < len(l.input) && !isWhitespace(l.input[l.pos]) && l.input[l.pos] != '=' && l.input[l.pos] != '>' && l.input[l.pos] != '/' {
		l.pos++
	}
	key := l.input[start:l.pos]

	l.skipWhitespace()
	if l.pos >= len(l.input) || l.input[l.pos] != '=' {
		return Attribute{Key: key, Value: ""}
	}
	l.pos++ // skip '='
	l.skipWhitespace()

	var value string
	if l.pos < len(l.input) && (l.input[l.pos] == '"' || l.input[l.pos] == '\'') {
		quote := l.input[l.pos]
		l.pos++
		start = l.pos
		for l.pos < len(l.input) && l.input[l.pos] != quote {
			l.pos++
		}
		value = l.input[start:l.pos]
		if l.pos < len(l.input) {
			l.pos++ // skip closing quote
		}
	} else {
		start = l.pos
		for l.pos < len(l.input) && !isWhitespace(l.input[l.pos]) && l.input[l.pos] != '>' && l.input[l.pos] != '/' {
			l.pos++
		}
		value = l.input[start:l.pos]
	}

	return Attribute{Key: key, Value: value}
}

func (l *Lexer) readEndTag() Token {
	l.pos += 2 // skip '</'
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '>' {
		l.pos++
	}
	tagName := l.input[start:l.pos]
	if l.pos < len(l.input) {
		l.pos++ // skip '>'
	}
	return Token{Type: EndTagToken, Data: strings.ToLower(tagName)}
}

func (l *Lexer) readComment() Token {
	l.pos += 4 // skip '<!--'
	end := strings.Index(l.input[l.pos:], "-->")
	if end == -1 {
		return Token{Type: CommentToken, Data: l.input[l.pos:]}
	}
	data := l.input[l.pos : l.pos+end]
	l.pos += end + 3
	return Token{Type: CommentToken, Data: data}
}

func (l *Lexer) readDoctype() Token {
	l.pos += 9 // skip '<!DOCTYPE'
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '>' {
		l.pos++
	}
	data := l.input[start:l.pos]
	if l.pos < len(l.input) {
		l.pos++
	}
	return Token{Type: DoctypeToken, Data: strings.TrimSpace(data)}
}

func (l *Lexer) ReadRawText(tagName string) string {
	closingTag := "</" + tagName + ">"
	idx := strings.Index(l.input[l.pos:], closingTag)
	if idx == -1 {
		result := l.input[l.pos:]
		l.pos = len(l.input)
		return result
	}
	result := l.input[l.pos : l.pos+idx]
	l.pos += idx
	return result
}
