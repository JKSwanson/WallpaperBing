package main

import (
	"WallpaperBing/modules"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func checkerColor(x, y, _, _ int) color.Color {
	xr := x / 10
	yr := y / 10

	if xr%2 == yr%2 {
		return color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff}
	} else {
		return color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xff}
	}
}

func (imgObj *ImagesWallpaperObject) makeRow(id int, file string) fyne.CanvasObject {

	filename := filepath.Base(file)
	button := widget.NewButton(filename, func() {
		imgObj.chooseImage(id)
	})

	preview := canvas.NewImageFromFile(file)
	iconHeight := button.MinSize().Height
	preview.SetMinSize(fyne.NewSize(iconHeight*3.2, iconHeight*1.5))

	return container.New(
		layout.NewBorderLayout(nil, nil, preview, nil),
		preview, button)
}

func (imgObj *ImagesWallpaperObject) makeSidebarList() []fyne.CanvasObject {
	var items []fyne.CanvasObject

	for id, name := range imgObj.SliceOfNamesImage {
		items = append(items, imgObj.makeRow(id, imgObj.MapImages[name].Filepath))
	}

	return items
}

type toolbarLabel struct {
}

func (t *toolbarLabel) ToolbarObject() fyne.CanvasObject {
	label = widget.NewLabel("filename")
	return label
}

const (
	WIDTH  = 980
	HEIGHT = 680
)

var (
	image              *canvas.Image
	label              *widget.Label
	descr              *widget.Label
	setWallpaperButton *widget.Button
	slideShowStarted   bool
)

func (imgObj *ImagesWallpaperObject) previousImage() {
	if imgObj.currentId == 0 {
		return
	}
	imgObj.chooseImage(imgObj.currentId - 1)
}

func (imgObj *ImagesWallpaperObject) nextImage() {
	if imgObj.currentId == imgObj.lenSet-1 {
		return
	}

	imgObj.chooseImage(imgObj.currentId + 1)
}

func (imgObj *ImagesWallpaperObject) chooseImage(id int) {
	path := imgObj.MapImages[imgObj.SliceOfNamesImage[id]].Filepath
	descrStr := imgObj.MapImages[imgObj.SliceOfNamesImage[id]].Description
	label.SetText(filepath.Base(path))
	descr.SetText(descrStr)
	image.File = path
	setWallpaperButton.OnTapped = func() {
		if slideShowStarted {
			slideShowStarted = false
			dialog.ShowInformation("Information", "Slideshow stopped!", imgObj.win)
		}
		e := modules.SetWallpaper(
			imgObj.MapImages[imgObj.SliceOfNamesImage[imgObj.currentId]].Filepath)
		if e != nil {
			fmt.Println(e)
			dialog.ShowError(e, imgObj.win)

		} else {
			dialog.ShowInformation("Information", "Your wallpaper is now set. Go check it!", imgObj.win)
		}
	}
	canvas.Refresh(image)
	imgObj.currentId = id
}

type ImagesWallpaperObject struct {
	MapImages         map[string]*modules.ImageParameters
	SliceOfNamesImage []string
	lenSet            int
	currentId         int
	win               fyne.Window
}

func SlideShowFunc(slideShow *bool, imgObj *ImagesWallpaperObject, timeInSec *int) {
	id := imgObj.currentId
	for *slideShow {
		id++
		e := modules.SetWallpaper(
			imgObj.MapImages[imgObj.SliceOfNamesImage[id]].Filepath)
		if e != nil {
			fmt.Println(e)
			return
		}
		time.Sleep(time.Duration(*timeInSec) * time.Second)
	}
	fmt.Println("Slideshow stopped!")
}

func main() {
	imgParameters := &map[string]*modules.ImageParameters{}

	usr, _ := user.Current()
	dirpath := filepath.Join(usr.HomeDir, "ImagesBing")
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		_ = os.Mkdir(dirpath, 0755)
		fmt.Println("Directory created")
	}

	jsonFileName := filepath.Join(dirpath, "Cache.json")

	if _, err := os.Stat(jsonFileName); err == nil {

		jsBytesFromFile, e := ioutil.ReadFile(jsonFileName)

		if e != nil {
			fmt.Println(e)
			return
		}

		e = json.Unmarshal(jsBytesFromFile, &imgParameters)

		if e != nil {
			fmt.Println(e)
		}
	}
	e := modules.GetImageXML(imgParameters)
	if e != nil {
		fmt.Println(e)
	}
	var imgWallSet = &ImagesWallpaperObject{
		MapImages: *imgParameters,
		lenSet:    len(*imgParameters),
		currentId: 0,
	}
	imgWallSet.SliceOfNamesImage = make([]string, 0, imgWallSet.lenSet)

	for name, value := range imgWallSet.MapImages {
		filename, e := modules.DownloadImage(value.Url, name, dirpath)
		imgWallSet.SliceOfNamesImage = append(imgWallSet.SliceOfNamesImage, name)
		if exist := imgWallSet.MapImages[name].Filepath; exist == "" {
			imgWallSet.MapImages[name].Filepath = filename
		}
		if e != nil {
			if e.Error() == "File already exists" {
				continue
			} else {
				fmt.Println(e)
				return
			}
		}
	}

	jsByte, _ := json.MarshalIndent(imgParameters, "", "   ")

	_ = ioutil.WriteFile(jsonFileName, jsByte, 0644)
	sort.Sort(sort.Reverse(sort.StringSlice(imgWallSet.SliceOfNamesImage)))

	imageApp := app.New()
	imgWallSet.win = imageApp.NewWindow("GoImages")

	imageApp.Settings().SetTheme(theme.DarkTheme())

	vlist := imgWallSet.makeSidebarList()

	navBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.NavigateBackIcon(), imgWallSet.previousImage),
		widget.NewToolbarSpacer(),
		&toolbarLabel{},
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.NavigateNextIcon(), imgWallSet.nextImage))

	fileList := container.NewVScroll(container.NewVBox(vlist...))

	checkers := canvas.NewRasterWithPixels(checkerColor)

	image = canvas.NewImageFromFile(
		imgWallSet.MapImages[imgWallSet.SliceOfNamesImage[imgWallSet.currentId]].Filepath)
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(WIDTH*0.5, HEIGHT*0.5))

	descr = widget.NewLabel(
		imgWallSet.MapImages[imgWallSet.SliceOfNamesImage[imgWallSet.currentId]].Description)
	descr.Alignment = fyne.TextAlignCenter
	descr.Wrapping = fyne.TextWrapWord

	setWallpaperButton = widget.NewButton("Set Wallpaper", func() {

	})
	var timeSlideShowInSec int
	slideShowStarted = false
	inputTime := widget.NewSelect(
		[]string{"30 sec", "1 min", "5 min", "10 min"},
		func(s string) {
			timeSlideShowInSec, _ = strconv.Atoi(strings.Split(s, " ")[0])
			if timeSlideShowInSec != 30 {
				timeSlideShowInSec *= 60
			}
			fmt.Println("selected", s)
		})

	slideShowButtonStart := widget.NewButton(
		"Start Slideshow",
		func() {
			if timeSlideShowInSec != 0 {
				if !slideShowStarted {
					slideShowStarted = true
					go SlideShowFunc(&slideShowStarted, imgWallSet, &timeSlideShowInSec)
				} else {
					dialog.ShowInformation("Information", "Slideshow already started!", imgWallSet.win)
				}
			} else {
				dialog.ShowInformation("Information", "Select time", imgWallSet.win)
			}
		},
	)
	slideShowButtonStop := widget.NewButton(
		"Stop Slideshow",
		func() {
			if slideShowStarted {
				slideShowStarted = false
				dialog.ShowInformation("Information", "Slideshow stopped!", imgWallSet.win)
			}
		},
	)
	imgWallSet.chooseImage(0)

	slideShowContainer := container.New(
		layout.NewBorderLayout(nil, nil, nil, nil),
		canvas.NewRectangle(color.RGBA{R: 0x15, G: 0x15, B: 0x15, A: 1}),
		container.NewCenter(container.NewHBox(
			inputTime, slideShowButtonStart, slideShowButtonStop,
		)),
	)
	imgDescrAndWallButton := container.New(
		layout.NewBorderLayout(descr, setWallpaperButton, nil, nil),
		descr, container.NewCenter(setWallpaperButton),
	)
	imgButtonLayout := container.New(
		layout.NewBorderLayout(nil, imgDescrAndWallButton, nil, nil),
		checkers, image, imgDescrAndWallButton,
	)

	cont := container.New(
		layout.NewBorderLayout(navBar, slideShowContainer, fileList, nil),
		navBar, slideShowContainer, fileList, imgButtonLayout,
	)
	imgWallSet.win.SetContent(cont)
	imgWallSet.win.Resize(fyne.NewSize(WIDTH, HEIGHT))

	imgWallSet.win.ShowAndRun()

}
