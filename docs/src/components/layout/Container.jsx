export default function Container({ children, className = '' }) {
  return (
    <div className={`container-max ${className}`}>
      {children}
    </div>
  )
}
