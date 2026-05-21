const COMMUNITIES = [
  {
    name: 'GitHub Discussions',
    desc: 'Ask questions, share ideas, and discuss Jabline features with the core team.',
    url: 'https://github.com/Jabline-lang/Jabline/discussions',
    icon: '🐙',
    color: 'text-jb-300',
  },
  {
    name: 'Report Issues',
    desc: 'Found a bug or have a feature request? Open an issue on GitHub.',
    url: 'https://github.com/Jabline-lang/Jabline/issues',
    icon: '🐛',
    color: 'text-red-400',
  },
  {
    name: 'Discord',
    desc: 'Real-time chat, help, and announcements from the Jabline team.',
    url: '#',
    icon: '💬',
    color: 'text-indigo-400',
  },
  {
    name: 'Package Registry',
    desc: 'Explore community packages and publish your own to the Jabline registry.',
    url: 'https://github.com/Jabline-lang/registry',
    icon: '📦',
    color: 'text-amber-400',
  },
]

const CONTRIBUTE = [
  { icon: '📖', title: 'Improve Docs',    desc: 'Fix typos, add examples, write guides.',         href: 'https://github.com/Jabline-lang/Jabline', label: 'Edit on GitHub' },
  { icon: '🐛', title: 'Report Bugs',     desc: 'Open an issue with steps to reproduce.',         href: 'https://github.com/Jabline-lang/Jabline/issues', label: 'Open Issue' },
  { icon: '⭐', title: 'Contribute Code', desc: 'Good-first-issues, features, VM optimization.',  href: 'https://github.com/Jabline-lang/Jabline', label: 'Fork & PR' },
  { icon: '📦', title: 'Publish Packages', desc: 'Build libraries for the community.',            href: 'https://github.com/Jabline-lang/registry', label: 'Registry' },
  { icon: '✍️', title: 'Write Tutorials', desc: 'Share knowledge through blog posts.',            href: '#', label: 'Blog Guidelines' },
  { icon: '💬', title: 'Help Others',     desc: 'Answer questions in Discussions & Discord.',     href: 'https://github.com/Jabline-lang/Jabline/discussions', label: 'Discussions' },
]

export default function Community({ setCurrentPage }) {
  return (
    <div className="pt-20 min-h-screen">
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <span className="accent-badge mb-4 inline-block">Community</span>
          <h1 className="text-jb-50 font-black mb-3">Join the Community</h1>
          <p className="text-jb-400 text-xl max-w-xl">
            Jabline is open source and community-driven. Connect with fellow developers and help shape the language.
          </p>
        </div>
      </div>

      <div className="container-max py-16 space-y-20">
        {/* Community Spaces */}
        <section>
          <h2 className="text-jb-50 font-bold mb-2">Community Spaces</h2>
          <p className="text-jb-400 mb-8">Multiple channels to connect with the Jabline community.</p>
          <div className="grid sm:grid-cols-2 gap-4">
            {COMMUNITIES.map(c => (
              <a
                key={c.name}
                href={c.url}
                target="_blank"
                rel="noopener noreferrer"
                className="jb-card jb-card-accent p-6 group flex gap-4"
              >
                <div className="text-4xl flex-shrink-0">{c.icon}</div>
                <div>
                  <h3 className="text-jb-100 font-semibold mb-1 group-hover:text-accent transition-colors">{c.name}</h3>
                  <p className="text-jb-400 text-sm leading-relaxed">{c.desc}</p>
                </div>
              </a>
            ))}
          </div>
        </section>

        {/* Contribute */}
        <section>
          <h2 className="text-jb-50 font-bold mb-2">How to Contribute</h2>
          <p className="text-jb-400 mb-8">Every contribution matters — from typo fixes to core VM improvements.</p>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {CONTRIBUTE.map(item => (
              <div key={item.title} className="jb-card p-6 flex flex-col">
                <div className="text-3xl mb-4">{item.icon}</div>
                <h3 className="text-jb-100 font-semibold mb-2">{item.title}</h3>
                <p className="text-jb-400 text-sm leading-relaxed mb-4 flex-1">{item.desc}</p>
                <a href={item.href} target="_blank" rel="noopener noreferrer"
                   className="text-accent text-sm font-medium hover:text-accent-light transition-colors">
                  {item.label} →
                </a>
              </div>
            ))}
          </div>
        </section>

        {/* Code of Conduct */}
        <section className="jb-card border-accent/20 p-8">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
            <div className="text-4xl">🤝</div>
            <div className="flex-1">
              <h3 className="text-jb-50 font-bold text-xl mb-2">Code of Conduct</h3>
              <p className="text-jb-400 max-w-xl">
                All Jabline spaces are committed to a welcoming and inclusive environment for everyone, regardless of background or experience level.
              </p>
            </div>
            <a href="https://github.com/Jabline-lang/Jabline/blob/main/CODE_OF_CONDUCT.md"
               target="_blank" rel="noopener noreferrer"
               className="jb-btn jb-btn-secondary text-sm flex-shrink-0">
              Read CoC →
            </a>
          </div>
        </section>

        {/* Open source CTA */}
        <section className="text-center py-8">
          <div className="hero-glow">
            <h2 className="text-jb-50 font-black mb-4 text-4xl">Open source, forever.</h2>
            <p className="text-jb-400 text-lg mb-8 max-w-md mx-auto">
              Jabline is MIT-licensed and will always be free. Star us on GitHub!
            </p>
            <a href="https://github.com/Jabline-lang/Jabline" target="_blank" rel="noopener noreferrer"
               className="jb-btn jb-btn-primary text-base px-8 py-3 inline-flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 16 16">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
              </svg>
              Star on GitHub
            </a>
          </div>
        </section>
      </div>
    </div>
  )
}
