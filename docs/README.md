# Mining Documentation

Welcome to the documentation for the Mining project. This folder contains detailed information about the API, CLI, architecture, and development processes.

## GitHub Pages Documentation

The full documentation is built with [MkDocs](https://www.mkdocs.org/) and [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme, and is available at:

**https://snider.github.io/Mining/**

## Local Development

### Prerequisites

- Python 3.x
- pip

### Setup & Serve

1. Install dependencies:

```bash
pip install -r docs/requirements.txt
```

2. Serve the documentation locally:

```bash
mkdocs serve
```

The documentation will be available at `http://127.0.0.1:8000/`

### Building

To build the static site:

```bash
mkdocs build
```

The built site will be in the `site/` directory.

## Legacy Documentation

- [**API Documentation**](API.md): Detailed information about the RESTful API endpoints, request/response formats, and Swagger usage.
- [**CLI Documentation**](CLI.md): A comprehensive guide to the Command Line Interface, including command descriptions and usage examples.
- [**Architecture Guide**](ARCHITECTURE.md): An overview of the project's design, including the modular `ManagerInterface`, core packages, and data flow.
- [**Development Guide**](DEVELOPMENT.md): Instructions for contributors on how to build, test, and release the project.

## Project Structure

```
docs/
├── index.md                    # Home page
├── getting-started/            # Getting started guides
├── cli/                        # CLI command reference
├── api/                        # API documentation
├── dashboard/                  # Web dashboard docs
├── desktop/                    # Desktop application docs
├── development/                # Development guides
├── architecture/               # Architecture documentation
├── pools/                      # Pool integration docs
├── miners/                     # Miner-specific documentation
├── troubleshooting/            # Troubleshooting guides
├── stylesheets/
│   └── extra.css              # Custom CSS
└── requirements.txt           # Python dependencies
```

## Writing Documentation

### Markdown Extensions

This project uses PyMdown Extensions which provide additional features:

- **Admonitions**: `!!! note`, `!!! warning`, `!!! tip`, etc.
- **Code blocks**: Syntax highlighting with line numbers
- **Tabs**: Tabbed content blocks
- **Task lists**: GitHub-style checkboxes
- **Emojis**: `:smile:`
- **Mermaid diagrams**: Flow charts and diagrams

### Example Admonition

```markdown
!!! tip "Mining Tip"
    Make sure to check your GPU temperature regularly!
```

### Example Code Block

````markdown
```go title="main.go" linenums="1" hl_lines="2 3"
package main

import "fmt"

func main() {
    fmt.Println("Hello, Mining!")
}
```
````

### Example Tabbed Content

```markdown
=== "Linux"
    ```bash
    ./miner-ctrl serve
    ```

=== "Windows"
    ```powershell
    miner-ctrl.exe serve
    ```

=== "macOS"
    ```bash
    ./miner-ctrl serve
    ```
```

## Contributing

When adding new documentation:

1. Create markdown files in the appropriate directory
2. Add the new page to the `nav:` section in `mkdocs.yml`
3. Follow the existing style and structure
4. Test locally with `mkdocs serve`
5. Submit a pull request

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch. The deployment is handled by the `.github/workflows/docs.yml` workflow.

## Resources

- [MkDocs Documentation](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [PyMdown Extensions](https://facelessuser.github.io/pymdown-extensions/)
- [Mermaid Diagrams](https://mermaid-js.github.io/mermaid/)
