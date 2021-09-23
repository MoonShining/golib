package expression

import (
	"strings"
)

const (
	and    = "&"
	or     = "|"
	lparen = "("
	rparen = ")"
	id     = "i"
)

var operators = map[string]struct{}{
	and:    {},
	or:     {},
	lparen: {},
	rparen: {},
}

type token struct {
	Literal string
}

type Node struct {
	Op    string
	Val   string
	Left  *Node
	Right *Node
}

// expr: logic ([and|or] logic)*
// logic: id | paren
// paren: ( expr )
func Parse(text string) (*Node, error) {
	toks := tokens(text)
	return (&parser{toks: toks}).parseExpr()
}

type parser struct {
	toks []token
	cur  int
}

func (p *parser) parseExpr() (*Node, error) {
	l, err := p.parseLogic()
	if err != nil {
		return nil, err
	}
	if p.cur >= len(p.toks) {
		return l, nil
	}
	var node *Node
	for p.cur < len(p.toks) {
		curTok := p.toks[p.cur].Literal
		if curTok == and || curTok == or {
			node = &Node{Op: curTok, Left: l}
			p.cur++
			r, err := p.parseLogic()
			if err != nil {
				return nil, err
			}
			node.Right = r
			l = node
		} else {
			break
		}
	}
	return node, nil
}

func (p *parser) parseLogic() (*Node, error) {
	t := p.toks[p.cur]
	if !isOperator(t.Literal) {
		p.cur++
		return &Node{Val: t.Literal}, nil
	}
	return p.parseParen()
}

func (p *parser) parseParen() (*Node, error) {
	p.cur++
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.cur++ // skip )
	return node, nil
}

func tokens(text string) []token {
	tokens := []token{}
	chars := strings.Split(text, "")
	i := 0
	for i < len(chars) {
		c := chars[i]
		if isBlank(c) {
			i++
			continue
		}
		if isOperator(c) {
			tokens = append(tokens, token{Literal: c})
			i++
			continue
		}
		str := readString(chars, i)
		tokens = append(tokens, token{Literal: str})
		i += len(str)
	}
	return tokens
}

func isOperator(r string) bool {
	_, ok := operators[r]
	return ok
}

func isBlank(r string) bool {
	return r == " "
}

func readString(chars []string, startIdx int) string {
	i := startIdx
	for i < len(chars) {
		charactor := chars[i]
		if !isBlank(charactor) && !isOperator(charactor) {
			i++
		} else {
			break
		}
	}
	return strings.Join(chars[startIdx:i], "")
}
