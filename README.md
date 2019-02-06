# Crust subscription server

### Setup

Copy `.env.example` to `.env` and make proper modifications for your 
local environment.

Please check the options available with `go run cmd/cli/*.go -h`.

### Running in local environment for development

Everything should be set and ready to run with `make realize`. This
utilizes realize tool that monitors codebase for changes and restarts
api http server for every file change. 

### Making changes

Please refer to each project's style guidelines and guidelines for submitting patches and additions.
In general, we follow the "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes
