package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
)

func main() {

	r := regexp.MustCompile(`\d+x\d+`)
	if (len(os.Args) == 1) || (r.FindString(os.Args[1]) == "") {
		fmt.Println(`Angiv venligst en opløsning. Eksempel: "./laernyeord.exe 1920x1080"`)
		os.Exit(1)
	}

	stringResolution := strings.SplitN(os.Args[1], "x", -1)

	x, _ := strconv.ParseInt(stringResolution[0], 10, 32)
	y, _ := strconv.ParseInt(stringResolution[1], 10, 32)

	hRes := int(x)
	vRes := int(y)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Angiv ord: ")
		ord, _ := reader.ReadString('\n')

		r := regexp.MustCompile(`[\w\-\'æøå ]{2,}`)

		ord = r.FindString(ord)

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		response, err := client.Get("https://ordnet.dk/ddo/ordbog?query=" + ord)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		doc, _ := goquery.NewDocumentFromReader(response.Body)

		var citater []string
		doc.Find(".citat").Each(func(i int, s *goquery.Selection) {
			if unicode.IsUpper(rune(s.Text()[0])) {
				citater = append(citater, s.Text())
			}
		})

		if len(citater) > 5 {
			citater = citater[:5]
		}

		if len(citater) == 0 {

			var forslag []string
			doc.Find("#more-alike-list-short li a").Each(func(i int, s *goquery.Selection) {
				forslag = append(forslag, s.Text())
			})

			forslagString := `Ingen resultater for "` + ord + `". Forslag: `

			if len(forslag) > 5 {
				forslag = forslag[:5]
			}

			if len(forslag) == 0 {
				forslagString = "Kunne ikke finde nogle citater."
			} else {
				for _, f := range forslag {
					forslagString += f + ", "
				}
				forslagString = forslagString[:len(forslagString)-2]
			}

			fmt.Println(forslagString)

		} else {

			font, err := truetype.Parse(gobold.TTF)
			if err != nil {
				log.Fatal(err)
			}

			face := truetype.NewFace(font, &truetype.Options{Size: float64(hRes / 80)})

			start := hRes / 16
			step := hRes / 48

			dc := gg.NewContext(hRes, vRes)
			dc.SetFontFace(face)
			dc.SetRGB(0, 0, 0)
			dc.Clear()
			dc.SetRGB(1, 1, 1)
			dc.DrawStringAnchored("-"+ord, float64(step), float64(start+step), 0, 1)

			font, err = truetype.Parse(goitalic.TTF)
			face = truetype.NewFace(font, &truetype.Options{Size: float64(hRes / 100)})
			dc.SetFontFace(face)

			for i, citat := range citater {
				endDot := ""
				if e := string(citat[len(citat)-1]); !((e == "!") || (e == ".") || (e == "?")) {
					endDot = "."
				}
				dc.DrawStringAnchored(`"`+citat+endDot+`"`, float64(step), float64(start+int(2.3*float64(step))+i*step), 0, 1)
			}
			dc.SavePNG(ord + ".png")
		}
	}
}
