# GitHub Profile Comments

Add a comment system to your GitHub profile README.

## Usage

### 1. Add to your profile README

```markdown
[![Comments](https://github-comment.injun.dev/api/user/YOUR_USERNAME/svg?theme=white)](https://github-comment.injun.dev/YOUR_USERNAME)
```

Replace `YOUR_USERNAME` with your GitHub username.

### 2. Login

Visit `https://github-comment.injun.dev/YOUR_USERNAME` and log in with GitHub OAuth.

### 3. Leave a comment

Write a comment (max 200 characters) on any profile.

## Themes

Change the appearance by modifying the `theme` parameter:

- `theme=white` - Light background (default)
- `theme=black` - Dark background
- `theme=transparent` - Transparent background

Example:
```markdown
[![Comments](https://github-comment.injun.dev/api/user/in-jun/svg?theme=black)](https://github-comment.injun.dev/in-jun)
```

## Features

- Real-time comment system
- Like/Dislike functionality
- Owner can highlight favorite comments with a star
- Minimal design with sharp edges and monochrome colors
- Pretendard font
- Responsive layout

## Examples

| Theme | Preview |
|-------|---------|
| White | [![Example](https://github-comment.injun.dev/api/user/in-jun/svg?theme=white)](https://github-comment.injun.dev/in-jun) |
| Black | [![Example](https://github-comment.injun.dev/api/user/in-jun/svg?theme=black)](https://github-comment.injun.dev/in-jun) |
| Transparent | [![Example](https://github-comment.injun.dev/api/user/in-jun/svg?theme=transparent)](https://github-comment.injun.dev/in-jun) |

## Tech Stack

**Backend:** Go, Gin, PostgreSQL, JWT authentication

**Frontend:** Vanilla HTML/CSS/JavaScript

**Deployment:** Docker, injunweb

## License

MIT
