package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	endpointURL  = flag.String("endpoint", "https://mulled.github.io/api/v1/", "API Endpoint")
	showBuilders = flag.Bool("builders", false, "Shows available builders and exits")
)

type image struct {
	Image       string `json:"image"`
	Packager    string `json:"packager"`
	Homepage    string `json:"homepage"`
	Description string `json:"description"`
	Versions    []struct {
		Version  string `json:"version"`
		Revision string `json:"revision"`
		Size     string `json:"size"`
		Date     string `json:"date"`
	}
}

type images []image

func (i images) Len() int {
	return len(i)
}

func (i images) Less(x, y int) bool {
	a, b := i[x], i[y]

	return a.Packager < b.Packager ||
		(a.Packager == b.Packager && a.Image < b.Image)
}

func (i images) Swap(x, y int) {
	i[x], i[y] = i[y], i[x]
}

type builder struct {
	Name               string `json:"name"`
	Homepage           string `json:"homepage"`
	ExplicitVersioning bool   `json:"explicitVersioning",omitempty`
}

func main() {
	flag.Parse()

	switch {

	case *showBuilders:
		loadAndShowBuilders()
	case flag.NArg() == 0:
		loadAndShowAllImages()
	default:
		for _, s := range flag.Args() {
			loadAndShowImage(s)
		}
	}
}

func loadAndShowBuilders() {
	w := tabwriter.NewWriter(os.Stdout, 15, 0, 4, ' ', 0)
	defer w.Flush()

	resp, err := http.Get(*endpointURL + "builders.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var builders []builder
	if err := json.NewDecoder(resp.Body).Decode(&builders); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Fprintln(w, "NAME\tHOMEPAGE\tEXPLICIT VERSIONING")

	for _, el := range builders {
		var evMessage string
		if el.ExplicitVersioning {
			evMessage = "YES"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", el.Name, el.Homepage, evMessage)
	}
}
func loadAndShowImage(img string) {
	url := fmt.Sprintf("%simages/%s.json", *endpointURL, img)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var message image
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("\n\n%s\n%s\n\n", strings.ToUpper(message.Image), strings.Repeat("=", len(message.Image)))

	fmt.Printf("%s\n\n%s\n\n", message.Homepage, message.Description)

	w := tabwriter.NewWriter(os.Stdout, 15, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REVISION\tVERSION\tSIZE\tDATE")

	for _, el := range message.Versions {
		t, err := time.Parse(time.RFC3339, el.Date)
		if err != nil {
			fmt.Fprintf(w, "%s\n", err)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", el.Revision, el.Version, el.Size, t.Format(time.Stamp))
	}
	w.Flush()
	fmt.Printf("\n\n")
}

func loadAndShowAllImages() {
	w := tabwriter.NewWriter(os.Stdout, 15, 0, 2, ' ', 0)

	resp, err := http.Get(*endpointURL + "images.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var message images
	if err := dec.Decode(&message); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Fprintln(w, "PACKAGER\tNAME\tREVISION\tVERSION")

	sort.Sort(message)
	for _, el := range message {
		for _, v := range el.Versions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", el.Packager, el.Image, v.Revision, v.Version)
		}
	}
	w.Flush()
}
