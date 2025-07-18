package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/sqweek/dialog"
)

var (
    darkTheme *material.Theme
    lightTheme *material.Theme
)

type AppState struct {
    theme   *material.Theme

    openBlueprintFolder  widget.Clickable
    blueprintPath string
    blueprintPathChan chan string
    blueprintSelected bool

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

    themeToggle widget.Clickable
    darkMode bool

}

func newAppState() *AppState{
    return &AppState{
        theme: material.NewTheme(),
        
        blueprintPath: "No Blueprint selected",
        blueprintPathChan: make(chan string, 1),

        imageDirPath: "No image Directory selected",
        imageDirPathChan: make(chan string, 1),

        destPath: "No destination path selected",
        destPathChan: make(chan string, 1),

        statusMessage:  "Please select all paths",
        longOpChannel: make(chan string, 1),

        darkMode: false,
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

    lightTheme = material.NewTheme()
    darkTheme = material.NewTheme()
    darkTheme.Palette = material.Palette{
        Bg:     color.NRGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xFF},
        Fg:     color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, 
        ContrastBg: color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xFF},
        ContrastFg: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
    }
    
    for {

        select{
        case path := <-state.blueprintPathChan:
            if path == ""{
                state.blueprintPath = "File selction cancelled or not file selected"
            } else if len(path) > 7 && path[:7] == "Error: "{
                state.blueprintPath = path
            }else {
                state.blueprintPath = path
                state.statusMessage = "Blueprint folder selected. Ready for next step"
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
            
            if state.themeToggle.Clicked(gtx) {
                state.darkMode  = !state.darkMode
            }

            th := lightTheme

            if state.darkMode {
                th = darkTheme
            }

            if state.openBlueprintFolder.Clicked(gtx){
                state.statusMessage= "Opening blueprint directory"
                window.Invalidate()
                go func() {
                    filepath, err := dialog.Directory().Title("Select Blueprint Directory").Browse()
                    handleDialogResult(filepath, err, state.blueprintPathChan, "Blueprint Folder Selection")
                    state.blueprintSelected = true
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
                if !state.blueprintSelected || !state.imageDirSelected || ! state.destPathSelected{
                    state.statusMessage = "Error: Please select all valid paths bevore moving images"
                } else{
                    state.statusMessage = "Proccessing... Finding files"
                    window.Invalidate()

                    go func(blueprint, imgDirP, destP string){
                        replicateBlueprintFromSource(blueprint, imgDirP, destP)
                        state.statusMessage = "Finished!"
                    }(state.blueprintPath, state.imageDirPath, state.destPath)
                    state.statusMessage = "Moving Images";
                }
            }
            
                bg := th.Bg
                paint.Fill(gtx.Ops, bg)
            
                layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H4(th, "Image Mover Utility")
					title.Color = color.NRGBA{R: 0, G: 80, B: 127, A: 255}
					title.Alignment = text.Middle
					return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, title.Layout)
				}),
				layout.Rigid(rowWithLabelAndButton(th, "1. Blueprint Directory:", state.blueprintPath, &state.openBlueprintFolder, "Select...")),
				layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
				layout.Rigid(rowWithLabelAndButton(th, "2. Image Directory:", state.imageDirPath, &state.openImageDirButton, "Select...")),
				layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
                layout.Rigid(rowWithLabelAndButton(th, "3. Destination Direcctory", state.destPath, &state.openDestFolderButton, "Select...")),
                layout.Rigid(layout.Spacer{Height: unit.Dp(25)}.Layout),
                layout.Rigid(func(gtx layout.Context) layout.Dimensions{
                    btn:= material.Button(th, &state.moveImgsButton, "Move Images")
                    return layout.Center.Layout(gtx, btn.Layout)
                }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					statusLabel := material.Body1(th, state.statusMessage)
					statusLabel.Alignment = text.Middle
					return layout.Center.Layout(gtx, statusLabel.Layout)
				}),
                layout.Rigid(func(gtx layout.Context) layout.Dimensions{
                    btn:= material.Button(th, &state.themeToggle, "Theme")
                    return layout.Center.Layout(gtx, btn.Layout)
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
