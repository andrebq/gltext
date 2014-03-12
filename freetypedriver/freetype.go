// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package driver

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"github.com/go-gl/glh"
	"github.com/go-gl/gltext"
	"image"
	"io"
	"io/ioutil"
)

func init() {
	gltext.RegisterDriver("freetype-go", LoadTruetype)
	gltext.RegisterDriver("default", LoadTruetype)
}

func LoadTruetype(r io.Reader, scale int32, low, high rune, dir gltext.Direction) (*gltext.FontConfig, *image.RGBA, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	// Read the truetype font.
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, nil, err
	}

	// Create our FontConfig type.
	var fc gltext.FontConfig
	fc.Dir = dir
	fc.Low = low
	fc.High = high
	fc.Glyphs = make(gltext.Charset, high-low+1)

	// Create an image, large enough to store all requested glyphs.
	//
	// We limit the image to 16 glyphs per row. Then add as many rows as
	// needed to encompass all glyphs, while making sure the resulting image
	// has power-of-two dimensions.
	gc := int32(len(fc.Glyphs))
	glyphsPerRow := int32(16)
	glyphsPerCol := (gc / glyphsPerRow) + 1

	gb := ttf.Bounds(scale)
	gw := (gb.XMax - gb.XMin)
	gh := (gb.YMax - gb.YMin) + 5
	iw := glh.Pow2(uint32(gw * glyphsPerRow))
	ih := glh.Pow2(uint32(gh * glyphsPerCol))

	rect := image.Rect(0, 0, int(iw), int(ih))
	img := image.NewRGBA(rect)

	// Use a freetype context to do the drawing.
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ttf)
	c.SetFontSize(float64(scale))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	// Iterate over all relevant glyphs in the truetype font and
	// draw them all to the image buffer.
	//
	// For each glyph, we also create a corresponding Glyph structure
	// for our Charset. It contains the appropriate glyph coordinate offsets.
	var gi int
	var gx, gy int32

	for ch := low; ch <= high; ch++ {
		index := ttf.Index(ch)
		metric := ttf.HMetric(scale, index)

		fc.Glyphs[gi].Advance = int(metric.AdvanceWidth)
		fc.Glyphs[gi].X = int(gx)
		fc.Glyphs[gi].Y = int(gy)
		fc.Glyphs[gi].Width = int(gw)
		fc.Glyphs[gi].Height = int(gh)

		pt := freetype.Pt(int(gx), int(gy)+int(c.PointToFix32(float64(scale))>>8))
		c.DrawString(string(ch), pt)

		if gi%16 == 0 {
			gx = 0
			gy += gh
		} else {
			gx += gw
		}

		gi++
	}

	return &fc, img, err
}
