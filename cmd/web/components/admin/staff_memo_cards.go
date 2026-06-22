package components

import "fmt"

func staffMemoLinkClickScript(memoID string) string {
	return fmt.Sprintf(`on click
  set #staff-memo-accept-%s.disabled to false
  set #staff-memo-reject-%s.disabled to false`, memoID, memoID)
}
