PaperShift Conversion Engines
PaperShift separates conversion orchestration from engine-specific execution.
LibreOffice
The LibreOffice engine supports document, spreadsheet, and presentation input formats for the conversion pairs it advertises.
Current source formats include:
```text
doc
docx
odt
rtf
txt
html
md
xls
xlsx
ods
csv
tsv
ppt
pptx
odp
```
The current implementation advertises PDF output.
ImageMagick
The ImageMagick engine handles image-family conversions.
Current image formats include:
```text
png
jpg
webp
gif
tiff
bmp
svg
avif
```
PDF
The PDF engine supports selected PDF-to-text, document, HTML, and image conversions.
Current outputs include:
```text
txt
docx
odt
html
png
jpg
webp
```
Image output currently renders the first PDF page.
Capability discovery
Supported pairs are exposed through:
```text
GET /api/v1/capabilities
```
The capability response should be treated as the source of truth for clients.
Adding support
A format is not a usable conversion capability merely because it is registered.
A complete addition requires:
format registration
correct detection
an engine that advertises the pair
successful conversion
output validation
tests
Security requirements
Engines execute external programs against untrusted files.
Implementations must:
use context-aware command execution
use isolated workspaces
avoid user-controlled executable paths
validate generated output
propagate process failures
honor job cancellation and timeouts
avoid writing outside the assigned workspace
