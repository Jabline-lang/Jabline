import { useState, useEffect, useRef } from 'react'

export default function Navbar({ currentPage, setCurrentPage, onSearch }) {
  const [scrolled, setScrolled] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [showSearch, setShowSearch] = useState(false)
  const searchRef = useRef(null)

  const navItems = [
    { id: 'guide',     label: 'Learn' },
    { id: 'docs',      label: 'Docs' },
    { id: 'stdlib',    label: 'Stdlib' },
    { id: 'playground', label: 'Playground' },
    { id: 'examples', label: 'Examples' },
    { id: 'blog',      label: 'Blog' },
    { id: 'community', label: 'Community' },
  ]

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 10)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    if (showSearch) searchRef.current?.focus()
  }, [showSearch])

  useEffect(() => {
    setMobileOpen(false)
    setShowSearch(false)
  }, [currentPage])

  return (
    <header
      className={`fixed top-0 inset-x-0 z-50 transition-all duration-300 ${
        scrolled
          ? 'bg-jb-900/95 backdrop-blur-md border-b border-jb-700'
          : 'bg-transparent border-b border-transparent'
      }`}
    >
      <div className="container-wide">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <button
            onClick={() => setCurrentPage('home')}
            className="flex items-center gap-2.5 group"
            aria-label="Jabline home"
          >
            {/* JB logo mark */}
            <div className="relative w-8 h-8 flex-shrink-0">
              <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-full">
                <rect width="32" height="32" rx="7" fill="#3d7eff" fillOpacity="0.15"/>
                  <rect x="0.5" y="0.5" width="31" height="31" rx="6.5" stroke="#3d7eff" strokeOpacity="0.5"/>
                  <text x="5" y="22" fontFamily="JetBrains Mono, monospace" fontWeight="700" fontSize="14" fill="#6aa3ff">JB</text>
              </svg>
            </div>
            <span className="font-bold text-lg text-jb-50 group-hover:text-accent transition-colors">
              Jabline
            </span>
            <span className="hidden sm:inline-block text-xs font-mono px-1.5 py-0.5 rounded bg-jb-750 border border-jb-650 text-jb-300">
              v0.6.0
            </span>
          </button>

          {/* Desktop Nav */}
          <nav className="hidden md:flex items-center gap-1" aria-label="Main navigation">
            {navItems.map(item => (
              <button
                key={item.id}
                onClick={() => setCurrentPage(item.id)}
                className={`nav-link ${currentPage === item.id ? 'active' : ''} ${
                  item.id === 'playground'
                    ? 'text-accent hover:text-accent-light font-semibold'
                    : ''
                }`}
              >
                {item.label}
                {item.id === 'playground' && (
                  <span className="ml-1 text-xs bg-accent/10 border border-accent/25 text-accent px-1 rounded">
                    ▶
                  </span>
                )}
              </button>
            ))}
          </nav>

          {/* Right side actions */}
          <div className="flex items-center gap-2">
            {/* Search button */}
            <button
              onClick={() => setShowSearch(v => !v)}
              className="jb-btn jb-btn-ghost p-2 text-jb-300 hover:text-jb-50"
              aria-label="Search"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
              </svg>
            </button>

            {/* GitHub */}
            <a
              href="https://github.com/Jabline-lang/Jabline"
              target="_blank"
              rel="noopener noreferrer"
              className="hidden sm:flex items-center gap-2 jb-btn jb-btn-secondary text-sm"
              aria-label="GitHub"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" fill="currentColor" viewBox="0 0 16 16">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
              </svg>
              <span>GitHub</span>
            </a>

            {/* Mobile hamburger */}
            <button
              onClick={() => setMobileOpen(v => !v)}
              className="md:hidden p-2 text-jb-300 hover:text-jb-50 transition"
              aria-label="Toggle menu"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                {mobileOpen
                  ? <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12"/>
                  : <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16"/>
                }
              </svg>
            </button>
          </div>
        </div>

        {/* Search bar */}
        {showSearch && (
          <div className="pb-4 animate-fade-in">
            <div className="relative">
              <svg className="absolute left-3 top-1/2 -translate-y-1/2 text-jb-400 w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
              </svg>
              <input
                ref={searchRef}
                type="text"
                placeholder="Search documentation..."
                onChange={e => onSearch?.(e.target.value)}
                className="w-full pl-10 pr-4 py-2.5 rounded-lg text-sm bg-jb-800 border border-jb-650 text-jb-100 placeholder-jb-400 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30"
              />
              <kbd className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-jb-400 font-mono bg-jb-750 px-1.5 py-0.5 rounded">
                ESC
              </kbd>
            </div>
          </div>
        )}
      </div>

      {/* Mobile menu */}
      {mobileOpen && (
        <div className="md:hidden border-t border-jb-700 bg-jb-900 animate-fade-in">
          <nav className="container-wide py-4 flex flex-col gap-1">
            {navItems.map(item => (
              <button
                key={item.id}
                onClick={() => setCurrentPage(item.id)}
                className={`text-left px-4 py-3 rounded-lg text-sm font-medium transition-colors ${
                  currentPage === item.id
                    ? 'bg-accent/10 text-accent'
                    : 'text-jb-300 hover:text-jb-100 hover:bg-jb-800'
                }`}
              >
                {item.label}
              </button>
            ))}
            <a
              href="https://github.com/Jabline-lang/Jabline"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-2 text-center jb-btn jb-btn-secondary text-sm"
            >
              GitHub
            </a>
          </nav>
        </div>
      )}
    </header>
  )
}
