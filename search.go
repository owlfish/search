/*
The search library provides a simple query language for searching records.

The query language features are:

 * boat whale - must contain both `boat` and `whale`
 * boat OR whale - must contain either `boat` or `whale`
 * boat whale OR shark - must contain `boat` and either `whale` or `shark`
 * boat whale NOT shark - must contain both `boat` and `whale` and not contain `shark`
 * "floating boat" whale - must contain the phrase "floating boat" and the word `whale`
 * boat whale tag:book - must contain both `boat` and `whale` and the `tag` field must contain `book`
 * boat tag:book OR tag:"published leaflet" - must contain the word `boat` and the `tag` field must either have `book` or the phrase `published leaflet`
 * boat OR NOT (tag:book OR tag:leaflet) - must contain 'boat' or the tag field must not contain 'book' or 'leaflet'

Such queries are parsed using the QueryParser function, which returns a Query
object.  Query objects are able to search any object that implements the
Searchable interface.

*/
package search

import (
	// "log"
	"strings"
	"unicode"
)

/*
Searchable objects can be searched by a Query to see if they match.
*/
type Searchable interface {
	/*
		Contains returns true if the phrase is present in the object, optionally restricted to the given field.
	*/
	Contains(field, phrase string) (present bool)
}

/*
SearchableFunc allows functions to implement the Searchable interface.
*/
type SearchableFunc func(field, phrase string) (present bool)

/*
Contains calls the SearchableFunc
*/
func (sf SearchableFunc) Contains(field, phrase string) (present bool) {
	return sf(field, phrase)
}

/*
SearchableStringSlice makes a slice of strings Searchable.

Each string in the slice is tested against the Query and returns true if any
matches.
*/
func SearchableStringSlice(record []string) SearchableFunc {
	return func(field, phrase string) bool {
		for _, str := range record {
			if strings.Contains(str, phrase) {
				return true
			}
		}
		return false
	}
}

/*
SearchableString makes a string Searchable.

The query will be tested against the string, returning true if it matches.
*/
func SearchableString(record string) SearchableFunc {
	return func(field, phrase string) bool {
		return strings.Contains(record, phrase)
	}
}

/*
A filter function is part of a Query that executes searches.

The filter function calls Search on the Searchable interface and tells the
Query whether it matches.

`match` is true if the Searchable does match the filter.
*/
type filter func(Searchable) (match bool)

/*
A Query object is returned by QueryPraser to handle executing seraches.

The Search method takes an object implementing the Searchable inteface and
returns whether it matches the query.
*/
type Query interface {
	/*
		Execute the query against the Searchable object s.

		Match is true if the searchable object satisfies the query.
	*/
	Search(s Searchable) (match bool)
}

// filters implements the Query interface for the package
type filters []filter

// Filters default to AND - as soon as one term doesn't match, return false
func (q filters) Search(s Searchable) (result bool) {
	for _, filt := range q {
		if !filt(s) {
			// log.Printf("Search of filter %v returned false\n", i)
			return false
		}
	}
	return true
}

// mustContain returns true if the Searchable matches the field and phrase
func mustContain(field, phrase string) filter {
	// log.Printf("Adding must contain %v:%v\n", field, phrase)
	return func(s Searchable) bool {
		if s.Contains(field, phrase) {
			// log.Printf("Must contain %v:%v returns true\n", field, phrase)
			return true
		}
		// log.Printf("Must contain %v:%v returns false\n", field, phrase)
		return false
	}
}

// mustContain returns true if the Searchable does not match the field and phrase
func mustNotContain(field, phrase string) filter {
	// log.Printf("Adding must NOT contain %v:%v\n", field, phrase)
	return func(s Searchable) bool {
		if s.Contains(field, phrase) {
			return false
		}
		return true
	}
}

// orFilter tries each subfilter until one matches.  If none match it returns false
func orFilter(subfilters ...filter) filter {
	// log.Printf("Adding OR filter with %v\n", subfilters)
	return func(s Searchable) bool {
		for _, f := range subfilters {
			if f(s) {
				// log.Printf("orFilter %v returned true\n", i)
				return true
			}
		}
		return false
	}
}

// notFilter runs each subfilter as AND and then inverts the result
func notFilter(subfilters ...filter) filter {
	// log.Printf("Adding NOT filter with %v\n", subfilters)
	return func(s Searchable) bool {
		for _, f := range subfilters {
			// If the result is false, then the AND is false, so we return true
			if !f(s) {
				// log.Printf("notFilter %v returned false, returning true\n", i)
				return true
			}
		}
		// All results were true, so we return false
		// log.Printf("All not filters returned true, returning false\n")
		return false
	}
}

type queryParserFrame struct {
	filters   filters
	orPhrase  bool
	notPhrase bool
}

/*
QueryParser truns a string such as "book whale" into a Query.
*/
func QueryParser(query string) (q Query) {
	var phraseStart, phraseEnd int
	var orPhrase, notPhrase, inquote bool

	query = strings.TrimSpace(query)

	results := make(filters, 0, 5)

	stack := make([]queryParserFrame, 0, 2)

	popStack := func() {
		// Do nothing if there is nothing on the stack.
		if len(stack) == 0 {
			return
		}
		stackFrame := stack[len(stack)-1]
		// log.Printf("Popping stack: %v\n", stackFrame)
		stack = stack[:len(stack)-1]
		// Stick the nested results into the previous frame
		bracketResults := results
		results = stackFrame.filters
		orPhrase = stackFrame.orPhrase
		notPhrase = stackFrame.notPhrase

		// We have just closed brackets - now need to add the contents into the main results.
		// To do this we need to know whether they are NOT or OR or default AND
		if orPhrase {
			// Try and build an OR with the previous phrase
			if len(results) > 0 {
				previousFilter := results[len(results)-1]
				// Is this a compound OR NOT search?
				if notPhrase {
					// log.Printf("Adding in the OR with NOT the bracketResults.Search %v\n", bracketResults)
					results[len(results)-1] = orFilter(previousFilter, notFilter(bracketResults...))
				} else {
					// log.Printf("Adding in the OR with the bracketResults.Search %v\n", bracketResults)
					results[len(results)-1] = orFilter(previousFilter, bracketResults.Search)
				}
			} else {
				// Suppress the OR and search for it
				// log.Printf("Suppressing OR and adding %v as AND\n", bracketResults)
				results = append(results, bracketResults.Search)
			}
		} else if notPhrase {
			// log.Printf("Adding bracket results %v as a NOT AND\n", bracketResults)
			results = append(results, notFilter(bracketResults...))
		} else {
			// log.Printf("Adding bracket results %v as an AND\n", bracketResults)
			results = append(results, bracketResults.Search)
		}

		orPhrase = false
		notPhrase = false
	}

	pushStack := func() {
		stackFrame := queryParserFrame{
			filters:   results,
			orPhrase:  orPhrase,
			notPhrase: notPhrase,
		}
		// log.Printf("Pushing stack: %v\n", stackFrame)
		stack = append(stack, stackFrame)
		results = make(filters, 0, 5)
		orPhrase = false
		notPhrase = false
	}

	// Closure to handle any found search phrases
	// The closure ensures that the same logic is used inside and outside of the loop
	phraseHandler := func() {
		if phraseStart < phraseEnd {
			phraseValue := query[phraseStart : phraseEnd+1]
			if phraseValue == "OR" {
				// Treat the next phrase as an OR with the previous one
				orPhrase = true
			} else if phraseValue == "NOT" {
				// Treat next phrase as a must not contain
				notPhrase = true
			} else {
				fieldBreak := strings.Index(phraseValue, ":")
				var fieldName, fieldValue string
				if fieldBreak > 0 {
					fieldName = phraseValue[:fieldBreak]
					fieldValue = phraseValue[fieldBreak+1:]
					// Remove any stray quotes, handles the form title:"A book"
					fieldValue = strings.Replace(fieldValue, "'", "", -1)
					fieldValue = strings.Replace(fieldValue, "\"", "", -1)
				} else {
					fieldValue = phraseValue
				}
				if orPhrase {
					// Try and build an OR with the previous phrase
					if len(results) > 0 {
						previousFilter := results[len(results)-1]
						// Is this a compound OR NOT search?
						if notPhrase {
							results[len(results)-1] = orFilter(previousFilter, mustNotContain(fieldName, fieldValue))
						} else {
							results[len(results)-1] = orFilter(previousFilter, mustContain(fieldName, fieldValue))
						}
					} else {
						// Suppress the OR and search for it
						results = append(results, mustContain(fieldName, fieldValue))
					}
				} else if notPhrase {
					results = append(results, mustNotContain(fieldName, fieldValue))
				} else {
					results = append(results, mustContain(fieldName, fieldValue))
				}
				orPhrase = false
				notPhrase = false
			}
		}
	}

	for pos, char := range query {
		if unicode.IsSpace(char) {
			if !inquote {
				// End of a phrase, spit it out.
				phraseHandler()
				phraseStart = pos + 1
			} else {
				phraseEnd = pos
			}
		} else if pos == phraseStart {
			// Begining of a new phrase.
			// Assume we are going to consume a character
			phraseStart++
			if !inquote && (char == '"' || char == '\'') {
				inquote = true
			} else if !inquote && char == '(' {
				pushStack()
			} else if !inquote && char == ')' {
				phraseEnd = pos - 1
				phraseHandler()
				phraseStart = pos + 1
				popStack()
			} else {
				// We didn't consume a character, so keep where we are
				phraseStart--
			}
			phraseEnd = pos
		} else {
			if inquote && (char == '"' || char == '\'') {
				inquote = false
				phraseEnd = pos - 1
			} else if !inquote && (char == '"' || char == '\'') {
				// Quote part way through the phrase, e.g. title:"A book"
				inquote = true
			} else if !inquote && char == ')' {
				phraseEnd = pos - 1
				phraseHandler()
				phraseStart = pos + 1
				popStack()
			} else {
				phraseEnd = pos
			}
		}
	}
	// End of all phrases, spit it out.
	phraseHandler()

	// Close any still open brackets
	for _ = range stack {
		// log.Printf("Handling un-closed stack\n")
		popStack()
	}

	return results
}
