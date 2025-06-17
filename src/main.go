package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
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
    excelPathSelecte bool

    openImageDirButton widget.Clickable
    imageDirPath string
    imageDirPathChan chan string
    imageDirSelected bool

    openDestFolderButton widget.Clickable
    destPath string
    destPathChan chan string
    destPathSelected bool

    moveImgsButton widget.Clickable
    statusMessage string
    longOpChannel chan string
}

func newAppState() *AppState{
    return &AppState{
        theme: material.NewTheme(),
        
        excelPath: "No Excel file selected",
        excelPathChan: make(chan string, 1),

        imageDirPath: "No image Directory selected",
        imageDirPathChan: make(chan string, 1),

        destPath: "No destination path selected",
        destPathChan: make(chan string, 1),

        statusMessage:  "Please select all paths",
        longOpChannel: make(chan string, 1),

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
        case path := <-state.imageDirPathChan:
            state.imageDirSelected = false
            if path == "" {
                state.imageDirPath = "Directory selection cancelled "
            } else if len(path) > 7 && path[:7] == "Error: "{
                state.imageDirPath = path
            } else{
                state.imageDirPath = path
                state.imageDirSelected = true
                state.statusMessage  = "Image direcctory selected"
            }
        window.Invalidate()
        case path := <-state.destPathChan:
            state.destPathSelected = false
            if path ==""{
                state.destPath = "Destination path selection cancelled"
            } else if len(path)> 7 && path[:7] == "Error: "{
                state.destPath = path
            } else {
                state.destPath = path
                state.destPathSelected = true
                state.statusMessage = "Destination folder selected"
            }
        window.Invalidate()
        case opStatus := <-state.longOpChannel:
            state.statusMessage = opStatus
            window.Invalidate()
        default:
        }

        switch e := window.Event().(type) {
        case app.DestroyEvent:
            return e.Err
        case app.FrameEvent:
            gtx := app.NewContext(&ops, e)

            if state.openExcelButton.Clicked(gtx){
                state.statusMessage= "Opening Excel file"
                window.Invalidate()
                go func() {
                    filepath, err := dialog.File().Title("Select Excel File").Filter("Excel Files", "xlsx","xls").Load()
                    handleDialogResult(filepath, err, state.excelPathChan, "Excel File Selection")
                    state.excelPathSelecte = true
                }()
            }

            if state.openImageDirButton.Clicked(gtx){
                state.statusMessage="Opening image directory"
                window.Invalidate()

                go func(){

                    dirPath, err := dialog.Directory().Title("Select img Directory").Browse()
                    handleDialogResult(dirPath, err, state.imageDirPathChan, "Selecting Image Directory")
                    state.imageDirSelected = true

                }()
            }
            if state.openDestFolderButton.Clicked(gtx){
                state.statusMessage="Opening Destination Folder"
                window.Invalidate()

                go func(){
                    destPath, err := dialog.Directory().Title("Select Destination Folder").Browse()
                    handleDialogResult(destPath, err, state.destPathChan, "Selecting Destination Folder")
                    state.destPathSelected = true
                }()
            }

            if state.moveImgsButton.Clicked(gtx){
                if !state.excelPathSelecte || !state.imageDirSelected || ! state.destPathSelected{
                    state.statusMessage = "Error: Please select all valid paths bevore moving images"
                } else{
                    state.statusMessage = "Proccessing... Getting image list..."
                    window.Invalidate()

                    go func(excelP, imgDirP, destP string){
                        imageList := getImageList(excelP)
                        if imageList == nil {
                            state.longOpChannel<- "Error: Could not get image list"
                            return
                        }
                        if len(imageList) == 0{
                            state.longOpChannel<-"No image references found in the Excel File"
                            return
                        }
                        moveFiles(imageList, imgDirP, destP)
                    }(state.excelPath, state.imageDirPath, state.destPath)
                }
            }
                layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H4(state.theme, "Image Mover Utility")
					title.Color = color.NRGBA{R: 0, G: 80, B: 127, A: 255}
					title.Alignment = text.Middle
					return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, title.Layout)
				}),
				layout.Rigid(rowWithLabelAndButton(state.theme, "1. Excel File:", state.excelPath, &state.openExcelButton, "Select...")),
				layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
				layout.Rigid(rowWithLabelAndButton(state.theme, "2. Image Directory:", state.imageDirPath, &state.openImageDirButton, "Select...")),
				layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
                layout.Rigid(rowWithLabelAndButton(state.theme, "3. Destination Direcctory", state.destPath, &state.openDestFolderButton, "Select...")),
                layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),
                layout.Rigid(func(gtx layout.Context) layout.Dimensions{
                    btn:= material.Button(state.theme, &state.moveImgsButton, "Move Images")
                    return layout.Center.Layout(gtx, btn.Layout)
                }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					statusLabel := material.Body1(state.theme, state.statusMessage)
					statusLabel.Alignment = text.Middle
					return layout.Center.Layout(gtx, statusLabel.Layout)
				}),
			)			// --- End UI Layout ---

			e.Frame(gtx.Ops)
        }

    }
}


func handleDialogResult(pathOrDir string, err error, resultChan chan string, logContext string){
    if err != nil{
        if err == dialog.ErrCancelled {
            log.Println(logContext, "cancelled")
            resultChan <- ""
        } else {
            log.Println("Error during ", logContext, ":", err)
            resultChan <- "Error: " + err.Error()
        }
        return
    }
    if pathOrDir == ""{
        log.Println(logContext, "returned empty path with no error")
        resultChan <- ""
        return
    }
    log.Println(logContext, "selected: ", pathOrDir)
    resultChan <- pathOrDir
}

func rowWithLabelAndButton(th *material.Theme, descText, pathText string, btn *widget.Clickable, btnText string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(th, descText).Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
					layout.Flexed(1, material.Body2(th, pathText).Layout), // Path text takes available space
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(material.Button(th, btn, btnText).Layout),
				)
			}),
		)
	}
}
