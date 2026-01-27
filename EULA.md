# End User License Agreement (EULA) for docs-drift

**Effective Date:** January 26, 2026
**Version:** 1.0

## Agreement to Terms

By using the docs-drift GitHub Action ("the Action"), you agree to be bound by the terms of this End User License Agreement ("EULA") and the MIT License under which the source code is distributed.

## License Grant

Subject to your compliance with this EULA, georg-nikola ("Developer") grants you a non-exclusive, worldwide, royalty-free license to use the Action in your GitHub repositories for the purpose of validating code examples in documentation.

## Permitted Uses

You may use the Action to:
- Validate code examples in Markdown documentation files
- Integrate the Action into your CI/CD workflows
- Run the Action in repositories you own or have permission to modify
- Use the Action in both public and private repositories

## Code Execution Notice

**IMPORTANT - INTENTIONAL BEHAVIOR:** The Action is specifically designed to execute code examples found in your documentation files. This is the core, intended functionality of the tool, not a bug or security vulnerability.

By using this Action, you acknowledge and accept that:

1. **Code Execution**: The Action will extract and execute code blocks from your Markdown files in the languages you specify (JavaScript, Python, Bash, Go, Ruby)
2. **Your Responsibility**: You are solely responsible for the content of code examples in your documentation
3. **Execution Environment**: Code runs in isolated processes within your GitHub Actions runner environment
4. **Security Measures**: The Action implements security controls:
   - Isolated process execution (separate from main runner)
   - Configurable timeouts (default: 30 seconds per block)
   - Restricted environment variables
   - No network access by default
   - Temporary file cleanup after execution
5. **Limitations**: While security measures are in place, the Action cannot prevent all potential issues arising from code execution
6. **Trust Requirement**: You should only use this Action in repositories you control and trust

## Restrictions

You may NOT:
- Use the Action for malicious purposes or to execute untrusted code
- Attempt to circumvent or disable security features
- Use the Action to execute code in repositories you don't control without authorization
- Hold the Developer liable for damages arising from your use of the Action
- Redistribute the Action under a different name or claim authorship

## Support and Maintenance

Support is provided on a best-effort basis through:
- **GitHub Issues**: https://github.com/georg-nikola/docs-drift/issues
- **Documentation**: README.md and repository documentation
- **Response Time**: We aim to respond within 3-5 business days

The Developer is not obligated to provide support but will make reasonable efforts to assist users and address reported issues.

## No Warranty

THE ACTION IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NONINFRINGEMENT.

THE DEVELOPER DOES NOT WARRANT THAT:
- The Action will meet your requirements
- The Action will be error-free, secure, or uninterrupted
- Code execution will not cause issues in your CI environment
- Results will be accurate in all circumstances

## Limitation of Liability

TO THE MAXIMUM EXTENT PERMITTED BY APPLICABLE LAW, IN NO EVENT SHALL THE DEVELOPER BE LIABLE FOR ANY SPECIAL, INCIDENTAL, INDIRECT, OR CONSEQUENTIAL DAMAGES WHATSOEVER (INCLUDING, WITHOUT LIMITATION, DAMAGES FOR LOSS OF PROFITS, BUSINESS INTERRUPTION, LOSS OF INFORMATION, OR ANY OTHER PECUNIARY LOSS) ARISING OUT OF THE USE OF OR INABILITY TO USE THE ACTION, EVEN IF THE DEVELOPER HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.

Your use of the Action is at your sole risk.

## Data Privacy and Security

The Action:
- **Does NOT collect** any user data, telemetry, or analytics
- **Does NOT transmit** data to external services
- **Does NOT store** data outside your GitHub Actions environment
- **Executes entirely locally** within your CI runner

All code execution and validation occurs within your GitHub Actions environment. No data leaves your control.

## Intellectual Property

- The Action's source code is licensed under the MIT License
- You retain all rights to your documentation and code examples
- The Developer retains all rights to the Action's branding and name
- This EULA does not grant you rights to use the Developer's trademarks

## Updates and Modifications

The Developer may release updates, bug fixes, and new versions of the Action. Your continued use of updated versions constitutes acceptance of any EULA modifications, which will be communicated through:
- Updated EULA.md file in the repository
- Release notes on GitHub
- Marketplace listing updates

You are responsible for reviewing EULA changes.

## Termination

This license is effective until terminated.

The license terminates automatically if you:
- Breach any terms of this EULA
- Use the Action for prohibited purposes
- Fail to comply with the MIT License terms

Upon termination, you must cease all use of the Action.

## Third-Party Dependencies

The Action relies on third-party runtimes and tools (Node.js, Python, Bash, Go, Ruby) which are subject to their own licenses and terms. You are responsible for ensuring compliance with those terms in your environment.

## Export Compliance

You agree to comply with all applicable export and import laws and regulations. You will not export, re-export, or transfer the Action to prohibited countries or individuals.

## Governing Law

This EULA shall be governed by and construed in accordance with the laws of the jurisdiction where the Developer resides, without regard to its conflict of law provisions.

Any disputes arising from this EULA shall be resolved in the courts of competent jurisdiction in that location.

## Severability

If any provision of this EULA is held to be unenforceable or invalid, that provision will be enforced to the maximum extent possible, and the other provisions will remain in full force and effect.

## Entire Agreement

This EULA, together with the MIT License, constitutes the entire agreement between you and the Developer regarding the Action and supersedes all prior agreements and understandings.

## Contact Information

For questions about this EULA, please:
- Open an issue: https://github.com/georg-nikola/docs-drift/issues
- Review the license: https://github.com/georg-nikola/docs-drift/blob/main/LICENSE
- View documentation: https://github.com/georg-nikola/docs-drift/blob/main/README.md

## Acknowledgment

BY USING THE ACTION, YOU ACKNOWLEDGE THAT YOU HAVE READ THIS EULA, UNDERSTAND IT, AND AGREE TO BE BOUND BY ITS TERMS AND CONDITIONS.

---

**Last Updated:** January 26, 2026
**Version:** 1.0
**Source Code License:** MIT License
**Repository:** https://github.com/georg-nikola/docs-drift
