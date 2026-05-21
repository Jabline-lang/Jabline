import { useState, useEffect, useRef } from 'react'

export default function DocsSidebar({ sections }) {
  const [activeId, setActiveId] = useState(sections[0]?.id || '')
  const observerRef = useRef(null)
  const isClickingRef = useRef(false)

  // Flatten all IDs for observation
  const allIds = sections.flatMap(s => [s.id, ...(s.children?.map(c => c.id) || [])])

  useEffect(() => {
    // Disconnect previous observer
    if (observerRef.current) observerRef.current.disconnect()

    // Build observer: fires when a section crosses the top ~25% of the viewport
    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (isClickingRef.current) return // ignore during smooth scroll

        // Find the topmost visible section
        const visible = entries
          .filter(e => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)

        if (visible.length > 0) {
          setActiveId(visible[0].target.id)
        }
      },
      {
        // Trigger when section enters top 30% of viewport
        rootMargin: '-80px 0px -65% 0px',
        threshold: 0,
      }
    )

    // Observe every section element
    allIds.forEach(id => {
      const el = document.getElementById(id)
      if (el) observerRef.current.observe(el)
    })

    return () => observerRef.current?.disconnect()
  }, [sections])

  const handleClick = (id) => {
    setActiveId(id)
    isClickingRef.current = true

    const el = document.getElementById(id)
    if (el) {
      const offset = 88 // navbar height
      const top = el.getBoundingClientRect().top + window.scrollY - offset
      window.scrollTo({ top, behavior: 'smooth' })
    }

    // Re-enable observer after scroll animation
    setTimeout(() => { isClickingRef.current = false }, 800)
  }

  return (
    <aside className="hidden lg:block w-52 flex-shrink-0">
      <div className="sticky top-24">
        <p className="text-xs font-semibold uppercase tracking-widest text-jb-500 mb-3 px-3">
          On this page
        </p>
        <nav className="space-y-0.5">
          {sections.map((section) => (
            <div key={section.id}>
              <button
                onClick={() => handleClick(section.id)}
                className={`sidebar-link ${activeId === section.id ? 'active' : ''}`}
              >
                {section.label}
              </button>
              {section.children && (
                <div className="ml-3 space-y-0.5">
                  {section.children.map(child => (
                    <button
                      key={child.id}
                      onClick={() => handleClick(child.id)}
                      className={`sidebar-link text-xs ${activeId === child.id ? 'active' : ''}`}
                    >
                      {child.label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          ))}
        </nav>
      </div>
    </aside>
  )
}
