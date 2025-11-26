# Commit Formatting Instructions

- **Message format**. Follow the 50/72 rule:

    - First line: Maximum 50 characters (but can extend to ~60 if
      necessary), starting with an emoji from the list below, plus a
      single space.
    - Blank line after the first line.
    - Body text: Wrap at 72 characters.

- **Emoji usage**. Start each commit message with the most specific
  emoji from this list:

    - Work in progress: 🚧
    - Code generation: 🎉
    - New feature: ✨
    - Bugfix: 🐛
    - Documentation: 📚
    - Documenting source code: ✏
    - Performance: 🐎
    - Design/CSS: 🎨
    - Adding/modifying a test: ✅
    - Improvement: ⚡
    - Refactor code: 🔨
    - Removing code/files: 🗑
    - Uncertain change: 🙏
    - Continuous Integration: 👷
    - Enforce guidelines: 👮
    - Security: 🔒
    - General dependency modification: ❗
    - Removing a dependency: ➖
    - Adding a dependency: ➕
    - Upgrading dependencies: ⬆
    - Downgrading dependencies: ⬇
    - Linting: 🎀
    - Translation: 🌐
    - Critical hotfix: 🚑
    - Deployment: 🚀
    - Git stuff: 🐙
    - Fixing on MacOS: 🍎
    - Fixing on Linux: 🐧
    - Fixing on Windows: 🖼
    - Analytics, tracking code, or telemetry (logs, traces, metrics): 📈
    - Docker: 🐳
    - Configuration files: 🔧
    - Package: 📦
    - Dangerous code necessary: 🦂
    - Reverting changes: ⏪
    - Breaking changes: 💥
    - Code review changes: 👌
    - Accessibility: ♿
    - Move/copy/rename files: 🚜

- Run this specific command to see real examples from this repository:

    ```bash
    git log -n 30 --pretty=format:"%s%n---" main
    ```
