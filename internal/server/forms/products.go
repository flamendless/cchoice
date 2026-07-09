package forms

type ProductSlugPath struct {
	Slug string `param:"slug" validate:"required"`
}
