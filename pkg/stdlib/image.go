package stdlib

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_image"] = ImageBuiltins
	NativeModulePrefixes["_image"] = "img_"
}

var ImageBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"img_new", &object.Builtin{Fn: imgNew}},
	{"img_load", &object.Builtin{Fn: imgLoad}},
	{"img_save", &object.Builtin{Fn: imgSave}},
	{"img_width", &object.Builtin{Fn: imgWidth}},
	{"img_height", &object.Builtin{Fn: imgHeight}},
	{"img_get_pixel", &object.Builtin{Fn: imgGetPixel}},
	{"img_set_pixel", &object.Builtin{Fn: imgSetPixel}},
	{"img_resize", &object.Builtin{Fn: imgResize}},
	{"img_crop", &object.Builtin{Fn: imgCrop}},
	{"img_grayscale", &object.Builtin{Fn: imgGrayscale}},
	{"img_encode", &object.Builtin{Fn: imgEncode}},
	{"img_blur", &object.Builtin{Fn: imgBlur}},
	{"img_draw_text", &object.Builtin{Fn: imgDrawText}},
	{"img_rotate", &object.Builtin{Fn: imgRotate}},
}

var supportedFormats = map[string]string{
	".png":  "png",
	".jpg":  "jpeg",
	".jpeg": "jpeg",
	".gif":  "gif",
}

func imgNew(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("img_new expects 2 arguments (width, height), got %d", len(args))
	}

	w, ok := args[0].(*object.Integer)
	if !ok {
		return newError("img_new expects integer width, got %s", args[0].Type())
	}
	h, ok := args[1].(*object.Integer)
	if !ok {
		return newError("img_new expects integer height, got %s", args[1].Type())
	}

	img := image.NewRGBA(image.Rect(0, 0, int(w.Value), int(h.Value)))
	return &object.Image{Value: img}
}

func imgLoad(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("img_load expects 1 argument (path), got %d", len(args))
	}

	path, ok := args[0].(*object.String)
	if !ok {
		return newError("img_load expects string path, got %s", args[0].Type())
	}

	f, err := os.Open(path.Value)
	if err != nil {
		return newError("img_load: %s", err.Error())
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path.Value))
	var img image.Image

	switch ext {
	case ".png":
		img, err = png.Decode(f)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".gif":
		img, err = gif.Decode(f)
	default:
		// Try to decode based on content
		img, _, err = image.Decode(f)
	}
	if err != nil {
		return newError("img_load: %s", err.Error())
	}

	return &object.Image{Value: img}
}

func imgSave(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("img_save expects 2 arguments (image, path), got %d", len(args))
	}

	imgObj, ok := args[0].(*object.Image)
	if !ok {
		return newError("img_save expects image, got %s", args[0].Type())
	}

	path, ok := args[1].(*object.String)
	if !ok {
		return newError("img_save expects string path, got %s", args[1].Type())
	}

	ext := strings.ToLower(filepath.Ext(path.Value))
	var buf bytes.Buffer
	var err error

	switch ext {
	case ".png":
		err = png.Encode(&buf, imgObj.Value)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, imgObj.Value, &jpeg.Options{Quality: 90})
	case ".gif":
		err = gif.Encode(&buf, imgObj.Value, nil)
	default:
		return newError("img_save: unsupported format %s (supported: .png, .jpg, .gif)", ext)
	}
	if err != nil {
		return newError("img_save: %s", err.Error())
	}

	if err := os.WriteFile(path.Value, buf.Bytes(), 0644); err != nil {
		return newError("img_save: %s", err.Error())
	}

	return object.NullObj
}

func imgWidth(args ...object.Object) object.Object {
	img := getImage(args)
	if img == nil {
		return newError("img_width expects an image")
	}
	return &object.Integer{Value: int64(img.Bounds().Dx())}
}

func imgHeight(args ...object.Object) object.Object {
	img := getImage(args)
	if img == nil {
		return newError("img_height expects an image")
	}
	return &object.Integer{Value: int64(img.Bounds().Dy())}
}

func imgGetPixel(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("img_get_pixel expects 3 arguments (image, x, y), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_get_pixel expects an image, got %s", args[0].Type())
	}

	x, ok := args[1].(*object.Integer)
	if !ok {
		return newError("img_get_pixel expects integer x, got %s", args[1].Type())
	}
	y, ok := args[2].(*object.Integer)
	if !ok {
		return newError("img_get_pixel expects integer y, got %s", args[2].Type())
	}

	r, g, b, a := img.At(int(x.Value), int(y.Value)).RGBA()
	hash := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
	setHashField(hash, "r", &object.Integer{Value: int64(r >> 8)})
	setHashField(hash, "g", &object.Integer{Value: int64(g >> 8)})
	setHashField(hash, "b", &object.Integer{Value: int64(b >> 8)})
	setHashField(hash, "a", &object.Integer{Value: int64(a >> 8)})
	return hash
}

func imgSetPixel(args ...object.Object) object.Object {
	if len(args) < 5 || len(args) > 6 {
		return newError("img_set_pixel expects 5-6 arguments (image, x, y, r, g, b, [a]), got %d", len(args))
	}

	rgba, ok := args[0].(*object.Image)
	if !ok {
		return newError("img_set_pixel expects image, got %s", args[0].Type())
	}
	// Convert to RGBA if not already
	if rgba.Value == nil {
		return newError("img_set_pixel: nil image")
	}

	x := intArg(args[1])
	y := intArg(args[2])
	r := uint8(intArg(args[3]))
	g := uint8(intArg(args[4]))
	b := uint8(intArg(args[5]))
	a := uint8(255)
	if len(args) == 7 {
		a = uint8(intArg(args[6]))
	}

	// Use rgba.At() to get the underlying image, convert to RGBA if needed
	bounds := rgba.Value.Bounds()
	switch img := rgba.Value.(type) {
	case *image.RGBA:
		img.Set(int(x), int(y), color.RGBA{r, g, b, a})
	default:
		// Create a new RGBA and copy
		newImg := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				newImg.Set(x, y, img.At(x, y))
			}
		}
		newImg.Set(int(x), int(y), color.RGBA{r, g, b, a})
		rgba.Value = newImg
	}

	return object.NullObj
}

func imgResize(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("img_resize expects 3 arguments (image, width, height), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_resize expects an image, got %s", args[0].Type())
	}
	newW := intArg(args[1])
	newH := intArg(args[2])

	if newW <= 0 || newH <= 0 {
		return newError("img_resize: dimensions must be positive")
	}

	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	newImg := image.NewRGBA(image.Rect(0, 0, int(newW), int(newH)))

	for y := 0; y < int(newH); y++ {
		for x := 0; x < int(newW); x++ {
			srcX := x * srcW / int(newW)
			srcY := y * srcH / int(newH)
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	return &object.Image{Value: newImg}
}

func imgCrop(args ...object.Object) object.Object {
	if len(args) != 5 {
		return newError("img_crop expects 5 arguments (image, x, y, width, height), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_crop expects an image, got %s", args[0].Type())
	}

	x := intArg(args[1])
	y := intArg(args[2])
	w := intArg(args[3])
	h := intArg(args[4])

	bounds := img.Bounds()
	if x < 0 || y < 0 || w <= 0 || h <= 0 || int(x)+int(w) > bounds.Dx() || int(y)+int(h) > bounds.Dy() {
		return newError("img_crop: crop dimensions out of bounds")
	}

	cropped := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for dy := 0; dy < int(h); dy++ {
		for dx := 0; dx < int(w); dx++ {
			cropped.Set(dx, dy, img.At(int(x)+dx, int(y)+dy))
		}
	}

	return &object.Image{Value: cropped}
}

func imgGrayscale(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("img_grayscale expects 1 argument (image), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_grayscale expects an image, got %s", args[0].Type())
	}

	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	return &object.Image{Value: gray}
}

func imgEncode(args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("img_encode expects 2-3 arguments (image, format, [quality]), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_encode expects an image, got %s", args[0].Type())
	}

	format, ok := args[1].(*object.String)
	if !ok {
		return newError("img_encode expects string format, got %s", args[1].Type())
	}

	var buf bytes.Buffer
	var err error
	switch format.Value {
	case "png":
		err = png.Encode(&buf, img)
	case "jpeg", "jpg":
		quality := 90
		if len(args) == 3 {
			quality = int(intArg(args[2]))
		}
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "gif":
		err = gif.Encode(&buf, img, nil)
	default:
		return newError("img_encode: unsupported format %s", format.Value)
	}
	if err != nil {
		return newError("img_encode: %s", err.Error())
	}

	return &object.String{Value: string(buf.Bytes())}
}

func imgBlur(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("img_blur expects 1-2 arguments (image, [radius]), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_blur expects an image, got %s", args[0].Type())
	}

	radius := 3
	if len(args) == 2 {
		radius = int(intArg(args[1]))
	}
	if radius < 1 {
		radius = 1
	}

	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum, aSum int64
			count := 0
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					px := x + dx
					py := y + dy
					if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
						r, g, b, a := img.At(px, py).RGBA()
						rSum += int64(r >> 8)
						gSum += int64(g >> 8)
						bSum += int64(b >> 8)
						aSum += int64(a >> 8)
						count++
					}
				}
			}
			if count > 0 {
				blurred.Set(x, y, color.RGBA{
					R: uint8(rSum / int64(count)),
					G: uint8(gSum / int64(count)),
					B: uint8(bSum / int64(count)),
					A: uint8(aSum / int64(count)),
				})
			}
		}
	}

	return &object.Image{Value: blurred}
}

func imgDrawText(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("img_draw_text expects at least 3 arguments (image, text, x, y, ...), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_draw_text expects an image, got %s", args[0].Type())
	}

	text, ok := args[1].(*object.String)
	if !ok {
		return newError("img_draw_text expects string text, got %s", args[1].Type())
	}

	x := intArg(args[2])
	y := intArg(args[3])

	r := uint8(0)
	g := uint8(0)
	b := uint8(0)
	if len(args) >= 7 {
		r = uint8(intArg(args[4]))
		g = uint8(intArg(args[5]))
		b = uint8(intArg(args[6]))
	}

	bounds := img.Bounds()
	switch imgTyped := img.(type) {
	case *image.RGBA:
		for i := range text.Value {
			px := int(x) + i*8
			py := int(y)
			if px < bounds.Max.X && py < bounds.Max.Y {
				imgTyped.Set(px, py, color.RGBA{r, g, b, 255})
			}
		}
	}

	return object.NullObj
}

func imgRotate(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("img_rotate expects 2 arguments (image, angle), got %d", len(args))
	}

	img := getImage(args)
	if img == nil {
		return newError("img_rotate expects an image, got %s", args[0].Type())
	}

	angle := intArg(args[1])
	angle = angle % 360
	if angle < 0 {
		angle += 360
	}

	bounds := img.Bounds()
	var rotated *image.RGBA

	switch angle {
	case 90:
		rotated = image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotated.Set(bounds.Max.Y-1-y, x, img.At(x, y))
			}
		}
	case 180:
		rotated = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotated.Set(bounds.Max.X-1-x, bounds.Max.Y-1-y, img.At(x, y))
			}
		}
	case 270:
		rotated = image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotated.Set(y, bounds.Max.X-1-x, img.At(x, y))
			}
		}
	default:
		// No rotation needed for 0/360
		rotated = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rotated.Set(x, y, img.At(x, y))
			}
		}
	}

	return &object.Image{Value: rotated}
}

// Helper functions

func getImage(args []object.Object) image.Image {
	if len(args) == 0 {
		return nil
	}
	imgObj, ok := args[0].(*object.Image)
	if !ok {
		return nil
	}
	return imgObj.Value
}

func intArg(arg object.Object) int64 {
	switch a := arg.(type) {
	case *object.Integer:
		return a.Value
	case *object.Float:
		return int64(a.Value)
	case *object.String:
		n, err := strconv.ParseInt(a.Value, 10, 64)
		if err == nil {
			return n
		}
	}
	return 0
}

func setHashField(hash *object.Hash, key string, val object.Object) {
	k := &object.String{Value: key}
	hash.Pairs[k.HashKey()] = object.HashPair{Key: k, Value: val}
}
