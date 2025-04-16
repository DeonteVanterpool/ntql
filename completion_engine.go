package ntql

import (
	// "regexp"

	"fmt"

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
	subjectTrie   *trie.Trie
	connectorTrie *trie.Trie
	tagTrie       *trie.Trie

	lexer *Lexer
}

func NewAutocompleteEngine(tags []string) *CompletionEngine {
	return &CompletionEngine{
		subjectTrie:   NewMegaTrie().subjectTrie,
		connectorTrie: NewConnectorTrie(),
		tagTrie:       NewTagTrie(tags),
	}
}

func NewConnectorTrie() *trie.Trie {
	connectorTrie := trie.New()
	for _, connector := range connectorTypes {
		connectorTrie.Put([]string{connector.String()}, connector)
	}

	return connectorTrie
}

func NewTagTrie(tags []string) *trie.Trie {
	tagTrie := trie.New()
	for _, tag := range tags {
		tagTrie.Put([]string{tag}, tag)
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
				fmt.Errorf("Unexpected Error: %v", err.Error())
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
			for _, tag := range e.tagTrie.Search([]string{input}).Results {
				suggestions = append(suggestions, tag.Value.(string))
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

func (e *CompletionEngine) buildTagTrie(tags []string) *trie.Trie {
	tagTrie := trie.New()
	for _, tag := range tags {
		tagTrie.Put([]string{tag}, tag)
	}
	return tagTrie
}

func (e *CompletionEngine) buildConnectorTrie(subject Subject) *trie.Trie {
	connectorTrie := trie.New()
	for _, connector := range connectorTypes {
		connectorTrie.Put([]string{connector.String()}, connector.String())
	}
	return connectorTrie
}

func (e *CompletionEngine) suggestFromSubject(subject Subject, verb string) ([]string, error) {
	verbTrie := e.buildVerbTrie(subject)
	verbs := make([]string, 0)
	for _, v := range verbTrie.Search([]string{verb}).Results {
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
