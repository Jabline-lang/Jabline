export default function Button({ children, variant = 'primary', onClick, href, external = false, className = '', disabled = false }) {
  const classes = {
    primary:   'jb-btn jb-btn-primary',
    secondary: 'jb-btn jb-btn-secondary',
    ghost:     'jb-btn jb-btn-ghost',
    accent:    'jb-btn jb-btn-primary',
  }

  const cls = `${classes[variant] || classes.primary} ${className} ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`

  if (href) {
    const props = external ? { target: '_blank', rel: 'noopener noreferrer' } : {}
    return (
      <a href={href} className={cls} {...props}>
        {children}
      </a>
    )
  }

  return (
    <button onClick={onClick} className={cls} disabled={disabled}>
      {children}
    </button>
  )
}
