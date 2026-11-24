---
name: versioning
description: Use when the user asks about creating or choosing a version, Git tag, or release.
---

# Versioning and Releases

When asked about a version, tag, or release:

1. Analyze the changes since the previous version.
2. Recommend a Semantic Versioning level and explain it briefly: major for incompatible changes, minor for backward-compatible functionality, or patch for backward-compatible fixes (including bug fixes).
3. Suggest a version and a descriptive tag message that starts with `Release vX.Y.Z`, replacing `X.Y.Z` with the recommended version.
4. Before creating a version tag, check the current branch. Only create a version tag on `main`; if on another branch, do not create it.
5. Never push a version tag automatically.
6. Provide the commands to push and delete the tag:
   - Push: `git push origin vX.Y.Z`
   - Delete locally: `git tag -d vX.Y.Z`
   - If the tag has been pushed, delete it remotely with `git push origin --delete vX.Y.Z`
