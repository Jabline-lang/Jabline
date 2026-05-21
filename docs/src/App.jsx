import { useState, useEffect } from 'react'
import Navbar from './components/layout/Navbar'
import Footer from './components/layout/Footer'
import Feedback from './components/common/Feedback'
import Home from './pages/Home'
import Guide from './pages/Guide'
import Docs from './pages/Docs'
import Stdlib from './pages/Stdlib'
import Examples from './pages/Examples'
import Blog from './pages/Blog'
import Community from './pages/Community'
import Playground from './pages/Playground'

function App() {
  const [currentPage, setCurrentPage] = useState('home')
  const [searchTerm, setSearchTerm] = useState('')
  const [playgroundCode, setPlaygroundCode] = useState(null)

  // Always dark mode
  useEffect(() => {
    document.documentElement.classList.add('dark')
  }, [])

  // Scroll to top on page change
  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'instant' })
  }, [currentPage])

  const navigateToPlayground = (code) => {
    if (code) setPlaygroundCode(code)
    setCurrentPage('playground')
  }

  const renderPage = () => {
    switch (currentPage) {
      case 'home':
        return <Home setCurrentPage={setCurrentPage} />
      case 'guide':
        return <Guide setCurrentPage={setCurrentPage} />
      case 'docs':
        return <Docs setCurrentPage={setCurrentPage} />
      case 'stdlib':
        return <Stdlib setCurrentPage={setCurrentPage} />
      case 'examples':
        return <Examples setCurrentPage={setCurrentPage} onTryCode={navigateToPlayground} />
      case 'blog':
        return <Blog searchTerm={searchTerm} />
      case 'community':
        return <Community setCurrentPage={setCurrentPage} />
      case 'playground':
        return <Playground setCurrentPage={setCurrentPage} initialCode={playgroundCode} />
      default:
        return <Home setCurrentPage={setCurrentPage} />
    }
  }

  // Pages that show the footer
  const showFooter = currentPage !== 'playground'

  return (
    <div className="min-h-screen flex flex-col" style={{ background: 'var(--bg-base)', color: 'var(--text-primary)' }}>
      <Navbar
        currentPage={currentPage}
        setCurrentPage={setCurrentPage}
        onSearch={setSearchTerm}
      />

      <main className="flex-1 animate-fade-in" key={currentPage}>
        {renderPage()}
      </main>

      {showFooter && <Footer setCurrentPage={setCurrentPage} />}
      <Feedback />
    </div>
  )
}

export default App
