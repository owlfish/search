package search

import (
	"strings"
	"testing"
)

var testMaterial = SearchableStringSlice([]string{`Raw testing subject pingo goes here.`,
	`Raw body of the test message goes here.
More than one line exists!`,
})

type testSearchObject struct {
	Title string
	Body  string
}

func (tso *testSearchObject) Contains(field, phrase string) (present bool) {
	switch field {
	case "title":
		return strings.Contains(tso.Title, phrase)
	case "body":
		return strings.Contains(tso.Body, phrase)
	default:
		if strings.Contains(tso.Title, phrase) {
			return true
		}
		return strings.Contains(tso.Body, phrase)
	}
}

var testFieldMaterial = &testSearchObject{
	Title: "Once upon a very merry time",
	Body:  "A bettle battle fought in a bottle",
}

var testCases = []struct {
	Name      string
	Condition string
	Result    bool
	Records   Searchable
}{
	{
		"testSearchNoMatch",
		"frog",
		false,
		testMaterial,
	},
	{
		"testSearchSimpleMatch",
		"test",
		true,
		testMaterial,
	},
	{
		"testSearchTwoWordsNoMatch",
		"test frog",
		false,
		testMaterial,
	},
	{
		"testSearchTwoWordsMatch",
		"test pingo",
		true,
		testMaterial,
	},
	{
		"testSearchWordAndPhraseMatch",
		"test 'subject pingo'",
		true,
		testMaterial,
	},
	{
		"testSearchORMatch",
		"frog OR test",
		true,
		testMaterial,
	},
	{
		"testSearchORPhraseMatch",
		"frog OR 'subject pingo'",
		true,
		testMaterial,
	},
	{
		"testSearchORNoMatch",
		"frog OR boil",
		false,
		testMaterial,
	},
	{
		"testSearchORThreeMatch",
		"frog OR boil OR test",
		true,
		testMaterial,
	},
	{
		"testSearchORThreeNoMatch",
		"frog OR boil OR witch",
		false,
		testMaterial,
	},
	{
		"testSearchLowerOrNoMatch",
		"test or frog",
		false,
		testMaterial,
	},
	{
		"testSearchNotPresent",
		"NOT frog",
		true,
		testMaterial,
	},
	{
		"testSearchNotPresent2",
		"test NOT frog",
		true,
		testMaterial,
	},
	{
		"testSearchNotPresentNotFound",
		"boil NOT frog",
		false,
		testMaterial,
	},
	{
		"testSearchNotPresentOr",
		"boil OR test NOT frog",
		true,
		testMaterial,
	},
	{
		"testSearchNotPresentOrNotFound",
		"boil OR witch NOT frog",
		false,
		testMaterial,
	},
	{
		"testEarlyOr",
		"OR test",
		true,
		testMaterial,
	},
	{
		"testEarlyOrNotFound",
		"OR frog",
		false,
		testMaterial,
	},
	{
		"testLateORNotFound",
		"frog OR",
		false,
		testMaterial,
	},
	{
		"testLateOR",
		"test OR",
		true,
		testMaterial,
	},
	{
		"testFieldAny",
		"merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldAny2",
		"battle",
		true,
		testFieldMaterial,
	},
	{
		"testFieldTitle",
		"title:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldTitleNotFound",
		"title:battle",
		false,
		testFieldMaterial,
	},
	{
		"testFieldBody",
		"body:battle",
		true,
		testFieldMaterial,
	},
	{
		"testFieldBodyNotFound",
		"body:merry",
		false,
		testFieldMaterial,
	},
	{
		"testFieldOrNoField",
		"body:merry OR battle",
		true,
		testFieldMaterial,
	},
	{
		"testFieldOrNoField2",
		"battle OR body:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldOrField",
		"title:merry OR body:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldOrField2",
		"body:merry OR title:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldNotField",
		"title:merry NOT body:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldNotField2",
		"NOT body:merry title:merry",
		true,
		testFieldMaterial,
	},
	{
		"testFieldNotField3",
		"title:merry NOT body:battle",
		false,
		testFieldMaterial,
	},
	{
		"testFieldWithQuotes",
		"'title:upon a very'",
		true,
		testFieldMaterial,
	},
	{
		"testFieldWithQuotes2",
		"title:'upon a very'",
		true,
		testFieldMaterial,
	},
	{
		"testFieldWithQuotes3",
		"title:\"upon a very\"",
		true,
		testFieldMaterial,
	},
	{
		"testFieldWithQuotesOR",
		"frog OR title:\"upon a very\"",
		true,
		testFieldMaterial,
	},
	{
		"testFieldWithQuotesORNot",
		"frog OR title:\"upon a very\" NOT body:'battle fought'",
		false,
		testFieldMaterial,
	},
}

func TestSearch(t *testing.T) {
	for _, test := range testCases {
		query := QueryParser(test.Condition)
		result := query.Search(test.Records)
		if result != test.Result {
			t.Errorf("%v failed, expected %v, got %v for search condition %v\n", test.Name, test.Result, result, test.Condition)
		}
	}
}
