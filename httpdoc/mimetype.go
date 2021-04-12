package httpdoc

//import (
//	"github.com/frk/httptest"
//	"github.com/frk/httptest/internal/comment"
//	"github.com/frk/httptest/internal/page"
//	"github.com/frk/httptest/internal/types"
//	"github.com/frk/tagutil"
//)
//
//// isSupportedMediaType reports whether or not the given mediatype is supported.
//func isSupportedMediaType(mediatype string) bool {
//	// TODO text/csv, application/x-www-form-urlencoded, text/plain
//	return mediatype == "application/json" ||
//		mediatype == "application/xml"
//}
//
//// addMarkupToValue marshals the given value according to its type and adds
//// html markup for syntax highlighting.
//func addMarkupToValue(value interface{}, mediatype string) (template.HTML, error) {
//	if !isSupportedMediaType(mediatype) {
//		return "", fmt.Errorf("unsupported media type %q", mediatype)
//	}
//
//	switch mediatype {
//	case "application/json":
//		data, err := json.MarshalIndent(value, "", "  ")
//		if err != nil {
//			return "", err
//		}
//		data, err = highlight.JSON(data, nil) // TODO elems
//		if err != nil {
//			return "", err
//		}
//		return template.HTML(data), nil
//	case "application/xml":
//		data, err := xml.MarshalIndent(value, "", "  ")
//		if err != nil {
//			return "", err
//		}
//		data, err = highlight.XML(data, nil) // TODO elems
//		if err != nil {
//			return "", err
//		}
//		return template.HTML(data), nil
//	}
//	// TODO can't reach?
//	return "", nil
//}
