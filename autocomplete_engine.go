package ntql

import (
	// "regexp"

	"github.com/shivamMg/trie"
)

type EngineState int

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

// CompletionEngine is a trie-based autocompletion engine for NTQL
type CompletionEngine struct {
	// checks to see if we're in a function call or not
	innerParens int

	subjectTrie *trie.Trie

	subject *Subject

	verbTrie *trie.Trie

	scanner *Scanner
}

func NewAutocompleteEngine(tags []string) *CompletionEngine {
	return &CompletionEngine{
		subjectTrie: NewMegaTrie().subjectTrie,
		innerParens: 0,
	}
}

func (e *CompletionEngine) Suggest(s string) ([]string, error) {
	if len(s) == 0 {
		return e.SuggestSubject("")
	}

	e.scanner = NewScanner(s)

	// at each dot, we must suggest a verb. at each open paren, we must suggest an object. at each close paren, we must suggest a connector. if outside of a method call, suggest a subject. if inside of a method call, suggest an object. Skip strings
	// dot -> suggest verb; open paren + inside method -> suggest object; open paren + outside method -> suggest subject; closing paren + lastcharspace -> suggest connector
	var lexemes []Lexeme
	for !e.scanner.atEnd() {
		lexeme, err := e.scanner.ScanLexeme()
		if err != nil {
			return nil, err
		}
		lexemes = append(lexemes, lexeme)
	}

	if lexemes[len(lexemes)-1][0] == '"' { // last lexeme string
		return []string{}, nil
	}
	switch lexemes[len(lexemes)-1] {
	case ".":
		return e.suggestFromSubject("")
	}
	switch lexemes[len(lexemes)-2] {
	case ".":
		subject, err := getSubject(string(lexemes[len(lexemes)-3])) // get lexeme before dot
		if err != nil {
			panic("unimplemented")
		}
		e.subject = subject
		return e.suggestFromSubject(string(lexemes[len(lexemes)-1]))
	}

	// before dot and outside of method call: suggest subject
	// after dot and outside of method call: suggest verb
	// inside of method call: suggest appropriate object from dtypes
	// last token object: suggest connector

	// TODO: Test expected tokens when subject blank, when verb blank, when object blank, when tag blank, etc.
	// End tokens: TokenBang, TokenDot, TokenLParen, TokenRParen
	// if last token is end token, and no error: suggest from expected tokens
	return nil, nil
}

func getSubject(s string) (*Subject, error) {
	for _, subject := range validSubjects {
		if toLowerCase(subject.Name) == toLowerCase(s) {
			return &subject, nil
		}
		for _, alias := range subject.Aliases {
			if toLowerCase(s) == toLowerCase(alias) {
				return &subject, nil
			}
		}
	}

	return nil, ErrInvalidToken{}
}

func (e *CompletionEngine) SuggestSubject(s string) ([]string, error) {
	subjects := make([]string, 0)
	for _, subject := range e.subjectTrie.Search([]string{s}).Results {
		subjects = append(subjects, subject.Value.(Subject).Name)
	}

	return subjects, nil
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

type AutocompleteError struct {
	Code     AutocompleteErrorCode
	Position int
	Message  string
}

func NewAutocompleteError(code AutocompleteErrorCode, pos int, msg string) *AutocompleteError {
	return &AutocompleteError{
		Code:     code,
		Position: pos,
		Message:  msg,
	}
}

func (e *AutocompleteError) Error() string {
	return e.Message
}

type AutocompleteErrorCode int

const (
	InvalidCharacter AutocompleteErrorCode = iota
)
