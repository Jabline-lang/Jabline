package object

import (
	"fmt"
	"image"
)

type Image struct {
	Value image.Image
}

func (img *Image) Type() ObjectType { return IMAGE_OBJ }
func (img *Image) Inspect() string {
	if img.Value == nil {
		return "image(nil)"
	}
	bounds := img.Value.Bounds()
	return fmt.Sprintf("image(%dx%d)", bounds.Dx(), bounds.Dy())
}
