package services

import (
	"testing"

	"cchoice/internal/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeImportExternalLinks_PreservesExistingWhenColumnsMissing(t *testing.T) {
	t.Parallel()

	existing := []ExternalPlatformLinkInput{
		{Platform: enums.EXTERNAL_PLATFORM_LAZADA.String(), URL: "https://lazada.test"},
	}
	links, err := mergeImportExternalLinks(map[string]string{}, map[string]int{}, existing)
	require.NoError(t, err)
	assert.Equal(t, existing, links)
}

func TestMergeImportExternalLinks_UpdatesOnlyProvidedPlatforms(t *testing.T) {
	t.Parallel()

	headerMap := map[string]int{
		colExternalLinkLazada: 0,
	}
	row := map[string]string{
		colExternalLinkLazada: "https://lazada.test/new",
	}

	links, err := mergeImportExternalLinks(row, headerMap, []ExternalPlatformLinkInput{
		{Platform: enums.EXTERNAL_PLATFORM_SHOPEE.String(), URL: "https://shopee.test/old"},
	})
	require.NoError(t, err)
	require.Len(t, links, 2)

	linkByPlatform := make(map[string]string, len(links))
	for _, link := range links {
		linkByPlatform[link.Platform] = link.URL
	}
	assert.Equal(t, "https://lazada.test/new", linkByPlatform[enums.EXTERNAL_PLATFORM_LAZADA.String()])
	assert.Equal(t, "https://shopee.test/old", linkByPlatform[enums.EXTERNAL_PLATFORM_SHOPEE.String()])
}

func TestMergeImportExternalLinks_SkipsBlankCellsInProvidedColumns(t *testing.T) {
	t.Parallel()

	headerMap := map[string]int{
		colExternalLinkLazada: 0,
		colExternalLinkTiktok: 1,
		colExternalLinkShopee: 2,
	}
	row := map[string]string{
		colExternalLinkLazada: "https://lazada.test/new",
		colExternalLinkTiktok: "",
		colExternalLinkShopee: "",
	}

	links, err := mergeImportExternalLinks(row, headerMap, []ExternalPlatformLinkInput{
		{Platform: enums.EXTERNAL_PLATFORM_SHOPEE.String(), URL: "https://shopee.test/old"},
	})
	require.NoError(t, err)
	require.Len(t, links, 2)

	linkByPlatform := make(map[string]string, len(links))
	for _, link := range links {
		linkByPlatform[link.Platform] = link.URL
	}
	assert.Equal(t, "https://lazada.test/new", linkByPlatform[enums.EXTERNAL_PLATFORM_LAZADA.String()])
	assert.Equal(t, "https://shopee.test/old", linkByPlatform[enums.EXTERNAL_PLATFORM_SHOPEE.String()])
}

func TestMergeImportString_SkipsBlankValues(t *testing.T) {
	t.Parallel()

	row := map[string]string{colDescription: ""}
	got := mergeImportString(row, colDescription, "keep me", ImportBlankSkip)
	assert.Equal(t, "keep me", got)
}

func TestMergeImportString_ReadOnlyIgnoresRowValue(t *testing.T) {
	t.Parallel()

	row := map[string]string{colImageURL: "https://cdn.test/new.jpg"}
	got := mergeImportString(row, colImageURL, "existing/path.jpg", ImportReadOnly)
	assert.Equal(t, "existing/path.jpg", got)
}

func TestImportBlankBehavior_ImageColumnsAreReadOnly(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ImportReadOnly, importBlankBehavior(colImageURL))
	assert.Equal(t, ImportReadOnly, importBlankBehavior(colThumbnailURL))
}
