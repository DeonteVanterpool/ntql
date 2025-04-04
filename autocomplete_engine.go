package tbql

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

// CompletionEngine is a trie-based autocompletion engine for TBQL
type CompletionEngine struct {
	// checks to see if we're in a function call or not
	innerParens int

	subjectTrie *trie.Trie

	subject Subject

	verbTrie *trie.Trie

	lexer *Lexer
}

func NewAutocompleteEngine(tags []string) *CompletionEngine {
	return &CompletionEngine{
		subjectTrie: NewMegaTrie().subjectTrie,
		innerParens: 0,
	}
}

func (e *CompletionEngine) Suggest(s string) ([]string, error) {
	if len(s) == 0 {
		// return e.Subject()
	}

	e.lexer = NewLexer(s)

    // at each space, we must suggest something else. at each dot, we must suggest a verb. at each open paren, we must suggest an object. at each close paren, we must suggest a connector. if outside of a method call, suggest a subject. if inside of a method call, suggest an object. Skip strings
	_, err := e.lexer.Lex()
	if err != nil {
		switch err.(type) {
		case ErrEndOfInput:
			return nil, nil
		default:
			return nil, err
		}
	}
	if !e.lexer.atEnd() {

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
