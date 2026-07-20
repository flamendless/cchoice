package metrics

import (
	"strings"
	"unicode"
)

const maxEventValueLen = 128
const highCardinalityValueLen = 32

// Client event names.
const (
	EventAdminVisit   = "admin_visit"
	EventAdminExec    = "admin_exec"
	EventCustomerVisit = "customer_visit"
	EventCustomerExec  = "customer_exec"
	EventAnonVisit    = "anon_visit"
	EventAnonExec     = "anon_exec"

	EventAddToCart              = "add_to_cart"
	EventSearchDesktop          = "search_desktop"
	EventSearchMobile           = "search_mobile"
	EventMobileSearchToggle     = "mobile_search_toggle"
	EventOrderTrackSearch       = "order_track_search"
	EventProductClick           = "product_click"
	EventPromoProductClick      = "promo_product_click"
	EventCategorySidePanelClick = "category_side_panel_click"
	EventBrandsSidePanelClick   = "brands_side_panel_click"
	EventProductExternalLinkClick = "product_external_link_click"
	EventCheckedPaymentMethod   = "checked_payment_method"
)

// Visit values.
const (
	VisitHomepage       = "homepage"
	VisitProduct        = "product"
	VisitSearch         = "search"
	VisitCart           = "cart"
	VisitOrderTracker   = "order tracker"
	VisitTerms          = "terms"
	VisitPrivacy        = "privacy"
	VisitReturn         = "return"
	VisitChangelogs     = "changelogs"
	VisitTech           = "tech"
	VisitForgotPassword = "forgot password"
	VisitResetPassword  = "reset password"
	VisitPaymentSuccess = "payment success"
	VisitPaymentCancel  = "payment cancel"
	VisitCPointsHome    = "cpoints home"
	VisitCPointsRedeem  = "cpoints redeem"
	VisitPlatforms      = "platforms"
	VisitQuotation      = "quotation"
	VisitCPointsGenerate = "cpoints generate"
	VisitCPointsCode    = "cpoints code"
)

// Action values.
const (
	ExecCheckout          = "checkout"
	ExecRemoveFromCart    = "remove from cart"
	ExecLogin             = "login"
	ExecRegister          = "register"
	ExecForgotPassword    = "forgot password"
	ExecResetPassword     = "reset password"
	ExecRedeemCPoints     = "redeem cpoints"
	ExecAddQuotationLine  = "add quotation line"
	ExecRemoveQuotationLine = "remove quotation line"
	ExecSubmitQuotation   = "submit quotation"
	ExecCreateHoliday     = "create holiday"
	ExecUpdateHoliday     = "update holiday"
	ExecDeleteHoliday     = "delete holiday"
	ExecCreateBrand       = "create brand"
	ExecUpdateBrand       = "update brand"
	ExecDeleteBrand       = "delete brand"
	ExecCreateCategory    = "create category"
	ExecUpdateCategory    = "update category"
	ExecDeleteCategory    = "delete category"
	ExecCreatePromo       = "create promo"
	ExecUpdatePromo       = "update promo"
	ExecDeletePromo       = "delete promo"
	ExecUpdateOrderStatus = "update order status"
	ExecSaveMemo          = "save memo"
	ExecSendMemoEmails    = "send memo emails"
	ExecCreateTrackedLink = "create tracked link"
	ExecDeleteTrackedLink = "delete tracked link"
	ExecRunExport         = "run export"
	ExecSubmitTimeOff     = "submit time off"
	ExecDeleteProduct     = "delete product"
)

var allowedClientEvents = map[string]struct{}{
	EventAdminVisit:               {},
	EventAdminExec:                {},
	EventCustomerVisit:            {},
	EventCustomerExec:             {},
	EventAnonVisit:                {},
	EventAnonExec:                 {},
	EventAddToCart:                {},
	EventSearchDesktop:            {},
	EventSearchMobile:             {},
	EventMobileSearchToggle:       {},
	EventOrderTrackSearch:         {},
	EventProductClick:             {},
	EventPromoProductClick:        {},
	EventCategorySidePanelClick:   {},
	EventBrandsSidePanelClick:     {},
	EventProductExternalLinkClick: {},
	EventCheckedPaymentMethod:     {},
}

var highCardinalityEvents = map[string]struct{}{
	EventSearchDesktop:    {},
	EventSearchMobile:     {},
	EventOrderTrackSearch: {},
}

func IsAllowedClientEvent(event string) bool {
	_, ok := allowedClientEvents[event]
	return ok
}

func SanitizeClientEventValue(event, value string) string {
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)

	maxLen := maxEventValueLen
	if _, ok := highCardinalityEvents[event]; ok {
		maxLen = highCardinalityValueLen
	}

	if len(value) > maxLen {
		value = value[:maxLen]
	}

	return value
}
