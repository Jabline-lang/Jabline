import { useState, useEffect } from 'react'

const CATEGORY_COLORS = {
  'Release':   'bg-accent/10 text-accent border-accent/30',
  'Tutorial':  'bg-blue-500/10 text-blue-400 border-blue-500/30',
  'Deep Dive': 'bg-violet-500/10 text-violet-400 border-violet-500/30',
  'Update':    'bg-green-500/10 text-green-400 border-green-500/30',
  'News':      'bg-orange-500/10 text-orange-400 border-orange-500/30',
}

export default function Blog({ searchTerm = '' }) {
  const [activeCategory, setActiveCategory] = useState('All')
  const [blogs, setBlogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [showDash, setShowDash] = useState(false)
  
  // Dashboard state
  const [newBlog, setNewBlog] = useState({
    title: '', excerpt: '', category: 'News', featured: false
  })

  const fetchBlogs = async () => {
    try {
      const res = await fetch('/api/blogs')
      const data = await res.json()
      setBlogs(data)
    } catch (err) {
      console.error('Failed to fetch blogs', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchBlogs()
  }, [])

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this post?')) return
    try {
      await fetch(`/api/blogs/${id}`, { method: 'DELETE' })
      setBlogs(blogs.filter(b => b.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  const handleCreate = async (e) => {
    e.preventDefault()
    const payload = {
      ...newBlog,
      date: new Date().toISOString().split('T')[0],
      author: 'Jabline Team',
      readTime: '3 min'
    }
    try {
      const res = await fetch('/api/blogs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      })
      const created = await res.json()
      setBlogs([created, ...blogs])
      setNewBlog({ title: '', excerpt: '', category: 'News', featured: false })
      setShowDash(false)
    } catch (err) {
      console.error(err)
    }
  }

  const categories = ['All', ...new Set(blogs.map(a => a.category))]

  const filtered = blogs.filter(a => {
    const matchSearch = !searchTerm ||
      a.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.excerpt.toLowerCase().includes(searchTerm.toLowerCase())
    const matchCat = activeCategory === 'All' || a.category === activeCategory
    return matchSearch && matchCat
  })

  const featured = filtered.find(a => a.featured)
  const rest     = filtered.filter(a => !a.featured || a.id !== featured?.id)

  return (
    <div className="pt-20 min-h-screen">
      {/* Header */}
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <div className="flex justify-between items-start">
            <div>
              <span className="accent-badge mb-4 inline-block">Blog</span>
              <h1 className="text-jb-50 font-black mb-3">Jabline Blog</h1>
              <p className="text-jb-400 text-xl max-w-xl">
                Releases, tutorials, deep dives, and news from the Jabline team.
              </p>
            </div>
            <button 
              onClick={() => setShowDash(!showDash)} 
              className="jb-btn jb-btn-secondary text-xs"
            >
              {showDash ? 'Close Dashboard' : 'Manage Blogs'}
            </button>
          </div>
        </div>
      </div>

      <div className="container-max py-12">
        {/* Admin Dashboard */}
        {showDash && (
          <div className="mb-12 p-6 jb-card border-accent/30 bg-accent/5">
            <h2 className="text-lg font-bold text-jb-50 mb-4 flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="var(--accent)"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/></svg>
              Blog Dashboard
            </h2>
            
            <form onSubmit={handleCreate} className="space-y-4 max-w-2xl">
              <div>
                <label className="block text-xs font-semibold text-jb-500 uppercase tracking-widest mb-1">Title</label>
                <input required value={newBlog.title} onChange={e => setNewBlog({...newBlog, title: e.target.value})} className="w-full bg-jb-900 border border-jb-800 rounded px-3 py-2 text-sm text-jb-100 outline-none focus:border-accent" />
              </div>
              <div>
                <label className="block text-xs font-semibold text-jb-500 uppercase tracking-widest mb-1">Excerpt</label>
                <textarea required value={newBlog.excerpt} onChange={e => setNewBlog({...newBlog, excerpt: e.target.value})} rows="3" className="w-full bg-jb-900 border border-jb-800 rounded px-3 py-2 text-sm text-jb-100 outline-none focus:border-accent" />
              </div>
              <div className="flex gap-4">
                <div className="flex-1">
                  <label className="block text-xs font-semibold text-jb-500 uppercase tracking-widest mb-1">Category</label>
                  <select value={newBlog.category} onChange={e => setNewBlog({...newBlog, category: e.target.value})} className="w-full bg-jb-900 border border-jb-800 rounded px-3 py-2 text-sm text-jb-100 outline-none focus:border-accent">
                    {Object.keys(CATEGORY_COLORS).map(c => <option key={c} value={c}>{c}</option>)}
                  </select>
                </div>
                <div className="flex items-center gap-2 mt-6">
                  <input type="checkbox" id="featured" checked={newBlog.featured} onChange={e => setNewBlog({...newBlog, featured: e.target.checked})} className="accent-accent" />
                  <label htmlFor="featured" className="text-sm text-jb-300 cursor-pointer">Featured Post</label>
                </div>
              </div>
              <button type="submit" className="jb-btn jb-btn-primary">Publish Post</button>
            </form>
          </div>
        )}

        {/* Category filter */}
        <div className="flex gap-2 flex-wrap mb-10">
          {categories.map(cat => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all border ${
                activeCategory === cat
                  ? 'bg-accent/15 text-accent border-accent/40'
                  : 'text-jb-400 border-jb-700 hover:border-jb-500 hover:text-jb-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="text-center py-20 text-jb-500">Loading posts...</div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-20 text-jb-500">No posts found.</div>
        ) : (
          <>
            {/* Featured Post */}
            {featured && (
              <div className="mb-12 group relative">
                {showDash && (
                  <button onClick={() => handleDelete(featured.id)} className="absolute top-4 right-4 z-10 p-2 bg-red-500/20 text-red-400 rounded hover:bg-red-500/40">
                    Delete
                  </button>
                )}
                <div className="absolute -inset-0.5 bg-gradient-to-r from-accent to-violet-500 rounded-2xl opacity-20 group-hover:opacity-40 blur transition duration-500"></div>
                <div className="relative jb-card p-8 sm:p-12 border-accent/30 bg-jb-950 h-full flex flex-col justify-center">
                  <div className="flex items-center gap-3 mb-4">
                    <span className={`text-xs px-2.5 py-1 rounded-full border ${CATEGORY_COLORS[featured.category] || CATEGORY_COLORS['Release']}`}>
                      {featured.category}
                    </span>
                    <span className="text-jb-500 text-sm">{featured.date}</span>
                    <span className="text-jb-600 text-sm hidden sm:inline">· {featured.readTime} read</span>
                  </div>
                  <h2 className="text-3xl sm:text-4xl font-bold text-jb-50 mb-4 group-hover:text-accent transition-colors">
                    {featured.title}
                  </h2>
                  <p className="text-jb-400 text-lg mb-6 max-w-3xl">
                    {featured.excerpt}
                  </p>
                  <div className="flex items-center gap-3 mt-auto">
                    <div className="w-8 h-8 rounded-full bg-accent/20 flex items-center justify-center text-accent font-bold text-sm">
                      JB
                    </div>
                    <span className="text-jb-300 text-sm font-medium">{featured.author}</span>
                  </div>
                </div>
              </div>
            )}

            {/* Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {rest.map(article => (
                <div key={article.id} className="jb-card p-6 flex flex-col relative group">
                  {showDash && (
                    <button onClick={() => handleDelete(article.id)} className="absolute top-4 right-4 z-10 p-1.5 bg-red-500/10 text-red-400 rounded hover:bg-red-500/30 text-xs">
                      Delete
                    </button>
                  )}
                  <div className="flex items-center gap-3 mb-4">
                    <span className={`text-xs px-2 py-0.5 rounded border ${CATEGORY_COLORS[article.category] || CATEGORY_COLORS['Tutorial']}`}>
                      {article.category}
                    </span>
                    <span className="text-jb-500 text-xs">{article.date}</span>
                  </div>
                  <h3 className="text-xl font-bold text-jb-50 mb-3 group-hover:text-accent transition-colors">
                    {article.title}
                  </h3>
                  <p className="text-jb-400 text-sm mb-6 flex-1">
                    {article.excerpt}
                  </p>
                  <div className="flex items-center justify-between mt-auto pt-4 border-t border-jb-800/50">
                    <span className="text-jb-400 text-xs">{article.author}</span>
                    <span className="text-jb-600 text-xs">{article.readTime} read</span>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
