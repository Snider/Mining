# Built with Claude Code

This project showcases what's possible when human creativity meets AI-powered development.

## The Team

<div class="grid cards" markdown>

-   :material-account:{ .lg .middle } **Snider**

    ---

    Project creator, architect, and the human behind the vision. Creator of the foundational libraries:

    - **[Borg](https://github.com/Snider/Borg)** - Encryption toolkit (SMSG, STMF, TIM)
    - **[Poindexter](https://github.com/Snider/Poindexter)** - KD-tree peer selection
    - **Enchantrix** - Additional tooling

    The ideas, direction, and core infrastructure came from Snider.

-   :material-robot:{ .lg .middle } **Claude (Opus 4.5)**

    ---

    AI development partner from Anthropic. Assisted with:

    - Code implementation and refactoring
    - Documentation and testing
    - UI component development
    - Bug fixing and optimization

</div>

## How It Was Built

This entire codebase was developed collaboratively using [Claude Code](https://claude.ai/code), Anthropic's CLI tool for AI-assisted development.

### The Development Process

1. **Vision & Architecture** - Snider defined the goals: a multi-miner management system with P2P capabilities, leveraging his encryption libraries
2. **Iterative Development** - Claude helped implement features, write tests, and refine code
3. **Real-time Feedback** - Features were built, tested, and adjusted through conversation
4. **Documentation** - This entire documentation site was generated collaboratively

### What Claude Helped Build

| Component | Contribution |
|-----------|--------------|
| **Go Backend** | REST API, miner management, SQLite persistence |
| **Angular Frontend** | Dashboard, profiles, console with ANSI colors |
| **P2P System** | Node identity, peer registry, WebSocket transport |
| **Testing** | E2E tests with Playwright, unit tests |
| **Documentation** | This MkDocs site with screenshots |

### What Snider Built

| Component | Description |
|-----------|-------------|
| **Borg Library** | SMSG encryption, STMF keypairs, TIM bundles |
| **Poindexter** | Multi-dimensional KD-tree for peer selection |
| **Project Vision** | The concept of a self-hosted mining dashboard |
| **Architecture Decisions** | P2P design, no cloud dependencies |

## Code Statistics

```
───────────────────────────────────────────────────────────────
Language            Files     Lines     Code    Comments
───────────────────────────────────────────────────────────────
Go                     45      8500     6800        450
TypeScript             35      4200     3500        200
HTML                   20       800      750         20
CSS                    15       600      550         30
───────────────────────────────────────────────────────────────
Total                 115     14100    11600        700
───────────────────────────────────────────────────────────────
```

## Lessons Learned

### What Works Well with AI

- **Boilerplate code** - Repetitive patterns, CRUD operations
- **Documentation** - Generating docs from code
- **Testing** - Writing test cases and E2E tests
- **Debugging** - Analyzing errors and suggesting fixes
- **Refactoring** - Improving code structure

### What Needs Human Direction

- **Architecture** - High-level design decisions
- **Security** - Cryptographic choices, threat modeling
- **User Experience** - Understanding real user needs
- **Domain Knowledge** - Mining-specific requirements
- **Library Selection** - Choosing the right tools

## Try Claude Code

Want to build something similar? Try Claude Code:

```bash
# Install Claude Code
npm install -g @anthropic-ai/claude-code

# Start a session
claude-code
```

Learn more at [claude.ai/code](https://claude.ai/code)

## Acknowledgments

Special thanks to:

- **Snider** - For the vision, the underlying libraries, and the opportunity to collaborate
- **Anthropic** - For building Claude and Claude Code
- **XMRig Team** - For the excellent mining software
- **TT-Miner Team** - For GPU mining support
- **Open Source Community** - For all the libraries that made this possible

---

<div style="text-align: center; margin-top: 2rem; color: #64748b;">
<em>Built with AI assistance, but powered by human creativity.</em>
</div>
