# BytePort

Infrastructure-as-code (IAC) deployment and UX generation platform for software developer portfolios. Reads an IAC file defining application structure, deploys from GitHub to AWS, and auto-generates portfolio site templates via LLM.

## Stack
- Language: Go (backend), TypeScript/JavaScript (frontend)
- Key deps: AWS SDK, GitHub API, OpenAI/LLM client
- Structure: `backend/` + `frontend/` monorepo

## Structure
- `backend/`: Go server handling IAC parsing, AWS deployment, LLM integration
- `frontend/`: Portfolio site frontend (TypeScript)
- `start` / `start.bat`: Dev startup scripts

## Key Patterns
- IAC-driven: single config file defines application + infra
- AWS deployment (EC2/ECS/Lambda depending on app type)
- LLM-generated portfolio widgets and object templates
- GitHub repo as the source of truth for app code

## Adding New Functionality
- Backend logic: `backend/` (Go modules)
- Frontend components: `frontend/`
- New IAC resource types: extend the IAC parser in `backend/`
- Run `./start` for local development
