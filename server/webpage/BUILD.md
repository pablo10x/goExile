# Webpage Build Instructions

When building the server, ensure Tailwind CSS is compiled.

## Steps

1. **Install dependencies** (one-time setup):
   ```bash
   cd server/webpage
   npm install
   ```

2. **Build Tailwind CSS** (before `go build`):
   ```bash
   npm run build:css
   ```

3. **Build the Go server**:
   ```bash
   go build
   ```

4. **Run the server**:
   ```bash
   ./server.exe
   ```

## Full Build Script (Windows)

```bash
cd server\webpage
npm install
npm run build:css
cd ..
go build
.\server.exe
```

## Full Build Script (Linux/Mac)

```bash
cd server/webpage
npm install
npm run build:css
cd ..
go build
./server
```

## Development Workflow

To make the development easier, you can run Tailwind CSS in watch mode:

```bash
cd server/webpage
npm run watch:css
```

This will automatically recompile CSS whenever you make changes to HTML files.
