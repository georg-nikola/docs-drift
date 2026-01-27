# Security Policy

## Overview

The security of docs-drift is a top priority. This document outlines our security policies, the intentional design of the tool, and how to report vulnerabilities.

## Intentional Code Execution

**IMPORTANT:** docs-drift is specifically designed to execute code examples from documentation files. This is not a bug or vulnerability—it's the core purpose of the tool.

### By Design
- ✅ Extracts code blocks from Markdown files
- ✅ Executes them in isolated processes
- ✅ Validates they run without errors
- ✅ Reports failures for documentation maintenance

This intentional behavior enables automated validation that documentation remains accurate and working.

## Security Measures

We implement multiple layers of security to ensure safe code execution:

### 1. Process Isolation
- Code runs in **separate OS processes**, not in-memory
- Each execution is isolated from the GitHub Actions runner
- Failed executions don't affect the runner environment

### 2. Resource Limits
- **Configurable timeouts** (default: 30 seconds per code block)
- Prevents infinite loops and runaway processes
- Automatic process termination on timeout

### 3. Environment Restrictions
- **Limited environment variables** passed to execution
- **No network access** by default in isolated mode (Python, Node.js)
- Temporary file cleanup after each execution

### 4. User Control
- Users explicitly configure which languages to validate
- `docs-drift:skip` directive to exclude specific code blocks
- `--changed-only` mode limits scope to modified files only

### 5. Transparency
- **Open source**: All code is publicly auditable
- Clear documentation of execution behavior
- No hidden functionality or telemetry

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x: (EOL)          |

We recommend always using the latest version.

## Reporting a Vulnerability

If you discover a security vulnerability in docs-drift, please report it responsibly.

### What to Report

**Do report:**
- Security vulnerabilities in the execution isolation
- Bypass of timeout mechanisms
- Unauthorized access to runner environment
- Code execution outside intended scope
- Vulnerabilities in dependencies

**Do NOT report as security issues:**
- Intentional code execution (this is by design)
- Code examples in your own documentation failing
- Expected behavior documented in README

### How to Report

**For security vulnerabilities:**

1. **Preferred**: Use GitHub Security Advisories
   - Go to: https://github.com/georg-nikola/docs-drift/security/advisories
   - Click "Report a vulnerability"
   - Provide detailed information (see below)

2. **Alternative**: Email (if GitHub Security is unavailable)
   - Create an issue titled "Security: [Brief Description]"
   - Mark it as private/security-related
   - We'll respond within 48 hours

### What to Include

Please provide:
- **Description**: Clear explanation of the vulnerability
- **Impact**: What could an attacker achieve?
- **Reproduction**: Step-by-step instructions to reproduce
- **Version**: Which version(s) are affected?
- **Proof of Concept**: Code/commands demonstrating the issue
- **Suggested Fix**: If you have ideas (optional)

### Response Timeline

We aim to respond to security reports with the following timeline:

- **Initial Response**: Within 48 hours
- **Severity Assessment**: Within 5 business days
- **Fix Development**: Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Next planned release
- **Disclosure**: After fix is released and users have time to update (typically 7-14 days)

### Coordinated Disclosure

We follow coordinated disclosure practices:

1. You report the vulnerability privately
2. We acknowledge and assess severity
3. We develop and test a fix
4. We release the fix in a new version
5. We publish a security advisory
6. You may publish your findings after disclosure

We appreciate researchers who follow responsible disclosure practices and will credit them in advisories (unless they prefer to remain anonymous).

## Security Best Practices for Users

To use docs-drift securely:

### 1. Trust Your Documentation
- Only use docs-drift in repositories you control
- Review code examples before adding them to docs
- Don't validate documentation from untrusted sources

### 2. Use Configuration Wisely
- Only enable languages you actually use
- Set appropriate timeouts for your use case
- Use `docs-drift:skip` for examples that shouldn't run

### 3. Leverage CI Features
- Use `--changed-only` in PRs to limit scope
- Review documentation changes in PR reviews
- Monitor CI logs for unexpected behavior

### 4. Keep Updated
- Use the latest version of docs-drift
- Subscribe to releases for security updates
- Review CHANGELOG.md for security fixes

### 5. Protect Secrets
- Never put secrets in documentation examples
- Use GitHub Secrets for sensitive data
- Be aware code examples run in your CI environment

## Known Limitations

We're transparent about what docs-drift can and cannot do:

### By Design Limitations
- **Executes code**: This is intentional, not a vulnerability
- **Trusts documentation**: Assumes you control your docs
- **CI environment**: Runs with the permissions of your CI runner

### Technical Limitations
- Cannot prevent all resource exhaustion (despite timeouts)
- Cannot validate code that requires external services
- Cannot guarantee isolation on all platforms (OS-dependent)

## Security Updates

Security updates are released as:
- **Patch versions** (0.3.x) for security fixes
- **GitHub Security Advisories** for severe issues
- **CHANGELOG.md** entries clearly marked as security fixes

Subscribe to releases to be notified of security updates:
https://github.com/georg-nikola/docs-drift/releases

## Third-Party Dependencies

docs-drift relies on system runtimes:
- Node.js (for JavaScript validation)
- Python 3 (for Python validation)
- Bash (for shell validation)
- Go (for Go validation)
- Ruby (for Ruby validation)

**Security responsibility:**
- We're responsible for how we use these runtimes
- You're responsible for keeping them updated in your CI environment
- Vulnerabilities in the runtimes themselves should be reported to their respective maintainers

## Security Checklist

Before using docs-drift in production:

- [ ] Read and understand this security policy
- [ ] Review the EULA and understand code execution behavior
- [ ] Configure only necessary languages
- [ ] Set appropriate timeouts for your use case
- [ ] Use `--changed-only` in PR validation workflows
- [ ] Subscribe to security advisories
- [ ] Keep docs-drift updated to latest version
- [ ] Review documentation changes in PRs
- [ ] Never include secrets in documentation examples

## Questions?

For security-related questions that aren't vulnerabilities:
- Review the README.md documentation
- Check existing GitHub Issues
- Open a new Discussion for clarification

For vulnerabilities, always use the private reporting methods described above.

---

**Last Updated:** January 26, 2026
**Security Contact:** Via GitHub Security Advisories
**Repository:** https://github.com/georg-nikola/docs-drift
