<a name="unreleased"></a>
## [Unreleased]


<a name="release-v1.1.21"></a>
## [release-v1.1.21] - 2026-04-27
### Bugfix
- Add CC for admin password reset and add updated_at
- Add updated_at to tbl_staff_roles
- Fix pages not rendering and add dedicated not found page
- Fix quantity
- Cart quantity
- Handle case if product has no image yet during update
- Remove product status filter for slug
- Add meta dates to product categories
- Product price saving
- Promo edit image upload and status
- Validate promo start/end dates
- Promo edit and delete
- Add missing patch

### Docs
- Add LICENSE
- Release v1.1.20

### Feature
- WIP Product Inventory
- Tracked links
- Product page
- Show promo banners in shop page
- WIP promos

### Maintenance
- Update tracked link name
- Add services in envvars
- Create image upload and preview component
- Little fixes
- use encoded brand id

### Server
- Implement homepage filters
- Handle redirection
- Fix product images
- Fix duplicating products due to product images
- Update CDN keys
- Implement product slugs
- Implement product edit
- Update image max file size upload to 3MB
- More fixes for promos
- Fix some bugs and update password length
- Allow unchanged image in brand update
- Brand management system image upload

### Web
- Test adsense
- Add platforms page
- Implement brand filtering in homepage
- Add manage product inventories
- Hardcode bosch in brand side panel
- Add dev ribbon
- Fix cart number
- Fix add to cart
- Improve product page SEO and design
- Update product listing columns and fix shop product image size
- Add dates validation
- Fix confirm password show/hide
- Refactor admin cards
- Embed youtube video in shop page
- Remove extra confirm password


<a name="release-v1.1.20"></a>
## [release-v1.1.20] - 2026-04-15
### AI
- Update

### Feature
- WIP Brand management

### Maintenance
- Apply code reviews
- Refactor contains check in scan_receipt
- Address static analysis reports
- Update enum usage

### Script
- Fix regex
- Create scf
- Update betteralign excludes
- Handle deleted/renamed file edge case in has trailing whitespace

### Server
- Add shopee pay
- Add LocalForgotPassword testing env var

### Web
- Organize superuser cards
- Simplify superuser home
- Display safe env vars in superuser portal


<a name="dev-v0.1.4"></a>
## [dev-v0.1.4] - 2026-04-13
### AI
- Update
- Explicitly state mage commands

### Feature
- Forgot password

### Maintenance
- Separate structs to files
- Static analysis
- Add more staff logs and modularize it via constants

### Server
- Implement product deletion

### Web
- Show logged in customer in shop page
- Add back to shop
- Customers table


<a name="release-v1.1.19"></a>
## [release-v1.1.19] - 2026-04-10
### AI
- Add admin services section
- Add LLM rule

### CICD
- Fix apt
- Add libvips and pkg-config
- Fix
- Add pkg-config
- Add libvips

### Feature
- Holidays

### Maintenance
- Fix formatting
- Update errs
- Organize enum files

### Script
- Update hasExtChanges
- Simplify hooks
- Add commit-msg hook that checks commit prefix

### Server
- Invoke StaffLogsService and update http errors and redirects

### Web
- Edit holidays


<a name="release-v1.1.18"></a>
## [release-v1.1.18] - 2026-04-06
### Bugfix
- Image size
- Remove header content type for changelog handler

### Docs
- Add cpoints diagram
- Release v1.1.17

### Feature
- Customer verification via OTP in email
- QR code
- C-Points generation
- WIP C-Points
- WIP customer login
- Thumbnail service

### Maintenance
- Use constants for mobile number prefix and add validations
- Add IService
- Add audit logging for report exports
- Update db interface naming
- Add OutputFormat enum
- Optimize map caching by pre-allocating province slice capacity
- Inline err assignments into if expressions
- Move AttendanceService and ReportService to shared Server struct
- Use enum for AppEnv instead of strings

### Performance
- Store CDN URL in product images

### Script
- Add servecustomer in magefile
- Exclude png file in trailing whitespace check

### Server
- Add customer status
- Implement rate limit for customer endpoints
- Implement HMAC based cpoints

### Web
- Block c-points for unverified users
- Update mobile prefix
- Logged in user should not be able to visit login page
- Style customer profile
- Improve and unify styles across admin and customer pages
- Add customer cpoints and redeem pages
- Add confirm password and fix bugs
- Display uploaded image preview
- Add product specs to admin products listing


<a name="release-v1.1.17"></a>
## [release-v1.1.17] - 2026-03-19
### Docs
- Release v1.1.16

### Feature
- Staff management
- Role-based endpoints
- Admin staff logs
- Products table
- Product upload
- WIP product upload

### Maintenance
- More migration to services
- More migration to product service
- Create location service
- Always pass encoded ID to services
- More services
- Create product images and brands services
- Update templ file
- Move PHP to constants
- Unify regexes

### Script
- Add trailing whitespace check

### Server
- Move create product to service

### Web
- Add more data in reports
- Update styles


<a name="release-v1.1.16"></a>
## [release-v1.1.16] - 2026-03-14
### Bugfix
- Remove upsert location

### Docs
- Update guide and fix typo
- Release v1.1.15

### Script
- Optimize gen* by checking for changes
- Fixed go build flag ordering and tag formatting

### Server
- Report - add total days and late data

### Web
- Update table when a filter is changed


<a name="release-v1.1.15"></a>
## [release-v1.1.15] - 2026-03-12
### Docs
- Release v1.1.14

### Feature
- XLSX report

### Web
- Add staff filter


<a name="release-v1.1.14"></a>
## [release-v1.1.14] - 2026-03-11
### Bugfix
- Path in templ generate not working

### Maintenance
- Update allowed origins

### Web
- Update metrics to trigger once
- Unify success and error messages to banners
- Update export csv to stream
- Rename metrics events for admin
- Update metrics events for admin
- Update date selectors


<a name="release-v1.1.13"></a>
## [release-v1.1.13] - 2026-03-10
### Bugfix
- Time out location now stored

### Feature
- Admin profile page

### Script
- Optimize start time by specifying templ path

### Web
- Display staff attendance record in inverted table
- Separate profile header
- Rename lunch break in/out -> start/end


<a name="release-v1.1.12"></a>
## [release-v1.1.12] - 2026-03-08
### Docs
- Update issue templates

### Feature
- WIP export attendance
- Implement lunch break tracking
- WIP time off
- Display location separately as a component
- Add location and useragent
- WIP products admin page

### Maintenance
- Organize utils
- Cleanup and simplifications
- Update goqite table

### Script
- Update magefile and air configs

### Server
- Be consistent with time storing
- Separate time in/out status
- Update haversine distance
- Separate useragent for time in and time out
- Update shop radius
- Use PH time
- Fix time in/out bugs
- Add required in shop
- Record staff access

### Web
- Add superuser time off
- Add staff time off
- Add staff portal
- Display staff id as encoded
- Display haversine distance
- Update location display
- Display location
- Add spinning animation to buttons
- Do not require location service for superusers
- Add indicator for refresh


<a name="release-v1.1.11"></a>
## [release-v1.1.11] - 2026-02-24
### Feature
- Location based time in/out enablement
- Admin page for time tracking


<a name="release-v1.1.10"></a>
## [release-v1.1.10] - 2026-02-24
### CICD
- More PR Check fixes
- Update PR Check
- Add PR Check

### Deps
- Downgrade broken templ
- Update go and libs

### Docs
- Update PR template checklist
- Add PR template
- Update README.md

### Maintenance
- Fix breaking changes
- Code reviews
- Cleanup CHANGELOGS

### Performance
- Improve prefetching
- Implement prefetching

### Script
- Verbose magefile

### Web
- Add T&C and PP page
- Prefetch product thumbnail on mouseenter


<a name="release-v1.1.9"></a>
## [release-v1.1.9] - 2026-02-17
### Bugfix
- Add to cart count

### Web
- Animate cart on count update
- Add AddToCart in promo product


<a name="release-v1.1.8"></a>
## [release-v1.1.8] - 2026-02-14
### Web
- Add metrics to promo product click and checked payment method


<a name="release-v1.1.7"></a>
## [release-v1.1.7] - 2026-02-12
### Server
- Fix random promo product not really being randomized

### Web
- Create maintenance page for 404


<a name="release-v1.1.6"></a>
## [release-v1.1.6] - 2026-02-12
### Server
- Use product name instead of time for promo metrics


<a name="release-v1.1.5"></a>
## [release-v1.1.5] - 2026-02-12
### Script
- Add checkMigrations

### Server
- Add missing Inc call in metrics


<a name="release-v1.1.4"></a>
## [release-v1.1.4] - 2026-02-12
### Server
- Add random promo product metrics
- Add random promo product

### Web
- Update promo position
- Update random promo product banner pos
- Display random promo product banner
- Decrease CORSeal logo


<a name="release-v1.1.3"></a>
## [release-v1.1.3] - 2026-02-05
### Web
- Add COR seal


<a name="release-v1.1.2"></a>
## [release-v1.1.2] - 2026-02-02
### Docs
- Add list.md for tracking script execution

### Server
- Update sale
- Implement product name sanitization


<a name="release-v1.1.1"></a>
## [release-v1.1.1] - 2026-01-26
### Server
- Update checkout to use discounts


<a name="release-v1.1.0"></a>
## [release-v1.1.0] - 2026-01-26
### Server
- Update sale discount style
- Implement CSV input for creating product sales
- Address more lighthouse reports


<a name="release-v1.0.11"></a>
## [release-v1.0.11] - 2026-01-15
### Server
- Set cache age for js files
- Address lighthouse reports
- Add robots.txt
- WIP product sales


<a name="release-v1.0.10"></a>
## [release-v1.0.10] - 2025-12-31
### Server
- Implement basic event metrics


<a name="release-v1.0.9"></a>
## [release-v1.0.9] - 2025-12-31
### CICD
- Change email step for dev

### Server
- Implement http requests prometheus metrics
- Use basic auth for metrics endpoint
- Address preload warnings


<a name="dev-v0.1.3"></a>
## [dev-v0.1.3] - 2025-12-29
### CICD
- Update tag


<a name="dev-v0.1.2"></a>
## [dev-v0.1.2] - 2025-12-29
### CICD
- Change email step for dev

<a name="release-v1.0.8"></a>
## [release-v1.0.8] - 2025-12-29
### CICD
- Add sender email
- Remove CC
- Test release notes email

<a name="release-v1.0.7"></a>
## [release-v1.0.7] - 2025-12-29
### Web
- On search click, display the modal


<a name="release-v1.0.6"></a>
## [release-v1.0.6] - 2025-12-29
### Script
- Add svgs in migrate images to cloudflare script

### Server
- Add delivery ETA in tracker page

### Web
- Improve header and searchbar (desktop and mobile)


<a name="release-v1.0.5"></a>
## [release-v1.0.5] - 2025-12-25
### Script
- Update dev

### Server
- Add waze

### Web
- Add waze
- Add gmaps in store section
- Add searchbox for order tracking


<a name="dev-v0.1.1"></a>
## [dev-v0.1.1] - 2025-12-25
### Bugfix
- Use git rev-list to correctly get latest tag

### Script
- Fix serveweb

### Server
- Orders routes
- Refactor how settings are used
- Implement changelogs parsing
- Allow cloudflare in CSP

### Web
- Order tracker page
- Add tatara and powercraft + coming soon ribbon in brand logos
- Add mobile number and email in order confirmation template
- Improve changelogs view
- More responsive fixes and more brand logos


<a name="release-v1.0.4"></a>
## [release-v1.0.4] - 2025-12-18
### Bugfix
- Use git rev-list to correctly get latest tag

### Server
- Implement changelogs parsing
- Allow cloudflare in CSP

### Web
- Improve changelogs view


<a name="release-v1.0.3"></a>
## [release-v1.0.3] - 2025-12-17
### Server
- Allow cloudflare in CSP


<a name="release-v1.0.2"></a>
## [release-v1.0.2] - 2025-12-17
### Web
- More responsive fixes and more brand logs


<a name="release-v1.0.1"></a>
## [release-v1.0.1] - 2025-12-16
### Server
- Fix searchbar
- WIP fixing searchbar
- Send email receipt for prod


<a name="release-v1.0.0"></a>
## [release-v1.0.0] - 2025-12-15
### Docs
- Update env
- dev-0.1.0

### Feature
- Mobile responsiveness

### Script
- Implement cmd/web for fast web changes
- Update prod
- Update dbbackup
- Allow local backup

### Server
- Update routes
- Update routes
- Fix IP binding
- WIP webhook

### Web
- Add prices in homepage and product details in modal


<a name="dev-v0.1.0"></a>
## [dev-v0.1.0] - 2025-12-13
### Deps
- Update packages

### Docs
- Update changelogs

### Feature
- Job/Queue system for email
- Order confirmation mail
- Mail
- Add delivery ETA
- Checkout -> Success flow

### Maintenance
- Cleanup
- Remove thumbnail data
- Use image extension enums and minimize variants
- Update test_linode
- Use CDN URL for email logo

### Server
- Add GrabPay
- Fix modal image viewer and use now the CDN
- Use Cloudflare Image instead of Linode Object Storage
- Add NCR in province selection
- Sort available payment methods
- Add other available payment methods
- Use business conf as pickup location
- Implement free shipping

### Web
- Update saving/restoring of shipping form


<a name="dev-v0.0.9"></a>
## [dev-v0.0.9] - 2025-11-22
### Deps
- Update

### Maintenance
- Log in test shipping
- Update enums
- Update errs
- Reduce hardcoded strings and use enums
- Add test_gvision
- Update errs
- Update image loading

### Script
- Update dbbackup.sh
- Update paths
- Copy script to usr path
- Add dbbackup
- Update genchlog
- Update scripts

### Server
- Use external api logs
- Add external api logs table
- Implement brand routes
- Update googlevision usage
- Implement receipt image -> data
- Add store

### Web
- Add fullname in checkout page
- Add brands in side panel
- Update layout of checkout page


<a name="dev-v0.0.8"></a>
## [dev-v0.0.8] - 2025-11-15
### Bugfix
- Make others section on top

### Deps
- Update templ version
- Upgrade dependencies

### Docs
- Update changelogs

### Feature
- Indicate active category section in side panel
- Scroll to category section from side panel
- Implement canceled payment page

### Maintenance
- Organize errors
- Update cmd and add testInteg

### Script
- Make parse products use public bucket
- Add migrate_images_linode
- Update test linode
- Update scripts to pull git tags

### Server
- Cache busting for 404 assets
- Update ref number generation
- Use product image for paymongo checkout page
- Use public bucket for product images and brand logos
- Implement separate buckets for public and private assets
- Use linode fs
- Use S3 URL for brand logo
- Add cmd for testing linode object storage connection
- Implement linode storage and organize env vars
- Add to session the shipping info
- Add orders in database

### Web
- Increase footer z-index
- Only include checked items in cart page for calculation
- Improve cart page styles
- Display product weight


<a name="dev-v0.0.7"></a>
## [dev-v0.0.7] - 2025-10-20
### Bugfix
- Address dropdown default getting value

### Feature
- Add version checking in browser
- Implement db caching of geocode
- Integrate shipping fee calculation in carts page
- CCS-57 CCS-56 CCS-53 Geocoding, Lalamove, C-Choice services
- Implement form sanitization
- CCS-53 WIP Lalamove integration
- Add prometheus and grafana

### Maintenance
- Refactor and add better errors
- Cleanup and improvements
- Cleanup
- Use std http method and any keyword

### Script
- Update bosch weight column
- CCS-54 Add golangci-lint
- CCS-54 Introduce magefile
- Update genimages

### Server
- Integrate weight
- Add weight and weight unit in product specs
- CCS-53 More shipping service work
- CCS-55 Update product images structure
- Use Prometheus for metrics
- CCS-52 Implement headers caching and metrics

### Web
- Add delivery fee loading animation when calculating


<a name="dev-v0.0.6"></a>
## [dev-v0.0.6] - 2025-08-26
### Bugfix
- CCS-51 Fix quantity increase/decrease behavior
- CCS-44 Use custom name for line items
- CCS-39 Exclude products with no valid image in home page
- Refresh on cart page should not re-create checkoutline
- Reuse created checkout

### CICD
- CCS-13 Remove push event for running lint workflow

### Deps
- Upgrade packages
- CCS-29 Upgrade deps

### Docs
- Generate changelogs
- Move ENCODE_SALT to required section

### Featue
- CCS-43 Display all payment methods

### Feature
- Complete cart page -> paymongo page
- CCS-46 Add other input in shipping information
- CCS-46 Display shipping address
- CCS-49 Add other payment methods images
- CCS-43 Display available payment methods
- CCS-38 Implement cart page summary
- CCS-35 Add minus/plus buttons for quantity in cart line
- CCS-26 Add product image, price, and total in cart page

### Maintenance
- Apply sc
- Simplify frontend code for shipping
- Rename file
- Fix build tags
- Rename parse_xlsx -> parse_products
- CCS-28 Add IsLocal and IsProd
- CCS-28 Centralize conf
- Patch recursive thumbnailify

### Performance
- CCS-50 Cache and singleflight address
- CCS-37 Address some issues reported by Lighthouse
- CCS-30 Utilize build tags to omit libvips

### Script
- CCS-47 WIP parse_map
- Add migration in deploy
- Default tailwind bin
- Update deps

### Server
- Replace fatal -> error in handlers
- Header cart count should reflect unique products, not total quantity
- CCS-45 Add cash on delivery in tbl_settings
- Use SQL functions to simplify quantity changes
- CCS-32 Implement checkoutline deletion
- CCS-33 Improve logs by using log tag
- CCS-31 Implement embedded, static, and stub fs mode
- CCS-31 Stub static files
- CCS-31 Move static to a package

### Web
- Update position and style of error banner
- Make the cart page a form
- CCS-41 Make header responsive
- CCS-38 Make cart page summary sticky
- CCS-36 Dynamic product image modal height
- CCS-34 Add checkbox in cart items
- CCS-32 Add trash button in cart line
- Improve empty cart page; Add cart summary bar


<a name="dev-v0.0.5"></a>
## [dev-v0.0.5] - 2025-07-22
### CICD
- Update dev_deploy

### Docs
- dev-v0.0.4 changelogs


<a name="dev-v0.0.4"></a>
## [dev-v0.0.4] - 2025-07-22
### CICD
- CCS-13 Try if update can be omitted
- CCS-13 Use libvips-dev
- CCS-13 Install libvips in workflow
- CCS-13 Apply nolint directive
- CCS-13 Fix path
- CCS-13 Explicitly pass args
- CCS-13 Update golangci config
- CCS-13 Update golangci-lint workflow
- CCS-13 Update golangci version to 8
- CCS-13 Update golangci config
- CCS-13 Integrate golangci-lint-action

### Feature
- Add to cart and cart page

### Maintenance
- CCS-27 Apply lint fixes
- Separate some routes to other files

### Script
- Add golangci in deps and testall
- Add genchlog


<a name="dev-v0.0.3"></a>
## [dev-v0.0.3] - 2025-07-20
### Deps
- Remove pnpm. Use tailwind as standalone
- Update htmx -> 2.0.6
- Add goutil

### Docs
- Update script commit tag
- CCS-20 Update guide
- Add genimages in setup steps
- Add env PAYMONGO_SUCCESS_URL
- Add feature
- Add PAYMONGO_API_KEY
- Update changelogs

### Feature
- CCS-23 Add IEncode.Name
- CCS-23 Implement SQIDS encoder/decoder
- Get available payment methods

### Maintenance
- Add ERR_ENV_VAR_REQUIRED
- Remove unnecessary parallel in unit tests
- Move enums to their domain packages
- Use constants for some repeating strings

### Script
- CCS-25 Add cmd for encoding/decoding
- CCS-20 Add MacOS support
- CCS-12 Add deps_debian
- Add db; Update logging; Update testall
- Move prealloc
- Reduce verbosity in sc()

### Server
- Add route for getting cart count
- WIP session manager for storing checkout lines
- Use HX-Redirect instead of http.Redirect to solve CORS
- Implement WIP checkout handler
- Add checkouts migration
- WIP PayMongo integration

### Web
- Add count in header cart
- Submit product id in modal add to cart
- Update site title to display app env
- Add add to cart button in product modal
- Update active search logic
- Add history:restore in home page elements
- Move Base to HomePage in preparation for ProductPage
- Group hover cart
- Add "X" clear button in search bar when there is a search query
- Add search more results component and add color transitions


<a name="dev-v0.0.2"></a>
## [dev-v0.0.2] - 2025-06-30
### CICD
- Separate deploy and send-email jobs


<a name="dev-v0.0.1"></a>
## [dev-v0.0.1] - 2025-06-30
### CICD
- Include sending of email
- Only deploy on matching tags
- Update workflow
- Add github workflow for deployment

### Deps
- Replace sql-migrate with goose
- Update htmx and others

### Docs
- Add NOTES
- Update commit labels
- Create Deps tag
- Add license for brand logos
- Add commit topics in README

### Maintenance
- Apply static analysis
- Fix git submodule
- Fix benchmark functions
- Add errs for commands
- Update changelogs

### Script
- Add deploy.sh
- Dynamic pnpx path
- Separate generation of images from cleandb
- Build goose
- Add libvips module in deps
- Rename testall -> testsum; Add simple testall
- Add prof
- Implement cmd/convert_images
- Rename process_images -> thumbnailify_images

### Server
- Update tbl_products_fts
- WIP FTS
- Implement prefix in base64 encoding

### Web
- Fix icon spinner in searchbar
- Improve searchbar and search results behavior
- Show search results
- Add footer wrap
- WIP active search
- Improve post home content styling
- Fix styling in post home content sections
- Rename other to post home content
- Add "About Us" and "Partners" sections


<a name="v0.0.1"></a>
## v0.0.1 - 2025-06-12
### Config
- update chglog

### Deps
- Update tailwind and templ

### Maintenance
- Update go version and apply shellcheck
- Fix warnings from errcheck
- Update deps
- Some changes regarding HTTP2
- Add modernize in sc
- QoL improvements
- More simplifications
- Fix margin of subcategory row
- Fix thumbnail logic with trimming
- update prod function
- Minor revision
- Apply modernize
- Update go version and fix function usage

### Performance
- Update images packages implementation
- Reimplement rendering of images
- Optimization + improvements of category and subcategories
- Optimize placeholder image
- optimize thumbnails by specifying size
- optimize by using cache control
- optimize the following:
- Implement caching for thumbnailify handler
- Implement http2 protocols
- Update benchmark loops

### Script
- Fix grep to use case-insensitive flag
- Add xtrace in bash commands
- Update prod script

### Server
- add changelogs route
- Fix SlugToTile as the caser should not be shared between goroutines
- Implement pagination for category sections
- use chi's middleware.Compress

### Tool
- add git-chglog

### Toolings
- Add usestdlibvars and fatcontext and update packages
- Update sql-migrate tool
- Move tools to tool directive

### Web
- Update logo to include full text
- Working services section
- WIP services section
- Add services to footer
- Finally
- Update meta tags
- Move modal close button
- Update scrollbar
- Fix image showing behavior
- Implement image viewer modal
- WIP image viewer modal
- Eager load elements instead of on intersect.
- Disable horizontal scrolling
- Remove px in logo sizes
- Improve accessibility by adding alt tags


[Unreleased]: https://github.com/flamendless/cchoice/compare/release-v1.1.14...HEAD
[release-v1.1.14]: https://github.com/flamendless/cchoice/compare/release-v1.1.13...release-v1.1.14
[release-v1.1.13]: https://github.com/flamendless/cchoice/compare/release-v1.1.12...release-v1.1.13
[release-v1.1.12]: https://github.com/flamendless/cchoice/compare/release-v1.1.11...release-v1.1.12
[release-v1.1.11]: https://github.com/flamendless/cchoice/compare/release-v1.1.10...release-v1.1.11
[release-v1.1.10]: https://github.com/flamendless/cchoice/compare/release-v1.1.9...release-v1.1.10
[release-v1.1.9]: https://github.com/flamendless/cchoice/compare/release-v1.1.8...release-v1.1.9
[release-v1.1.8]: https://github.com/flamendless/cchoice/compare/release-v1.1.7...release-v1.1.8
[release-v1.1.7]: https://github.com/flamendless/cchoice/compare/release-v1.1.6...release-v1.1.7
[release-v1.1.6]: https://github.com/flamendless/cchoice/compare/release-v1.1.5...release-v1.1.6
[release-v1.1.5]: https://github.com/flamendless/cchoice/compare/release-v1.1.4...release-v1.1.5
[release-v1.1.4]: https://github.com/flamendless/cchoice/compare/release-v1.1.3...release-v1.1.4
[release-v1.1.3]: https://github.com/flamendless/cchoice/compare/release-v1.1.2...release-v1.1.3
[release-v1.1.2]: https://github.com/flamendless/cchoice/compare/release-v1.1.1...release-v1.1.2
[release-v1.1.1]: https://github.com/flamendless/cchoice/compare/release-v1.1.0...release-v1.1.1
[release-v1.1.0]: https://github.com/flamendless/cchoice/compare/release-v1.0.11...release-v1.1.0
[release-v1.0.11]: https://github.com/flamendless/cchoice/compare/release-v1.0.10...release-v1.0.11
[release-v1.0.10]: https://github.com/flamendless/cchoice/compare/release-v1.0.9...release-v1.0.10
[release-v1.0.9]: https://github.com/flamendless/cchoice/compare/dev-v0.1.3...release-v1.0.9
[dev-v0.1.3]: https://github.com/flamendless/cchoice/compare/dev-v0.1.2...dev-v0.1.3
[dev-v0.1.2]: https://github.com/flamendless/cchoice/compare/release-v1.0.8...dev-v0.1.2
[release-v1.0.8]: https://github.com/flamendless/cchoice/compare/release-v1.0.7...release-v1.0.8
[release-v1.0.7]: https://github.com/flamendless/cchoice/compare/release-v1.0.6...release-v1.0.7
[release-v1.0.6]: https://github.com/flamendless/cchoice/compare/release-v1.0.5...release-v1.0.6
[release-v1.0.5]: https://github.com/flamendless/cchoice/compare/dev-v0.1.1...release-v1.0.5
[dev-v0.1.1]: https://github.com/flamendless/cchoice/compare/release-v1.0.4...dev-v0.1.1
[release-v1.0.4]: https://github.com/flamendless/cchoice/compare/release-v1.0.3...release-v1.0.4
[release-v1.0.3]: https://github.com/flamendless/cchoice/compare/release-v1.0.2...release-v1.0.3
[release-v1.0.2]: https://github.com/flamendless/cchoice/compare/release-v1.0.1...release-v1.0.2
[release-v1.0.1]: https://github.com/flamendless/cchoice/compare/release-v1.0.0...release-v1.0.1
[release-v1.0.0]: https://github.com/flamendless/cchoice/compare/dev-v0.1.0...release-v1.0.0
[dev-v0.1.0]: https://github.com/flamendless/cchoice/compare/dev-v0.0.9...dev-v0.1.0
[dev-v0.0.9]: https://github.com/flamendless/cchoice/compare/dev-v0.0.8...dev-v0.0.9
[dev-v0.0.8]: https://github.com/flamendless/cchoice/compare/dev-v0.0.7...dev-v0.0.8
[dev-v0.0.7]: https://github.com/flamendless/cchoice/compare/dev-v0.0.6...dev-v0.0.7
[dev-v0.0.6]: https://github.com/flamendless/cchoice/compare/dev-v0.0.5...dev-v0.0.6
[dev-v0.0.5]: https://github.com/flamendless/cchoice/compare/dev-v0.0.4...dev-v0.0.5
[dev-v0.0.4]: https://github.com/flamendless/cchoice/compare/dev-v0.0.3...dev-v0.0.4
[dev-v0.0.3]: https://github.com/flamendless/cchoice/compare/dev-v0.0.2...dev-v0.0.3
[dev-v0.0.2]: https://github.com/flamendless/cchoice/compare/dev-v0.0.1...dev-v0.0.2
[dev-v0.0.1]: https://github.com/flamendless/cchoice/compare/v0.0.1...dev-v0.0.1
