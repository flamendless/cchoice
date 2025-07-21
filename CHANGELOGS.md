<a name="unreleased"></a>
## [Unreleased]


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


[Unreleased]: https://github.com/flamendless/cchoice/compare/dev-v0.0.3...HEAD
[dev-v0.0.3]: https://github.com/flamendless/cchoice/compare/dev-v0.0.2...dev-v0.0.3
[dev-v0.0.2]: https://github.com/flamendless/cchoice/compare/dev-v0.0.1...dev-v0.0.2
[dev-v0.0.1]: https://github.com/flamendless/cchoice/compare/v0.0.1...dev-v0.0.1
