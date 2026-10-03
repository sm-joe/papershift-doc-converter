package formats

var registry = []Format{
	{
		ID:        "pdf",
		Name:      "PDF",
		Extension: ".pdf",
		MIMETypes: []string{"application/pdf"},
		Family:    FamilyDocument,
	},
	{
		ID:        "doc",
		Name:      "Microsoft Word",
		Extension: ".doc",
		MIMETypes: []string{"application/msword"},
		Family:    FamilyDocument,
	},
	{
		ID:        "docx",
		Name:      "Microsoft Word",
		Extension: ".docx",
		MIMETypes: []string{
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		Family: FamilyDocument,
	},
	{
		ID:        "odt",
		Name:      "OpenDocument Text",
		Extension: ".odt",
		MIMETypes: []string{"application/vnd.oasis.opendocument.text"},
		Family:    FamilyDocument,
	},
	{
		ID:        "rtf",
		Name:      "Rich Text Format",
		Extension: ".rtf",
		MIMETypes: []string{"application/rtf", "text/rtf"},
		Family:    FamilyDocument,
	},
	{
		ID:        "txt",
		Name:      "Plain Text",
		Extension: ".txt",
		MIMETypes: []string{"text/plain"},
		Family:    FamilyDocument,
	},
	{
		ID:        "html",
		Name:      "HTML",
		Extension: ".html",
		MIMETypes: []string{"text/html"},
		Family:    FamilyDocument,
	},
	{
		ID:        "md",
		Name:      "Markdown",
		Extension: ".md",
		MIMETypes: []string{"text/markdown", "text/plain"},
		Family:    FamilyDocument,
	},

	{
		ID:        "xls",
		Name:      "Microsoft Excel",
		Extension: ".xls",
		MIMETypes: []string{"application/vnd.ms-excel"},
		Family:    FamilySpreadsheet,
	},
	{
		ID:        "xlsx",
		Name:      "Microsoft Excel",
		Extension: ".xlsx",
		MIMETypes: []string{
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		Family: FamilySpreadsheet,
	},
	{
		ID:        "ods",
		Name:      "OpenDocument Spreadsheet",
		Extension: ".ods",
		MIMETypes: []string{"application/vnd.oasis.opendocument.spreadsheet"},
		Family:    FamilySpreadsheet,
	},
	{
		ID:        "csv",
		Name:      "CSV",
		Extension: ".csv",
		MIMETypes: []string{"text/csv"},
		Family:    FamilySpreadsheet,
	},
	{
		ID:        "tsv",
		Name:      "TSV",
		Extension: ".tsv",
		MIMETypes: []string{"text/tab-separated-values", "text/plain"},
		Family:    FamilySpreadsheet,
	},

	{
		ID:        "ppt",
		Name:      "Microsoft PowerPoint",
		Extension: ".ppt",
		MIMETypes: []string{"application/vnd.ms-powerpoint"},
		Family:    FamilyPresentation,
	},
	{
		ID:        "pptx",
		Name:      "Microsoft PowerPoint",
		Extension: ".pptx",
		MIMETypes: []string{
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		},
		Family: FamilyPresentation,
	},
	{
		ID:        "odp",
		Name:      "OpenDocument Presentation",
		Extension: ".odp",
		MIMETypes: []string{"application/vnd.oasis.opendocument.presentation"},
		Family:    FamilyPresentation,
	},

	{
		ID:        "png",
		Name:      "PNG",
		Extension: ".png",
		MIMETypes: []string{"image/png"},
		Family:    FamilyImage,
	},
	{
		ID:        "jpg",
		Name:      "JPEG",
		Extension: ".jpg",
		MIMETypes: []string{"image/jpeg"},
		Family:    FamilyImage,
	},
	{
		ID:        "webp",
		Name:      "WebP",
		Extension: ".webp",
		MIMETypes: []string{"image/webp"},
		Family:    FamilyImage,
	},
	{
		ID:        "gif",
		Name:      "GIF",
		Extension: ".gif",
		MIMETypes: []string{"image/gif"},
		Family:    FamilyImage,
	},
	{
		ID:        "tiff",
		Name:      "TIFF",
		Extension: ".tiff",
		MIMETypes: []string{"image/tiff"},
		Family:    FamilyImage,
	},
	{
		ID:        "bmp",
		Name:      "BMP",
		Extension: ".bmp",
		MIMETypes: []string{"image/bmp"},
		Family:    FamilyImage,
	},
	{
		ID:        "svg",
		Name:      "SVG",
		Extension: ".svg",
		MIMETypes: []string{"image/svg+xml"},
		Family:    FamilyImage,
	},
	{
		ID:        "avif",
		Name:      "AVIF",
		Extension: ".avif",
		MIMETypes: []string{"image/avif"},
		Family:    FamilyImage,
	},
}

func All() []Format {
	result := make([]Format, len(registry))
	copy(result, registry)

	return result
}

func Get(id string) (Format, bool) {
	for _, format := range registry {
		if format.ID == id {
			return format, true
		}
	}

	return Format{}, false
}
