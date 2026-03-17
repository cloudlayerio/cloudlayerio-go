package cloudlayer

// APIVersion represents the CloudLayer.io API version.
// The API version must be explicitly specified when creating a client.
type APIVersion string

const (
	// V1 is API version 1. Conversion endpoints return raw binary by default.
	V1 APIVersion = "v1"
	// V2 is API version 2. Conversion endpoints return Job objects by default.
	V2 APIVersion = "v2"
)

// PDFFormat represents a PDF page size format.
type PDFFormat string

// PDF page size format constants.
const (
	FormatLetter  PDFFormat = "letter"
	FormatLegal   PDFFormat = "legal"
	FormatTabloid PDFFormat = "tabloid"
	FormatLedger  PDFFormat = "ledger"
	FormatA0      PDFFormat = "a0"
	FormatA1      PDFFormat = "a1"
	FormatA2      PDFFormat = "a2"
	FormatA3      PDFFormat = "a3"
	FormatA4      PDFFormat = "a4"
	FormatA5      PDFFormat = "a5"
	FormatA6      PDFFormat = "a6"
)

// ImageType represents an output image format.
type ImageType string

// Image format constants.
const (
	ImagePNG  ImageType = "png"
	ImageJPEG ImageType = "jpeg"
	ImageJPG  ImageType = "jpg"
	ImageWebP ImageType = "webp"
	ImageSVG  ImageType = "svg"
)

// JobStatus represents the status of a conversion job.
type JobStatus string

// Job status constants.
const (
	JobPending JobStatus = "pending"
	JobSuccess JobStatus = "success"
	JobError   JobStatus = "error"
)

// JobType represents the type of a conversion job.
type JobType string

// Job type constants.
const (
	JobHTMLPDF       JobType = "html-pdf"
	JobHTMLImage     JobType = "html-image"
	JobURLPDF        JobType = "url-pdf"
	JobURLImage      JobType = "url-image"
	JobTemplatePDF   JobType = "template-pdf"
	JobTemplateImage JobType = "template-image"
	JobDOCXPDF       JobType = "docx-pdf"
	JobDOCXHTML      JobType = "docx-html"
	JobImagePDF      JobType = "image-pdf"
	JobPDFImage      JobType = "pdf-image"
	JobPDFDOCX       JobType = "pdf-docx"
	JobPDFMerge      JobType = "merge-pdf"
)

// WaitUntil represents a page load event to wait for.
type WaitUntil string

// Page load event constants.
const (
	WaitLoad             WaitUntil = "load"
	WaitDOMContentLoaded WaitUntil = "domcontentloaded"
	WaitNetworkIdle0     WaitUntil = "networkidle0"
	WaitNetworkIdle2     WaitUntil = "networkidle2"
)
