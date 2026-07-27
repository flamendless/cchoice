package services

import (
	"context"
	"strings"

	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
)

func ParsePromoLinkFields(
	ctx context.Context,
	promoType enums.PromoType,
	linkType enums.PromoLinkType,
	trackedLinkID string,
	linkURL string,
	trackedLink *TrackedLinkService,
) (PromoLinkFields, error) {
	trackedLinkID = strings.TrimSpace(trackedLinkID)
	linkURL = strings.TrimSpace(linkURL)
	if linkType == enums.PROMO_LINK_TYPE_UNDEFINED {
		linkType = enums.PROMO_LINK_TYPE_NONE
	}

	if promoType == enums.PROMO_TYPE_BANNER_VIDEO {
		if linkType != enums.PROMO_LINK_TYPE_NONE || trackedLinkID != "" || linkURL != "" {
			return PromoLinkFields{}, errs.ErrPromoLinkImageOnly
		}
		return PromoLinkFields{}, nil
	}

	switch linkType {
	case enums.PROMO_LINK_TYPE_NONE:
		return PromoLinkFields{}, nil
	case enums.PROMO_LINK_TYPE_TRACKED:
		if trackedLinkID == "" {
			return PromoLinkFields{}, errs.ErrPromoLinkRequired
		}
		if linkURL != "" {
			return PromoLinkFields{}, errs.ErrPromoLinkConflict
		}
		link, err := trackedLink.GetTrackedLinkByID(ctx, trackedLinkID)
		if err != nil {
			return PromoLinkFields{}, err
		}
		if link == nil || link.Status == enums.TRACKED_LINK_STATUS_DELETED {
			return PromoLinkFields{}, errs.ErrNotFound
		}
		return PromoLinkFields{TrackedLinkID: trackedLinkID}, nil
	case enums.PROMO_LINK_TYPE_CUSTOM:
		if linkURL == "" {
			return PromoLinkFields{}, errs.ErrPromoLinkRequired
		}
		if trackedLinkID != "" {
			return PromoLinkFields{}, errs.ErrPromoLinkConflict
		}
		if !isValidPromoLinkURL(linkURL) {
			return PromoLinkFields{}, errs.ErrPromoLinkInvalid
		}
		return PromoLinkFields{LinkURL: linkURL}, nil
	default:
		return PromoLinkFields{}, errs.ErrPromoLinkInvalid
	}
}

func isValidPromoLinkURL(linkURL string) bool {
	if utils.IsExternalURL(linkURL) {
		return true
	}
	return strings.HasPrefix(linkURL, "/")
}

func PromoLinkTypeFromFields(fields PromoLinkFields) enums.PromoLinkType {
	if fields.TrackedLinkID != "" {
		return enums.PROMO_LINK_TYPE_TRACKED
	}
	if fields.LinkURL != "" {
		return enums.PROMO_LINK_TYPE_CUSTOM
	}
	return enums.PROMO_LINK_TYPE_NONE
}
