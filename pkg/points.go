package pkg

import (
	"errors"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	font2 "golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"strconv"
	"strings"
)

var font *truetype.Font
var face font2.Face
var marker image.Image
var markerView image.Image

type DefinitionPoint struct {
	X        float64
	Y        float64
	Angle    float64
	HasAngle bool
	Number   int
}

func PointFromArray(arr []string) (*DefinitionPoint, error) {
	if len(arr) < 3 {
		return nil, errors.New("invalid point array")
	}
	var angle float64 = 0
	hasAngle := false

	x, _ := strconv.ParseFloat(arr[0], 64)
	y, _ := strconv.ParseFloat(arr[1], 64)
	number, _ := strconv.Atoi(arr[2])

	if len(arr) > 3 {
		hasAngle = true
		angle, _ = strconv.ParseFloat(arr[3], 64)
	}

	return &(DefinitionPoint{
		X: x, Y: y, Angle: angle, HasAngle: hasAngle, Number: number,
	}), nil
}

func ParsePoints(query string) []DefinitionPoint {
	var points []DefinitionPoint
	strs := strings.Split(query, ",")

	for _, str := range strs {
		coords := strings.Split(str, "_")

		if point, err := PointFromArray(coords); err == nil {
			points = append(points, *point)
		}
	}

	return points
}

func GetPointsCache(points []DefinitionPoint) string {
	cache := ""
	for _, p := range points {
		cache += fmt.Sprintf("%f_%f_%d_%f:", p.X, p.Y, p.Number, p.Angle)
	}
	return cache
}

func InitFont() error {
	var err error

	font, err = truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	face = truetype.NewFace(font, &truetype.Options{Size: 16})
	marker, err = gg.LoadPNG("./assets/marker.png")
	if err != nil {
		return err
	}

	markerView, err = gg.LoadPNG("./assets/marker-view.png")
	if err != nil {
		return err
	}

	return nil
}

func DrawMarker(ctx *gg.Context, _x, _y float64, number int, inUV bool) {
	var x, y float64

	if inUV {
		x = _x * float64(ctx.Width())
		y = _y * float64(ctx.Height())
	} else {
		x = _x
		y = _y
	}

	ctx.DrawImageAnchored(marker, int(x), int(y), 0.5, 0.5)
	ctx.SetFontFace(face)
	ctx.SetRGB(1, 1, 1)
	ctx.Fill()
	ctx.DrawStringAnchored(strconv.Itoa(number), x, y-7, 0.5, 0.5)
}

func DrawMarkerView(ctx *gg.Context, _x, _y, angle float64, number int, inUV bool) {
	var x, y float64

	if inUV {
		x = _x * float64(ctx.Width())
		y = _y * float64(ctx.Height())
	} else {
		x = _x
		y = _y
	}

	ctx.RotateAbout(gg.Radians(angle), x, y)
	ctx.DrawImageAnchored(markerView, int(x), int(y), 0.5, 0.5)
	ctx.RotateAbout(gg.Radians(-angle), x, y)
	ctx.SetFontFace(face)
	ctx.SetRGB(1, 1, 1)
	ctx.Fill()
	ctx.DrawStringAnchored(strconv.Itoa(number), x, y-2, 0.5, 0.5)
}

func DrawPoints(fileName string, points []DefinitionPoint, inUV bool) {
	img, _ := gg.LoadImage(fileName)
	dc := gg.NewContextForImage(img)

	for _, p := range points {
		if p.HasAngle {
			DrawMarkerView(dc, p.X, p.Y, p.Angle, p.Number, inUV)
		} else {
			DrawMarker(dc, p.X, p.Y, p.Number, inUV)
		}
	}

	dc.SavePNG(fileName)
}
