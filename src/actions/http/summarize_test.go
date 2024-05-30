package http

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"oh-my-chat/src/utils"
)

type testCase struct {
	desc      string
	response  []byte
	expect    string
	fields    SummarizeFields
	separator Separator
}

var (
	testCase1 testCase = testCase{
		desc: "something",
		response: []byte(`{
      "age":37,
      "children": ["Sara","Alex","Jack"],
      "fav.movie": "Deer Hunter",
      "friends": [
        {"age": 44, "first": "Dale", "last": "Murphy"},
        {"age": 68, "first": "Roger", "last": "Craig"},
        {"age": 47, "first": "Jane", "last": "Murphy"}
      ],
      "name": {"first": "Tom", "last": "Anderson"}
      }`),
		expect: utils.NewStringBuilder().
			NextLine("Age: 37").
			NextLine("Name: Tom").
			NextLine("Dale Last name: Murphy").
			String(),
		fields: SummarizeFields{
			SummarizeField{Name: "Age", Path: "age"},
			SummarizeField{Name: "Name", Path: "name.first"},
			SummarizeField{Name: "Dale Last name", Path: "friends.0.last"},
		},
		separator: ColonStyle,
	}
	testCase2 = testCase{
		desc: "something2",
		response: []byte(`{
      "age":37,
      "children": ["Sara","Alex","Jack"],
      "fav.movie": "Deer Hunter",
      "friends": [
        {"age": 44, "first": "Dale", "last": "Murphy"},
        {"age": 68, "first": "Roger", "last": "Craig"},
        {"age": 47, "first": "Jane", "last": "Murphy"}
      ],
      "name": {"first": "Tom", "last": "Anderson"}
      }`),
		expect: utils.NewStringBuilder().
			NextLine("Friends: Dale, Roger, Jane").
			NextLine("Name: ommitted").
			String(),
		fields: SummarizeFields{
			SummarizeField{Name: "Friends", Path: "friends.#.first"},
			SummarizeField{Name: "Name", Path: "name"},
		},
		separator: ColonStyle,
	}
	testCase3 = testCase{
		desc: "something3",
		response: []byte(`{
      "age":37,
      "children": ["Sara","Alex","Jack"],
      "fav.movie": "Deer Hunter",
      "friends": [
        {"age": 34, "first": "Alice", "last": "Smith"},
        {"age": 29, "first": "Bob", "last": "Johnson"},
        {"age": 47, "first": "Jane", "last": "Williams"},
        {"age": 52, "first": "John", "last": "Brown"},
        {"age": 63, "first": "Eve", "last": "Jones"},
        {"age": 38, "first": "Dale", "last": "Murphy"},
        {"age": 44, "first": "Roger", "last": "Craig"},
        {"age": 41, "first": "Alice", "last": "Johnson"},
        {"age": 53, "first": "Bob", "last": "Smith"},
        {"age": 27, "first": "Jane", "last": "Murphy"},
        {"age": 39, "first": "John", "last": "Williams"},
        {"age": 48, "first": "Eve", "last": "Brown"}
      ],
      "name": {"first": "Tom", "last": "Anderson"}
      }`),
		expect: utils.NewStringBuilder().
			NextLine("Friends Ages 34, 29, 47, 52, 63, 38, 44, 41, 53, 27, ...").
			NextLine("Children Sara, Alex, Jack").
			String(),
		fields: SummarizeFields{
			SummarizeField{Name: "Friends Ages", Path: "friends.#.age"},
			SummarizeField{Name: "Children", Path: "children"},
		},
		separator: WriteSpaceStyle,
	}
	testCase4 testCase = testCase{
		desc: "custom style separator",
		response: []byte(`{
      "age":37,
      "children": ["Sara","Alex","Jack"],
      "fav.movie": "Deer Hunter",
      "friends": [
        {"age": 44, "first": "Dale", "last": "Murphy"},
        {"age": 68, "first": "Roger", "last": "Craig"},
        {"age": 47, "first": "Jane", "last": "Murphy"}
      ],
      "name": {"first": "Tom", "last": "Anderson"}
      }`),
		expect: utils.NewStringBuilder().
			NextLine("Age-> 37").
			NextLine("Name-> Tom").
			String(),
		fields: SummarizeFields{
			SummarizeField{Name: "Age", Path: "age"},
			SummarizeField{Name: "Name", Path: "name.first"},
		},
		separator: Separator("-> "),
	}
)

func TestSummarize(t *testing.T) {
	assert := assert.New(t)

	for _, _case := range []testCase{
		testCase1,
		testCase2,
		testCase3,
	} {
		t.Run(_case.desc, func(t *testing.T) {
			result := Summarize(_case.response, _case.fields, _case.separator)
			assert.Equal(_case.expect, result)
		})
	}

}