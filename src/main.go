package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/sqweek/dialog"
)

type AppState struct {
    theme   *material.Theme
    openExcelButton  widget.Clickable
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
    var ops op.Ops
    state := newAppState()
    var gtx layout.Context

    for {

        select{
        case path := <-state.excelPathChan:
            if path == ""{
                state.excelPath = "File selction cancelled or not file selected"
            } else if len(path) > 7 && path[:7] == "Error: "{
                state.excelPath = path
            }else {
                state.excelPath = path
                state.statusMessage = "Excel file selected. Ready for next step"
            }
            window.Invalidate()
        default:
        }

        switch e:= window.Event().(type) {
        case app.DestroyEvent:
            return e.Err
        case app.FrameEvent:
            gtx = app.NewContext(&ops, e)                 

            if state.openExcelButton.Clicked(gtx){

                state.statusMessage = "Opening file dialog"

                window.Invalidate()

                go func(){
                    filepath, err := dialog.File().Filter("Excel Files", "xlsx", "xls").Load()

                    if err != nil {
                        if err == dialog.ErrCancelled {
                            log.Println("File Selection cancelled")
                            state.excelPathChan <- ""
                        }else {
                            log.Println("Error Selecting file: ", err)
                            state.excelPathChan <- "Error: " + err.Error()
                        }
                        return
                    }
                    log.Println("Selected File: ", filepath)
                    state.excelPathChan <- filepath
                }()
            }
                layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceSides, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H4(state.theme, "Image Mover Utility")
					title.Color = color.NRGBA{R: 0, G: 80, B: 127, A: 255}
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Button(state.theme, &state.openExcelButton, "1. Select Excel File").Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Display the selected path or status
					pathLabel := material.Body1(state.theme, "Excel Path: "+state.excelPath)
					return pathLabel.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					statusLabel := material.Body2(state.theme, state.statusMessage)
					return statusLabel.Layout(gtx)
				}),

				// TODO: Add buttons and displays for Image Directory and Destination Path
				// TODO: Add a "Move Images" button and trigger your moveFiles logic
			)
			// --- End UI Layout ---

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
