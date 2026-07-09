package forms_test

import (
	"testing"

	"cchoice/internal/constants"
	"cchoice/internal/server/forms"

	"github.com/stretchr/testify/assert"
)

func TestCategorySectionQuery_EffectiveLimit(t *testing.T) {
	tests := []struct {
		name  string
		query forms.CategorySectionQuery
		want  int
	}{
		{
			name:  "zero uses default",
			query: forms.CategorySectionQuery{},
			want:  constants.DefaultLimitCategories,
		},
		{
			name:  "negative uses default",
			query: forms.CategorySectionQuery{Limit: -1},
			want:  constants.DefaultLimitCategories,
		},
		{
			name:  "below minimum clamps up",
			query: forms.CategorySectionQuery{Limit: 3},
			want:  constants.DefaultLimitCategories,
		},
		{
			name:  "at minimum",
			query: forms.CategorySectionQuery{Limit: constants.DefaultLimitCategories},
			want:  constants.DefaultLimitCategories,
		},
		{
			name:  "above minimum",
			query: forms.CategorySectionQuery{Limit: 20},
			want:  20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.query.EffectiveLimit())
		})
	}
}
