package expression

import "testing"

func TestToken(t *testing.T) {
	for _, text := range []string{
		"aa&bb|cc&dd",

		"aa&bb|(cc&dd)",
		"(aa&bb)|cc&dd",
		"aa&(bb|cc)&dd",

		"(aa&bb|cc)&dd",
		"aa&(bb|cc&dd)",

		"(aa&bb)|(cc&dd)",
	} {
		tokens := tokens(text)
		p := parser{toks: tokens}
		n, err := p.parseExpr()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v", n)
	}
}
