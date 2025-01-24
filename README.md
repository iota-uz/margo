# Margo CMS

Margo is a Markdown-driven Content Management System (CMS) built with Go, TailwindCSS, and Templ. It leverages file-based routing and provides developers with a robust, scalable, and intuitive framework for building high-performance websites. Margo is a part of the larger **IOTA SDK ecosystem**, a modular ERP solution designed for diverse industries like finance, manufacturing, and warehouse management.

## Key Features

- **Markdown-Driven Content**: Create and manage content in Markdown files, enabling a simple and effective way to write and update content.
- **File-Based Routing**: Automatically maps files to routes, making navigation structure intuitive and developer-friendly.
- **Go-Powered**: Benefit from the performance, concurrency, and security of the Go programming language.
- **TailwindCSS Integration**: Use TailwindCSS to style your projects with ease, ensuring a modern and consistent design system.
- **Templ Components**: Use Templ for fast, modular, and reusable UI components.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/dl/) (v1.20 or later)
- [Node.js](https://nodejs.org/en/download/) (for TailwindCSS and related tooling)
- [templ](https://github.com/a-h/templ) (install via `go install github.com/a-h/templ/cmd/templ@latest`)
- **Margo**: Install Margo via:

```bash
go get github.com/iota-uz/margo
```

## Getting Started

### 1. Initialize the Project

Create a new Go project and import Margo:

```bash
mkdir my-margo-project
cd my-margo-project
go mod init github.com/your-username/my-margo-project
```

Install dependencies:

```bash
go get github.com/iota-uz/margo
npm install
```

### 2. Register Components

Margo components are provided via `layouts` and `registry`. Example setup:

```go
package ui

import (
	"github.com/iota-uz/margo/registry"

	"github.com/username/your-project/components/button"
	"github.com/username/your-project/components/primitives"
	"github.com/username/your-project/components/typography"
)

func BlogLayout() registry.Layout {
	layout := registry.NewLayout("Blog")

	// typography
	layout.Register("h1", typography.H1)
	layout.Register("h2", typography.H2)
	layout.Register("h3", typography.H3)
	layout.Register("h4", typography.H4)
	layout.Register("h5", typography.H5)
	layout.Register("h6", typography.H6)

	// primitives
	layout.Register("a", primitives.Link)
	layout.Register("ul", primitives.List)
	layout.Register("li", primitives.Item)
	layout.Register("img", primitives.Image)

	// buttons
	layout.Register("ButtonPrimary", button.Primary)
	layout.Register("ButtonSecondary", button.Secondary)
	layout.Register("ButtonDanger", button.Danger)

	return layout
}

func Registry() registry.Registry {
	r := registry.New()
	r.RegisterLayout(BlogLayout())
	return r
}
```

### 3. Run the Development Server

Generate Templ files:

```bash
templ generate
```

Start the Go server:

```bash
go run .
```

Run Tailwind in development mode:

```bash
npx tailwindcss -i ./assets/styles.css -o ./public/styles.css --watch
```

Open your browser and navigate to `http://localhost:8080`.

## Project Structure

```plaintext
my-margo-project/
├── assets/         # TailwindCSS input files
├── components/     # Reusable Templ components
├── content/        # Markdown files for pages
├── layouts/        # Layout templates
├── public/         # Generated CSS and static assets
├── routes/         # File-based routing logic
├── main.go         # Application entry point
└── tailwind.config.js # TailwindCSS configuration
```

## Creating Content

Add Markdown files in the `content/` directory. File names map directly to routes. For example:

- `content/index.md` -> `/`
- `content/about.md` -> `/about`

### Frontmatter

Use YAML frontmatter for metadata:

```markdown
---
title: "About Us"
description: "Learn more about our mission and values."
---

Welcome to our website!
```

## File-Based Routing

Routes are automatically generated based on the directory structure in `content/`. Customize routing logic in the `routes/` package if needed.

## Extending Margo

### Adding Components

Create reusable Templ components in the `components/` directory. For example:

```go
package components

import (
	"github.com/iota-uz/margo/registry"
	"github.com/a-h/templ"
)

func ButtonPrimary(props struct {
	Text string
	Href string
}) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<a href="` + props.Href + `" class="btn-primary">` + props.Text + `</a>`))
		return err
	})
}

func Register(layout registry.Layout) {
	layout.Register("ButtonPrimary", ButtonPrimary)
}
```

Use the component in your templates:

```templ
@components.ButtonPrimary({
	Text: "Learn More",
	Href: "/about",
})
```

### Styling with TailwindCSS

Define your styles in `assets/styles.css` and extend Tailwind in `tailwind.config.js` as needed.

### Custom Logic

Add custom middleware, handlers, or utilities in `main.go` or the `routes/` package.

## Deployment

1. Build the Go application:

```bash
go build -o margo
```

2. Compile TailwindCSS for production:

```bash
npx tailwindcss -i ./assets/styles.css -o ./public/styles.css --minify
```

3. Deploy the `margo` binary and the `public/` directory to your server or hosting provider.

## Testing

### Run Unit Tests

Use the Go testing framework:

```bash
go test ./...
```

### Linting and Formatting

Format your Go code:

```bash
go fmt ./...
```

Lint your code:

```bash
golangci-lint run
```

## Community and Support

- **Discussions**: Join the conversation on [GitHub Discussions](https://github.com/iota-uz/margo/discussions).
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/iota-uz/margo/issues).
- **IOTA SDK**: Explore the larger ecosystem at [IOTA SDK](https://github.com/iota-uz/iota-sdk).

## License

Margo is licensed under the [MIT License](LICENSE).

---

Happy coding with Margo!

