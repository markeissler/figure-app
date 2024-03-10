package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDigitCount(t *testing.T) {
	testCases := []struct {
		name  string
		given int
		want  int
	}{
		{
			name:  "1 --> 1",
			given: 1,
			want:  1,
		},
		{
			name:  "12 --> 2",
			given: 12,
			want:  2,
		},
		{
			name:  "99 --> 2",
			given: 99,
			want:  2,
		},
		{
			name:  "102 --> 3",
			given: 102,
			want:  3,
		},
		{
			name:  "215 --> 3",
			given: 215,
			want:  3,
		},
		{
			name:  "999 --> 3",
			given: 999,
			want:  3,
		},
		{
			name:  "1000 --> 4",
			given: 1000,
			want:  4,
		},
		{
			name:  "10030 --> 5",
			given: 10030,
			want:  5,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := DigitCount(testCase.given)
			assert.Equal(t, testCase.want, actual)
		})
	}
}

func TestFirstOrBlank(t *testing.T) {
	testCases := []struct {
		name  string
		given []string
		want  string
	}{
		{
			name: "one empty string",
			given: []string{
				"",
			},
			want: "",
		},
		{
			name: "one non-empty string",
			given: []string{
				"pomegranate",
			},
			want: "pomegranate",
		},
		{
			name: "two non-empty strings",
			given: []string{
				"apple",
				"banana",
			},
			want: "apple",
		},
		{
			name: "one blank string",
			given: []string{
				"    ",
			},
			want: "",
		},
		{
			name:  "nil input",
			given: nil,
			want:  "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := FirstOrBlank(testCase.given...)
			assert.Equal(t, testCase.want, actual)
		})
	}
}
