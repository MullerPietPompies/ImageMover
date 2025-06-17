package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type AppState struct {
    theme   *material.Theme
    openExcelutton  widget.Clickable
    excelPath string
    excelPathChan chan string
    statusMessage string
}

func newAppState() *AppState{
    return &AppState{
        theme: material.NewTheme(),
        excelPath: "No Excel file selected",
        excelPathChan: make(chan string, 1),
        statusMessage: "Click Button to select excel file",

    }
}

func main() {

    go func ()  {
        window := new(app.Window)
        err := run(window)
        if err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()
    app.Main()

	fmt.Println("Welcome to the Image Mover Util!")
	fmt.Println("--------------------------------")

	fmt.Println("Enter The Excel File Path: ")

	var path string
	fmt.Scanln(&path)
	imageList := getImageList(path)

	var imageDir string
	fmt.Println("Select image directory")
	fmt.Scanln(&imageDir)

	var destinationPath string
	fmt.Println("Choose Destination Folder: ")
	fmt.Scanln(&destinationPath)

	fmt.Printf("Moving Images! \n")
	moveFiles(imageList, imageDir, destinationPath)

	fmt.Println("----------------------")
	fmt.Println("Thank you for using this utility")
}

func run(window *app.Window) error {
    theme := material.NewTheme()
    var ops op.Ops

    for {
        switch e := window.Event().(type){
        case app.DestroyEvent:
            return e.Err
        case app.FrameEvent:
            gtx := app.NewContext(&ops, e)
            title := material.H1(theme, "Hello there")
            maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
            title.Color = maroon
            title.Alignment = text.Middle
            
            title.Layout(gtx)
            drawRect(&ops)
            e.Frame(gtx.Ops)
            
        
        }
    }
}

func drawRect(ops *op.Ops){
    defer clip.Rect{Max: image.Pt(100,100)}.Push(ops).Pop()
    paint.ColorOp{Color: color.NRGBA{R: 0x80, A: 0xFF}}.Add(ops)
    paint.PaintOp{}.Add(ops)
}

func drawRedRect10PixelsRight(ops *op.Ops) {
	defer op.Offset(image.Pt(100, 0)).Push(ops).Pop()
	drawRect(ops)
}
