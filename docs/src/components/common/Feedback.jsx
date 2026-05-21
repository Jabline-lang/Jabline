import { useState } from 'react'

export default function Feedback() {
  const [open, setOpen] = useState(false)
  const [sent, setSent]   = useState(false)
  const [rating, setRating] = useState(null)
  const [text, setText]   = useState('')

  const handleSubmit = (e) => {
    e.preventDefault()
    setSent(true)
    setTimeout(() => { setOpen(false); setSent(false); setRating(null); setText('') }, 2000)
  }

  return (
    <>
      {/* Feedback trigger */}
      <div className="fixed bottom-6 right-6 z-40">
        <button
          onClick={() => setOpen(v => !v)}
          className="jb-btn jb-btn-primary shadow-jb text-sm px-4 py-2.5 flex items-center gap-2"
          aria-label="Send feedback"
        >
          <span>💬</span>
          Feedback
        </button>
      </div>

      {/* Feedback panel */}
      {open && (
        <div className="fixed bottom-20 right-6 z-50 w-80 jb-card shadow-jb-lg animate-slide-up">
          <div className="p-5">
            {sent ? (
              <div className="text-center py-4">
                <div className="text-3xl mb-3">🎉</div>
                <p className="text-jb-100 font-semibold">Thanks for the feedback!</p>
              </div>
            ) : (
              <form onSubmit={handleSubmit}>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-jb-100 font-semibold text-sm">Share Feedback</h3>
                  <button type="button" onClick={() => setOpen(false)}
                    className="text-jb-500 hover:text-jb-300 transition-colors text-lg leading-none">
                    ×
                  </button>
                </div>
                <p className="text-jb-400 text-xs mb-4">How was your experience with the docs?</p>

                {/* Rating */}
                <div className="flex gap-2 mb-4">
                  {['😞', '😐', '😊', '🤩'].map((emoji, i) => (
                    <button
                      key={i}
                      type="button"
                      onClick={() => setRating(i)}
                      className={`flex-1 py-2 rounded-lg text-xl border transition-all ${
                        rating === i
                          ? 'border-accent bg-accent/10'
                          : 'border-jb-700 hover:border-jb-500'
                      }`}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>

                <textarea
                  value={text}
                  onChange={e => setText(e.target.value)}
                  placeholder="Tell us more (optional)..."
                  className="w-full px-3 py-2.5 rounded-lg text-sm resize-none mb-4"
                  rows={3}
                />
                <button type="submit" className="jb-btn jb-btn-primary w-full justify-center text-sm">
                  Send Feedback
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </>
  )
}
