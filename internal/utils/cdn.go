package utils

import (
	"fmt"
	"strings"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
)

// cdnDefaultVariant is the only named variant configured on the Cloudflare
// Images account. Delivery URLs end with it, and it is the segment replaced by a
// resize transform.
const cdnDefaultVariant = "public"

// cdnTransform resizes without ever upscaling and lets the CDN negotiate the
// encoding (AVIF/WebP) from the request's Accept header.
const cdnTransform = "fit=scale-down,format=auto"

// CDNImageWidth rewrites a Cloudflare Images delivery URL so the image is
// downscaled to width at the edge. URLs that are not CDN delivery URLs, and all
// URLs when flexible variants are not enabled for the account, are returned
// unchanged.
func CDNImageWidth(rawURL string, width int) string {
	if width <= 0 || !conf.Conf().CloudflareImages.FlexibleVariants {
		return rawURL
	}

	base, variant, ok := cutVariant(rawURL)
	if !ok || variant != cdnDefaultVariant {
		return rawURL
	}

	return fmt.Sprintf("%s/w=%d,%s", base, width, cdnTransform)
}

// CDNImageSrcSet builds a srcset for the given candidate widths. It returns an
// empty string when resizing is unavailable, so callers can leave the attribute
// off entirely rather than emitting identical candidates.
func CDNImageSrcSet(rawURL string, widths ...int) string {
	if len(widths) == 0 || !conf.Conf().CloudflareImages.FlexibleVariants {
		return ""
	}

	if _, variant, ok := cutVariant(rawURL); !ok || variant != cdnDefaultVariant {
		return ""
	}

	candidates := make([]string, 0, len(widths))
	for _, width := range widths {
		if width <= 0 {
			continue
		}
		candidates = append(candidates, fmt.Sprintf("%s %dw", CDNImageWidth(rawURL, width), width))
	}
	if len(candidates) == 0 {
		return ""
	}

	return strings.Join(candidates, ", ")
}

// LogoWithTextURL and LogoOnlyURL are used both by the header markup and by the
// preload hints in the document head. They must agree exactly, otherwise the
// browser downloads the logo twice.
func LogoWithTextURL() string {
	return CDNImageWidth(constants.PathSVGLogoWithCompleteText, constants.WidthLogoWithText)
}

func LogoOnlyURL() string {
	return CDNImageWidth(constants.PathSVGLogoOnly, constants.WidthLogoOnly)
}

func cutVariant(rawURL string) (string, string, bool) {
	if !strings.HasPrefix(rawURL, constants.OriginImageCDN+"/") {
		return "", "", false
	}

	idx := strings.LastIndex(rawURL, "/")
	if idx <= 0 {
		return "", "", false
	}

	return rawURL[:idx], rawURL[idx+1:], true
}
