package cloudlayer

// Margin represents page margins with CSS units or pixel values.
type Margin struct {
	Top    *LayoutDimension `json:"top,omitempty"`
	Right  *LayoutDimension `json:"right,omitempty"`
	Bottom *LayoutDimension `json:"bottom,omitempty"`
	Left   *LayoutDimension `json:"left,omitempty"`
}

// Viewport represents a browser viewport configuration.
// Note: the JSON field name is "viewPort" (capital P) to match the legacy API.
type Viewport struct {
	Width             int     `json:"width"`
	Height            int     `json:"height"`
	DeviceScaleFactor float64 `json:"deviceScaleFactor,omitempty"`
	IsMobile          bool    `json:"isMobile,omitempty"`
	HasTouch          bool    `json:"hasTouch,omitempty"`
	IsLandscape       bool    `json:"isLandscape,omitempty"`
}

// Cookie represents a browser cookie to set before page navigation.
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	URL      *string `json:"url,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	Path     *string `json:"path,omitempty"`
	Expires  *int    `json:"expires,omitempty"`
	HTTPOnly *bool   `json:"httpOnly,omitempty"`
	Secure   *bool   `json:"secure,omitempty"`
	SameSite *string `json:"sameSite,omitempty"`
}

// Authentication represents HTTP basic authentication credentials.
type Authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Batch represents a batch of URLs for batch conversion.
type Batch struct {
	URLs []string `json:"urls"`
}

// HeaderFooterTemplate configures a PDF header or footer.
type HeaderFooterTemplate struct {
	Method         *string           `json:"method,omitempty"`
	Selector       *string           `json:"selector,omitempty"`
	Margin         *Margin           `json:"margin,omitempty"`
	Style          map[string]string `json:"style,omitempty"`
	ImageStyle     map[string]string `json:"imageStyle,omitempty"`
	Template       *string           `json:"template,omitempty"`
	TemplateString *string           `json:"templateString,omitempty"`
}

// PreviewOptions configures preview image generation for a conversion.
type PreviewOptions struct {
	Width               *int       `json:"width,omitempty"`
	Height              *int       `json:"height,omitempty"`
	Type                *ImageType `json:"type,omitempty"`
	Quality             int        `json:"quality"`
	MaintainAspectRatio *bool      `json:"maintainAspectRatio,omitempty"`
}

// WaitForSelectorOptions configures waiting for a DOM selector.
type WaitForSelectorOptions struct {
	Selector string                     `json:"selector"`
	Options  *WaitForSelectorSubOptions `json:"options,omitempty"`
}

// WaitForSelectorSubOptions are additional options for WaitForSelector.
type WaitForSelectorSubOptions struct {
	Visible *bool `json:"visible,omitempty"`
	Hidden  *bool `json:"hidden,omitempty"`
	Timeout *int  `json:"timeout,omitempty"`
}

// PDFOptions configures PDF-specific output settings.
type PDFOptions struct {
	PrintBackground *bool                  `json:"printBackground,omitempty"`
	Format          *PDFFormat             `json:"format,omitempty"`
	Margin          *Margin                `json:"margin,omitempty"`
	HeaderTemplate  *HeaderFooterTemplate  `json:"headerTemplate,omitempty"`
	FooterTemplate  *HeaderFooterTemplate  `json:"footerTemplate,omitempty"`
	GeneratePreview *GeneratePreviewOption `json:"generatePreview,omitempty"`
}

// ImageOptions configures image-specific output settings.
type ImageOptions struct {
	ImageType       *ImageType             `json:"imageType,omitempty"`
	Quality         *int                   `json:"quality,omitempty"`
	Trim            *bool                  `json:"trim,omitempty"`
	Transparent     *bool                  `json:"transparent,omitempty"`
	GeneratePreview *GeneratePreviewOption `json:"generatePreview,omitempty"`
}

// PuppeteerOptions configures browser rendering behavior.
type PuppeteerOptions struct {
	WaitUntil              *WaitUntil              `json:"waitUntil,omitempty"`
	WaitForFrame           *bool                   `json:"waitForFrame,omitempty"`
	WaitForFrameAttachment *bool                   `json:"waitForFrameAttachment,omitempty"`
	WaitForFrameNavigation *WaitUntil              `json:"waitForFrameNavigation,omitempty"`
	WaitForFrameImages     *bool                   `json:"waitForFrameImages,omitempty"`
	WaitForFrameSelector   *WaitForSelectorOptions `json:"waitForFrameSelector,omitempty"`
	WaitForSelector        *WaitForSelectorOptions `json:"waitForSelector,omitempty"`
	PreferCSSPageSize      *bool                   `json:"preferCSSPageSize,omitempty"`
	Scale                  *float64                `json:"scale,omitempty"`
	Height                 *LayoutDimension        `json:"height,omitempty"`
	Width                  *LayoutDimension        `json:"width,omitempty"`
	Landscape              *bool                   `json:"landscape,omitempty"`
	PageRanges             *string                 `json:"pageRanges,omitempty"`
	AutoScroll             *bool                   `json:"autoScroll,omitempty"`
	ViewPort               *Viewport               `json:"viewPort,omitempty"`
	TimeZone               *string                 `json:"timeZone,omitempty"`
	EmulateMediaType       *NullableString         `json:"emulateMediaType,omitempty"`
}

// URLOptions configures URL-based conversion input.
type URLOptions struct {
	URL            *string         `json:"url,omitempty"`
	Authentication *Authentication `json:"authentication,omitempty"`
	Batch          *Batch          `json:"batch,omitempty"`
	Cookies        []Cookie        `json:"cookies,omitempty"`
}

// HTMLOptions configures HTML-based conversion input.
type HTMLOptions struct {
	// HTML is the base64-encoded HTML content. Required.
	// Use [EncodeHTML] to encode raw HTML strings.
	HTML string `json:"html"`
}

// TemplateOptions configures template-based conversion input.
// Note: the "name" field is provided by [BaseOptions] which is always
// embedded alongside TemplateOptions in composite types.
type TemplateOptions struct {
	TemplateID *string                `json:"templateId,omitempty"`
	Template   *string                `json:"template,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// BaseOptions are common options shared by all conversion endpoints.
type BaseOptions struct {
	Name      *string        `json:"name,omitempty"`
	Timeout   *int           `json:"timeout,omitempty"`
	Delay     *int           `json:"delay,omitempty"`
	Filename  *string        `json:"filename,omitempty"`
	Inline    *bool          `json:"inline,omitempty"`
	Async     *bool          `json:"async,omitempty"`
	Storage   *StorageOption `json:"storage,omitempty"`
	Webhook   *string        `json:"webhook,omitempty"`
	APIVer    *string        `json:"apiVer,omitempty"`
	ProjectID *string        `json:"projectId,omitempty"`
}

// StorageRequestOptions identifies a specific storage configuration by ID.
type StorageRequestOptions struct {
	ID string `json:"id"`
}

// URLToPDFOptions are options for [Client.URLToPDF].
type URLToPDFOptions struct {
	URLOptions
	PDFOptions
	PuppeteerOptions
	BaseOptions
}

// URLToImageOptions are options for [Client.URLToImage].
type URLToImageOptions struct {
	URLOptions
	ImageOptions
	PuppeteerOptions
	BaseOptions
}

// HTMLToPDFOptions are options for [Client.HTMLToPDF].
type HTMLToPDFOptions struct {
	HTMLOptions
	PDFOptions
	PuppeteerOptions
	BaseOptions
}

// HTMLToImageOptions are options for [Client.HTMLToImage].
type HTMLToImageOptions struct {
	HTMLOptions
	ImageOptions
	PuppeteerOptions
	BaseOptions
}

// TemplateToPDFOptions are options for [Client.TemplateToPDF].
type TemplateToPDFOptions struct {
	TemplateOptions
	PDFOptions
	PuppeteerOptions
	BaseOptions
}

// TemplateToImageOptions are options for [Client.TemplateToImage].
type TemplateToImageOptions struct {
	TemplateOptions
	ImageOptions
	PuppeteerOptions
	BaseOptions
}

// DOCXToPDFOptions are options for [Client.DOCXToPDF].
type DOCXToPDFOptions struct {
	File *FileInput `json:"-"`
	BaseOptions
}

// DOCXToHTMLOptions are options for [Client.DOCXToHTML].
type DOCXToHTMLOptions struct {
	File *FileInput `json:"-"`
	BaseOptions
}

// PDFToDOCXOptions are options for [Client.PDFToDOCX].
type PDFToDOCXOptions struct {
	File *FileInput `json:"-"`
	BaseOptions
}

// MergePDFsOptions are options for [Client.MergePDFs].
// Merge uses URL-based input, not file uploads.
type MergePDFsOptions struct {
	URLOptions
	BaseOptions
}

// ListTemplatesOptions are optional query parameters for [Client.ListTemplates].
type ListTemplatesOptions struct {
	Type     *string `json:"type,omitempty"`
	Category *string `json:"category,omitempty"`
	Tags     *string `json:"tags,omitempty"`
	Expand   *bool   `json:"expand,omitempty"`
}
