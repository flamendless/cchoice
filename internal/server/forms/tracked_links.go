package forms

type TrackedLinkPath struct {
	Slug string `param:"slug" validate:"required"`
}

type TrackedLinkUTMQuery struct {
	UTMSource   string `form:"utm_source"`
	UTMMedium   string `form:"utm_medium"`
	UTMCampaign string `form:"utm_campaign"`
}

type TrackedLinkQRPath struct {
	ID string `param:"id" validate:"required"`
}
