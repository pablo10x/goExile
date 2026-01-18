package utils

import (
	"fmt"
	"net/http"

	"github.com/pterm/pterm"
)

// PrintStatus writes a basic server ready message to the response.
func PrintStatus(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Exile Master Server Ready\n")
}

// PrintBanner displays the server title.
func PrintBanner() {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("EXILE MASTER SERVER")
}

// PrintSection displays a stylized section header.
func PrintSection(title string, status string, isSuccess bool) {
	if isSuccess {
		pterm.Success.Println(title)
	} else {
		pterm.Warning.Println(title)
	}
}

// PrintSubItem displays a nested item.
func PrintSubItem(text string) {
	pterm.Info.WithPrefix(pterm.Prefix{
		Text:  "└",
		Style: pterm.NewStyle(pterm.FgGray),
	}).Println(text)
}

// PrintStartupComplete displays the final server ready status with API endpoints using Pterm.
func PrintStartupComplete(port string) {
	pterm.Println() // Spacer

	// Define data for the panel
	apiLink := fmt.Sprintf("http://127.0.0.1:%s", port)
	healthLink := fmt.Sprintf("http://127.0.0.1:%s/health", port)
	statsLink := fmt.Sprintf("http://127.0.0.1:%s/api/stats", port)

	// Create a bullet list for endpoints
	endpoints := pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: "API Access", TextStyle: pterm.NewStyle(pterm.FgCyan), Bullet: "•"},
		{Level: 1, Text: apiLink, TextStyle: pterm.NewStyle(pterm.FgLightWhite)},
		{Level: 0, Text: "Health Monitor", TextStyle: pterm.NewStyle(pterm.FgCyan), Bullet: "•"},
		{Level: 1, Text: healthLink, TextStyle: pterm.NewStyle(pterm.FgLightWhite)},
		{Level: 0, Text: "Metrics Feed", TextStyle: pterm.NewStyle(pterm.FgCyan), Bullet: "•"},
		{Level: 1, Text: statsLink, TextStyle: pterm.NewStyle(pterm.FgLightWhite)},
	})

	content, _ := endpoints.Srender()

	// Wrap in a box
	box := pterm.DefaultBox.
		WithTitle("SYSTEM ONLINE").
		WithTitleBottomRight().
		WithRightPadding(10).
		WithLeftPadding(2).
		WithTopPadding(1).
		WithBottomPadding(1).
		Sprint(content)

	pterm.Println(box)
	
pterm.Info.Println("Press Ctrl+C to stop the server")
}