# GitHub Actions Workflow

This project uses GitHub Actions to automatically build and publish Docker images to GitHub Container Registry (ghcr.io).

## Workflow Overview

The workflow (`/.github/workflows/docker.yml`) automatically:

- **Builds** the Docker image on pushes to main/master, pull requests, and tags
- **Publishes** to GitHub Container Registry (`ghcr.io`)
- **Tags** images with:
  - Branch name for branch pushes
  - PR number for pull requests
  - Semantic version tags (v1.0.0, v1.0, v1) for version tags
  - `latest` tag for the default branch
- **Caches** build layers for faster builds
- **Supports** multi-platform builds (linux/amd64, linux/arm64)

## Usage

### Pulling the Image

After the workflow runs, pull the image using:

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull the image
docker pull ghcr.io/vibes/draft-board:latest
```

Replace `vibes/draft-board` with your actual repository path.

### Using Specific Tags

```bash
# Latest from main branch
docker pull ghcr.io/vibes/draft-board:latest

# Specific version
docker pull ghcr.io/vibes/draft-board:v1.0.0

# From a specific branch
docker pull ghcr.io/vibes/draft-board:feature-branch
```

### Running the Container

```bash
docker run -d \
  --name draft-board \
  -p 8080:8080 \
  ghcr.io/vibes/draft-board:latest
```

## Version Tags

To create a versioned release:

1. Create and push a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The workflow will automatically build and tag:
   - `ghcr.io/vibes/draft-board:v1.0.0`
   - `ghcr.io/vibes/draft-board:v1.0`
   - `ghcr.io/vibes/draft-board:v1`

## Permissions

The workflow uses `GITHUB_TOKEN` which is automatically provided by GitHub Actions. Make sure your repository has:

- **Packages: write** permission (enabled by default)
- **Contents: read** permission (enabled by default)

## Manual Trigger

You can manually trigger the workflow from the GitHub Actions tab:

1. Go to **Actions** tab
2. Select **Build and Publish Docker Image**
3. Click **Run workflow**

## Viewing Published Packages

View your published images at:
```
https://github.com/vibes/draft-board/pkgs/container/draft-board
```

(Replace `vibes/draft-board` with your repository path)

