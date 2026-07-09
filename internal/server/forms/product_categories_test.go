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
			name:  "zero uses shop default",
			query: forms.CategorySectionQuery{},
			want:  constants.DefaultShopCategorySectionsPerPage,
		},
		{
			name:  "negative uses shop default",
			query: forms.CategorySectionQuery{Limit: -1},
			want:  constants.DefaultShopCategorySectionsPerPage,
		},
		{
			name:  "explicit shop page size",
			query: forms.CategorySectionQuery{Limit: constants.DefaultShopCategorySectionsPerPage},
			want:  constants.DefaultShopCategorySectionsPerPage,
		},
		{
			name:  "above default",
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
