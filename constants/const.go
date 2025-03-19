package constants

import (
	"fmt"

	"github.com/mbndr/figlet4go"
)

const info string = `
	Maintainer	: prashant.nandipati@zigram.tech
	Version		: v0.1b
`

func PrintTitle() {
	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render("KubePromPush")
	fmt.Println(renderStr)
	fmt.Print(info)
}
