# Go Search / Query library

This library turns queries such as `"Go for gold" type:book OR type:article` into a Query object that returns true or false if a Searchable object matches.

## Example

```
type texts struct {
	Title string
	Type  string
}

func (eo *texts) Contains(field, phrase string) (present bool) {
	switch field {
	case "type":
		return strings.Contains(eo.Type, phrase)
	default:
		return strings.Contains(eo.Title, phrase)
	}
}

func ExampleSearchable() {
	text1 := &texts{Title: "Go for gold and other tales", Type: "book"}
	text2 := &texts{Title: "Not written, Go for gold and nothing more", Type: "article"}
	text3 := &texts{Title: "Going for gold and nothing more", Type: "book"}

	query := QueryParser(`"Go for gold" type:book OR type:article`)

	fmt.Printf("text1 - match: %v\n", query.Search(text1))
	fmt.Printf("text2 - match: %v\n", query.Search(text2))
	fmt.Printf("text3 - match: %v\n", query.Search(text3))
	// Output:
	// text1 - match: true
	// text2 - match: true
	// text3 - match: false
}
```

## Documentation

Documentation is provided in [GoDoc](https://godoc.org/github.com/owlfish/search)

