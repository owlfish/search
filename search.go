/*
The search library provides a simple query language for searching records.

The query language features are:

 * boat whale - must contain both `boat` and `whale`
 * boat OR whale - must contain either `boat` or `whale`
 * boat whale OR shark - must contain `boat` and either `whale` or `shark`
 * boat whale NOT shark - must contain both `boat` and `whale` and not contain `shark`
 * "floating boat" whale - must contain the phrase "floating boat" and the word `whale`
 * boat whale tag:book - must contain both `boat` and `whale` and the `tag` field must contain the word `book`
 * boat tag:book OR tag:"published leaflet" - must contain the word `boat` and either the `tag` field must have the word `book` or the phrase `published leaflet`

Such queries are parsed using the QueryParser function, which returns a Query
object.  Query objects are able to search any object that implements the
Searchable interface.

*/
package search

import (
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

func (q filters) Search(s Searchable) (result bool) {
	for _, filt := range q {
		localResult := filt(s)
		if !localResult {
			return false
		}
	}
	return true
}

// mustContain returns true if the Searchable matches the field and phrase
func mustContain(field, phrase string) filter {
	return func(s Searchable) bool {
		if s.Contains(field, phrase) {
			return true
		}
		return false
	}
}

// mustContain returns true if the Searchable does not match the field and phrase
func mustNotContain(field, phrase string) filter {
	return func(s Searchable) bool {
		if s.Contains(field, phrase) {
			return false
		}
		return true
	}
}

// orFilter tries each subfilter until one matches.  If none match it returns false
func orFilter(subfilters ...filter) filter {
	return func(s Searchable) bool {
		for _, f := range subfilters {
			if f(s) {
				return true
			}
		}
		return false
	}
}

/*
QueryParser truns a string such as "book whale" into a Query.
*/
func QueryParser(query string) (q Query) {
	var phraseStart, phraseEnd int
	var orPhrase, notPhrase, inquote bool

	results := make(filters, 0, 5)

	// Closure to handle any found search phrases
	// The closure ensures that the same logic is used inside and outside of the loop
	phraseHandler := func() {
		if !(phraseStart == phraseEnd+1) {
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
						results[len(results)-1] = orFilter(previousFilter, mustContain(fieldName, fieldValue))
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
			} else {
				phraseEnd = pos
			}
		}
	}
	// End of all phrases, spit it out.
	phraseHandler()
	return results
}
