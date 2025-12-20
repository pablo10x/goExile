package utils

import (
	"fmt"
	"net/http"
)

// ANSI color codes for terminal styling
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBgBlue  = "\033[44m"
)

func PrintStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Exile Master Server Ready\n")
}

func PrintBanner() {
	banner := colorCyan + colorBold + `
    ╔══════╗╚═╗  ╚═╝╚═╝╚══════╗╚══════╗
    ║  ╔═══╝ ╚╗╔╝  ║  ║  ╔════╝║  ╔═══╝
    ║  ╚═══╗  ╚╝   ║  ║  ║     ║  ╚═══╗
    ║  ╔═══╝  ╔╝╚╗ ║  ║  ║     ║  ╔═══╝
    ║  ╚═══╗ ╔╝  ╚╗║  ║  ╚════╗║  ╚═══╗
    ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝` + colorReset + `
` + colorDim + `    ─────────────────────────────────────` + colorReset + `
` + colorWhite + `       Master Server ` + colorCyan + `v1.0.0` + colorReset + colorDim + ` │ ` + colorGreen + `Ready` + colorReset + `
` + colorDim + `    ─────────────────────────────────────` + colorReset + `
`
	fmt.Println(banner)
}

func PrintSection(title string, status string, isSuccess bool) {
	statusColor := colorGreen
	statusIcon := "✓"
	if !isSuccess {
		statusColor = colorYellow
		statusIcon = "○"
	}
	fmt.Printf("  %s%s%s %s%s%s\n", statusColor, statusIcon, colorReset, colorWhite, title, colorReset)
}

func PrintSubItem(text string) {
	fmt.Printf("    %s└─%s %s%s%s\n", colorDim, colorReset, colorDim, text, colorReset)
}

func PrintStartupComplete(port string) {
	fmt.Println()
	fmt.Printf("  %s%s▸ Server Ready%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("  %s────────────────────────────────────────%s\n", colorDim, colorReset)
	fmt.Printf("  %s●%s API      %shttp://localhost:%s%s\n", colorGreen, colorReset, colorCyan, port, colorReset)
	fmt.Printf("  %s●%s Health   %shttp://localhost:%s/health%s\n", colorGreen, colorReset, colorCyan, port, colorReset)
	fmt.Printf("  %s●%s Stats    %shttp://localhost:%s/api/stats%s\n", colorGreen, colorReset, colorCyan, port, colorReset)
	fmt.Println()
	fmt.Printf("  %sPress Ctrl+C to stop%s\n", colorDim, colorReset)
	fmt.Printf("  %s────────────────────────────────────────%s\n", colorDim, colorReset)
	fmt.Println()
}
