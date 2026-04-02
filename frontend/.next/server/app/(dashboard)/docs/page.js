(()=>{var e={};e.id=722,e.ids=[722],e.modules={47849:e=>{"use strict";e.exports=require("next/dist/client/components/action-async-storage.external")},72934:e=>{"use strict";e.exports=require("next/dist/client/components/action-async-storage.external.js")},55403:e=>{"use strict";e.exports=require("next/dist/client/components/request-async-storage.external")},54580:e=>{"use strict";e.exports=require("next/dist/client/components/request-async-storage.external.js")},94749:e=>{"use strict";e.exports=require("next/dist/client/components/static-generation-async-storage.external")},45869:e=>{"use strict";e.exports=require("next/dist/client/components/static-generation-async-storage.external.js")},20399:e=>{"use strict";e.exports=require("next/dist/compiled/next-server/app-page.runtime.prod.js")},72254:e=>{"use strict";e.exports=require("node:buffer")},6005:e=>{"use strict";e.exports=require("node:crypto")},15673:e=>{"use strict";e.exports=require("node:events")},88849:e=>{"use strict";e.exports=require("node:http")},22286:e=>{"use strict";e.exports=require("node:https")},47261:e=>{"use strict";e.exports=require("node:util")},90686:(e,t,r)=>{"use strict";r.r(t),r.d(t,{GlobalError:()=>n.a,__next_app__:()=>m,originalPathname:()=>p,pages:()=>d,routeModule:()=>u,tree:()=>c}),r(23562),r(68032),r(82053),r(26083),r(96560);var o=r(23191),s=r(88716),a=r(37922),n=r.n(a),i=r(95231),l={};for(let e in i)0>["default","tree","pages","GlobalError","originalPathname","__next_app__","routeModule"].indexOf(e)&&(l[e]=()=>i[e]);r.d(t,l);let c=["",{children:["(dashboard)",{children:["docs",{children:["__PAGE__",{},{page:[()=>Promise.resolve().then(r.bind(r,23562)),"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/(dashboard)/docs/page.tsx"]}]},{}]},{layout:[()=>Promise.resolve().then(r.bind(r,68032)),"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/(dashboard)/layout.tsx"]}]},{layout:[()=>Promise.resolve().then(r.bind(r,82053)),"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/layout.tsx"],error:[()=>Promise.resolve().then(r.bind(r,26083)),"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/error.tsx"],"not-found":[()=>Promise.resolve().then(r.bind(r,96560)),"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/not-found.tsx"]}],d=["/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/(dashboard)/docs/page.tsx"],p="/(dashboard)/docs/page",m={require:r,loadChunk:()=>Promise.resolve()},u=new o.AppPageRouteModule({definition:{kind:s.x.APP_PAGE,page:"/(dashboard)/docs/page",pathname:"/docs",bundlePath:"",filename:"",appPaths:[]},userland:{loaderTree:c}})},84279:(e,t,r)=>{Promise.resolve().then(r.bind(r,12816))},12816:(e,t,r)=>{"use strict";r.r(t),r.d(t,{default:()=>C});var o=r(10326),s=r(17577),a=r(33104),n=r(33071),i=r(54432),l=r(90772),c=r(567),d=r(65685),p=r(3634),m=r(56625),u=r(45691),y=r(92498),h=r(32130),g=r(88378),x=r(58038),f=r(7027),b=r(88307),v=r(39183);/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let j=(0,r(62881).Z)("BookOpen",[["path",{d:"M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z",key:"vv98re"}],["path",{d:"M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z",key:"1cyq3y"}]]);var w=r(66307),P=r(43810);let N=[{id:"quick-start",title:"Quick Start Guide",icon:p.Z,category:"getting-started",description:"Get up and running with BytePort in minutes",content:`# Quick Start Guide

Welcome to BytePort! This guide will help you deploy your first application in just a few minutes.

## Prerequisites

- A Git repository with your application code
- A BytePort account (sign up at byteport.dev)
- Basic knowledge of your application's tech stack

## Step 1: Connect Your Repository

1. Navigate to the Projects page
2. Click "New Project"
3. Select your Git provider (GitHub, GitLab, or Custom)
4. Enter your repository URL
5. Choose your default branch (usually 'main' or 'master')

## Step 2: Configure Your Project

1. Give your project a descriptive name
2. Add a brief description (optional)
3. Select your framework/runtime
4. Upload an NVMS configuration file (optional)
5. Enable or disable automatic deployments

## Step 3: Deploy

1. Click "Create Project" to finalize setup
2. Navigate to your project's detail page
3. Click "Deploy Now" to trigger your first deployment
4. Monitor the deployment progress in real-time

## Next Steps

- Set up environment variables for your application
- Configure custom domains
- Set up monitoring and alerts
- Explore advanced deployment options`,examples:[{title:"Example NVMS Configuration",language:"yaml",code:`name: my-app
version: 1.0.0
runtime: nodejs-18
build:
  command: npm run build
  output: dist
deploy:
  type: static
  provider: aws
  region: us-east-1
environment:
  NODE_ENV: production`}]},{id:"projects",title:"Managing Projects",icon:m.Z,category:"guides",description:"Learn how to create and manage your projects",content:`# Managing Projects

Projects in BytePort represent your applications and their deployment configurations.

## Creating a Project

Projects are created through a three-step wizard:

### Step 1: Repository Connection
- Choose your Git provider
- Provide the repository URL
- Select the default branch

### Step 2: Configuration
- Set project name and description
- Choose framework/runtime
- Upload NVMS configuration (optional)
- Configure auto-deployment settings

### Step 3: Review
- Verify all settings
- Create the project

## Project Settings

### Environment Variables
Manage environment variables that will be injected into your deployments:
- Add/remove variables
- Mark variables as secret
- Copy values securely

### Git Integration
- View connection status
- Update repository URL
- Reconnect if needed

### Deployment Settings
- Configure build commands
- Set deployment triggers
- Manage deployment branches

## Archiving and Deleting Projects

### Archive
- Temporarily disable a project
- Keep all data and history
- Can be restored later

### Delete
- Permanently remove a project
- Deletes all associated deployments
- Cannot be undone`,examples:[{title:"Environment Variables Setup",language:"bash",code:`# Set environment variables via CLI
byteport env set DATABASE_URL="postgresql://..."
byteport env set API_KEY="sk-..." --secret

# List all environment variables
byteport env list

# Remove an environment variable
byteport env unset API_KEY`}]},{id:"deployments",title:"Deployments",icon:u.Z,category:"guides",description:"Deploy and manage your applications",content:`# Deployments

Deployments are instances of your application running on BytePort infrastructure.

## Triggering Deployments

### Manual Deployment
1. Navigate to your project
2. Click "Deploy Now"
3. Select branch and commit (optional)
4. Monitor deployment progress

### Automatic Deployment
Enable automatic deployments to deploy on every push:
- Configure in project settings
- Specify branches to watch
- Set up deployment hooks

## Deployment Lifecycle

1. **Building** - Source code is fetched and built
2. **Deploying** - Application is deployed to infrastructure
3. **Running** - Application is live and serving traffic
4. **Failed** - Deployment encountered an error
5. **Terminated** - Deployment has been stopped

## Monitoring Deployments

### Real-time Logs
- View build and runtime logs
- Filter by severity
- Search log content
- Export logs for analysis

### Metrics
- CPU usage
- Memory consumption
- Network traffic
- Request rates
- Error rates

### Alerts
Set up alerts for:
- Deployment failures
- High resource usage
- Error rate thresholds
- Downtime incidents`,examples:[{title:"Deploy via CLI",language:"bash",code:`# Deploy latest commit from main branch
byteport deploy --project=my-app

# Deploy specific commit
byteport deploy --project=my-app --commit=abc123

# Deploy with environment override
byteport deploy --project=my-app --env=staging

# Watch deployment progress
byteport deploy --project=my-app --follow`}]},{id:"api-reference",title:"API Reference",icon:y.Z,category:"api",description:"Complete API documentation for BytePort",content:`# API Reference

The BytePort API provides programmatic access to all platform features.

## Authentication

All API requests require authentication via API key:

\`\`\`
Authorization: Bearer YOUR_API_KEY
\`\`\`

Get your API key from Settings > API Keys.

## Base URL

\`\`\`
https://api.byteport.dev/v1
\`\`\`

## Rate Limits

- 1000 requests per hour for standard plans
- 5000 requests per hour for pro plans
- Rate limit headers included in responses

## Projects API

### List Projects
\`\`\`
GET /projects
\`\`\`

### Create Project
\`\`\`
POST /projects
\`\`\`

### Get Project
\`\`\`
GET /projects/:id
\`\`\`

### Update Project
\`\`\`
PATCH /projects/:id
\`\`\`

### Delete Project
\`\`\`
DELETE /projects/:id
\`\`\`

## Deployments API

### List Deployments
\`\`\`
GET /deployments
\`\`\`

### Create Deployment
\`\`\`
POST /deployments
\`\`\`

### Get Deployment
\`\`\`
GET /deployments/:id
\`\`\`

### Get Deployment Logs
\`\`\`
GET /deployments/:id/logs
\`\`\`

### Get Deployment Metrics
\`\`\`
GET /deployments/:id/metrics
\`\`\``,examples:[{title:"Create Project via API",language:"javascript",code:`const response = await fetch('https://api.byteport.dev/v1/projects', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'My App',
    repository_url: 'https://github.com/user/repo',
    branch: 'main',
    framework: 'nextjs',
    auto_deploy: true
  })
});

const project = await response.json();
console.log('Project created:', project.id);`},{title:"List Deployments",language:"python",code:`import requests

headers = {
    'Authorization': 'Bearer YOUR_API_KEY'
}

response = requests.get(
    'https://api.byteport.dev/v1/deployments',
    headers=headers,
    params={'project_id': 'proj_123', 'limit': 10}
)

deployments = response.json()
for deployment in deployments['data']:
    print(f"{deployment['id']}: {deployment['status']}")`}]},{id:"cli-reference",title:"CLI Reference",icon:h.Z,category:"api",description:"Command-line interface documentation",content:`# CLI Reference

The BytePort CLI provides a powerful command-line interface for managing your deployments.

## Installation

\`\`\`bash
# Using npm
npm install -g @byteport/cli

# Using yarn
yarn global add @byteport/cli

# Using homebrew (macOS)
brew install byteport
\`\`\`

## Authentication

\`\`\`bash
# Login interactively
byteport login

# Login with API key
byteport login --api-key YOUR_API_KEY
\`\`\`

## Commands

### Projects

\`\`\`bash
# List all projects
byteport projects list

# Create a new project
byteport projects create

# View project details
byteport projects info <project-id>

# Delete a project
byteport projects delete <project-id>
\`\`\`

### Deployments

\`\`\`bash
# Deploy a project
byteport deploy --project=<project-id>

# List deployments
byteport deployments list

# View deployment logs
byteport logs <deployment-id>

# View deployment metrics
byteport metrics <deployment-id>

# Stop a deployment
byteport deployments stop <deployment-id>
\`\`\`

### Environment Variables

\`\`\`bash
# List environment variables
byteport env list --project=<project-id>

# Set an environment variable
byteport env set KEY=value --project=<project-id>

# Set a secret environment variable
byteport env set KEY=value --secret --project=<project-id>

# Delete an environment variable
byteport env unset KEY --project=<project-id>
\`\`\``,examples:[{title:"Complete Deployment Workflow",language:"bash",code:`# Login to BytePort
byteport login

# Create a new project
byteport projects create \\
  --name "My App" \\
  --repo "https://github.com/user/repo" \\
  --branch "main"

# Set environment variables
byteport env set DATABASE_URL="postgresql://..." --secret
byteport env set NODE_ENV="production"

# Deploy the project
byteport deploy --follow

# View logs
byteport logs --tail 100

# Check metrics
byteport metrics --duration 1h`}]},{id:"nvms-config",title:"NVMS Configuration",icon:g.Z,category:"advanced",description:"Advanced configuration with NVMS files",content:`# NVMS Configuration

NVMS (Network Virtual Machine Specification) files provide advanced configuration for your deployments.

## File Format

NVMS files can be written in YAML or JSON format.

## Structure

\`\`\`yaml
name: string          # Project name
version: string       # Configuration version
runtime: string       # Runtime environment
build:
  command: string     # Build command
  output: string      # Build output directory
  env: object        # Build environment variables
deploy:
  type: string       # Deployment type (static, server, container)
  provider: string   # Cloud provider
  region: string     # Deployment region
  instances: number  # Number of instances
  scaling:
    min: number      # Minimum instances
    max: number      # Maximum instances
    target_cpu: number # Target CPU percentage
environment:
  object             # Runtime environment variables
health:
  path: string       # Health check endpoint
  interval: number   # Check interval in seconds
  timeout: number    # Timeout in seconds
\`\`\`

## Examples

### Static Site
\`\`\`yaml
name: my-static-site
version: 1.0.0
runtime: static
build:
  command: npm run build
  output: dist
deploy:
  type: static
  provider: aws
  region: us-east-1
\`\`\`

### Node.js Server
\`\`\`yaml
name: my-api
version: 1.0.0
runtime: nodejs-18
build:
  command: npm run build
  output: dist
deploy:
  type: server
  provider: aws
  region: us-east-1
  instances: 2
  scaling:
    min: 1
    max: 5
    target_cpu: 70
environment:
  NODE_ENV: production
  PORT: 3000
health:
  path: /health
  interval: 30
  timeout: 5
\`\`\``,examples:[{title:"Full NVMS Configuration Example",language:"yaml",code:`name: production-api
version: 2.0.0
runtime: nodejs-18

build:
  command: npm ci && npm run build
  output: dist
  env:
    NODE_ENV: production

deploy:
  type: server
  provider: aws
  region: us-east-1
  instances: 3
  scaling:
    min: 2
    max: 10
    target_cpu: 75
    target_memory: 80

environment:
  NODE_ENV: production
  PORT: 3000
  LOG_LEVEL: info

health:
  path: /api/health
  interval: 30
  timeout: 5
  healthy_threshold: 2
  unhealthy_threshold: 3

networking:
  ports:
    - 3000
  domains:
    - api.example.com
  ssl: true

resources:
  cpu: 2
  memory: 4096
  disk: 20480

monitoring:
  enabled: true
  metrics:
    - cpu
    - memory
    - requests
  alerts:
    - type: error_rate
      threshold: 5
      duration: 300`}]},{id:"security",title:"Security Best Practices",icon:x.Z,category:"advanced",description:"Secure your deployments and data",content:`# Security Best Practices

Follow these guidelines to keep your deployments secure.

## API Keys

- Never commit API keys to source control
- Rotate keys regularly (every 90 days)
- Use separate keys for different environments
- Revoke unused keys immediately
- Use read-only keys when possible

## Environment Variables

- Mark sensitive values as secrets
- Never log secret values
- Use separate variables for each environment
- Rotate secrets regularly
- Audit who has access to secrets

## Network Security

- Enable HTTPS for all deployments
- Use SSL certificates from trusted providers
- Configure CORS policies appropriately
- Implement rate limiting
- Use IP allowlisting when possible

## Access Control

- Follow principle of least privilege
- Use role-based access control
- Enable multi-factor authentication
- Review access logs regularly
- Revoke access for departed team members

## Compliance

- Enable audit logging
- Regular security reviews
- Keep dependencies updated
- Follow industry standards (SOC2, GDPR, etc.)
- Implement data retention policies

## Monitoring

- Set up security alerts
- Monitor for unusual activity
- Review deployment logs
- Track failed authentication attempts
- Monitor resource usage for anomalies`,examples:[{title:"Secure Environment Setup",language:"bash",code:`# Use secret management for sensitive values
byteport env set DB_PASSWORD="..." --secret
byteport env set API_KEY="..." --secret

# Use read-only API keys for CI/CD
byteport api-keys create \\
  --name "CI/CD Key" \\
  --scope "read" \\
  --expires "90d"

# Enable audit logging
byteport settings update \\
  --audit-logging enabled \\
  --log-retention 90d`}]}];function C(){let[e,t]=(0,s.useState)(""),[r,p]=(0,s.useState)("all"),[m,u]=(0,s.useState)(N[0]),[h,g]=(0,s.useState)(null),x=N.filter(t=>{let o=t.title.toLowerCase().includes(e.toLowerCase())||t.description.toLowerCase().includes(e.toLowerCase())||t.content.toLowerCase().includes(e.toLowerCase()),s="all"===r||t.category===r;return o&&s}),C=(e,t)=>{navigator.clipboard.writeText(e),g(t),setTimeout(()=>g(null),2e3)};return(0,o.jsxs)("div",{className:"flex flex-1 flex-col overflow-hidden",children:[o.jsx(a.x,{title:"Documentation",subtitle:"Learn how to build and deploy with BytePort",action:(0,o.jsxs)(l.z,{variant:"outline",onClick:()=>window.open("https://byteport.dev/docs","_blank"),children:[o.jsx(f.Z,{className:"h-4 w-4 mr-2"}),"Full Documentation"]})}),(0,o.jsxs)("section",{className:"flex flex-1 overflow-hidden",children:[(0,o.jsxs)("div",{className:"w-80 border-r border-border bg-muted/30",children:[(0,o.jsxs)("div",{className:"p-4 space-y-4",children:[(0,o.jsxs)("div",{className:"relative",children:[o.jsx(b.Z,{className:"absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"}),o.jsx(i.I,{placeholder:"Search documentation...",value:e,onChange:e=>t(e.target.value),className:"pl-10"})]}),o.jsx("div",{className:"flex flex-wrap gap-2",children:[{id:"all",label:"All Docs"},{id:"getting-started",label:"Getting Started"},{id:"guides",label:"Guides"},{id:"api",label:"API Reference"},{id:"advanced",label:"Advanced"}].map(e=>o.jsx(c.C,{variant:r===e.id?"default":"outline",className:"cursor-pointer",onClick:()=>p(e.id),children:e.label},e.id))})]}),o.jsx(d.x,{className:"h-[calc(100vh-250px)]",children:(0,o.jsxs)("div",{className:"p-4 space-y-1",children:[x.map(e=>o.jsx("button",{onClick:()=>u(e),className:`w-full text-left rounded-lg p-3 transition-colors ${m?.id===e.id?"bg-primary text-primary-foreground":"hover:bg-muted"}`,children:(0,o.jsxs)("div",{className:"flex items-start gap-3",children:[o.jsx(e.icon,{className:"h-5 w-5 mt-0.5 flex-shrink-0"}),(0,o.jsxs)("div",{className:"flex-1 min-w-0",children:[o.jsx("div",{className:"font-medium truncate",children:e.title}),o.jsx("div",{className:`text-xs mt-1 line-clamp-2 ${m?.id===e.id?"text-primary-foreground/80":"text-muted-foreground"}`,children:e.description})]}),o.jsx(v.Z,{className:"h-4 w-4 flex-shrink-0"})]})},e.id)),0===x.length&&(0,o.jsxs)("div",{className:"text-center py-8 text-muted-foreground",children:[o.jsx(j,{className:"h-12 w-12 mx-auto mb-3 opacity-50"}),o.jsx("p",{className:"text-sm",children:"No documentation found"}),o.jsx("p",{className:"text-xs mt-1",children:"Try adjusting your search"})]})]})})]}),o.jsx("div",{className:"flex-1 overflow-y-auto",children:m?(0,o.jsxs)("div",{className:"p-8 max-w-4xl mx-auto",children:[(0,o.jsxs)("div",{className:"mb-6",children:[(0,o.jsxs)("div",{className:"flex items-center gap-3 mb-2",children:[o.jsx(m.icon,{className:"h-8 w-8 text-primary"}),o.jsx("h1",{className:"text-3xl font-bold",children:m.title})]}),o.jsx("p",{className:"text-muted-foreground",children:m.description})]}),o.jsx(n.Zb,{className:"p-6 mb-6",children:o.jsx("div",{className:"prose prose-sm dark:prose-invert max-w-none",children:m.content.split("\n").map((e,t)=>e.startsWith("# ")?o.jsx("h1",{className:"text-2xl font-bold mt-6 mb-4",children:e.substring(2)},t):e.startsWith("## ")?o.jsx("h2",{className:"text-xl font-semibold mt-5 mb-3",children:e.substring(3)},t):e.startsWith("### ")?o.jsx("h3",{className:"text-lg font-medium mt-4 mb-2",children:e.substring(4)},t):e.startsWith("```")?null:""===e.trim()?o.jsx("br",{},t):e.startsWith("- ")?o.jsx("li",{className:"ml-4",children:e.substring(2)},t):o.jsx("p",{className:"mb-2",children:e},t))})}),m.examples&&m.examples.length>0&&(0,o.jsxs)("div",{className:"space-y-4",children:[o.jsx("h2",{className:"text-2xl font-bold",children:"Examples"}),m.examples.map((e,t)=>(0,o.jsxs)(n.Zb,{className:"overflow-hidden",children:[(0,o.jsxs)("div",{className:"flex items-center justify-between bg-muted px-4 py-3 border-b",children:[(0,o.jsxs)("div",{className:"flex items-center gap-2",children:[o.jsx(y.Z,{className:"h-4 w-4"}),o.jsx("span",{className:"font-medium",children:e.title}),o.jsx(c.C,{variant:"outline",className:"text-xs",children:e.language})]}),o.jsx(l.z,{size:"sm",variant:"ghost",onClick:()=>C(e.code,`${m.id}-${t}`),children:h===`${m.id}-${t}`?(0,o.jsxs)(o.Fragment,{children:[o.jsx(w.Z,{className:"h-4 w-4 mr-2 text-green-500"}),"Copied"]}):(0,o.jsxs)(o.Fragment,{children:[o.jsx(P.Z,{className:"h-4 w-4 mr-2"}),"Copy"]})})]}),o.jsx("div",{className:"p-4 bg-slate-950 text-slate-50",children:o.jsx("pre",{className:"text-sm overflow-x-auto",children:o.jsx("code",{children:e.code})})})]},t))]})]}):o.jsx("div",{className:"flex items-center justify-center h-full",children:(0,o.jsxs)("div",{className:"text-center",children:[o.jsx(j,{className:"h-16 w-16 mx-auto mb-4 text-muted-foreground"}),o.jsx("h2",{className:"text-2xl font-semibold mb-2",children:"Select a documentation topic"}),o.jsx("p",{className:"text-muted-foreground",children:"Choose a topic from the sidebar to get started"})]})})})]})]})}},33104:(e,t,r)=>{"use strict";r.d(t,{x:()=>a});var o=r(10326);r(17577);var s=r(75445);function a({title:e,subtitle:t,action:r,className:a,...n}){return o.jsx("div",{className:(0,s.cn)("border-b border-dark-surfaceVariant bg-dark-surfaceContainerLow px-6 py-6",a),...n,children:(0,o.jsxs)("div",{className:"flex items-center justify-between",children:[(0,o.jsxs)("div",{className:"flex-1",children:[o.jsx("h1",{className:"text-2xl font-bold text-dark-onSurface",children:e}),t&&o.jsx("p",{className:"mt-1 text-sm text-dark-onSurfaceVariant",children:t})]}),r&&o.jsx("div",{className:"ml-4 flex items-center gap-2",children:r})]})})}},567:(e,t,r)=>{"use strict";r.d(t,{C:()=>i});var o=r(10326);r(17577);var s=r(79360),a=r(75445);let n=(0,s.j)("inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",{variants:{variant:{default:"border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80",secondary:"border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",destructive:"border-transparent bg-destructive text-destructive-foreground shadow hover:bg-destructive/80",outline:"text-foreground",success:"border-transparent bg-green-500 text-white shadow hover:bg-green-600",warning:"border-transparent bg-yellow-500 text-white shadow hover:bg-yellow-600"}},defaultVariants:{variant:"default"}});function i({className:e,variant:t,...r}){return o.jsx("div",{className:(0,a.cn)(n({variant:t}),e),...r})}},33071:(e,t,r)=>{"use strict";r.d(t,{Ol:()=>i,SZ:()=>c,Zb:()=>n,aY:()=>d,eW:()=>p,ll:()=>l});var o=r(10326),s=r(17577),a=r(75445);let n=s.forwardRef(({className:e,...t},r)=>o.jsx("div",{ref:r,className:(0,a.cn)("rounded-xl border bg-card text-card-foreground shadow",e),...t}));n.displayName="Card";let i=s.forwardRef(({className:e,...t},r)=>o.jsx("div",{ref:r,className:(0,a.cn)("flex flex-col space-y-1.5 p-6",e),...t}));i.displayName="CardHeader";let l=s.forwardRef(({className:e,...t},r)=>o.jsx("h3",{ref:r,className:(0,a.cn)("font-semibold leading-none tracking-tight",e),...t}));l.displayName="CardTitle";let c=s.forwardRef(({className:e,...t},r)=>o.jsx("p",{ref:r,className:(0,a.cn)("text-sm text-muted-foreground",e),...t}));c.displayName="CardDescription";let d=s.forwardRef(({className:e,...t},r)=>o.jsx("div",{ref:r,className:(0,a.cn)("p-6 pt-0",e),...t}));d.displayName="CardContent";let p=s.forwardRef(({className:e,...t},r)=>o.jsx("div",{ref:r,className:(0,a.cn)("flex items-center p-6 pt-0",e),...t}));p.displayName="CardFooter"},54432:(e,t,r)=>{"use strict";r.d(t,{I:()=>n});var o=r(10326),s=r(17577),a=r(75445);let n=s.forwardRef(({className:e,type:t,...r},s)=>o.jsx("input",{type:t,className:(0,a.cn)("flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",e),ref:s,...r}));n.displayName="Input"},66307:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("CircleCheck",[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["path",{d:"m9 12 2 2 4-4",key:"dzmm74"}]])},92498:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("Code",[["polyline",{points:"16 18 22 12 16 6",key:"z7tu5w"}],["polyline",{points:"8 6 2 12 8 18",key:"1eg1df"}]])},43810:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("Copy",[["rect",{width:"14",height:"14",x:"8",y:"8",rx:"2",ry:"2",key:"17jyea"}],["path",{d:"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2",key:"zix9uf"}]])},7027:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("ExternalLink",[["path",{d:"M15 3h6v6",key:"1q9fwt"}],["path",{d:"M10 14 21 3",key:"gplh6r"}],["path",{d:"M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6",key:"a6xqqp"}]])},56625:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("GitBranch",[["line",{x1:"6",x2:"6",y1:"3",y2:"15",key:"17qcm7"}],["circle",{cx:"18",cy:"6",r:"3",key:"1h7g24"}],["circle",{cx:"6",cy:"18",r:"3",key:"fqmcym"}],["path",{d:"M18 9a9 9 0 0 1-9 9",key:"n2h4wq"}]])},58038:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("Shield",[["path",{d:"M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z",key:"oel41y"}]])},3634:(e,t,r)=>{"use strict";r.d(t,{Z:()=>o});/**
 * @license lucide-react v0.424.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */let o=(0,r(62881).Z)("Zap",[["path",{d:"M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z",key:"1xq2db"}]])},23562:(e,t,r)=>{"use strict";r.r(t),r.d(t,{$$typeof:()=>n,__esModule:()=>a,default:()=>i});var o=r(68570);let s=(0,o.createProxy)(String.raw`/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/(dashboard)/docs/page.tsx`),{__esModule:a,$$typeof:n}=s;s.default;let i=(0,o.createProxy)(String.raw`/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/frontend/app/(dashboard)/docs/page.tsx#default`)},68032:(e,t,r)=>{"use strict";r.r(t),r.d(t,{default:()=>o.c});var o=r(92931)}};var t=require("../../../webpack-runtime.js");t.C(e);var r=e=>t(t.s=e),o=t.X(0,[421,721,254,603,124],()=>r(90686));module.exports=o})();