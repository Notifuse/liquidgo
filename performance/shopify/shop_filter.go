package shopify

import (
	"fmt"
	"regexp"
	"strings"
)

// ShopFilter provides Shopify-specific template filters
type ShopFilter struct{}

// AssetUrl generates an asset URL
func (f *ShopFilter) AssetUrl(input string) string {
	return fmt.Sprintf("/files/1/[shop_id]/[shop_id]/assets/%s", input)
}

// GlobalAssetUrl generates a global asset URL
func (f *ShopFilter) GlobalAssetUrl(input string) string {
	return fmt.Sprintf("/global/%s", input)
}

// ShopifyAssetUrl generates a Shopify asset URL
func (f *ShopFilter) ShopifyAssetUrl(input string) string {
	return fmt.Sprintf("/shopify/%s", input)
}

// ScriptTag generates a script tag
func (f *ShopFilter) ScriptTag(url string) string {
	return fmt.Sprintf(`<script src="%s" type="text/javascript"></script>`, url)
}

// StylesheetTag generates a stylesheet link tag
func (f *ShopFilter) StylesheetTag(url string, media ...string) string {
	mediaAttr := "all"
	if len(media) > 0 {
		mediaAttr = media[0]
	}
	return fmt.Sprintf(`<link href="%s" rel="stylesheet" type="text/css"  media="%s"  />`, url, mediaAttr)
}

// LinkTo generates an anchor tag
func (f *ShopFilter) LinkTo(link, url string, title ...string) string {
	titleAttr := ""
	if len(title) > 0 {
		titleAttr = title[0]
	}
	return fmt.Sprintf(`<a href="%s" title="%s">%s</a>`, url, titleAttr, link)
}

// ImgTag generates an image tag
func (f *ShopFilter) ImgTag(url string, alt ...string) string {
	altAttr := ""
	if len(alt) > 0 {
		altAttr = alt[0]
	}
	return fmt.Sprintf(`<img src="%s" alt="%s" />`, url, altAttr)
}

// LinkToVendor generates a link to a vendor
func (f *ShopFilter) LinkToVendor(vendor interface{}) string {
	if vendor == nil || vendor == "" {
		return "Unknown Vendor"
	}
	v := fmt.Sprint(vendor)
	return f.LinkTo(v, f.UrlForVendor(v), v)
}

// LinkToType generates a link to a product type
func (f *ShopFilter) LinkToType(productType interface{}) string {
	if productType == nil || productType == "" {
		return "Unknown Vendor"
	}
	t := fmt.Sprint(productType)
	return f.LinkTo(t, f.UrlForType(t), t)
}

// UrlForVendor generates a URL for a vendor
func (f *ShopFilter) UrlForVendor(vendorTitle string) string {
	return fmt.Sprintf("/collections/%s", toHandle(vendorTitle))
}

// UrlForType generates a URL for a product type
func (f *ShopFilter) UrlForType(typeTitle string) string {
	return fmt.Sprintf("/collections/%s", toHandle(typeTitle))
}

// ProductImgUrl generates a product image URL with size
func (f *ShopFilter) ProductImgUrl(url string, style ...string) (string, error) {
	re := regexp.MustCompile(`\Aproducts/([\w\-\_]+)\.(\w{2,4})`)
	matches := re.FindStringSubmatch(url)
	if matches == nil {
		return "", fmt.Errorf("filter \"size\" can only be called on product images")
	}

	styleParam := "small"
	if len(style) > 0 {
		styleParam = style[0]
	}

	switch styleParam {
	case "original":
		return "/files/shops/random_number/" + url, nil
	case "grande", "large", "medium", "compact", "small", "thumb", "icon":
		return fmt.Sprintf("/files/shops/random_number/products/%s_%s.%s", matches[1], styleParam, matches[2]), nil
	default:
		return "", fmt.Errorf("valid parameters for filter \"size\" are: original, grande, large, medium, compact, small, thumb and icon")
	}
}

// DefaultPagination generates default pagination HTML
func (f *ShopFilter) DefaultPagination(paginate map[string]interface{}) string {
	var html []string

	if prev, ok := paginate["previous"].(map[string]interface{}); ok {
		title := fmt.Sprint(prev["title"])
		url := fmt.Sprint(prev["url"])
		html = append(html, fmt.Sprintf(`<span class="prev">%s</span>`, f.LinkTo(title, url)))
	}

	if parts, ok := paginate["parts"].([]interface{}); ok {
		currentPage := fmt.Sprint(paginate["current_page"])
		for _, p := range parts {
			if part, ok := p.(map[string]interface{}); ok {
				title := fmt.Sprint(part["title"])
				isLink, _ := part["is_link"].(bool)

				if isLink {
					url := fmt.Sprint(part["url"])
					html = append(html, fmt.Sprintf(`<span class="page">%s</span>`, f.LinkTo(title, url)))
				} else if title == currentPage {
					html = append(html, fmt.Sprintf(`<span class="page current">%s</span>`, title))
				} else {
					html = append(html, fmt.Sprintf(`<span class="deco">%s</span>`, title))
				}
			}
		}
	}

	if next, ok := paginate["next"].(map[string]interface{}); ok {
		title := fmt.Sprint(next["title"])
		url := fmt.Sprint(next["url"])
		html = append(html, fmt.Sprintf(`<span class="next">%s</span>`, f.LinkTo(title, url)))
	}

	return strings.Join(html, " ")
}

// Pluralize returns singular or plural word based on input
func (f *ShopFilter) Pluralize(input interface{}, singular, plural string) string {
	var num int
	switch v := input.(type) {
	case int:
		num = v
	case int64:
		num = int(v)
	case float64:
		num = int(v)
	default:
		return plural
	}

	if num == 1 {
		return singular
	}
	return plural
}

// toHandle converts a string to a URL-friendly handle
func toHandle(str string) string {
	result := strings.ToLower(str)
	result = strings.ReplaceAll(result, "'", "")
	result = strings.ReplaceAll(result, "\"", "")
	result = strings.ReplaceAll(result, "(", "")
	result = strings.ReplaceAll(result, ")", "")
	result = strings.ReplaceAll(result, "[", "")
	result = strings.ReplaceAll(result, "]", "")

	// Replace non-word characters with dashes
	re := regexp.MustCompile(`\W+`)
	result = re.ReplaceAllString(result, "-")

	// Remove leading and trailing dashes
	result = strings.Trim(result, "-")

	return result
}
