# Showcase Outreach Strategy

Internal planning document for potential "Used By" showcase candidates. These are production-grade TUI tools that could benefit from tuikit-go, modeled after successful tools like lazygit and k9s.

## Candidate Tools (10-15)

### Tier 1: Direct Contact (High Fit)

1. **[lazygit](https://github.com/jesseduffield/lazygit)** — Git client with interactive staging, stashing, rebasing. Could leverage tuikit's table/layout for enhanced diff views and keybinding registry.

2. **[k9s](https://github.com/derailed/k9s)** — Kubernetes cluster manager with real-time resource monitoring. Perfect fit for tuikit's table sorting/filtering and theme system.

3. **[glow](https://github.com/charmbracelet/glow)** — Markdown renderer for terminal. Could use tuikit's layout engine for multi-pane document navigation.

4. **[nats-top](https://github.com/nats-io/nats-top)** — NATS broker monitoring dashboard. Ideal for tuikit's poller + table + status bar pattern.

5. **[ctop](https://github.com/bcicen/ctop)** — Container monitoring (Docker/containerd). Heavy table/filtering use case matching tuikit's core strengths.

### Tier 2: Community Ecosystem (Medium Fit)

6. **[lazydocker](https://github.com/jesseduffield/lazydocker)** — Docker client companion to lazygit. Table-driven UI with sidebar, natural fit for DualPane layout.

7. **[gotop](https://github.com/xxxserxxx/gotop)** — System resource monitor (CPU, memory, disks). Uses sparklines and dynamic refresh—native tuikit utilities.

8. **[fzf](https://github.com/junegunn/fzf)** — Fuzzy finder. While mature, tuikit's SelectOne/MultiSelect CLI primitives align with core use case.

9. **[jq](https://github.com/jqlang/jq)** — JSON query processor. Upcoming interactive mode could leverage tuikit's JSON viewer components.

10. **[bat](https://github.com/sharkdp/bat)** — Cat clone with syntax highlighting. Interactive pager mode could use tuikit's layout + theming.

### Tier 3: Emerging/Niche (Lower Fit)

11. **[bottom](https://github.com/ClementtsaC/bottom)** — System monitor with modular layout. Tuikit's DualPane could improve side-by-side process/disk views.

12. **[dust](https://github.com/bootandy/dust)** — Disk usage analyzer. TreeView + sorting could benefit from tuikit's table patterns.

13. **[ripgrep-all](https://github.com/phiresky/ripgrep-all)** — Search tool with result browser. FilterBar + ListView combo natural for tuikit.

14. **[tealdeer](https://github.com/dbrgn/tealdeer)** — tldr page client with interactive navigation. CLI primitives perfect fit.

15. **[sd](https://github.com/chmln/sd)** — Regex stream editor with preview mode. Could use tuikit's ConfigEditor for search/replace patterns.

## Outreach Goals

- **Primary**: Showcase tuikit-go's ability to ship complete TUI tools quickly with minimal boilerplate
- **Secondary**: Build community adoption and gather real-world usage patterns
- **Tertiary**: Identify missing features or pain points from existing TUI ecosystems

## Notes

- Focus Tier 1 first—high-visibility tools with clear architectural alignment
- Mention specific tuikit features (table sorting, DualPane, CLI primitives, self-update) that solve their current pain points
- Lead with examples from gitstream-tui and cryptstream-tui as proof of concept
- Time outreach to align with major releases or security updates when tool maintainers are engaged
