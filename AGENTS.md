# AGENTS.md yt-transcribe  

> **purpose** – This file is the onboarding manual for every AI assistant (Claude, Cursor, GPT, etc.) and every human who edits this repository.

---

## 0. Project overview

yt-transcribe is a Go application heavily utilizing the OpenAI API for audio transcription and summarization. It servers HTML templates that interact with an async worker running in a goroutine. Key components:

- **cmd/runserver.go**: Entry point for the web server, defines routes and starts the background worker.
- **cmd/transcribe.go**: Entry point for the CLI transcription command.
- **internal/ai/**: Contains OpenAI-related functionality, including audio transcription and summarization.
- **internal/fetch/youtube.go**: Handles downloading YouTube videos and extracting audio using yt-dlp.
- **internal/http/server.go**: Implements the HTTP server and handlers for the web interface.
- **internal/http/worker.go**: Implements a background worker that processes transcription requests asynchronously.
- **internal/http/templates/**: Contains HTML templates for the web interface.

**Golden rule**: When unsure about implementation details or requirements, ALWAYS consult the developer rather than making assumptions.

---

## 1. Build, test & utility commands

You can use the regular Go commands to build, test, format and run the application. Below is the list of commands we expose to the user and you can use to test the application:

```bash
go run main.go runserver    # Start the development server
go run main.go transcribe   # Transcribe a YouTube video by its URL
go run main.go version      # Show the current version of the application
```

---

## 2. Coding standards

*  **Golang**: Use `gofmt` for formatting and `go vet` for static analysis. 
*  **Formatting**: Standard formatting rules as defined by `gofmt`.
*  **Docker**: Whenever adding new dependencies that depend on system utilities, ensure they are included in the `Dockerfile`.

---

## 3. Project layout & Core Components

| Directory         | Description                                                         |
| ----------------- | ------------------------------------------------------------------- |
| `src/components/` | React components                                                    |
| `src/lib/`        | Utilities, classes and logic that can be re-used between components |
| `src/routes/`     | React router routes                                                 |

---

## 4. Anchor comments

Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that we can easily `grep` for.

### Guidelines:

- Use `AIDEV-NOTE:`, `AIDEV-TODO:`, or `AIDEV-QUESTION:` (all-caps prefix) for comments aimed at AI and developers.
- Keep them concise (≤ 120 chars).
- **Important:** Before scanning files, always first try to **locate existing anchors** `AIDEV-*` in relevant subdirectories.
- **Update relevant anchors** when modifying associated code.
- **Do not remove `AIDEV-NOTE`s** without explicit human instruction.
- Make sure to add relevant anchor comments, whenever a file or piece of code is:
  * too long, or
  * too complex, or
  * very important, or
  * confusing, or
  * could have a bug unrelated to the task you are currently working on.
```

---

## 5. Common pitfalls

---

## 6. Domain-Specific Terminology

---

## AI Assistant Workflow: Step-by-Step Methodology

When responding to user instructions, the AI assistant (Claude, Cursor, GPT, etc.) should follow this process to ensure clarity, correctness, and maintainability:

1. **Consult Relevant Guidance**: When the user gives an instruction, consult the relevant instructions from `AGENTS.md` file.
2. **Clarify Ambiguities**: Based on what you could gather, see if there's any need for clarifications. If so, ask the user targeted questions before proceeding.
3. **Break Down & Plan**: Break down the task at hand and chalk out a rough plan for carrying it out, referencing project conventions and best practices.
4. **Trivial Tasks**: If the plan/request is trivial, go ahead and get started immediately.
5. **Non-Trivial Tasks**: Otherwise, present the plan to the user for review and iterate based on their feedback.
6. **Track Progress**: Use a to-do list (internally, or optionally in a `TODOS.md` file) to keep track of your progress on multi-step or complex tasks.
7. **If Stuck, Re-plan**: If you get stuck or blocked, return to step 3 to re-evaluate and adjust your plan.
8. **Update Documentation**: Once the user's request is fulfilled, update relevant anchor comments (`AIDEV-NOTE`, etc.) in the files you touched and the `AGENTS.md` file.
9. **User Review**: After completing the task, ask the user to review what you've done, and repeat the process as needed.
10. **Session Boundaries**: If the user's request isn't directly related to the current context and can be safely started in a fresh session, suggest starting from scratch to avoid context confusion.
