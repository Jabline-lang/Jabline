import { useEffect } from 'react'
import hljs from 'highlight.js'

/**
 * Hook para resaltar código en bloques
 */
export function useHighlight(code) {
  useEffect(() => {
    hljs.highlightAll()
  }, [code])
}

/**
 * Hook para manejar el scroll suave
 */
export function useScrollToTop() {
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [])
}

/**
 * Hook para manejar eventos de teclado
 */
export function useKeyDown(key, callback) {
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === key) {
        callback()
      }
    }
    
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [key, callback])
}
