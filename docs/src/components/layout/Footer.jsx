export default function Footer({ setCurrentPage }) {
  const year = new Date().getFullYear()

  const sections = [
    {
      title: 'Language',
      links: [
        { label: 'Learn', page: 'guide' },
        { label: 'Documentation', page: 'docs' },
        { label: 'Standard Library', page: 'stdlib' },
        { label: 'Playground', page: 'playground' },
        { label: 'Examples', page: 'examples' },
      ],
    },
    {
      title: 'Community',
      links: [
        { label: 'Blog', page: 'blog' },
        { label: 'Community', page: 'community' },
        { label: 'GitHub Discussions', href: 'https://github.com/Jabline-lang/Jabline/discussions' },
        { label: 'Issues', href: 'https://github.com/Jabline-lang/Jabline/issues' },
      ],
    },
    {
      title: 'Resources',
      links: [
        { label: 'GitHub', href: 'https://github.com/Jabline-lang/Jabline' },
        { label: 'Package Registry', href: 'https://github.com/Jabline-lang/registry' },
        { label: 'Changelog', href: 'https://github.com/Jabline-lang/Jabline/releases' },
        { label: 'MIT License', href: 'https://github.com/Jabline-lang/Jabline/blob/main/LICENSE' },
      ],
    },
  ]

  return (
    <footer className="border-t border-jb-700 mt-20">
      {/* Top accent line */}
      <div className="h-px bg-gradient-to-r from-transparent via-accent/40 to-transparent" />

      <div className="bg-jb-950">
        <div className="container-wide py-16">
          <div className="grid md:grid-cols-4 gap-12 mb-12">
            {/* Brand */}
            <div>
              <div className="flex items-center gap-2 mb-4">
                <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-8 h-8">
                  <rect width="32" height="32" rx="7" fill="#3d7eff" fillOpacity="0.15"/>
                  <rect x="0.5" y="0.5" width="31" height="31" rx="6.5" stroke="#3d7eff" strokeOpacity="0.5"/>
                  <text x="5" y="22" fontFamily="JetBrains Mono, monospace" fontWeight="700" fontSize="14" fill="#6aa3ff">JB</text>
                </svg>
                <span className="font-bold text-jb-50 text-lg">Jabline</span>
              </div>
              <p className="text-jb-400 text-sm leading-relaxed mb-4">
                A compiled, cloud-native programming language with a custom bytecode VM. Built for speed, concurrency, and developer experience.
              </p>
              {/* Social links */}
              <div className="flex items-center gap-3">
                <a
                  href="https://github.com/Jabline-lang/Jabline"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-jb-400 hover:text-accent transition-colors"
                  aria-label="GitHub"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
                  </svg>
                </a>
              </div>
            </div>

            {/* Link sections */}
            {sections.map((section) => (
              <div key={section.title}>
                <h4 className="text-jb-200 font-semibold text-sm mb-4 uppercase tracking-widest">
                  {section.title}
                </h4>
                <ul className="space-y-2.5">
                  {section.links.map((link) => (
                    <li key={link.label}>
                      {link.page ? (
                        <button
                          onClick={() => setCurrentPage?.(link.page)}
                          className="text-jb-400 hover:text-accent text-sm transition-colors"
                        >
                          {link.label}
                        </button>
                      ) : (
                        <a
                          href={link.href}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-jb-400 hover:text-accent text-sm transition-colors"
                        >
                          {link.label}
                        </a>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>

          {/* Bottom bar */}
          <div className="border-t border-jb-800 pt-8 flex flex-col sm:flex-row items-center justify-between gap-4">
            <p className="text-jb-500 text-sm">
              © {year} Jabline Language. Released under the MIT License.
            </p>
            <div className="flex items-center gap-4">
              <span className="text-jb-600 text-xs font-mono">v0.6.0</span>
              <span className="text-jb-700">·</span>
              <span className="text-jb-600 text-xs font-mono">Built with ❤️</span>
            </div>
          </div>
        </div>
      </div>
    </footer>
  )
}
