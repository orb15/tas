package system

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"

	"tas/internal/model"
	"tas/internal/util"

	"github.com/fogleman/gg"
)

const (
	uwpParserString     = `([a-zA-Z0-9\- \.']{1,}) ([0-9]{4}) ([A-EX]{1})([0-9A]{1})([0-9A-F]{5})\-([0-9A-F])`
	uwpOnlyParserString = `([A-EX]{1}[0-9A]{1}[0-9A-F]{5}\-[0-9A-F].*[A-Z]{1,})`
)

var (
	uwpParserRegex     = regexp.MustCompile(uwpParserString)
	uwpOnlyParserRegex = regexp.MustCompile(uwpOnlyParserString)
)

// ------------------------------------------------------
// Font Manager
// ------------------------------------------------------

type FontManager struct {
	TitleFont         font.Face
	UWPHeaderFace     font.Face
	TableFont         font.Face
	IntermedTableFont font.Face
	SmallTableFont    font.Face
	DiagramFont       font.Face
}

func NewFontManager() (*FontManager, error) {

	//title font
	titleFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-ExtraBold.ttf"), 64)
	if err != nil {
		return nil, err
	}

	//UWP Header
	uwpHeaderFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-Bold.ttf"), 48)
	if err != nil {
		return nil, err
	}

	//table font
	tableFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-Regular.ttf"), 30)
	if err != nil {
		return nil, err
	}

	//table font
	intermedTableFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-Regular.ttf"), 20)
	if err != nil {
		return nil, err
	}

	//table font - small
	smallTableFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-Regular.ttf"), 14)
	if err != nil {
		return nil, err
	}

	//table font
	diagramFace, err := loadAndScaleFont(filepath.Join("assets", "fonts", "genos_static", "Genos-Regular.ttf"), 20)
	if err != nil {
		return nil, err
	}

	return &FontManager{
		TitleFont:         titleFace,
		UWPHeaderFace:     uwpHeaderFace,
		TableFont:         tableFace,
		IntermedTableFont: intermedTableFace,
		SmallTableFont:    smallTableFace,
		DiagramFont:       diagramFace,
	}, nil
}

func loadAndScaleFont(path string, scale float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fontObj, _ := sfnt.Parse(fontBytes)
	face, _ := opentype.NewFace(fontObj, &opentype.FaceOptions{
		Size: scale,
		DPI:  72,
	})

	return face, nil
}

// ------------------------------------------------------
// Textuure Manager
// ------------------------------------------------------
type OffsetPattern struct {
	Image   image.Image
	OffsetX int // Left edge of texture in world coordinates
	OffsetY int // Top edge of texture in world coordinates
}

// ColorAt fulfills the required gg.Pattern interface using absolute math
func (op *OffsetPattern) ColorAt(x, y int) color.Color {
	bounds := op.Image.Bounds()

	// Translate global canvas (x, y) down to local image (lx, ly)
	lx := x - op.OffsetX
	ly := y - op.OffsetY

	// Strict RepeatNone out-of-bounds handling
	if lx < 0 || ly < 0 || lx >= bounds.Dx() || ly >= bounds.Dy() {
		return color.Transparent
	}

	return op.Image.At(bounds.Min.X+lx, bounds.Min.Y+ly)
}

type TextureManager struct {
	asteroidImages []image.Image
	planetImages   []image.Image
	testImage      image.Image
	usedPlanets    map[int]struct{}
	usedAteroids   map[int]struct{}
	dice           util.Dice
}

func NewTextureManager() (*TextureManager, error) {

	path := filepath.Join("assets", "textures", "planet_*.png")
	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("failed to glob rocky textures: %w", err)
	}
	planetImages := make([]image.Image, 0, len(matches))
	for _, m := range matches {
		img, err := gg.LoadImage(m)
		if err != nil {
			return nil, fmt.Errorf("failed to load rocky texture %s: %w", m, err)
		}
		planetImages = append(planetImages, img)
	}

	path = filepath.Join("assets", "textures", "asteroid_*.png")
	matches, err = filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("failed to glob asteroid textures: %w", err)
	}
	asteroidImages := make([]image.Image, 0, len(matches))
	for _, m := range matches {
		img, err := gg.LoadImage(m)
		if err != nil {
			return nil, fmt.Errorf("failed to load asteroid texture %s: %w", m, err)
		}
		asteroidImages = append(asteroidImages, img)
	}

	testImage, err := gg.LoadImage(filepath.Join("assets", "textures", "test.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to load test/fallback texture test.png: %w", err)
	}

	tm := &TextureManager{
		planetImages:   planetImages,
		asteroidImages: asteroidImages,
		testImage:      testImage,
		dice:           util.NewDice(),
	}

	return tm, nil
}

func (tm *TextureManager) PlanetImage() image.Image {

	planetImgCount := len(tm.planetImages)
	switch {
	case planetImgCount == 0:
		return tm.testImage
	case planetImgCount == 1:
		img := tm.planetImages[0]
		tm.planetImages = nil
		return img
	case planetImgCount > 1:
		roll := tm.dice.Dx(planetImgCount - 1)
		img := tm.planetImages[roll]
		remainder := tm.planetImages[:roll]
		if roll != planetImgCount-1 {
			remainder = append(remainder, tm.planetImages[roll+1:]...)
		}
		tm.planetImages = remainder
		return img
	}

	return tm.testImage
}

func (tm *TextureManager) AsteroidImage() image.Image {
	asteroidImgCount := len(tm.asteroidImages)
	switch {
	case asteroidImgCount == 0:
		return tm.testImage
	case asteroidImgCount == 1:
		img := tm.asteroidImages[0]
		tm.asteroidImages = nil
		return img
	case asteroidImgCount > 1:
		roll := tm.dice.Dx(asteroidImgCount - 1)
		img := tm.asteroidImages[roll]
		remainder := tm.asteroidImages[:roll]
		if roll != asteroidImgCount-1 {
			remainder = append(remainder, tm.asteroidImages[roll+1:]...)
		}
		tm.asteroidImages = remainder
		return img
	}

	return tm.testImage
}

// ------------------------------------------------------
// UWP Helpers - note uwp already validated earlier to
// build system
// ------------------------------------------------------
type uwpDetails struct {
	uwp      string
	name     string
	hex      string
	starport string
	size     string
	atmo     string
	hydro    string
	pop      string
	gov      string
	law      string
	tech     string
}

func (u *uwpDetails) toTable() []string {
	s := make([]string, 8)
	s[0] = u.starport
	s[1] = u.size
	s[2] = u.atmo
	s[3] = u.hydro
	s[4] = u.pop
	s[5] = u.gov
	s[6] = u.law
	s[7] = u.tech

	return s
}

// I could bring in the JSOn data at this point, but I want ot very tightly control
// what is displayed on screen as layout is paramount
func parseUWPDetails(sys *model.SolarSystem) uwpDetails {

	//locate primary world to fetch UWP
	var uwp string
	for _, o := range sys.Orbitals {
		if !o.IsPrimary {
			continue
		}
		uwp = o.UWP
		break
	}

	//name is readily available
	details := uwpDetails{}
	details.name = sys.Name

	//get the small-form uwp
	matches := uwpOnlyParserRegex.FindStringSubmatch(uwp)
	details.uwp = matches[1]

	//parse uwp in detail
	matches = uwpParserRegex.FindStringSubmatch(uwp)

	//hex location
	details.hex = matches[2]

	var t string
	starport := matches[3]
	switch starport {
	case "A":
		t = "Excellent"
	case "B":
		t = "Good"
	case "C":
		t = "Average"
	case "D":
		t = "Poor"
	case "E":
		t = "Frontier"
	case "X":
		t = "None"
	}
	details.starport = fmt.Sprintf("Starport %s: %s", starport, t)

	size := matches[4]
	switch size {
	case "0":
		t = "less than 1000km (asteroid)"
	case "1":
		t = "1600km (Triton)"
	case "2":
		t = "3200km (Moon)"
	case "3":
		t = "4800km (Mercury)"
	case "4":
		t = "6400km (Mars)"
	case "5":
		t = "8000km"
	case "6":
		t = "9600km"
	case "7":
		t = "11200km"
	case "8":
		t = "12,800km (Earth)"
	case "9":
		t = "14,400km"
	case "A":
		t = "16000km"
	}
	details.size = fmt.Sprintf("Size %s: %s", size, t)

	core5 := matches[5]

	atmo := string(core5[0])
	switch atmo {
	case "0":
		t = "None (Moon)"
	case "1":
		t = "Trace (Mars)"
	case "2":
		t = "Very thin, Tainted"
	case "3":
		t = "Very thin"
	case "4":
		t = "Thin, Tainted"
	case "5":
		t = "Thin"
	case "6":
		t = "Standard (Earth)"
	case "7":
		t = "Standard, Tainted"
	case "8":
		t = "Dense"
	case "9":
		t = "Dense, Tainted"
	case "A":
		t = "Exotic"
	case "B":
		t = "Corrosive (Venus)"
	case "C":
		t = "Insidious"
	case "D":
		t = "Very Dense"
	case "E":
		t = "Low"
	case "F":
		t = "Unusual"
	}
	details.atmo = fmt.Sprintf("Atmosphere %s: %s", atmo, t)

	hydro := string(core5[1])
	switch hydro {
	case "0":
		t = "Desert"
	case "1":
		t = "Dry"
	case "2":
		t = "Few small seas"
	case "3":
		t = "Small seas and oceans"
	case "4":
		t = "Wet"
	case "5":
		t = "A large ocean"
	case "6":
		t = "Large oceans"
	case "7":
		t = "Earth-like"
	case "8":
		t = "Only islands"
	case "9":
		t = "Few islands"
	case "A":
		t = "Waterworld"
	}
	details.hydro = fmt.Sprintf("Hydrographics %s: %s", hydro, t)

	pop := string(core5[2])
	switch pop {
	case "0":
		t = "None"
	case "1":
		t = "Few"
	case "2":
		t = "Hundreds"
	case "3":
		t = "Thosands"
	case "4":
		t = "Tens of thousands"
	case "5":
		t = "Hundreds of thousands"
	case "6":
		t = "Millions"
	case "7":
		t = "Tens of millions"
	case "8":
		t = "Hundreds of millions"
	case "9":
		t = "Billions"
	case "A":
		t = "Tens of billions"
	case "B":
		t = "Hundreds of billions"
	case "C":
		t = "Trillions"
	}
	details.pop = fmt.Sprintf("Population %s:  %s", pop, t)

	gov := string(core5[3])
	switch gov {
	case "0":
		t = "None"
	case "1":
		t = "Company or Corporation"
	case "2":
		t = "Participating Democracy"
	case "3":
		t = "Self-Perpetuating Oligarchy"
	case "4":
		t = "Representative Democracy"
	case "5":
		t = "Feudal Technocracy"
	case "6":
		t = "Captive Government"
	case "7":
		t = "Balkanisation"
	case "8":
		t = "Civil Service Bureaucracy"
	case "9":
		t = "Impersonal Bureaucracy"
	case "A":
		t = "Charismatic Dictator"
	case "B":
		t = "Non-Charismatic Leader"
	case "C":
		t = "Charismatic Oligarchy"
	case "D":
		t = "Religious Dictatorship"
	case "E":
		t = "Religious Autocracy"
	case "F":
		t = "Totalitarian Oligarchy"
	}
	details.gov = fmt.Sprintf("Government %s: %s", gov, t)

	law := string(core5[4])
	switch law {
	case "0":
		t = "Almost no restrictions"
	case "1":
		t = "Poison, WMD, Battle Dress"
	case "2":
		t = "Energy weapons, Combat Armor"
	case "3":
		t = "Military weapons, Flak Armor"
	case "4":
		t = "Light assault, submachine guns, Cloth"
	case "5":
		t = "Personal concealable weapons, Mesh"
	case "6":
		t = "All firearms but shotguns"
	case "7":
		t = "Shotguns"
	case "8":
		t = "Bladed weapons & stunners, All armor"
	case "9":
		t = "All weapons and armor"
	}
	details.law = fmt.Sprintf("Law Level %s: %s", law, t)

	// tech level is easy
	details.tech = fmt.Sprintf("Tech Level %s", matches[6])

	return details
}

// ------------------------------------------------------
// Misc Helpers
// ------------------------------------------------------
func textAt(gtx *gg.Context, x float64, y float64, fontFace font.Face, color string, s string) {
	gtx.SetFontFace(fontFace)
	gtx.SetHexColor(color)
	gtx.DrawStringAnchored(s, x, y, 0, 0)
}

func textInBox(gtx *gg.Context, x float64, height float64, fontFace font.Face, color string, s string) {
	gtx.SetFontFace(fontFace)
	gtx.SetHexColor(color)
	y := height - gtx.FontHeight()/2
	gtx.DrawStringAnchored(s, x, y, 0, 0)
}

func renderTableText(gtx *gg.Context, x float64, y float64, fontFace font.Face, color string, table []string) {
	gtx.SetFontFace(fontFace)
	gtx.SetHexColor(color)
	fontHeight := gtx.FontHeight()

	for i, s := range table {
		gtx.DrawString(s, x, y+(fontHeight*float64(i)))
	}
}

func wrapTableTo(count int, table []string) []string {
	final := make([]string, 0, len(table))
	for _, s := range table {
		if len(s) <= count {
			final = append(final, s)
			continue
		}
		parts := strings.Split(s, " ")
		curLineLen := 0
		line := ""
		for _, p := range parts {
			if curLineLen < count {
				if curLineLen+len(p)+1 < count {
					if len(line) == 0 {
						line = p
						curLineLen = len(p)
					} else {
						line = line + " " + p
						curLineLen += len(p) + 1
					}
					continue
				}
				final = append(final, line)
				curLineLen = len(p)
				line = p
			}
		}
		if len(line) > 0 {
			final = append(final, line)
		}
	}
	return final
}

func rectBorder(gtx *gg.Context, minX, minY, width, height float64, borderWidth float64) {
	gtx.SetHexColor(borderColor)
	gtx.SetLineWidth(borderWidth)
	gtx.DrawRectangle(minX, minY, width, height)
	gtx.Stroke()
}

func displayImageInBrowser(ctx *util.TASContext, dc *gg.Context) {
	log := ctx.Logger()

	//create a channel to communicate the shutdown signal
	done := make(chan struct{})

	//set up endpoint handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			close(done) //close the channel once the function exits - on error or normally
		}()
		w.Header().Set("Content-Type", "image/png")

		//send the image directly to the browser
		if err := png.Encode(w, dc.Image()); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			log.Err(err).Msg("Failed to encode image")
			return
		}
	})

	//configure a custom server instance
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	//start the server
	go func() {
		log.Info().Msg("Server ready!")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	//give the server a bit to set up and then point the Windows default browser at the server
	go func() {
		time.Sleep(100 * time.Millisecond) // Give the server a tiny moment to bind to port 8080
		// WSL can execute Windows binaries (.exe) natively through its path translation!
		_ = exec.Command("cmd.exe", "/c", "start", "http://localhost:8080").Run()
	}()

	//block execution here until the render is complete
	<-done

	// 7. Trigger a quick graceful shutdown (allowing 1 second to flush network links)
	log.Info().Msg("Image served successfully. Shutting down webserver...")
	tctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := server.Shutdown(tctx); err != nil {
		log.Err(err).Msg("failed to cleanly shutdown webserver")
	}

	log.Info().Msg("Webserver shutdown cleanly")
}
