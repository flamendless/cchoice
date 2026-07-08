package services

import (
	"strings"
)

func pickCell(row map[string]string, col string, fallback string) string {
	return mergeImportString(row, col, fallback, ImportBlankSkip)
}

func importBlankBehavior(col string) ProductImportBlankBehavior {
	if behavior, ok := productImportColumnDefs[col]; ok {
		return behavior
	}
	return ImportBlankSkip
}

func columnInFile(headerMap map[string]int, col string) bool {
	_, ok := headerMap[col]
	return ok
}

func cellProvided(row map[string]string, col string) bool {
	return cellValue(row, col) != ""
}

func mergeImportString(
	row map[string]string,
	col string,
	existing string,
	behavior ProductImportBlankBehavior,
) string {
	switch behavior {
	case ImportReadOnly:
		return existing
	case ImportBlankApply:
		return cellValue(row, col)
	default:
		if val := cellValue(row, col); val != "" {
			return val
		}
		return existing
	}
}

func mergeImportExternalLinks(
	row map[string]string,
	headerMap map[string]int,
	existing []ExternalPlatformLinkInput,
) ([]ExternalPlatformLinkInput, error) {
	if !columnInFile(headerMap, colExternalLinkLazada) &&
		!columnInFile(headerMap, colExternalLinkTiktok) &&
		!columnInFile(headerMap, colExternalLinkShopee) {
		return existing, nil
	}

	linkMap := make(map[string]string, len(existing))
	for _, link := range existing {
		linkMap[strings.ToUpper(link.Platform)] = link.URL
	}

	for _, item := range productExternalPlatformColumns {
		if !columnInFile(headerMap, item.Column) {
			continue
		}
		url := cellValue(row, item.Column)
		if url == "" {
			continue
		}
		linkMap[item.Platform.String()] = url
	}

	links := make([]ExternalPlatformLinkInput, 0, len(linkMap))
	for platform, url := range linkMap {
		links = append(links, ExternalPlatformLinkInput{
			Platform: platform,
			URL:      url,
		})
	}

	if err := validateExternalPlatformLinks(links); err != nil {
		return nil, err
	}
	return links, nil
}
