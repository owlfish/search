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

var testFieldMaterialWithSpecialChars = &testSearchObject{
	Title: "Once upon a (very merry) time",
	Body:  "A bettle OR battle NOT fought in a bottle",
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
		"testSearchNoMatchExtraWhiteSpace",
		" frog ",
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
		"testSpecialCharInQuotes",
		"'bettle OR battle'",
		true,
		testFieldMaterialWithSpecialChars,
	},
	{
		"testSpecialCharInQuotes2",
		"'battle NOT fought'",
		true,
		testFieldMaterialWithSpecialChars,
	},
	{
		"testSpecialCharInQuotes3",
		"'(very merry) time'",
		true,
		testFieldMaterialWithSpecialChars,
	},
	{
		"testSpecialCharInQuotes",
		"NOT 'bettle OR battle'",
		false,
		testFieldMaterialWithSpecialChars,
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
	{
		"testOrNotSequenceNoMatch",
		"frog OR NOT body:'battle fought'",
		false,
		testFieldMaterial,
	},
	{
		"testOrNotSequenceMatch",
		"frog OR NOT body:'battle not fought'",
		true,
		testFieldMaterial,
	},
	{
		"testNotOrSequenceNoMatch",
		"frog NOT OR body:'battle fought'",
		false,
		testFieldMaterial,
	},
	{
		"testNotOrSequenceMatch",
		"frog OR NOT body:'battle not fought'",
		true,
		testFieldMaterial,
	},
	// Title: "Once upon a very merry time",
	// Body:  "A bettle battle fought in a bottle",
	{
		"testNotMultiFieldOrSequenceNoMatch1",
		"merry NOT body:'in a bottle' OR NOT body:'battle fought'",
		false,
		testFieldMaterial,
	},
	{
		"testNotMultiFieldOrSequenceMatch1",
		"merry NOT body:'in a bottle' OR NOT body:'battle not fought'",
		true,
		testFieldMaterial,
	},
	{
		"testNotMultiFieldOrSequenceMatch2",
		"merry NOT body:'in a not bottle' OR NOT body:'battle fought'",
		true,
		testFieldMaterial,
	},
	{
		"testOrBracketsSimpleMatch",
		"frog OR (battle fought)",
		true,
		testFieldMaterial,
	},
	{
		"testOrBracketsSimpleMatchFail",
		"frog OR (battle fought frog)",
		false,
		testFieldMaterial,
	},
	{
		"testOrBracketsMultipleORSimpleMatch",
		"upon (frog OR battle OR fought)",
		true,
		testFieldMaterial,
	},
	{
		"testOrBracketsNotMultipleORSimpleMatch",
		"upon NOT (frog OR battle OR fought)",
		false,
		testFieldMaterial,
	},
	{
		"testNotOrBracketsMultipleORSimpleMatch1",
		"upon OR NOT (frog OR battle OR fought)",
		true,
		testFieldMaterial,
	},
	{
		"testNotOrBracketsMultipleORSimpleMatch2",
		"frog OR NOT (frog OR battle OR fought)",
		false,
		testFieldMaterial,
	},
	{
		"nestedBrackets1",
		"frog OR (battle NOT (frog OR things))",
		true,
		testFieldMaterial,
	},
	{
		"nestedBracketsExtraSpaces1",
		"frog   OR (  battle  NOT  (  frog OR things ) )",
		true,
		testFieldMaterial,
	},
	{
		"nestedBracketsMissingClosure",
		"frog   OR (  battle  NOT  (  frog OR things )",
		true,
		testFieldMaterial,
	},
	{
		"nestedBracketsMissingTwoClosure",
		"frog   OR (  battle  NOT  (  frog OR things",
		true,
		testFieldMaterial,
	},
	{
		"extraneousOrInBrackets",
		"(OR battle)",
		true,
		testFieldMaterial,
	},
	{
		"extraneousOrInBrackets2",
		"OR (things)",
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

func TestSearchableString(t *testing.T) {
	a1 := SearchableString("The cat jumped over the mouse")
	q1 := QueryParser("cat jumped")
	if !q1.Search(a1) {
		t.Errorf("Error matching in SearchableString.\n")
	}
	q2 := QueryParser("cat fox")
	if q2.Search(a1) {
		t.Errorf("Error matching in SearchableString.\n")
	}
}
