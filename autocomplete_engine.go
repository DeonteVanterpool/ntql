package tbql

import (
	"regexp"

	"github.com/shivamMg/trie"
)

type EngineState int

const (
	StateSubject EngineState = iota
	StateVerb
	StateObject
	StateString
	StateTag
)

type ObjectType int

const (
	ObjectTypeNumber ObjectType = iota
	ObjectTypeString
	ObjectTypeBool
	ObjectTypeDate
	ObjectTypeDateTime
	ObjectTypeTag
)

type MegaTrie struct {
	subjectTrie *trie.Trie
}

func NewMegaTrie() *MegaTrie {
	subjectTrie := trie.New()
	for _, subject := range validSubjects {
		for _, alias := range subject.Aliases {
			subjectTrie.Put([]string{alias}, subject)
		}
		subjectTrie.Put([]string{subject.Name}, subject)
	}
	return &MegaTrie{
		subjectTrie: subjectTrie,
	}
}

// CompletionEngine is a trie-based autocompletion engine for TBQL
type CompletionEngine struct {
	// checks to see if we're in a function call or not
	innerParens int

	subjectTrie *trie.Trie

	subject Subject

	verbTrie *trie.Trie

    tokenizer *Lexer

	dtypes []ObjectType
}

func NewAutocompleteEngine(tags []string) *CompletionEngine {
	return &CompletionEngine{
        subjectTrie: NewMegaTrie().subjectTrie,
		innerParens: 0,
	}
}

func (e *CompletionEngine) Suggest(s string) ([]string, error) {
    if len(s) == 0 {
        return e.Subject()
    }

    if tokens[len(tokens)-1] == ' ' {

    }

    // subject: outside last function call's parentheses
    // function call: identifier followed by LPAREN
    // object: inside function call's parentheses, and previous token either a connector or LPAREN
    // string: inside function call's parentheses, and previous token either a connector or LPAREN
    // scan until rest of string has no dot, LPAREN, or space
}

func (e *CompletionEngine) buildVerbTrie(subject Subject) *trie.Trie {
	verbTrie := trie.New()
	for _, verb := range subject.ValidVerbs {
		verbTrie.Put([]string{verb.Name}, verb)
	}
	return verbTrie
}

func (e *CompletionEngine) suggestFromSubject(verb string) ([]string, error) {
	verbs := make([]string, 0)
	for _, v := range e.verbTrie.Search([]string{verb}).Results {
		verbs = append(verbs, v.Value.(Verb).Name)
	}

	return verbs, nil
}

func (e *CompletionEngine) setState(s EngineState) {
	e.state = s
}

type AutocompleteError struct {
    Code    AutocompleteErrorCode
    Position int
    Message string
}

func NewAutocompleteError(code AutocompleteErrorCode, pos int, msg string) *AutocompleteError {
    return &AutocompleteError{
        Code: code,
        Position: pos,
        Message: msg,
    }
}

func (e *AutocompleteError) Error() string {
    return e.Message
}

type AutocompleteErrorCode int

const (
    InvalidCharacter AutocompleteErrorCode = iota
)

func (e *CompletionEngine) Previous() (byte, error) {
	if e.pos == 0 {
		return '\x00', NewTokenizationError(EndOfInput, e.pos, "Reached beginning of input")
	}
	return e.s[e.pos-1], nil
}

func (e *CompletionEngine) matchRegexp(r *regexp.Regexp) bool {
	if r.MatchString(e.s[e.pos : e.pos+1]) {
		e.pos++
		return true
	}
	return false
}

func (e *CompletionEngine) match(c byte) bool {
	if e.s[e.pos] == c {
		e.pos++
		return true
	}
	return false
}

func (e *CompletionEngine) atEnd() bool {
	return e.pos >= len(e.s)
}

func (e *CompletionEngine) advance() (byte, error) {
	if e.atEnd() {
		return '\x00', NewTokenizationError(EndOfInput, e.pos, "Reached end of input")
	}
	e.pos++
	return e.s[e.pos-1], nil
}

func (e *CompletionEngine) Subject() ([]string, error) {
	if e.match('.') {
        e.setState(StateVerb)
        e.subject = e.subjectTrie.Search([]string{e.text}).Results[0].Value.(Subject)
        e.verbTrie = e.buildVerbTrie(e.subject)
		return e.Verb()
	}
	if e.atEnd() {
        subjects := make([]string, 0)
        for _, s := range e.subjectTrie.Search([]string{e.text}).Results {
            subjects = append(subjects, s.Value.(Subject).Name)
        }
		return subjects, nil
	}

	c, _ := e.advance() // Should not be an error here
	e.text += string(c)

	return e.Subject()
}

func (e *CompletionEngine) Verb() ([]string, error) {
    if e.match('(') {
        e.innerParens = 0
        return e.Object()
    }
    if e.atEnd() {
        return e.suggestFromSubject(e.text)
    }

    c, _ := e.advance() // Should not be an error here
    e.text += string(c)

    return e.Verb()
}

func (e *CompletionEngine) Object() ([]string, error) {
    if e.match('"') {
        return e.String()
    }
}

func (e *CompletionEngine) String() ([]string, error) {

}

func (e *CompletionEngine) Tag() ([]string, error) {

}

