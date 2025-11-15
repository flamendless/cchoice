<a name="unreleased"></a>
## [Unreleased]

### Bugfix
- Make others section on top

### Deps
- Update templ version
- Upgrade dependencies

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


[Unreleased]: https://github.com/flamendless/cchoice/compare/dev-v0.0.7...HEAD
[dev-v0.0.7]: https://github.com/flamendless/cchoice/compare/dev-v0.0.6...dev-v0.0.7
[dev-v0.0.6]: https://github.com/flamendless/cchoice/compare/dev-v0.0.5...dev-v0.0.6
[dev-v0.0.5]: https://github.com/flamendless/cchoice/compare/dev-v0.0.4...dev-v0.0.5
[dev-v0.0.4]: https://github.com/flamendless/cchoice/compare/dev-v0.0.3...dev-v0.0.4
[dev-v0.0.3]: https://github.com/flamendless/cchoice/compare/dev-v0.0.2...dev-v0.0.3
[dev-v0.0.2]: https://github.com/flamendless/cchoice/compare/dev-v0.0.1...dev-v0.0.2
[dev-v0.0.1]: https://github.com/flamendless/cchoice/compare/v0.0.1...dev-v0.0.1
