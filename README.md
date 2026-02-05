# GitHub Profile Guestbook

Add an interactive guestbook to your GitHub profile README.

## Usage

### 1. Add to your profile README

```markdown
[![Guestbook](https://github-profile-guestbook.injun.dev/api/user/YOUR_USERNAME/svg)](https://github-profile-guestbook.injun.dev/YOUR_USERNAME)
```

Replace `YOUR_USERNAME` with your GitHub username.

### 2. Login

Visit `https://github-profile-guestbook.injun.dev/YOUR_USERNAME` and log in with GitHub OAuth.

### 3. Leave a message

Write a message (max 200 characters) on any profile.

## Features

- **Automatic theme switching** - Adapts to light/dark mode automatically
- Leave messages on any GitHub profile (200 characters max)
- React with likes and dislikes
- Highlight your favorite messages with a star
- Clean, minimal design that fits any profile
- Fast and lightweight

## Example

[![Example](https://github-profile-guestbook.injun.dev/api/user/in-jun/svg)](https://github-profile-guestbook.injun.dev/in-jun)

*The widget automatically adapts to light/dark mode based on your system settings.*

## Tech Stack

**Backend:** Go, Gin, PostgreSQL, JWT authentication

**Frontend:** Vanilla HTML/CSS/JavaScript

**Deployment:** Docker, [INJUNWEB](https://github.com/injunweb)

## License

MIT

