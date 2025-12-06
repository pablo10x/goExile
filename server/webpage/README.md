# Game Server Registry Dashboard

Modern web dashboard for the Game Server Registry built with Tailwind CSS and Chart.js.

## Directory Structure

```
webpage/
├── src/
│   └── input.css          # Tailwind CSS input file
├── dist/
│   └── output.css         # Generated Tailwind CSS output
├── dashboard.html         # Main dashboard page
├── login.html            # Login page
├── package.json          # NPM configuration with Tailwind scripts
├── tailwind.config.js    # Tailwind CSS configuration
├── postcss.config.js     # PostCSS configuration
└── README.md             # This file
```

## Setup

### Prerequisites
- Node.js (v14 or higher)
- npm

### Installation

1. Navigate to the webpage directory:
```bash
cd server/webpage
```

2. Install dependencies:
```bash
npm install
```

3. Build Tailwind CSS:
```bash
npm run build:css
```

Or watch for changes during development:
```bash
npm run watch:css
```

## Development Workflow

1. **Edit HTML**: Modify `dashboard.html` or `login.html`
2. **Watch CSS**: Run `npm run watch:css` to auto-rebuild Tailwind CSS
3. **Test**: Open `http://localhost:8081` in your browser (when server is running)

## Build for Production

The Tailwind CSS output is automatically generated when you run:
```bash
npm run build:css
```

The compiled CSS file `dist/output.css` is then referenced in your HTML files.

## Tailwind CSS Customization

Edit `tailwind.config.js` to customize:
- Colors
- Fonts
- Breakpoints
- Plugins
- And more

## Important Notes

- The Go server serves static files from the `webpage/` directory
- Make sure to build Tailwind CSS before deploying
- The `dist/output.css` file should be included in your HTML via `<link rel="stylesheet" href="/dist/output.css">`

## Links

- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Chart.js Documentation](https://www.chartjs.org/)
