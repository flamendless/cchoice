<a name="unreleased"></a>
## [Unreleased]


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


[Unreleased]: https://github.com/flamendless/cchoice/compare/dev-v0.0.2...HEAD
[dev-v0.0.2]: https://github.com/flamendless/cchoice/compare/dev-v0.0.1...dev-v0.0.2
[dev-v0.0.1]: https://github.com/flamendless/cchoice/compare/v0.0.1...dev-v0.0.1
