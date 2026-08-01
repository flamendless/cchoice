package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

func resolveAvailableBrandSlug(
	ctx context.Context,
	q *queries.Queries,
	name string,
	excludeBrandID int64,
) (string, error) {
	base := utils.BrandSlug(name)
	if base == "" {
		return "", errs.ErrInvalidParams
	}

	for suffix := 0; suffix < 1000; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}

		var (
			exists bool
			err    error
		)
		if excludeBrandID > 0 {
			exists, err = q.BrandSlugExistsForOtherBrand(ctx, queries.BrandSlugExistsForOtherBrandParams{
				Slug: sql.NullString{String: candidate, Valid: true},
				ID:   excludeBrandID,
			})
		} else {
			exists, err = q.BrandSlugExists(ctx, sql.NullString{String: candidate, Valid: true})
		}
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", errors.New("could not resolve unique brand slug")
}
