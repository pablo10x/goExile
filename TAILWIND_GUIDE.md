# Quick Development Guide

## Your Setup is Ready! ✅

### What we've done:
1. ✅ Added recent dashboard changes to git
2. ✅ Set up Tailwind CSS with proper configuration
3. ✅ Reorganized webpage directory with clean structure
4. ✅ Created build scripts for easy CSS compilation
5. ✅ Verified the server builds and runs successfully

### Directory Structure:
```
server/webpage/
├── src/
│   └── input.css              # Your Tailwind CSS source (edit this)
├── dist/
│   └── output.css             # Generated CSS (don't edit directly)
├── dashboard.html             # Main dashboard
├── login.html                 # Login page
├── package.json               # NPM configuration
├── tailwind.config.js         # Tailwind theme customization
├── postcss.config.js          # PostCSS setup
└── BUILD.md                   # Build instructions
```

## Development Workflow:

### 1. Watch CSS Changes (Recommended):
```bash
cd server/webpage
npm run watch:css
```
This watches for changes and auto-compiles. Keep this running while developing.

### 2. Add New Classes/Components:
- Edit `src/input.css` for custom styles
- Use Tailwind classes directly in HTML
- The CSS auto-rebuilds thanks to watch mode

### 3. Customize Theme:
- Edit `tailwind.config.js` to change colors, fonts, breakpoints
- Example: Add a custom color, spacing, etc.
- Run `npm run build:css` to apply changes

### 4. Build for Production:
```bash
cd server/webpage
npm run build:css
cd ..
go build
./server.exe
```

## Tailwind CSS Cheat Sheet:

### Common Classes:
```html
<!-- Layout -->
<div class="flex justify-center items-center gap-4">
  <!-- Flex container -->
</div>

<!-- Colors -->
<p class="text-slate-100 bg-slate-900">Light text on dark background</p>

<!-- Spacing -->
<div class="p-6 m-4 gap-8">Padding and margin</div>

<!-- Typography -->
<h1 class="text-2xl font-bold text-slate-100">Heading</h1>

<!-- Responsive -->
<div class="w-full md:w-1/2 lg:w-1/3">Responsive widths</div>

<!-- Hover States -->
<button class="bg-blue-500 hover:bg-blue-600">Hover button</button>
```

## Next Steps:

1. **Modify `dashboard.html`**: Replace inline styles with Tailwind classes
2. **Update `login.html`**: Apply Tailwind styling
3. **Add components**: Use classes from `src/input.css`
4. **Test**: Run server and check `http://localhost:8081`

## Troubleshooting:

**CSS not updating?**
- Make sure `npm run watch:css` is running
- Check that `dashboard.html` links to `/dist/output.css`
- Rebuild: `npm run build:css`

**Server not serving CSS?**
- Verify the Go server has `webpage/` directory access
- Check browser DevTools → Network tab for CSS requests
- Ensure CSS file exists: `ls server/webpage/dist/output.css`

## Resources:
- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Tailwind Color Palette](https://tailwindcss.com/docs/customizing-colors)
- [Tailwind Components](https://tailwindcss.com/docs/reusing-styles)
