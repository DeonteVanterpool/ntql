package ntql

import (
	// "regexp"

	"fmt"

	trie "github.com/Vivino/go-autocomplete-trie"
)

type EngineState int

type MegaTrie struct {
	subjectTrie *trie.Trie
}

func NewMegaTrie() *MegaTrie {
	subjectTrie := trie.New()
	for _, subject := range validSubjects {
		for _, alias := range subject.Aliases {
			subjectTrie.Insert(alias)
		}
		subjectTrie.Insert(subject.Name)
	}
	return &MegaTrie{
		subjectTrie: subjectTrie,
	}
}

// CompletionEngine is a trie-based autocompletion engine for NTQL
type CompletionEngine struct {
	subjectTrie   *trie.Trie
	connectorTrie *trie.Trie
	tagTrie       *trie.Trie
	tags          []string
	subjects      []string
	verbs         []string
	connectors    []string

	lexer *Lexer
}

func GetValidSubjects() []string {
	subjects := make([]string, 0)
	for _, subject := range validSubjects {
		subjects = append(subjects, subject.Name)
		for _, alias := range subject.Aliases {
			subjects = append(subjects, alias)
		}
	}
	return subjects
}

func GetValidConnectors() []string {
	connectors := make([]string, 0)
	for _, connector := range connectorTypes {
		connectors = append(connectors, connector.String())
	}
	return connectors
}

func NewAutocompleteEngine(tags []string) *CompletionEngine {
	return &CompletionEngine{
		subjectTrie:   NewMegaTrie().subjectTrie,
		connectorTrie: NewConnectorTrie(),
		tagTrie:       NewTagTrie(tags),

		tags:       tags,
		subjects:   GetValidSubjects(),
		connectors: GetValidConnectors(),
	}
}

func NewConnectorTrie() *trie.Trie {
	connectorTrie := trie.New()
	for _, connector := range connectorTypes {
		connectorTrie.Insert(connector.String())
	}

	return connectorTrie
}

func NewTagTrie(tags []string) *trie.Trie {
	tagTrie := trie.New()
	for _, tag := range tags {
		tagTrie.Insert(tag, tag)
	}

	return tagTrie
}

func (e *CompletionEngine) Suggest(s string) ([]string, error) {
	if len(s) == 0 {
		return e.SuggestSubject("")
	}

	e.lexer = NewLexer(s)

	var lastToken Token
	var lastSubject *Subject = nil
	for {
		err := e.lexer.ScanToken()
		exit := false
		if err != nil {
			switch err.(type) {
			case ErrEndOfInput:
				exit = true
			case ErrInvalidSubject:
				err := err.(ErrInvalidSubject)
				return e.SuggestSubject(string(err.Lexeme))
			case ErrInvalidToken:
				return []string{}, nil
			default:
				return nil, fmt.Errorf("Unexpected Error: %v", err.Error())
			}
		}
		lastToken, err = e.lexer.lastToken()
		if err != nil {
			return e.SuggestSubject("")
		}
		if lastToken.Kind == TokenSubject {
			lastSubject, err = getSubject(string(lastToken.Literal))
			if err != nil { // invalid subject
				return e.SuggestSubject(string(lastToken.Literal))
			}
		}
		if exit {
			break
		}
	}

	if lastCharSpace(s) {
		switch lastToken.Kind {
		case TokenSubject, TokenVerb, TokenBang, TokenLParen:
			return []string{}, nil
		case TokenTag, TokenBool, TokenString, TokenInt, TokenDate, TokenDateTime, TokenRParen:
			return e.suggestConnector("")
		case TokenOr, TokenAnd:
			if e.lexer.insideMethodCall() {
				return e.suggestObjects(*lastSubject, "")
			} else {
				return e.SuggestSubject("")
			}
		case TokenDot:
			return e.suggestFromSubject(*lastSubject, "")
		default:
			panic("Unimplemented token type in switch statemen")
		}
	} else {
		switch lastToken.Kind {
		case TokenSubject:
			return e.SuggestSubject(lastToken.Literal)
		case TokenVerb:
			return e.suggestFromSubject(*lastSubject, lastToken.Literal)
		case TokenTag, TokenBool, TokenString, TokenInt, TokenDate, TokenDateTime:
			return e.suggestObjects(*lastSubject, lastToken.Literal)
		case TokenOr, TokenAnd, TokenRParen: // TODO: handle incomplete connector cases on error
			return []string{}, nil
		case TokenDot:
			return e.suggestFromSubject(*lastSubject, lastToken.Literal)
		case TokenBang, TokenLParen:
			if e.lexer.insideMethodCall() {
				return e.suggestObjects(*lastSubject, "")
			} else {
				return e.SuggestSubject("")
			}
		}
	}

	// TODO: Test expected tokens when subject blank, when verb blank, when object blank, when tag blank, etc.
	// End tokens: TokenBang, TokenDot, TokenLParen, TokenRParen
	// if last token is end token, and no error: suggest from expected tokens
	return nil, nil
}

func lastCharSpace(s string) bool {
	return s[len(s)-1] == ' '
}

func (e *CompletionEngine) suggestConnector(s string) ([]string, error) {
	if s == "" {
		return e.connectors, nil
	}
	connectors := make([]string, 0)
	for _, c := range connectorTypes {
		connectors = append(connectors, c.String())
	}

	return connectors, nil
}

func (e *CompletionEngine) suggestObjects(subject Subject, input string) ([]string, error) {
	suggestions := make([]string, 0)
	for _, subj := range subject.ValidTypes {
		switch subj {
		case DTypeTag:
			if input == "" {
				suggestions = append(suggestions, e.tags...)
				continue
			}
			for _, tag := range e.tagTrie.SearchAll(input) {
				suggestions = append(suggestions, tag)
			}
		case DTypeString, DTypeInt:
		case DTypeDate:
			suggestions = append(suggestions, "today", "yesterday", "tomorrow")
		case DTypeDateTime:
			suggestions = append(suggestions, "now")
		}
	}
	return suggestions, nil
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
	if s == "" {
		return e.subjects, nil
	}
	subjects := make([]string, 0)
	for _, subject := range e.subjectTrie.SearchAll(s) {
		subjects = append(subjects, subject)
	}

	return subjects, nil
}

func (e *CompletionEngine) buildVerbTrie(subject Subject) *trie.Trie {
	verbTrie := trie.New()
	verbs := make([]string, 0)
	for _, verb := range subject.ValidVerbs {
		verbs = append(verbs, verb.Name)
		verbTrie.Insert(verb.Name)
		for _, alias := range verb.Aliases {
			verbTrie.Insert(alias)
			verbs = append(verbs, alias)
		}
	}

	e.verbs = verbs
	return verbTrie
}

func (e *CompletionEngine) buildTagTrie(tags []string) *trie.Trie {
	tagTrie := trie.New()
	for _, tag := range tags {
		tagTrie.Insert(tag)
	}
	return tagTrie
}

func (e *CompletionEngine) buildConnectorTrie(subject Subject) *trie.Trie {
	connectorTrie := trie.New()
	for _, connector := range connectorTypes {
		connectorTrie.Insert(connector.String())
	}
	return connectorTrie
}

func (e *CompletionEngine) suggestFromSubject(subject Subject, verb string) ([]string, error) {

	if verb == "" {
		return e.verbs, nil
	}
	verbTrie := e.buildVerbTrie(subject)
	verbs := make([]string, 0)
	for _, v := range verbTrie.SearchAll(verb) {
		verbs = append(verbs, v)
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
