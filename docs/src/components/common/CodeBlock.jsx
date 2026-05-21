import { useEffect, useState } from 'react'
// Use npm package directly — no CDN dependency, full color control
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'

// Register languages once
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('json', json)

export default function CodeBlock({ code, language = 'jabline', filename = null, showHeader = true }) {
  const [copied, setCopied]         = useState(false)
  const [highlighted, setHighlighted] = useState('')

  useEffect(() => {
    // Jabline → JavaScript (closest grammar)
    const lang = language === 'jabline' ? 'javascript' : language
    try {
      if (hljs.getLanguage(lang)) {
        const result = hljs.highlight(code, { language: lang })
        setHighlighted(result.value)
      } else {
        setHighlighted(hljs.highlightAuto(code).value)
      }
    } catch {
      // Fallback to plain text
      setHighlighted(code.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'))
    }
  }, [code, language])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code)
    } catch {
      const el = document.createElement('textarea')
      el.value = code
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    }
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const langLabel = { jabline: 'jb', javascript: 'js', bash: 'sh', json: 'json', go: 'go' }[language] || language

  return (
    <div className="jb-code-block mb-4">
      {showHeader && (
        <div className="code-header">
          <div className="flex items-center gap-3">
            <div className="code-dots">
              <span className="dot dot-red" />
              <span className="dot dot-yellow" />
              <span className="dot dot-green" />
            </div>
            <span className="text-jb-500 text-xs font-mono">{filename || langLabel}</span>
          </div>
          <button
            onClick={handleCopy}
            className={`copy-btn ${copied ? 'copied' : ''}`}
            aria-label="Copy code"
          >
            {copied ? (
              <span className="flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="3">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/>
                </svg>
                Copied!
              </span>
            ) : (
              <span className="flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                  <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                </svg>
                Copy
              </span>
            )}
          </button>
        </div>
      )}
      <pre>
        <code
          className={`hljs language-${language === 'jabline' ? 'javascript' : language}`}
          dangerouslySetInnerHTML={{ __html: highlighted }}
        />
      </pre>
    </div>
  )
}
