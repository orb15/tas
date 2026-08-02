package system

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"

	h "tas/internal/cmd/helpers"
	"tas/internal/model"
	"tas/internal/util"

	"github.com/fogleman/gg"
)

const (
	//key image information
	canvasWidth           = 3000
	canvasHeight          = 2000
	margin                = 50
	borderWidth           = 5
	borderSpace           = 10
	canvasBackgroundColor = "000000"
	textColor             = "ffffff"
	borderColor           = "333333"
	sysDiagInnerColor     = canvasBackgroundColor

	//title area - all coords are _relative_ to the region!
	titleMinX = 0
	titleMinY = 0
	titleMaxX = titleMinX + canvasWidth - margin - margin
	titleMaxY = titleMinY + 100

	//define system diagram area constants - all coords are _relative_ to the region!
	sysDiagMinX           = 0
	sysDiagMinY           = titleMaxY + borderSpace
	sysDiagMaxX           = sysDiagMinX + 2500 - margin - margin
	sysDiagMaxY           = sysDiagMinY + 500
	sysDiagSunRadius      = 20
	sysDiagSunHaloOffset  = 5
	sysDiagOrbitPathWidth = 1

	//UWP area - all coords are _relative_ to the region!
	uwpMinX = sysDiagMaxX + borderSpace
	uwpMinY = sysDiagMinY
	uwpMaxX = titleMaxX
	uwpMaxY = sysDiagMaxY

	//data area - all coords are _relative_ to the region
	dataMinX = sysDiagMinX
	dataMinY = sysDiagMaxY + borderSpace
	dataMaxX = titleMaxX
	dataMaxY = canvasHeight - margin - margin
)

func produceImage(ctx *util.TASContext, sys *model.SolarSystem) {

	// create the font manager - hold all the fonts used in the image
	fonts, err := NewFontManager()
	if err != nil {
		ctx.Logger().Error().Msg("Error creating FontManager " + err.Error())
		return
	}

	// create the texture manager - hold all textures to apply to orbitals
	textureMgr, err := NewTextureManager()
	if err != nil {
		ctx.Logger().Error().Msg("Error creating textureManager " + err.Error())
		return
	}

	// establish the image and set background color
	fullCanvas := gg.NewContext(canvasWidth, canvasHeight)
	fullCanvas.SetHexColor(canvasBackgroundColor)
	fullCanvas.Clear()

	// establish margins. Origin is now (margin, margin) for this context and all
	// derived contexts
	fullCanvas.Push()
	fullCanvas.Translate(margin, margin)

	//render the elements
	drawTitle(fullCanvas, fonts, sys)
	drawSystemDiagram(fullCanvas, fonts, textureMgr, sys)
	drawUWPDetails(fullCanvas, fonts, sys)
	drawDataSpace(fullCanvas, fonts, sys)

	// need to pop back to the full image now that we have done our drawing everywhere
	fullCanvas.ClearPath()
	fullCanvas.Pop()

	// render the image in the browser
	displayImageInBrowser(ctx, fullCanvas)

	//also write to file if requested
	writeToFile, _ := ctx.Config().Flags.GetBool(util.ToFileFlagName)
	if writeToFile {
		fullCanvas.SavePNG(filepath.Join(h.OutputDirectoryName, sys.ToImageFilename()))
	}

}

func drawTitle(gtx *gg.Context, fonts *FontManager, sys *model.SolarSystem) {

	// push existing constraints onto the stack, ensure we restore the relative coordinate
	// system to where it was before we were called
	gtx.Push()
	defer gtx.Pop()

	// Calculate true width and height components
	width := float64(titleMaxX) - float64(titleMinX)
	height := float64(titleMaxY) - float64(titleMinY)

	// Move canvas origin to the start of the title block
	gtx.Translate(titleMinX, titleMinY)

	// Draw the border
	rectBorder(gtx, 0, 0, width, height, borderWidth)

	//get uwp, build title
	uwp := parseUWPDetails(sys)
	title := fmt.Sprintf("%s  %s    Hex: %s", uwp.name, uwp.uwp, uwp.hex)

	// render the title
	textInBox(gtx, titleMinX+(2*borderWidth), height, fonts.TitleFont, textColor, title)
}

func drawSystemDiagram(gtx *gg.Context, fonts *FontManager, textureMgr *TextureManager,
	sys *model.SolarSystem) {

	// push existing constraints onto the stack, ensure we restore the relative coordinate
	// system to where it was before we were called
	gtx.Push()
	defer gtx.Pop()

	// calculate width and height of the box
	width := float64(sysDiagMaxX) - float64(sysDiagMinX)
	height := float64(sysDiagMaxY) - float64(sysDiagMinY)

	// move canvas origin to the start of the diagram block
	gtx.Translate(sysDiagMinX, sysDiagMinY)

	// Draw the border
	rectBorder(gtx, 0, 0, width, height, borderWidth)

	// we can have up to 16 orbitals, so we need 16 cells in this space
	var left, right float64
	left = sysDiagMinX + borderWidth
	right = sysDiagMaxX - borderWidth
	diff := right - left
	delta := diff / 16
	gtx.SetHexColor(sysDiagInnerColor)
	gtx.SetLineWidth(1)

	// need 1 fewer vertical lines than desired cells
	for i := 1; i <= 16-1; i++ {
		gtx.DrawLine(sysDiagMinX+(delta*float64(i)), 0, delta*float64(i), height)
		gtx.Stroke()
	}

	// draw a horizontal line about 1/3rd the way down. Below this will be text, above it an image
	drop := .3 * height
	gtx.DrawLine(0, drop, right, drop)
	gtx.Stroke()

	// add data to boxes - start in first cell as there can be 16 orbitals so no sun will be drawn
	// I could add the sun but do not want the boxes to be any thinner
	var xStart, yStart, cellDelta float64
	xStart = 2 * borderWidth
	yStart = drop + 50
	cellDelta = 150 //manually adjusted

	for i, o := range sys.Orbitals {
		table := make([]string, 0, 4)
		table = append(table, fmt.Sprintf("Orbital: %d", o.OrbitalNumber))
		switch {
		case o.IsAsteroid:
			table = append(table, "Asteroid")
		case o.IsGasGiant:
			table = append(table, "Gas Giant")
		default:
			table = append(table, "Rocky Planet")
		}
		table = append(table, fmt.Sprintf("%.2fAU", o.OrbitalDistanceAU))
		table = append(table, fmt.Sprintf("Moons: %d", len((o.Moons))))
		x := xStart + (float64(i) * cellDelta)
		table = wrapTableTo(10, table)
		renderTableText(gtx, x, yStart, fonts.TableFont, textColor, table)
	}

	// draw planets filled with textures
	var radius float64 = 60
	var cy float64 = sysDiagMinY - 35
	var cx float64
	gtx.SetHexColor(sysDiagInnerColor)
	for i, o := range sys.Orbitals {

		gtx.Push()
		cx = (sysDiagMinX + 75) + (float64(i) * cellDelta)

		gtx.DrawCircle(cx, cy, radius)
		gtx.Clip()

		var img image.Image
		switch {
		case o.IsAsteroid:
			img = textureMgr.AsteroidImage()
		default:
			img = textureMgr.PlanetImage()
		}
		imgWidth := float64(img.Bounds().Dx())
		imgHeight := float64(img.Bounds().Dy())
		topLeftX := cx - (imgWidth / 2.0)
		topLeftY := cy - (imgHeight / 2.0)

		// 5. Natively paint the texture image onto the canvas.
		// The clipping mask ensures pixels outside the circle are discarded.
		gtx.DrawImage(img, int(topLeftX), int(topLeftY))

		gtx.ResetClip()
		gtx.Pop()

	}

}

func drawUWPDetails(gtx *gg.Context, fonts *FontManager, sys *model.SolarSystem) {
	// push existing constraints onto the stack, ensure we restore the relative coordinate
	// system to where it was before we were called
	gtx.Push()
	defer gtx.Pop()

	// Calculate width and height of the box
	width := float64(uwpMaxX) - float64(uwpMinX)
	height := float64(uwpMaxY) - float64(uwpMinY)

	// Move canvas origin to the start of the diagram block
	gtx.Translate(uwpMinX, uwpMinY)

	// Draw the border
	rectBorder(gtx, 0, 0, width, height, borderWidth)

	// get parsed uwp
	uwp := parseUWPDetails(sys)

	//custom place the uwp in the top center of the box
	textAt(gtx, borderWidth*2, 48, fonts.UWPHeaderFace, textColor, uwp.uwp)

	//convert UWP to a table, establish word wrap and render it
	table := uwp.toTable()
	table = wrapTableTo(32, table)
	renderTableText(gtx, borderWidth*2, 96, fonts.TableFont, textColor, table)
}

func drawDataSpace(gtx *gg.Context, fonts *FontManager, sys *model.SolarSystem) {
	// push existing constraints onto the stack, ensure we restore the relative coordinate
	// system to where it was before we were called
	gtx.Push()
	defer gtx.Pop()

	// calculate true width and height components
	width := float64(dataMaxX) - float64(dataMinX)
	height := float64(dataMaxY) - float64(dataMinY)

	// move canvas origin to the start of the title block
	gtx.Translate(dataMinX, dataMinY)

	// draw the border
	rectBorder(gtx, 0, 0, width, height, borderWidth)

	// establish space for at least the orbitals but not moons
	data := make([]string, 0, len(sys.Orbitals))

	// build the orbital data table
	name := sys.Name
	for _, o := range sys.Orbitals {
		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %d", name, o.OrbitalNumber))
		oType := ""
		switch {
		case o.IsAsteroid:
			oType = "Asteroid"
		case o.IsGasGiant:
			oType = "Gas Giant"
		case o.IsPrimary:
			uwp := parseUWPDetails(sys)
			oType = "Rocky, Primary: " + uwp.uwp
		default:
			oType = "Rocky"
		}
		line.WriteString(" " + oType)
		if o.IsPrimary {
			line.WriteString(fmt.Sprintf(" %.2fAU", o.OrbitalDistanceAU))
		} else {
			line.WriteString(fmt.Sprintf(" (%dkm) %.2fAU", o.Size, o.OrbitalDistanceAU))
		}
		data = append(data, line.String())
		for _, t := range o.Tags {
			data = append(data, fmt.Sprintf("Tag -> %s: %s", t.TagName, t.Description))
		}
		for i, m := range o.Moons {
			data = append(data, fmt.Sprintf("  Moon %d (%dkm)", i+1, m.Size))
			for _, t := range m.Tags {
				data = append(data, fmt.Sprintf("  Tag -> %s: %s", t.TagName, t.Description))
			}
		}
		data = append(data, "\n")
	}

	// at this point, the data table is built. It could vary wildly in size, so we need to build
	// some flexibility into this based on the number of rows we are trying to present
	limits := getOrbitalTableLimits(len(data), fonts)

	data = wrapTableTo(limits.wrapCol1At, data)

	// too many rows for left column
	if limits.use2Cols {
		col1 := data[:limits.maxRowsPerColumn]
		col2 := data[limits.maxRowsPerColumn:]
		renderTableText(gtx, borderWidth*2, 48, limits.fontToUse, textColor, col1)
		renderTableText(gtx, float64(limits.startCol2X), 48, limits.fontToUse, textColor, col2)
	} else {
		renderTableText(gtx, borderWidth*2, 48, limits.fontToUse, textColor, data)
	}
}

type orbitalTableSpecs struct {
	maxRowsPerColumn int
	use2Cols         bool
	fontToUse        font.Face
	wrapCol1At       int
	startCol2X       int
}

func getOrbitalTableLimits(tableLen int, fonts *FontManager) orbitalTableSpecs {

	//println("system details rowcount: ", tableLen)
	limits := orbitalTableSpecs{}
	switch {
	case tableLen <= 33:
		limits.use2Cols = false
		limits.fontToUse = fonts.TableFont
		limits.maxRowsPerColumn = 33
		limits.wrapCol1At = 103
		limits.startCol2X = 1200

	case tableLen <= 61:
		limits.use2Cols = true
		limits.fontToUse = fonts.TableFont
		limits.maxRowsPerColumn = 33
		limits.wrapCol1At = 103
		limits.startCol2X = 1400

	case tableLen <= 98:
		limits.use2Cols = true
		limits.fontToUse = fonts.IntermedTableFont
		limits.maxRowsPerColumn = 49
		limits.wrapCol1At = 103
		limits.startCol2X = 1200

	default: // this is a complete guess at this point
		limits.use2Cols = true
		limits.fontToUse = fonts.SmallTableFont
		limits.maxRowsPerColumn = 72
		limits.wrapCol1At = 125
		limits.startCol2X = 1100
	}
	return limits
}
