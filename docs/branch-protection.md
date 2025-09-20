# Branch Protection Setup

This document describes how to configure branch protection rules for the GoQSO repository to ensure all pull requests pass the required checks.

## GitHub Branch Protection Rules

To set up branch protection for the `main` branch:

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Branches**
3. Click **Add rule** next to "Branch protection rules"
4. Configure the following settings:

### Branch name pattern
```
main
```

### Protection rules to enable:

#### ✅ Require a pull request before merging
- **Required approvals**: 1
- ✅ Dismiss stale PR approvals when new commits are pushed
- ✅ Require review from code owners (if CODEOWNERS file exists)

#### ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging
- **Required status checks** (add these):
  - `Lint Go Code`
  - `Run Tests` 
  - `Build Application`
  - `Security Scan`

#### ✅ Require conversation resolution before merging

#### ✅ Require signed commits (optional but recommended)

#### ✅ Require linear history (optional)

#### ✅ Include administrators
- This ensures even administrators must follow the same rules

## Required Status Checks

The following GitHub Actions workflows must pass before a PR can be merged:

1. **Lint Go Code** - Runs golangci-lint with comprehensive linting rules
2. **Run Tests** - Executes unit tests with race detection and coverage reporting
3. **Build Application** - Ensures the application compiles successfully
4. **Security Scan** - Runs gosec security analysis

## Local Development

Before creating a pull request, developers should run:

```bash
# Run all checks locally
make check

# Or run individual checks
make fmt      # Format code
make vet      # Run go vet
make lint     # Run golangci-lint (installs if needed)
make test     # Run tests
make build    # Build application
```

## Bypassing Rules (Emergency Only)

In emergency situations, administrators can temporarily bypass rules by:

1. Disabling branch protection
2. Making necessary changes
3. Re-enabling branch protection

However, this should be avoided and any emergency changes should be reviewed in a subsequent PR.

## Status Check Configuration

The workflows are configured to:
- Only run when Go files are changed
- Cache Go modules for faster builds
- Generate coverage reports
- Upload build artifacts
- Provide detailed failure information

## Troubleshooting

### Common Issues:

1. **Status checks not appearing**: Ensure workflows have run at least once on the main branch
2. **Lint failures**: Run `make lint` locally to see detailed error messages
3. **Test failures**: Run `make test` locally for debugging
4. **Build failures**: Check for missing dependencies with `make deps`

### Adding New Status Checks:

1. Create the workflow in `.github/workflows/`
2. Test the workflow on a feature branch
3. Add the job name to required status checks in branch protection rules