package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposePictureResponse implements [Composer.ComposePictureResponse].
func (c *composer) ComposePictureResponse(response brevity.PictureResponse) NaturalLanguageResponse {
	info := c.ComposeCoreInformationFormat(response.Groups...)
	if response.Count == 0 {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", c.ComposeCallsigns(c.callsign), brevity.Clean),
			Speech:   fmt.Sprintf("%s, %s", c.ComposeCallsigns(c.callsign), brevity.Clean),
		}
	}

	groupCountFillIn := "single group."
	if response.Count > 1 {
		groupCountFillIn = fmt.Sprintf("%d groups.", response.Count)
	}

	info.Speech = strings.TrimSpace(info.Speech)
	info.Subtitle = strings.TrimSpace(info.Subtitle)

	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s %s", c.ComposeCallsigns(c.callsign), groupCountFillIn, info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s %s", c.ComposeCallsigns(c.callsign), groupCountFillIn, info.Speech),
	}
}
