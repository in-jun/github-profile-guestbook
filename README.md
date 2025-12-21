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

- **Automatic theme switching** - Adapts to light/dark mode automatically
- Leave comments on any GitHub profile (200 characters max)
- React with likes and dislikes
- Highlight your favorite comments with a star
- Clean, minimal design that fits any profile
- Fast and lightweight

## Example

[![Example](https://github-comment.injun.dev/api/user/in-jun/svg)](https://github-comment.injun.dev/in-jun)

*The widget automatically adapts to light/dark mode based on your system settings.*

## Tech Stack

**Backend:** Go, Gin, PostgreSQL, JWT authentication

**Frontend:** Vanilla HTML/CSS/JavaScript

**Deployment:** Docker, injunweb

## License

MIT
