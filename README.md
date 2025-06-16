# GitHub Profile Comments

Add a comment system to your GitHub profile README.

## Usage

### 1. Add to your profile README

```markdown
[![Comments](https://github-comment.injun.dev/api/user/YOUR_USERNAME/svg)](https://github-comment.injun.dev/YOUR_USERNAME)
```

Replace `YOUR_USERNAME` with your GitHub username.

### 2. Login

Visit `https://github-comment.injun.dev/YOUR_USERNAME` and log in with GitHub OAuth.

### 3. Leave a comment

Write a comment (max 200 characters) on any profile.

## Features

- **Automatic dark mode detection** - Adapts to your system theme via CSS `prefers-color-scheme`
- Real-time comment system
- Like/Dislike functionality
- Owner can highlight favorite comments with a star
- Minimal design with sharp edges and monochrome colors
- System native fonts for consistent rendering
- Responsive layout
- Optimized SVG generation with `strings.Builder`

## Example

[![Example](https://github-comment.injun.dev/api/user/in-jun/svg)](https://github-comment.injun.dev/in-jun)

*The widget automatically adapts to light/dark mode based on your system settings.*

## Tech Stack

**Backend:** Go, Gin, PostgreSQL, JWT authentication

**Frontend:** Vanilla HTML/CSS/JavaScript

**Deployment:** Docker, injunweb

## License

MIT
