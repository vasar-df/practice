package vasar

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/skin"
	"image"
	"image/draw"
	"image/png"
	"os"
)

var (
	// steve is the Vasar steve skin, used when the player has an unsupported skin.
	steve skin.Skin
	// bounds contains all possible bounds for skins.
	bounds [][][2][2]int
)

// init initializes the bounds for all possible skins.
func init() {
	skinBounds, err := os.ReadFile("assets/skins/bounds.json")
	if err != nil {
		panic(err)
	}
	geometry, err := os.ReadFile("assets/skins/geometry.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(skinBounds, &bounds)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile("assets/skins/steve.png", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, rect.Min, draw.Src)

	steve = skin.New(rect.Dx(), rect.Dy())
	steve.ModelConfig.Default = "geometry.humanoid.custom"
	steve.Model = geometry
	steve.Pix = rgba.Pix
}

// searchTransparency searches for transparency in the given skin, returning the found percentage. This percentage is
// between 0 and 1.
func searchTransparency(skin skin.Skin) (float64, error) {
	max := skin.Bounds().Size()
	sizeBounds, err := sizeSpecificBounds(max)
	if err != nil {
		return 0.0, err
	}

	var transparent int
	for _, bound := range sizeBounds {
		if bound[1][0] > max.X || bound[1][1] > max.Y {
			// Skip bounds that are, well, out of bounds.
			continue
		}
		for y := bound[0][1]; y <= bound[1][1]; y++ {
			for x := bound[0][0]; x <= bound[1][0]; x++ {
				if skin.Pix[((max.X*y)+x)*4+3] < 127 {
					transparent++
				}
			}
		}
	}
	return float64(transparent) / float64(max.X*max.Y), nil
}

// sizeSpecificBounds returns the size specific bounds for the given size.
func sizeSpecificBounds(size image.Point) ([][2][2]int, error) {
	switch size {
	case image.Point{X: 64, Y: 32}, image.Point{X: 64, Y: 64}:
		return bounds[0], nil
	case image.Point{X: 128, Y: 128}:
		return bounds[1], nil
	}
	return nil, fmt.Errorf("skin: unsupported skin size (%v)", size)
}
